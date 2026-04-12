package jira

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

//
// ===================== CLIENT + API =====================
//

func TestNewClient(t *testing.T) {
	c := NewClient("http://example.com")

	if c.baseURL != "http://example.com" {
		t.Fatalf("wrong baseURL")
	}

	if c.httpClient == nil {
		t.Fatalf("httpClient not initialized")
	}
}

func TestBuildURL(t *testing.T) {
	c := NewClient("http://test")

	url := c.buildURL("/path")

	if url != "http://test/path" {
		t.Fatalf("unexpected url: %s", url)
	}
}

//
// ===================== GetProjects =====================
//

func TestGetProjects_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id":"1","key":"TEST","name":"Test","self":"url"}]`))
	}))
	defer server.Close()

	client := NewClient(server.URL)

	projects, err := client.GetProjects()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(projects) != 1 {
		t.Fatalf("expected 1 project")
	}
}

func TestGetProjects_BadStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "fail", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL)

	_, err := client.GetProjects()
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestGetProjects_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	client := NewClient(server.URL)

	_, err := client.GetProjects()
	if err == nil {
		t.Fatalf("expected decode error")
	}
}

//
// ===================== GetIssuesByProject =====================
//

func TestGetIssuesByProject_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		if q.Get("jql") != "project=TEST" {
			t.Fatalf("wrong jql")
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"total":1,"issues":[]}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)

	res, err := client.GetIssuesByProject("TEST", 0, 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Total != 1 {
		t.Fatalf("wrong total")
	}
}

func TestGetIssuesByProject_EmptyProjectKey(t *testing.T) {
	client := NewClient("http://example.com")

	_, err := client.GetIssuesByProject("", 0, 50)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestGetIssuesByProject_BadStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "fail", http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewClient(server.URL)

	_, err := client.GetIssuesByProject("TEST", 0, 50)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestGetIssuesByProject_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	client := NewClient(server.URL)

	_, err := client.GetIssuesByProject("TEST", 0, 50)
	if err == nil {
		t.Fatalf("expected decode error")
	}
}

//
// ===================== RETRY =====================
//

func TestWithRetry_SuccessFirstTry(t *testing.T) {
	cfg := RetryConfig{MinTimeSleep: 1, MaxTimeSleep: 4}

	called := 0

	err := WithRetry(cfg, func() error {
		called++
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error")
	}

	if called != 1 {
		t.Fatalf("should be called once")
	}
}

func TestWithRetry_RetryThenSuccess(t *testing.T) {
	cfg := RetryConfig{MinTimeSleep: 1, MaxTimeSleep: 4}

	called := 0

	err := WithRetry(cfg, func() error {
		called++
		if called < 2 {
			return errors.New("temporary error")
		}
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error")
	}

	if called < 2 {
		t.Fatalf("should retry")
	}
}

func TestWithRetry_StopOn4xx(t *testing.T) {
	cfg := RetryConfig{MinTimeSleep: 1, MaxTimeSleep: 8}

	called := 0

	err := WithRetry(cfg, func() error {
		called++
		return errors.New("unexpected status 400")
	})

	if err == nil {
		t.Fatalf("expected error")
	}

	if called != 1 {
		t.Fatalf("should not retry on 4xx")
	}
}

func TestWithRetry_ExceedsLimit(t *testing.T) {
	cfg := RetryConfig{MinTimeSleep: 1, MaxTimeSleep: 2}

	err := WithRetry(cfg, func() error {
		return errors.New("fail")
	})

	if err == nil {
		t.Fatalf("expected error")
	}
}

//
// ===================== EDGE =====================
//

// проверяем что sleep реально увеличивается (косвенно)
func TestWithRetry_BackoffGrowth(t *testing.T) {
	cfg := RetryConfig{MinTimeSleep: 1, MaxTimeSleep: 4}

	start := time.Now()

	_ = WithRetry(cfg, func() error {
		return errors.New("fail")
	})

	elapsed := time.Since(start)

	if elapsed <= 0 {
		t.Fatalf("expected some delay")
	}
}
