package etlprocess

import (
	"context"
	"errors"
	"testing"

	jiraclient "github.com/microservices-development-hse/connector/internal/jira"
	jiramodels "github.com/microservices-development-hse/connector/internal/models/jira"
	"github.com/microservices-development-hse/connector/internal/workers"
)

type mockJiraClient struct {
	getProjectsFunc func() ([]jiramodels.ProjectResponse, error)
}

func (m *mockJiraClient) GetProjects() ([]jiramodels.ProjectResponse, error) {
	return m.getProjectsFunc()
}

func (m *mockJiraClient) GetIssuesByProject(string, int, int) (*jiramodels.IssueSearchResponse, error) {
	return nil, nil
}

type mockPool struct {
	runFunc func(ctx context.Context, projectKey string) ([]jiramodels.Issue, error)
}

func (m *mockPool) Run(ctx context.Context, projectKey string) ([]jiramodels.Issue, error) {
	return m.runFunc(ctx, projectKey)
}

func fastRetry() jiraclient.RetryConfig {
	return jiraclient.RetryConfig{MinTimeSleep: 1, MaxTimeSleep: 2}
}

func TestNewExtractor(t *testing.T) {
	e := NewExtractor(&mockJiraClient{}, fastRetry(), 100, 3)

	if e == nil {
		t.Fatal("expected extractor, got nil")
	}

	if e.threadCount != 3 {
		t.Fatalf("wrong threadCount")
	}

	if e.maxResults != 100 {
		t.Fatalf("wrong maxResults")
	}
}
func TestNewExtractor_DefaultPoolFactory(t *testing.T) {
	client := &mockJiraClient{}
	cfg := fastRetry()

	e := NewExtractor(client, cfg, 10, 2)

	pool := e.poolFactory()

	if pool == nil {
		t.Fatal("expected pool, got nil")
	}
}
func TestGetProjects_Success(t *testing.T) {
	client := &mockJiraClient{
		getProjectsFunc: func() ([]jiramodels.ProjectResponse, error) {
			return []jiramodels.ProjectResponse{
				{ID: "1", Key: "A"},
			}, nil
		},
	}

	e := NewExtractor(client, fastRetry(), 10, 1)

	res, err := e.GetProjects()
	if err != nil {
		t.Fatal(err)
	}

	if len(res) != 1 {
		t.Fatalf("expected 1 project")
	}
}

func TestGetProjects_Error(t *testing.T) {
	client := &mockJiraClient{
		getProjectsFunc: func() ([]jiramodels.ProjectResponse, error) {
			return nil, errors.New("fail")
		},
	}

	e := NewExtractor(client, fastRetry(), 10, 1)

	_, err := e.GetProjects()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetAllIssues_Success(t *testing.T) {
	e := NewExtractor(&mockJiraClient{}, fastRetry(), 10, 1)

	e.poolFactory = func() workers.PoolInterface {
		return &mockPool{
			runFunc: func(ctx context.Context, projectKey string) ([]jiramodels.Issue, error) {
				return []jiramodels.Issue{
					{ID: "1"},
					{ID: "2"},
				}, nil
			},
		}
	}

	res, err := e.GetAllIssues(context.Background(), "TEST")
	if err != nil {
		t.Fatal(err)
	}

	if len(res) != 2 {
		t.Fatalf("expected 2 issues")
	}
}

func TestGetAllIssues_Error(t *testing.T) {
	e := NewExtractor(&mockJiraClient{}, fastRetry(), 10, 1)

	e.poolFactory = func() workers.PoolInterface {
		return &mockPool{
			runFunc: func(ctx context.Context, projectKey string) ([]jiramodels.Issue, error) {
				return nil, errors.New("fail")
			},
		}
	}

	_, err := e.GetAllIssues(context.Background(), "TEST")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetAllIssues_Empty(t *testing.T) {
	e := NewExtractor(&mockJiraClient{}, fastRetry(), 10, 1)

	e.poolFactory = func() workers.PoolInterface {
		return &mockPool{
			runFunc: func(ctx context.Context, projectKey string) ([]jiramodels.Issue, error) {
				return []jiramodels.Issue{}, nil
			},
		}
	}

	res, err := e.GetAllIssues(context.Background(), "TEST")
	if err != nil {
		t.Fatal(err)
	}

	if len(res) != 0 {
		t.Fatal("expected empty slice")
	}
}
