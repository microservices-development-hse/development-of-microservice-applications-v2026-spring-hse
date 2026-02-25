package main

import (
	"github.com/microservices-development-hse/connector/config"
	"github.com/microservices-development-hse/connector/internal/database"
	"github.com/microservices-development-hse/connector/internal/logger"
)

func main() {
	if err := logger.Init(); err != nil {
		panic(err)
	}

	logger.Info("Application starting")

	cfg, err := config.Load("config.yaml")
	if err != nil {
		logger.Error(err.Error())
		return
	}

	if err := database.Init(cfg.DBSettings); err != nil {
		logger.Error(err.Error())
		return
	}

	if err := database.InitStatements(); err != nil {
		logger.Error(err.Error())
		return
	}

	logger.Info("Application started successfully")
}
