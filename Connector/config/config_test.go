package config

import (
	"os"
	"testing"
)

func TestLoad_ValidConfig(t *testing.T) {
	content := `
ProgramSettings:
  jiraUrl: "https://test"
  threadCount: 2
  issueInOneRequest: 100
  minTimeSleep: 100
  maxTimeSleep: 1000
  port: 8080

DBSettings:
  dbUser: "user"
  dbPassword: "pass"
  dbHost: "localhost"
  dbPort: 5432
  dbName: "db"
`

	tmpFile := "test_config.yaml"
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile)

	cfg, err := Load(tmpFile)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.ProgramSettings.ThreadCount != 2 {
		t.Errorf("expected threadCount=2, got %d", cfg.ProgramSettings.ThreadCount)
	}
}

func TestValidate_InvalidThreadCount(t *testing.T) {
	cfg := validConfig()
	cfg.ProgramSettings.ThreadCount = 0

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for threadCount < 1")
	}
}

func TestValidate_InvalidIssueCount(t *testing.T) {
	cfg := validConfig()
	cfg.ProgramSettings.IssueInOneRequest = 10

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for invalid issueInOneRequest")
	}
}

func TestValidate_InvalidSleep(t *testing.T) {
	cfg := validConfig()
	cfg.ProgramSettings.MinTimeSleep = 0

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for minTimeSleep <= 0")
	}
}

func TestValidate_MinGreaterThanMax(t *testing.T) {
	cfg := validConfig()
	cfg.ProgramSettings.MinTimeSleep = 2000
	cfg.ProgramSettings.MaxTimeSleep = 1000

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for min > max")
	}
}

func TestValidate_InvalidPort(t *testing.T) {
	cfg := validConfig()
	cfg.ProgramSettings.Port = 70000

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for invalid port")
	}
}

func TestValidate_EmptyJiraURL(t *testing.T) {
	cfg := validConfig()
	cfg.ProgramSettings.JiraURL = ""

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for empty JiraURL")
	}
}

func validConfig() *Config {
	return &Config{
		ProgramSettings: ProgramSettings{
			JiraURL:           "https://test",
			ThreadCount:       1,
			IssueInOneRequest: 100,
			MinTimeSleep:      100,
			MaxTimeSleep:      1000,
			Port:              8080,
		},
		DBSettings: DBSettings{
			User:     "user",
			Password: "pass",
			Host:     "localhost",
			Port:     5432,
			Name:     "db",
		},
	}
}
