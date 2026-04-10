package jira

import (
	"errors"
	"testing"
	"time"
)

func TestWithRetry_SuccessFirstTry(t *testing.T) {
	cfg := RetryConfig{MinTimeSleep: 1, MaxTimeSleep: 4}

	called := 0

	err := WithRetry(cfg, func() error {
		called++
		return nil
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if called != 1 {
		t.Fatalf("expected 1 call, got %d", called)
	}
}

func TestWithRetry_SuccessAfterRetries(t *testing.T) {
	cfg := RetryConfig{MinTimeSleep: 1, MaxTimeSleep: 4}

	called := 0

	err := WithRetry(cfg, func() error {
		called++
		if called < 3 {
			return errors.New("temporary error")
		}
		return nil
	})

	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if called != 3 {
		t.Fatalf("expected 3 calls, got %d", called)
	}
}

func TestWithRetry_StopOn4xx(t *testing.T) {
	cfg := RetryConfig{MinTimeSleep: 1, MaxTimeSleep: 4}

	called := 0

	err := WithRetry(cfg, func() error {
		called++
		return errors.New("unexpected status 404")
	})

	if err == nil {
		t.Fatal("expected error")
	}

	if called != 1 {
		t.Fatalf("expected 1 call (no retry), got %d", called)
	}
}

func TestWithRetry_ExceedsMaxRetries(t *testing.T) {
	cfg := RetryConfig{MinTimeSleep: 1, MaxTimeSleep: 2}

	called := 0

	err := WithRetry(cfg, func() error {
		called++
		return errors.New("always fails")
	})

	if err == nil {
		t.Fatal("expected error")
	}

	if called < 2 {
		t.Fatalf("expected multiple retries, got %d calls", called)
	}
}

func TestWithRetry_ExponentialBackoff(t *testing.T) {
	cfg := RetryConfig{MinTimeSleep: 10, MaxTimeSleep: 40}

	start := time.Now()

	_ = WithRetry(cfg, func() error {
		return errors.New("fail")
	})

	elapsed := time.Since(start)

	// ожидаем примерно 10 + 20 + 40 = ~70ms
	if elapsed < 60*time.Millisecond {
		t.Fatalf("expected backoff delay, got too fast: %v", elapsed)
	}
}
