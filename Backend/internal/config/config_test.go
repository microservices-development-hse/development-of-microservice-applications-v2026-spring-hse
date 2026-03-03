package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	type testCase struct {
		name        string
		yamlContent string
		expectError bool
	}

	tests := []testCase{
		{
			name: "Success: Valid config",
			yamlContent: `
DBSettings:
  dbUser: "admin"
  dbPassword: "password"
  dbHost: "127.0.0.1"
  dbPort: 5432
  dbName: "jira_db"
ProgramSettings:
  resourceTimeout: 5
  analyticsTimeout: 15
  bindAddress: "0.0.0.0"
  bindPort: 8000
`,
			expectError: false,
		},
		{
			name: "Fail: Missing required field",
			yamlContent: `
DBSettings:
  dbPassword: "123"
  dbHost: "localhost"
  dbPort: 5432
  dbNamee: "db"
ProgramSettings:
  resourceTimeout: 5
  analyticsTimeout: 15
  bindAddress: "0.0.0.0"
  bindPort: 8000
`,
			expectError: true,
		},
		{
			name: "Fail: Invalid Port",
			yamlContent: `
DBSettings:
  dbUser: "admin"
  dbPassword: "123"
  dbHost: "localhost"
  dbPort: 999999
  dbName: "db"
ProgramSettings:
  resourceTimeout: 5
  analyticsTimeout: 15
  bindAddress: "0.0.0.0"
  bindPort: 8000
`,
			expectError: true,
		},
		{
			name: "Fail: Invalid BindAddress",
			yamlContent: `
DBSettings:
  dbUser: "admin"
  dbPassword: "123"
  dbHost: "localhost"
  dbPort: 5432
  dbName: "db"
ProgramSettings:
  resourceTimeout: 5
  analyticsTimeout: 15
  bindAddress: "wrong-address"
  bindPort: 8000
`,
			expectError: true,
		},
		{
			name: "Fail: Invalid YAML syntax",
			yamlContent: `
DBSettings:
  dbUser: "admin"
  dbPort: : : : broken
`,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmpFile := "test_config.yaml"

			err := os.WriteFile(tmpFile, []byte(tc.yamlContent), 0644)
			if err != nil {
				t.Fatalf("Could not create temp file: %v", err)
			}

			t.Cleanup(func() {
				_ = os.Remove(tmpFile)
			})

			_, err = LoadConfig(tmpFile)
			if (err != nil) != tc.expectError {
				t.Errorf("Test '%s' failed: expected error: %v, got: %v", tc.name, tc.expectError, err)
			}
		})
	}

	t.Run("Fail: File not found", func(t *testing.T) {
		_, err := LoadConfig("non_existent_file.yaml")
		if err == nil {
			t.Error("Expected error for non-existent file, but got nil")
		}
	})
}
