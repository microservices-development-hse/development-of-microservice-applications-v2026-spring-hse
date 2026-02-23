package main

import (
	"fmt"

	"github.com/microservices-development-hse/backend/internal/logger"
	"github.com/microservices-development-hse/backend/internal/models"
	"github.com/microservices-development-hse/backend/internal/repository/postgres"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := logger.InitLogger(); err != nil {
		fmt.Printf("Failed to init logger: %v\n", err)
		return
	}

	dsn := "host=localhost user=postgres password=yourpassword dbname=postgres port=5432 sslmode=disable"

	db, closeDB, err := postgres.InitDB(dsn)
	if err != nil {
		logrus.Fatalf("Could not initialize database: %v", err)
	}

	defer closeDB()
	logrus.Info("Application started successfully")

	projectRepo := postgres.NewProjectRepository(db)
	newProject := &models.Project{
		Key:   "WEB",
		Title: "Frontend Angular App",
	}
	// --- ТЕСТ 1: Создание проекта ---
	logrus.Info("Starting test: CreateProject")
	if err := projectRepo.CreateProject(newProject); err != nil {
		logrus.Errorf("Test failed at creation: %v", err)
	}

	// --- ТЕСТ 2: Поиск по Ключу (то, что мы добавили) ---
	logrus.Info("Starting test: GetProjectByKey")
	project, err := projectRepo.GetProjectByKey("WEB")
	if err != nil {
		logrus.Errorf("Test failed at searching: %v", err)
	} else if project != nil {
		logrus.Infof("Found project: %s with ID: %d", project.Title, project.ID)
	}

	// --- ТЕСТ 3: Получение всех проектов ---
	logrus.Info("Starting test: GetAllProjects")
	all, err := projectRepo.GetAllProjects()
	if err != nil {
		logrus.Errorf("Test failed at getting all: %v", err)
	}
	logrus.Infof("Total projects in DB: %d", len(all))

	logrus.Info("All tests completed successfully!")
}
