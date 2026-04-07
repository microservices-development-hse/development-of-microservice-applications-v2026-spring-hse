package jira

import (
	"fmt"
	"strings"
	"time"

	"github.com/microservices-development-hse/connector/internal/logger"
)

type RetryConfig struct {
	MinTimeSleep int
	MaxTimeSleep int
}

type ErrNoRetry struct {
	Err error
}

func (e *ErrNoRetry) Error() string { return e.Err.Error() }

func WithRetry(cfg RetryConfig, fn func() error) error {
	var err error

	for currentSleep := cfg.MinTimeSleep; currentSleep <= cfg.MaxTimeSleep; currentSleep *= 2 {
		if err = fn(); err == nil {
			return nil
		}

		if strings.Contains(err.Error(), "unexpected status 4") {
			return err
		}

		logger.Warning("Jira request failed, retrying in %dms: %v", currentSleep, err)
		time.Sleep(time.Duration(currentSleep) * time.Millisecond)
	}

	err = fmt.Errorf("retry limit exceeded (maxSleep=%dms): %w", cfg.MaxTimeSleep, err)
	logger.Error("%s", err.Error())

	return err
}
