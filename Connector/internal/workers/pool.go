package workers

import (
	"context"
	"fmt"
	"sync"

	jiraclient "github.com/microservices-development-hse/connector/internal/jira"
	jiramodels "github.com/microservices-development-hse/connector/internal/models/jira"
)

type Pool struct {
	threadCount int
	client      jiraclient.ClientInterface
	retryConfig jiraclient.RetryConfig
	maxResults  int
}

type PoolInterface interface {
	Run(ctx context.Context, projectKey string) ([]jiramodels.Issue, error)
}

func NewPool(
	threadCount int,
	client jiraclient.ClientInterface,
	retryConfig jiraclient.RetryConfig,
	maxResults int,
) *Pool {
	return &Pool{
		threadCount: threadCount,
		client:      client,
		retryConfig: retryConfig,
		maxResults:  maxResults,
	}
}

func (p *Pool) Run(ctx context.Context, projectKey string) ([]jiramodels.Issue, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	firstBatch, err := p.client.GetIssuesByProject(projectKey, 0, 1)
	if err != nil {
		return nil, fmt.Errorf("get total: %w", err)
	}

	total := firstBatch.Total
	if total == 0 {
		return []jiramodels.Issue{}, nil
	}

	jobs := make(chan Job, total/p.maxResults+1)
	results := make(chan Result, total/p.maxResults+1)

	var wg sync.WaitGroup

	for i := 0; i < p.threadCount; i++ {
		w := NewWorker(
			ctx,
			i+1,
			jobs,
			results,
			p.client,
			p.retryConfig,
			projectKey,
			p.maxResults,
		)

		wg.Add(1)

		go func() {
			defer wg.Done()

			w.Start()
		}()
	}

	go func() {
		for start := 0; start < total; start += p.maxResults {
			select {
			case <-ctx.Done():
				close(jobs)
				return
			case jobs <- Job{StartAt: start}:
			}
		}

		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var allIssues []jiramodels.Issue

	for res := range results {
		if res.Err != nil {
			cancel()
			return nil, res.Err
		}

		allIssues = append(allIssues, res.Issues...)
	}

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
		return allIssues, nil
	}
}
