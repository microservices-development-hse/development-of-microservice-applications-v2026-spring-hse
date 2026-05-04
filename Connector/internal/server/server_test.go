package server

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jira "github.com/microservices-development-hse/connector/internal/jira"
)

func TestNew(t *testing.T) {
	client := jira.NewClient("http://test.com")
	retryConfig := jira.RetryConfig{MinTimeSleep: 100, MaxTimeSleep: 1000}

	var db *sql.DB = nil

	srv := New(8080, client, retryConfig, 100, db, 2)

	if srv.httpServer == nil {
		t.Fatal("httpServer is nil")
	}

	if srv.httpServer.Addr != ":8080" {
		t.Errorf("expected Addr :8080, got %s", srv.httpServer.Addr)
	}

	if srv.httpServer.ReadTimeout != 30*time.Second {
		t.Errorf("wrong ReadTimeout")
	}

	if srv.httpServer.WriteTimeout != 120*time.Second {
		t.Errorf("wrong WriteTimeout")
	}

	if srv.httpServer.IdleTimeout != 60*time.Second {
		t.Errorf("wrong IdleTimeout")
	}
}

func TestMiddleware_Logging(t *testing.T) {
	handler := loggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestMiddleware_Recovery(t *testing.T) {
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	handler := recoveryMiddleware(panicHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

func TestMiddleware_RecoveryNoPanic(t *testing.T) {
	normalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := recoveryMiddleware(normalHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestMiddleware_CORS(t *testing.T) {
	handler := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("missing CORS header")
	}

	if rec.Header().Get("Access-Control-Allow-Methods") != "GET, POST, OPTIONS" {
		t.Error("missing Allow-Methods header")
	}
}

func TestMiddleware_CORS_Options(t *testing.T) {
	called := false
	handler := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}

	if called {
		t.Error("OPTIONS request should not reach inner handler")
	}
}

func TestWithMiddleware_Chain(t *testing.T) {
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := withMiddleware(finalHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestShutdown_WithoutStart(t *testing.T) {
	client := jira.NewClient("http://test.com")
	retryConfig := jira.RetryConfig{MinTimeSleep: 100, MaxTimeSleep: 1000}

	var db *sql.DB = nil

	srv := New(9999, client, retryConfig, 100, db, 2)

	if err := srv.Shutdown(); err != nil {
		t.Errorf("Shutdown on non-started server should not fail: %v", err)
	}
}
