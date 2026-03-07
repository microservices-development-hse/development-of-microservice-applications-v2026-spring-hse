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
	//dsn := "host=localhost user=postgres password=yourpassword dbname=postgres port=5432 sslmode=disable"

	db, closeDB, err := postgres.InitDB(dsn)
	if err != nil {
		logrus.Fatalf("Could not initialize database: %v", err)
	}

	defer closeDB()

	repo := postgres.NewAnalyticsRepository(db)
	serv := service.NewAnalyticsService(repo)
	hand := handler.NewAnalyticsHandler(serv, repo)

	r := handler.NewRouter(hand)

	addr := fmt.Sprintf("%s:%d", cfg.ProgramSettings.BindAddress, cfg.ProgramSettings.BindPort)
	logrus.Infof("Server is starting at %s", addr)

	server := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	if err := server.ListenAndServe(); err != nil {
		logrus.Fatalf("Server failed: %v", err)
	}
}
