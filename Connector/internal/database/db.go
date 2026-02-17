package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/microservices-development-hse/connector/config"
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
		conn.Close()
		return fmt.Errorf("db.Ping: %w", err)
	}

	db = conn
	log.Println("[INFO] Database connection established")
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
	db = nil
	return err
}

func Ping() error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	return db.Ping()
}
