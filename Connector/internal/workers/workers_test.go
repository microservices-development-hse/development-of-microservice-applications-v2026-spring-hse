package workers

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	jiraclient "github.com/microservices-development-hse/connector/internal/jira"
	jiramodels "github.com/microservices-development-hse/connector/internal/models/jira"
)

type mockClient struct {
	getProjectsFunc func() ([]jiramodels.ProjectResponse, error)
	getIssuesFunc   func(projectKey string, startAt, maxResults int) (*jiramodels.IssueSearchResponse, error)
	mu              sync.Mutex
	callCount       int
}

func (m *mockClient) GetProjects() ([]jiramodels.ProjectResponse, error) {
	m.mu.Lock()

	defer m.mu.Unlock()

	if m.getProjectsFunc != nil {
		return m.getProjectsFunc()
	}

	return []jiramodels.ProjectResponse{}, nil
}

func (m *mockClient) GetIssuesByProject(projectKey string, startAt, maxResults int) (*jiramodels.IssueSearchResponse, error) {
	m.mu.Lock()

	defer m.mu.Unlock()

	m.callCount++
	if m.getIssuesFunc != nil {
		return m.getIssuesFunc(projectKey, startAt, maxResults)
	}

	return &jiramodels.IssueSearchResponse{Total: 0, Issues: []jiramodels.Issue{}}, nil
}

func retryConfig() jiraclient.RetryConfig {
	return jiraclient.RetryConfig{
		MinTimeSleep: 1,
		MaxTimeSleep: 5,
	}
}

func TestNewPool(t *testing.T) {
	client := &mockClient{}
	cfg := retryConfig()

	pool := NewPool(3, client, cfg, 100)
	if pool == nil {
		t.Fatal("NewPool returned nil")
	}

	if pool.threadCount != 3 {
		t.Errorf("expected threadCount 3, got %d", pool.threadCount)
	}

	if pool.maxResults != 100 {
		t.Errorf("expected maxResults 100, got %d", pool.maxResults)
	}
}

func TestPool_Run_Success(t *testing.T) {
	client := &mockClient{
		getIssuesFunc: func(projectKey string, startAt, maxResults int) (*jiramodels.IssueSearchResponse, error) {
			var issues []jiramodels.Issue

			switch startAt {
			case 0:
				if maxResults == 1 {
					return &jiramodels.IssueSearchResponse{Total: 4, Issues: []jiramodels.Issue{}}, nil
				}

				issues = []jiramodels.Issue{{Key: "TEST-1"}, {Key: "TEST-2"}}
			case 2:
				issues = []jiramodels.Issue{{Key: "TEST-3"}, {Key: "TEST-4"}}
			default:
				issues = []jiramodels.Issue{}
			}

			return &jiramodels.IssueSearchResponse{Total: 4, Issues: issues}, nil
		},
	}

	pool := NewPool(2, client, retryConfig(), 2)
	ctx := context.Background()

	issues, err := pool.Run(ctx, "TEST")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(issues) != 4 {
		t.Errorf("expected 4 issues, got %d", len(issues))
	}
}

