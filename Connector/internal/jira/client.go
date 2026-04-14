package jira

import (
	"net/http"
	"time"

	models "github.com/microservices-development-hse/connector/internal/models/jira"
)

type ClientInterface interface {
	GetIssuesByProject(projectKey string, startAt, maxResults int) (*models.IssueSearchResponse, error)
	GetProjects() ([]models.ProjectResponse, error)
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}
