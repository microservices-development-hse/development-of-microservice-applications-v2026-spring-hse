package etlprocess

import (
	"fmt"

	jiraclient "github.com/microservices-development-hse/connector/internal/jira"
	jiramodels "github.com/microservices-development-hse/connector/internal/models/jira"
)

type Extractor struct {
	client      *jiraclient.Client
	retryConfig jiraclient.RetryConfig
	maxResults  int
}

func NewExtractor(client *jiraclient.Client, retryConfig jiraclient.RetryConfig, maxResults int) *Extractor {
	return &Extractor{
		client:      client,
		retryConfig: retryConfig,
		maxResults:  maxResults,
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

func (e *Extractor) GetAllIssues(projectKey string) ([]jiramodels.Issue, error) {
	var allIssues []jiramodels.Issue

	startAt := 0

	for {
		var batch *jiramodels.IssueSearchResponse

		err := jiraclient.WithRetry(e.retryConfig, func() error {
			var err error

			batch, err = e.client.GetIssuesByProject(projectKey, startAt, e.maxResults)
			if err != nil {
				return fmt.Errorf("client.GetIssuesByProject: %w", err)
			}

			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("extractor: GetAllIssues at startAt=%d: %w", startAt, err)
		}

		allIssues = append(allIssues, batch.Issues...)

		startAt += len(batch.Issues)
		if startAt >= batch.Total {
			break
		}
	}

	return allIssues, nil
}
