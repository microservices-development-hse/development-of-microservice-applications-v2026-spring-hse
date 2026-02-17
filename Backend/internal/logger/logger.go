package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

const (
	filePermission = 0o666
	dirPermission  = 0o755
)

type CustomHook struct {
	File      *os.File
	logLevels []logrus.Level
}

func (hook *CustomHook) Levels() []logrus.Level {
	return hook.logLevels
}

func (hook *CustomHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		return err
	}

	_, err = hook.File.WriteString(line)

	return err
}

func InitLogger() error {
	logDir := "logs"

	if err := os.MkdirAll(logDir, dirPermission); err != nil {
		return fmt.Errorf("failed to create logs directory: %w", err)
	}

	allLogsPath := filepath.Join(logDir, "logs.log")
	errLogsPath := filepath.Join(logDir, "err_logs.log")

	allLogsFile, err := os.OpenFile(allLogsPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, filePermission)
	if err != nil {
		return fmt.Errorf("failed to open logs.log: %w", err)
	}

	errLogsFile, err := os.OpenFile(errLogsPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, filePermission)
	if err != nil {
		return fmt.Errorf("failed to open err_logs.log: %w", err)
	}

	logrus.SetOutput(allLogsFile)
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)

	errorLevels := []logrus.Level{
		logrus.WarnLevel,
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}

	logrus.AddHook(&CustomHook{
		File:      errLogsFile,
		logLevels: errorLevels,
	})

	logrus.AddHook(&CustomHook{
		File:      os.Stdout,
		logLevels: errorLevels,
	})

	return nil
}
