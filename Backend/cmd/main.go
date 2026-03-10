package main

import (
	"fmt"
	"net/http"

	"github.com/microservices-development-hse/backend/internal/config"
	"github.com/microservices-development-hse/backend/internal/handler"
	"github.com/microservices-development-hse/backend/internal/logger"
	"github.com/microservices-development-hse/backend/internal/repository/postgres"
	"github.com/microservices-development-hse/backend/internal/service"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := logger.InitLogger(); err != nil {
		fmt.Printf("Failed to init logger: %v\n", err)
		return
	}

	cfg, err := config.LoadConfig("../configs/config.yaml")
	if err != nil {
		logrus.Fatalf("Config error: %v", err)
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		cfg.DBSettings.DBHost,
		cfg.DBSettings.DBUser,
		cfg.DBSettings.DBPassword,
		cfg.DBSettings.DBName,
		cfg.DBSettings.DBPort,
	)

	db, closeDB, err := postgres.InitDB(dsn)
	if err != nil {
		logrus.Fatalf("Could not initialize database: %v", err)
	}

	defer closeDB()

	projectRepo := postgres.NewProjectRepository(db)
	analyticsRepo := postgres.NewAnalyticsRepository(db)

	projectService := service.NewProjectService(projectRepo)
	analyticsService := service.NewAnalyticsService(analyticsRepo)

	projectHandler := handler.NewProjectHandler(projectService)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsService)

	r := handler.NewRouter(cfg, projectHandler, analyticsHandler)

	addr := fmt.Sprintf("%s:%d", cfg.ProgramSettings.BindAddress, cfg.ProgramSettings.BindPort)
	logrus.Infof("Server is starting at %s", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		logrus.Fatalf("Server failed: %v", err)
	}
}
