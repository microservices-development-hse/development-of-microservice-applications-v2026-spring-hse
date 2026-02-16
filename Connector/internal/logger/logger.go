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

func Info(msg string) {
	if infoLogger != nil {
		infoLogger.Println(msg)
	}
}

func Warning(msg string) {
	if warningLogger != nil {
		warningLogger.Println(msg)
	}
}

func Error(msg string) {
	if errorLogger != nil {
		errorLogger.Println(msg)
	}
}
