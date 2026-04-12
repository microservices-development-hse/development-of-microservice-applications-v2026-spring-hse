package config

import (
	"os"
	"testing"
)

// ─── Load ───────────────────────────────────────────────────────────────────

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

	defer func() { _ = os.Remove(tmpFile) }()

	cfg, err := Load(tmpFile)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.ProgramSettings.ThreadCount != 2 {
		t.Errorf("expected threadCount=2, got %d", cfg.ProgramSettings.ThreadCount)
	}

	if cfg.ProgramSettings.JiraURL != "https://test" {
		t.Errorf("expected jiraUrl=https://test, got %s", cfg.ProgramSettings.JiraURL)
	}

	if cfg.DBSettings.User != "user" {
		t.Errorf("expected dbUser=user, got %s", cfg.DBSettings.User)
	}

	if cfg.DBSettings.Port != 5432 {
		t.Errorf("expected dbPort=5432, got %d", cfg.DBSettings.Port)
	}
}

func TestLoad_FileNotFound_FallsBackToExample(t *testing.T) {
	_, err := Load("nonexistent_path/config.yaml")
	if err == nil {
		t.Fatal("expected error when config file not found and no fallback")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	content := `invalid: [yaml: content`
	tmpFile := "test_invalid.yaml"

	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.Remove(tmpFile) }()

	_, err = Load(tmpFile)
	if err == nil {
		t.Fatal("expected error for invalid yaml")
	}
}

func TestLoad_ValidConfigFailsValidation(t *testing.T) {
	content := `
ProgramSettings:
  jiraUrl: ""
  threadCount: 1
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
	tmpFile := "test_invalid_validate.yaml"

	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.Remove(tmpFile) }()

	_, err = Load(tmpFile)
	if err == nil {
		t.Fatal("expected validation error for empty jiraUrl")
	}
}

// ─── Validate ───────────────────────────────────────────────────────────────

func TestValidate_ValidConfig(t *testing.T) {
	cfg := validConfig()

	if err := cfg.Validate(); err != nil {
		t.Errorf("expected no error for valid config, got %v", err)
	}
}

func TestValidate_EmptyJiraURL(t *testing.T) {
	cfg := validConfig()
	cfg.ProgramSettings.JiraURL = ""

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for empty JiraURL")
	}
}

func TestValidate_InvalidThreadCount(t *testing.T) {
	cfg := validConfig()
	cfg.ProgramSettings.ThreadCount = 0

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for threadCount < 1")
	}
}

func TestValidate_ThreadCountExactlyOne(t *testing.T) {
	cfg := validConfig()
	cfg.ProgramSettings.ThreadCount = 1

	if err := cfg.Validate(); err != nil {
		t.Errorf("expected no error for threadCount=1, got %v", err)
	}
}

func TestValidate_IssueInOneRequest_TooLow(t *testing.T) {
	cfg := validConfig()
	cfg.ProgramSettings.IssueInOneRequest = 10

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for issueInOneRequest < 50")
	}
}

func TestValidate_IssueInOneRequest_TooHigh(t *testing.T) {
	cfg := validConfig()
	cfg.ProgramSettings.IssueInOneRequest = 1001

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for issueInOneRequest > 1000")
	}
}

func TestValidate_IssueInOneRequest_MinBoundary(t *testing.T) {
	cfg := validConfig()
	cfg.ProgramSettings.IssueInOneRequest = 50

	if err := cfg.Validate(); err != nil {
		t.Errorf("expected no error for issueInOneRequest=50, got %v", err)
	}
}

func TestValidate_IssueInOneRequest_MaxBoundary(t *testing.T) {
	cfg := validConfig()
	cfg.ProgramSettings.IssueInOneRequest = 1000

	if err := cfg.Validate(); err != nil {
		t.Errorf("expected no error for issueInOneRequest=1000, got %v", err)
	}
}

func TestValidate_MinTimeSleep_Zero(t *testing.T) {
	cfg := validConfig()
	cfg.ProgramSettings.MinTimeSleep = 0

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for minTimeSleep <= 0")
	}
}

func TestValidate_MinTimeSleep_Negative(t *testing.T) {
	cfg := validConfig()
	cfg.ProgramSettings.MinTimeSleep = -1

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for minTimeSleep < 0")
	}
}

func TestValidate_MaxTimeSleep_Zero(t *testing.T) {
	cfg := validConfig()
	cfg.ProgramSettings.MaxTimeSleep = 0

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for maxTimeSleep <= 0")
	}
}

func TestValidate_MaxTimeSleep_Negative(t *testing.T) {
	cfg := validConfig()
	cfg.ProgramSettings.MaxTimeSleep = -1

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for maxTimeSleep < 0")
	}
}

func TestValidate_MinGreaterThanMax(t *testing.T) {
	cfg := validConfig()
	cfg.ProgramSettings.MinTimeSleep = 2000
	cfg.ProgramSettings.MaxTimeSleep = 1000

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for min > max")
	}
}

func TestValidate_MinEqualsMax(t *testing.T) {
	cfg := validConfig()
	cfg.ProgramSettings.MinTimeSleep = 1000
	cfg.ProgramSettings.MaxTimeSleep = 1000

	if err := cfg.Validate(); err != nil {
		t.Errorf("expected no error for min == max, got %v", err)
	}
}

func TestValidate_Port_Zero(t *testing.T) {
	cfg := validConfig()
	cfg.ProgramSettings.Port = 0

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for port=0")
	}
}

func TestValidate_Port_Negative(t *testing.T) {
	cfg := validConfig()
	cfg.ProgramSettings.Port = -1

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for port < 0")
	}
}

func TestValidate_Port_TooHigh(t *testing.T) {
	cfg := validConfig()
	cfg.ProgramSettings.Port = 70000

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for port > 65535")
	}
}

func TestValidate_Port_MinBoundary(t *testing.T) {
	cfg := validConfig()
	cfg.ProgramSettings.Port = 1

	if err := cfg.Validate(); err != nil {
		t.Errorf("expected no error for port=1, got %v", err)
	}
}

func TestValidate_Port_MaxBoundary(t *testing.T) {
	cfg := validConfig()
	cfg.ProgramSettings.Port = 65535

	if err := cfg.Validate(); err != nil {
		t.Errorf("expected no error for port=65535, got %v", err)
	}
}

// ─── helpers ────────────────────────────────────────────────────────────────

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
