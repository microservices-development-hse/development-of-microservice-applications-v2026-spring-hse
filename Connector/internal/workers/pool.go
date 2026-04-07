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
	client      *jiraclient.Client
	retryConfig jiraclient.RetryConfig
	maxResults  int
}

func NewPool(
	threadCount int,
	client *jiraclient.Client,
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

	// узнаём total
	firstBatch, err := p.client.GetIssuesByProject(projectKey, 0, 1)
	if err != nil {
		return nil, fmt.Errorf("get total: %w", err)
	}

	total := firstBatch.Total

	jobs := make(chan Job, total/p.maxResults+1)
	results := make(chan Result, total/p.maxResults+1)

	var wg sync.WaitGroup

	// воркеры
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

	// задачи (страницы)
	for start := 0; start < total; start += p.maxResults {
		select {
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return nil, ctx.Err()
		default:
			jobs <- Job{StartAt: start}
		}
	}
	close(jobs)

	// сбор результатов
	var allIssues []jiramodels.Issue

	for i := 0; i < total/p.maxResults+1; i++ {
		res := <-results

		if res.Err != nil {
			cancel()  // останавливаем всех воркеров
			wg.Wait() // ждём их завершения
			return nil, res.Err
		}

		allIssues = append(allIssues, res.Issues...)
	}

	wg.Wait()

	return allIssues, nil
}
