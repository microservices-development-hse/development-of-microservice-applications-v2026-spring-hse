package workers

import (
	"context"
	"fmt"

	jiraclient "github.com/microservices-development-hse/connector/internal/jira"
	jiramodels "github.com/microservices-development-hse/connector/internal/models/jira"
)

type Job struct {
	StartAt int
}

type Result struct {
	Issues []jiramodels.Issue
	Err    error
}

type Worker struct {
	id          int
	ctx         context.Context
	jobs        <-chan Job
	results     chan<- Result
	client      *jiraclient.Client
	retryConfig jiraclient.RetryConfig
	projectKey  string
	maxResults  int
}

func NewWorker(
	ctx context.Context,
	id int,
	jobs <-chan Job,
	results chan<- Result,
	client *jiraclient.Client,
	retryConfig jiraclient.RetryConfig,
	projectKey string,
	maxResults int,
) *Worker {
	return &Worker{
		id:          id,
		ctx:         ctx,
		jobs:        jobs,
		results:     results,
		client:      client,
		retryConfig: retryConfig,
		projectKey:  projectKey,
		maxResults:  maxResults,
	}
}

func (w *Worker) Start() {
	for {
		select {
		case <-w.ctx.Done():
			return

		case job, ok := <-w.jobs:
			if !ok {
				return
			}

			var batch *jiramodels.IssueSearchResponse

			err := jiraclient.WithRetry(w.retryConfig, func() error {
				var err error

				batch, err = w.client.GetIssuesByProject(
					w.projectKey,
					job.StartAt,
					w.maxResults,
				)
				if err != nil {
					return fmt.Errorf("GetIssuesByProject: %w", err)
				}

				return nil
			})

			if err != nil {
				w.results <- Result{Err: err}
				return
			}

			w.results <- Result{Issues: batch.Issues}
		}
	}
}
