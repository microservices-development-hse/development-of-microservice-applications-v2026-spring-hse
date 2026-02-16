package logger

import (
	"os"
	"testing"
)

func TestInitLogger(t *testing.T) {
	os.RemoveAll("logs")

	err := InitLogger()
	if err != nil {
		t.Fatalf("InitLogger() returned error: %v", err)
	}

	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		t.Error("Folder logs was not created")
	}

	files := []string{"logs/logs.log", "logs/err_logs.log"}
	for _, f := range files {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			t.Errorf("File %s was not created", f)
		}
	}
}
