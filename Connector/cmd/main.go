package main

import "github.com/microservices-development-hse/connector/internal/logger"

func main() {
	err := logger.Init()
	if err != nil {
		panic(err)
	}

	logger.Info("Application started")
}
