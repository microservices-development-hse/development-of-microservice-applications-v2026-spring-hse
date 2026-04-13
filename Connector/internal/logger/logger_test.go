package logger

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func resetLogger() {
	once = sync.Once{}
	infoLogger = nil
	warningLogger = nil
	errorLogger = nil
}

func TestInit_CreatesFiles(t *testing.T) {
	resetLogger()

	_ = os.RemoveAll("logs")

	err := Init()
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		t.Fatalf("logs directory not created")
	}

	if _, err := os.Stat(filepath.Join("logs", "logs.log")); os.IsNotExist(err) {
		t.Fatalf("logs.log not created")
	}

	if _, err := os.Stat(filepath.Join("logs", "err_logs.log")); os.IsNotExist(err) {
		t.Fatalf("err_logs.log not created")
	}
}

func TestInit_OnlyOnce(t *testing.T) {
	resetLogger()

	_ = os.RemoveAll("logs")

	err1 := Init()
	err2 := Init()

	if err1 != nil {
		t.Fatalf("first init failed: %v", err1)
	}

	if err2 != nil {
		t.Fatalf("second init should not fail: %v", err2)
	}
}

func TestLoggingFunctions_NoPanic(t *testing.T) {
	resetLogger()

	_ = os.RemoveAll("logs")

	err := Init()
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	Info("info test %d", 1)
	Warning("warning test %s", "abc")
	Error("error test")
}

func TestLogging_WritesToFiles(t *testing.T) {
	resetLogger()

	_ = os.RemoveAll("logs")

	if err := Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	Info("info message")
	Warning("warning message")
	Error("error message")

	logData, err := os.ReadFile(filepath.Join("logs", "logs.log"))
	if err != nil {
		t.Fatalf("read logs.log failed: %v", err)
	}

	errLogData, err := os.ReadFile(filepath.Join("logs", "err_logs.log"))
	if err != nil {
		t.Fatalf("read err_logs.log failed: %v", err)
	}

	if len(logData) == 0 {
		t.Fatalf("logs.log is empty")
	}

	if len(errLogData) == 0 {
		t.Fatalf("err_logs.log is empty")
	}
}

func TestInit_MkdirError(t *testing.T) {
	resetLogger()

	_ = os.MkdirAll("logs", 0755)
	_ = os.RemoveAll("logs")

	if err := os.WriteFile("logs", []byte{}, 0644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	defer func() { _ = os.Remove("logs") }()

	err := Init()
	if err == nil {
		t.Fatal("expected error when cannot create logs directory, got nil")
	}
}

func TestInit_LogFileCreateError(t *testing.T) {
	resetLogger()

	if err := os.MkdirAll("logs", 0444); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	defer func() { _ = os.Chmod("logs", 0755) }()
	defer func() { _ = os.RemoveAll("logs") }()

	err := Init()
	if err == nil {
		t.Fatal("expected error when cannot create log file, got nil")
	}
}

func TestInit_ErrLogFileCreateError(t *testing.T) {
	resetLogger()

	if err := os.MkdirAll("logs", 0755); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	defer func() { _ = os.RemoveAll("logs") }()

	t.Skip("skipping: difficult to mock file creation error for second file")
}

func TestInit_ErrLogFileCreateError_Alternative(t *testing.T) {
	resetLogger()

	if err := os.MkdirAll("logs", 0444); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	defer func() { _ = os.Chmod("logs", 0755) }()
	defer func() { _ = os.RemoveAll("logs") }()

	t.Skip("error branch for err_logs.log creation is practically unreachable")
}
