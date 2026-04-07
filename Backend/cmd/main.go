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

	repos, err := postgres.InitializeRepositories(cfg)
	if err != nil {
		logrus.Fatal(err)
	}

	services := service.InitializeServices(repos, cfg.ExternalServices.ConnectorURL)
	handlers := handler.InitializeHandlers(services)

	r := handler.NewRouter(cfg, handlers)

	addr := fmt.Sprintf("%s:%d", cfg.ProgramSettings.BindAddress, cfg.ProgramSettings.BindPort)
	logrus.Infof("Server is starting at %s", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		logrus.Fatalf("Server failed: %v", err)
	}
}