func TestPool_Run_FirstBatchError(t *testing.T) {
	client := &mockClient{
		getIssuesFunc: func(projectKey string, startAt, maxResults int) (*jiramodels.IssueSearchResponse, error) {
			return nil, errors.New("connection failed")
		},
	}

	pool := NewPool(2, client, retryConfig(), 50)
	ctx := context.Background()

	_, err := pool.Run(ctx, "TEST")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestPool_Run_EmptyProject(t *testing.T) {
	client := &mockClient{
		getIssuesFunc: func(projectKey string, startAt, maxResults int) (*jiramodels.IssueSearchResponse, error) {
			return &jiramodels.IssueSearchResponse{Total: 0, Issues: []jiramodels.Issue{}}, nil
		},
	}

	pool := NewPool(1, client, retryConfig(), 50)
	ctx := context.Background()

	issues, err := pool.Run(ctx, "EMPTY")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(issues) != 0 {
		t.Errorf("expected 0 issues, got %d", len(issues))
	}
}

func TestPool_Run_SinglePage(t *testing.T) {
	client := &mockClient{
		getIssuesFunc: func(projectKey string, startAt, maxResults int) (*jiramodels.IssueSearchResponse, error) {
			if startAt == 0 && maxResults == 1 {
				return &jiramodels.IssueSearchResponse{Total: 3, Issues: []jiramodels.Issue{}}, nil
			}

			return &jiramodels.IssueSearchResponse{
				Total: 3,
				Issues: []jiramodels.Issue{
					{Key: "TEST-1"},
					{Key: "TEST-2"},
					{Key: "TEST-3"},
				},
			}, nil
		},
	}

	pool := NewPool(1, client, retryConfig(), 10)
	ctx := context.Background()

	issues, err := pool.Run(ctx, "TEST")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(issues) != 3 {
		t.Errorf("expected 3 issues, got %d", len(issues))
	}
}

func TestPool_Run_WorkerError(t *testing.T) {
	called := 0
	client := &mockClient{
		getIssuesFunc: func(projectKey string, startAt, maxResults int) (*jiramodels.IssueSearchResponse, error) {
			called++
			if called == 1 {
				return &jiramodels.IssueSearchResponse{Total: 10, Issues: []jiramodels.Issue{}}, nil
			}

			return nil, errors.New("worker fetch error")
		},
	}

	pool := NewPool(2, client, retryConfig(), 5)
	ctx := context.Background()

	_, err := pool.Run(ctx, "TEST")
	if err == nil {
		t.Fatal("expected error from worker, got nil")
	}
}

func TestPool_Run_MultipleWorkers(t *testing.T) {
	client := &mockClient{
		getIssuesFunc: func(projectKey string, startAt, maxResults int) (*jiramodels.IssueSearchResponse, error) {
			if startAt == 0 && maxResults == 1 {
				return &jiramodels.IssueSearchResponse{Total: 20, Issues: []jiramodels.Issue{}}, nil
			}

			issues := make([]jiramodels.Issue, maxResults)

			for i := 0; i < maxResults; i++ {
				issues[i] = jiramodels.Issue{Key: "TEST"}
			}

			return &jiramodels.IssueSearchResponse{Total: 20, Issues: issues}, nil
		},
	}

	pool := NewPool(4, client, retryConfig(), 5)
	ctx := context.Background()

	issues, err := pool.Run(ctx, "TEST")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(issues) != 20 {
		t.Errorf("expected 20 issues, got %d", len(issues))
	}
}

func TestPool_Run_ContextCancelImmediately(t *testing.T) {
	client := &mockClient{
		getIssuesFunc: func(projectKey string, startAt, maxResults int) (*jiramodels.IssueSearchResponse, error) {
			return &jiramodels.IssueSearchResponse{Total: 100, Issues: []jiramodels.Issue{}}, nil
		},
	}

	pool := NewPool(2, client, retryConfig(), 50)

	ctx, cancel := context.WithCancel(context.Background())

	cancel()

	_, err := pool.Run(ctx, "TEST")
	if err == nil {
		t.Fatal("expected context error, got nil")
	}
}

func TestPool_Run_ContextCancelDuringJobCreation(t *testing.T) {
	client := &mockClient{
		getIssuesFunc: func(projectKey string, startAt, maxResults int) (*jiramodels.IssueSearchResponse, error) {
			time.Sleep(50 * time.Millisecond)

			if startAt == 0 && maxResults == 1 {
				return &jiramodels.IssueSearchResponse{Total: 100, Issues: []jiramodels.Issue{}}, nil
			}

			return &jiramodels.IssueSearchResponse{Total: 100, Issues: []jiramodels.Issue{{Key: "TEST-1"}}}, nil
		},
	}

	pool := NewPool(2, client, retryConfig(), 10)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()

	_, err := pool.Run(ctx, "TEST")
	if err == nil {
		t.Fatal("expected context error, got nil")
	}
}

func TestNewWorker(t *testing.T) {
	ctx := context.Background()
	jobs := make(chan Job)
	results := make(chan Result)
	client := &mockClient{}
	cfg := retryConfig()

	w := NewWorker(ctx, 5, jobs, results, client, cfg, "PROJ", 100)
	if w.id != 5 {
		t.Errorf("expected id 5, got %d", w.id)
	}

	if w.projectKey != "PROJ" {
		t.Errorf("expected projectKey PROJ, got %s", w.projectKey)
	}

	if w.maxResults != 100 {
		t.Errorf("expected maxResults 100, got %d", w.maxResults)
	}
}

func TestWorker_Start_ProcessesJobs(t *testing.T) {
	jobs := make(chan Job, 1)
	results := make(chan Result, 1)

	client := &mockClient{
		getIssuesFunc: func(projectKey string, startAt, maxResults int) (*jiramodels.IssueSearchResponse, error) {
			return &jiramodels.IssueSearchResponse{
				Total: 2,
				Issues: []jiramodels.Issue{
					{Key: "PROJ-1"},
					{Key: "PROJ-2"},
				},
			}, nil
		},
	}

	ctx := context.Background()
	w := NewWorker(ctx, 1, jobs, results, client, retryConfig(), "PROJ", 50)

	jobs <- Job{StartAt: 0}

	close(jobs)

	go w.Start()

	select {
	case res := <-results:
		if res.Err != nil {
			t.Errorf("unexpected error: %v", res.Err)
		}

		if len(res.Issues) != 2 {
			t.Errorf("expected 2 issues, got %d", len(res.Issues))
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for result")
	}
}

func TestWorker_Start_HandlesError(t *testing.T) {
	jobs := make(chan Job, 1)
	results := make(chan Result, 1)

	client := &mockClient{
		getIssuesFunc: func(projectKey string, startAt, maxResults int) (*jiramodels.IssueSearchResponse, error) {
			return nil, errors.New("jira api error")
		},
	}

	ctx := context.Background()
	w := NewWorker(ctx, 1, jobs, results, client, retryConfig(), "PROJ", 50)

	jobs <- Job{StartAt: 0}

	close(jobs)

	go w.Start()

	select {
	case res := <-results:
		if res.Err == nil {
			t.Fatal("expected error, got nil")
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for result")
	}
}

func TestWorker_Start_WithRetrySuccess(t *testing.T) {
	jobs := make(chan Job, 1)
	results := make(chan Result, 1)

	callCount := 0
	client := &mockClient{
		getIssuesFunc: func(projectKey string, startAt, maxResults int) (*jiramodels.IssueSearchResponse, error) {
			callCount++
			if callCount == 1 {
				return nil, errors.New("temporary error")
			}

			return &jiramodels.IssueSearchResponse{
				Total:  1,
				Issues: []jiramodels.Issue{{Key: "PROJ-1"}},
			}, nil
		},
	}

	ctx := context.Background()
	w := NewWorker(ctx, 1, jobs, results, client, jiraclient.RetryConfig{MinTimeSleep: 1, MaxTimeSleep: 2}, "PROJ", 50)

	jobs <- Job{StartAt: 0}

	close(jobs)

	go w.Start()

	select {
	case res := <-results:
		if res.Err != nil {
			t.Errorf("unexpected error: %v", res.Err)
		}

		if len(res.Issues) != 1 {
			t.Errorf("expected 1 issue, got %d", len(res.Issues))
		}

		if callCount < 2 {
			t.Errorf("expected retry, got %d calls", callCount)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for result")
	}
}

func TestWorker_Start_ContextCancel(t *testing.T) {
	jobs := make(chan Job)
	results := make(chan Result)

	ctx, cancel := context.WithCancel(context.Background())
	w := NewWorker(ctx, 1, jobs, results, &mockClient{}, retryConfig(), "PROJ", 50)

	cancel()

	done := make(chan struct{})

	go func() {
		w.Start()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker did not stop on context cancel")
	}
}

func TestWorker_Start_ClosedJobsChannel(t *testing.T) {
	jobs := make(chan Job)
	results := make(chan Result)

	ctx := context.Background()
	w := NewWorker(ctx, 1, jobs, results, &mockClient{}, retryConfig(), "PROJ", 50)

	close(jobs)

	done := make(chan struct{})

	go func() {
		w.Start()

		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker did not exit on closed jobs channel")
	}
}
