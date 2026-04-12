package etlprocess

import (
	"context"
	"fmt"

	"github.com/microservices-development-hse/connector/internal/logger"
	"github.com/microservices-development-hse/connector/internal/workers"

	jiraclient "github.com/microservices-development-hse/connector/internal/jira"
	jiramodels "github.com/microservices-development-hse/connector/internal/models/jira"
)

type Extractor struct {
	client      jiraclient.ClientInterface
	retryConfig jiraclient.RetryConfig
	maxResults  int
	threadCount int
	poolFactory func() workers.PoolInterface
}

type ExtractorInterface interface {
	GetProjects() ([]jiramodels.ProjectResponse, error)
	GetAllIssues(ctx context.Context, projectKey string) ([]jiramodels.Issue, error)
}

func NewExtractor(
	client jiraclient.ClientInterface,
	retryConfig jiraclient.RetryConfig,
	maxResults int,
	threadCount int,
) *Extractor {
	return &Extractor{
		client:      client,
		retryConfig: retryConfig,
		maxResults:  maxResults,
		threadCount: threadCount,
		poolFactory: func() workers.PoolInterface {
			return workers.NewPool(
				threadCount,
				client,
				retryConfig,
				maxResults,
			)
		},
	}
}

func (e *Extractor) GetProjects() ([]jiramodels.ProjectResponse, error) {
	var projects []jiramodels.ProjectResponse

	err := jiraclient.WithRetry(e.retryConfig, func() error {
		var err error

		projects, err = e.client.GetProjects()
		if err != nil {
			return fmt.Errorf("client.GetProjects: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("extractor: GetProjects: %w", err)
	}

	return projects, nil
}

func (e *Extractor) GetAllIssues(ctx context.Context, projectKey string) ([]jiramodels.Issue, error) {
	pool := e.poolFactory()

	logger.Info("extractor: start parallel fetch for project %q with %d workers",
		projectKey, e.threadCount)

	issues, err := pool.Run(ctx, projectKey)
	if err != nil {
		return nil, fmt.Errorf("extractor: GetAllIssues: %w", err)
	}

	logger.Info("extractor: fetched %d issues for project %q",
		len(issues), projectKey)

	return issues, nil
}
