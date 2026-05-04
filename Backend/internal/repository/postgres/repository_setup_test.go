package postgres

import (
	"fmt"
	"testing"

	"github.com/microservices-development-hse/backend/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type RepoTestEnv struct {
	DB  *gorm.DB
	Cfg *config.Config
}

func SetupRepoTestEnv(t *testing.T) *RepoTestEnv {
	cfg, err := config.LoadConfig("../../../configs/config.yaml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		cfg.DBSettings.DBHost,
		cfg.DBSettings.DBUser,
		cfg.DBSettings.DBPassword,
		cfg.DBSettings.DBName,
		cfg.DBSettings.DBPort,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to connect to test DB: %v", err)
	}

	db.Exec("TRUNCATE TABLE status_changes, issues, projects, authors CASCADE")

	return &RepoTestEnv{
		DB:  db,
		Cfg: cfg,
	}
}
