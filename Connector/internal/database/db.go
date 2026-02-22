package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"github.com/microservices-development-hse/connector/config"
	"github.com/microservices-development-hse/connector/internal/logger"
)

var db *sql.DB

func Init(cfg config.DBSettings) error {
	if db != nil {
		return fmt.Errorf("database already initialized")
	}

	connStr := fmt.Sprintf(
		"user=%s password=%s host=%s port=%d dbname=%s sslmode=disable",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name,
	)

	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("sql.Open: %w", err)
	}

	conn.SetMaxOpenConns(20)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(10 * time.Minute)

	if err := conn.Ping(); err != nil {
		if closeErr := conn.Close(); closeErr != nil {
			logger.Error(fmt.Sprintf("database ping failed: %v; additionally failed to close connection: %v", err, closeErr))
			return fmt.Errorf("db.Ping: %v; close: %w", err, closeErr)
		}

		logger.Error(fmt.Sprintf("db.Ping failed: %v", err))
		return fmt.Errorf("db.Ping: %w", err)
	}

	db = conn
	logger.Info("Database connection established")
	return nil
}

func GetDB() *sql.DB {
	if db == nil {
		panic("database: GetDB called before Init")
	}
	return db
}

func Close() error {
	if db == nil {
		return nil
	}
	err := db.Close()
	if err != nil {
		logger.Error(fmt.Sprintf("failed to close database: %v", err))
	}
	db = nil
	return err
}

func Ping() error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	return db.Ping()
}
