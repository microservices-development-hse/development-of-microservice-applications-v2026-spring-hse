package postgres

import (
	"fmt"

	"github.com/microservices-development-hse/backend/internal/config"
	gorm_postgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repositories struct {
	Project   *ProjectRepository
	Analytics *AnalyticsRepository
	Issue     *IssueRepository
	Author    *AuthorRepository
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		Project:   NewProjectRepository(db),
		Analytics: NewAnalyticsRepository(db),
		Issue:     NewIssueRepository(db),
		Author:    NewAuthorRepository(db),
	}
}

func InitializeRepositories(cfg *config.Config) (*Repositories, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		cfg.DBSettings.DBHost, cfg.DBSettings.DBUser, cfg.DBSettings.DBPassword,
		cfg.DBSettings.DBName, cfg.DBSettings.DBPort,
	)

	db, err := gorm.Open(gorm_postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	repos := NewRepositories(db)

	return repos, nil
}
