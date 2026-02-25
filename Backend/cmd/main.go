package main

import (
	"fmt"

	"github.com/microservices-development-hse/backend/internal/logger"
	"github.com/microservices-development-hse/backend/internal/repository/postgres"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := logger.InitLogger(); err != nil {
		fmt.Printf("Failed to init logger: %v\n", err)
		return
	}

	dsn := "host=localhost user=postgres password=yourpassword dbname=postgres port=5432 sslmode=disable"

	_, closeDB, err := postgres.InitDB(dsn)
	if err != nil {
		logrus.Fatalf("Could not initialize database: %v", err)
	}

	defer closeDB()

	logrus.Info("Application started successfully")

	//projectRepo := postgres.NewProjectRepository(db)
	//newProject := &models.Project{
	//	Key:   "WEB",
	//	Title: "Frontend Angular App",
	//}
}
