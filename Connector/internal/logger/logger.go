package logger

import (
	"io"
	"log"
	"os"
	"sync"
)

var (
	once          sync.Once
	infoLogger    *log.Logger
	warningLogger *log.Logger
	errorLogger   *log.Logger
)

func Init() error {
	var err error

	once.Do(func() {
		if e := os.MkdirAll("logs", 0755); e != nil {
			err = e
			return
		}

		logFile, e := os.OpenFile("logs/logs.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if e != nil {
			err = e
			return
		}

		errFile, e := os.OpenFile("logs/err_logs.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if e != nil {
			err = e
			return
		}

		infoLogger = log.New(logFile, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

		warningMultiWriter := io.MultiWriter(logFile, errFile, os.Stdout)
		warningLogger = log.New(warningMultiWriter, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)

		errorMultiWriter := io.MultiWriter(logFile, errFile, os.Stdout)
		errorLogger = log.New(errorMultiWriter, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	})

	return err
}

func Info(format string, args ...interface{}) {
	if infoLogger != nil {
		infoLogger.Printf(format, args...)
	}
}

func Warning(format string, args ...interface{}) {
	if warningLogger != nil {
		warningLogger.Printf(format, args...)
	}
}

func Error(format string, args ...interface{}) {
	if errorLogger != nil {
		errorLogger.Printf(format, args...)
	}
}
