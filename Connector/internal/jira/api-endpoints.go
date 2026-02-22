package jira

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	models "github.com/microservices-development-hse/connector/internal/models/jira"
)

func (c *Client) buildURL(path string) string {
	return c.baseURL + path
}

func (c *Client) GetProjects() ([]models.ProjectResponse, error) {
	endpoint := c.buildURL("/rest/api/2/project")

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var projects []models.ProjectResponse
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return projects, nil
}

func (c *Client) GetIssuesByProject(projectKey string, startAt, maxResults int) (*models.IssueSearchResponse, error) {
	if projectKey == "" {
		return nil, fmt.Errorf("projectKey cannot be empty")
	}

	jql := fmt.Sprintf("project=%s", projectKey)
	u, err := url.Parse(c.buildURL("/rest/api/2/search"))
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}

	query := u.Query()
	query.Set("jql", jql)
	query.Set("startAt", fmt.Sprintf("%d", startAt))
	query.Set("maxResults", fmt.Sprintf("%d", maxResults))
	query.Set("expand", "changelog")
	u.RawQuery = query.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result models.IssueSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}
