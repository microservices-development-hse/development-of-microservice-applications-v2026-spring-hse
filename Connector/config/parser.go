package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	p := c.ProgramSettings

	if p.JiraURL == "" {
		return fmt.Errorf("JiraURL is required (current value: empty)")
	}

	if p.ThreadCount < 1 {
		return fmt.Errorf("ThreadCount must be >= 1 (current value: %d)", p.ThreadCount)
	}

	if p.IssueInOneRequest < 50 || p.IssueInOneRequest > 1000 {
		return fmt.Errorf("IssueInOneRequest must be between 50 and 1000 (current value: %d)", p.IssueInOneRequest)
	}

	if p.MinTimeSleep <= 0 {
		return fmt.Errorf("MinTimeSleep must be > 0 (current value: %d)", p.MinTimeSleep)
	}

	if p.MaxTimeSleep <= 0 {
		return fmt.Errorf("MaxTimeSleep must be > 0 (current value: %d)", p.MaxTimeSleep)
	}

	if p.MinTimeSleep > p.MaxTimeSleep {
		return fmt.Errorf("MinTimeSleep (%d) must be <= MaxTimeSleep (%d)", p.MinTimeSleep, p.MaxTimeSleep)
	}

	if p.Port <= 0 || p.Port > 65535 {
		return fmt.Errorf("Port must be between 1 and 65535 (current value: %d)", p.Port)
	}

	return nil
}
