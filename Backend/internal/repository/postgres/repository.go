package postgres

import (
	"fmt"
	"time"

	"github.com/microservices-development-hse/backend/internal/config"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repositories struct {
	Project   *ProjectRepository
	Analytics *AnalyticsRepository
	Issue     *IssueRepository
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		Project:   NewProjectRepository(db),
		Analytics: NewAnalyticsRepository(db),
		Issue:     NewIssueRepository(db),
	}
}

func InitializeRepositories(cfg *config.Config) (*Repositories, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		cfg.DBSettings.DBHost, cfg.DBSettings.DBUser, cfg.DBSettings.DBPassword,
		cfg.DBSettings.DBName, cfg.DBSettings.DBPort,
	)

	// db, err := gorm.Open(gorm_postgres.Open(dsn), &gorm.Config{})
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to connect database: %w", err)
	// }
	var db *gorm.DB
	var err error

	for i := 0; i < 10; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}

		logrus.Infof("Database not ready, retrying in 5s... (Attempt %d/10)", i+1)
		time.Sleep(5 * time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after retries: %w", err)
	}

	repos := NewRepositories(db)

	return repos, nil
}
