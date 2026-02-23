package jira

import (
	"fmt"
	"time"

	"github.com/microservices-development-hse/connector/internal/logger"
)

type RetryConfig struct {
	MinTimeSleep int
	MaxTimeSleep int
}

func WithRetry(cfg RetryConfig, fn func() error) error {
	var err error

	for currentSleep := cfg.MinTimeSleep; currentSleep <= cfg.MaxTimeSleep; currentSleep *= 2 {
		if err = fn(); err == nil {
			return nil
		}

		logger.Warning(fmt.Sprintf("Jira request failed, retrying in %dms: %v", currentSleep, err))
		time.Sleep(time.Duration(currentSleep) * time.Millisecond)
	}

	err = fmt.Errorf("retry limit exceeded (maxSleep=%dms): %w", cfg.MaxTimeSleep, err)
	logger.Error(err.Error())

	return err
}
