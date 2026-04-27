package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/microservices-development-hse/backend/internal/config"
	"github.com/microservices-development-hse/backend/internal/models"
	"github.com/microservices-development-hse/backend/internal/repository/postgres"
	"github.com/microservices-development-hse/backend/internal/service"
	"github.com/stretchr/testify/assert"
	gormpg "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// setupTestDeps собирает реальную цепочку зависимостей
func setupTestDeps(t *testing.T) (*gorm.DB, *Handlers) {
	dsn := "host=localhost user=pguser password=pgpassword dbname=testdb port=5432 sslmode=disable"
	db, err := gorm.Open(gormpg.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Не удалось подключиться к тестовой БД: %v", err)
	}

	// Очистка таблиц перед тестом, чтобы избежать конфликтов ID
	db.Exec("TRUNCATE TABLE projects, analytics_snapshots CASCADE")

	// Инициализация слоев
	projRepo := postgres.NewProjectRepository(db)
	projSvc := service.NewProjectService(projRepo)

	analRepo := postgres.NewAnalyticsRepository(db)
	analSvc := service.NewAnalyticsService(analRepo)

	// Собираем Handlers
	h := &Handlers{
		Project:   NewProjectHandler(projSvc),
		Analytics: NewAnalyticsHandler(analSvc, projSvc),
	}

	return db, h
}

func TestRouter_FullIntegration_Analytics(t *testing.T) {
	db, h := setupTestDeps(t)
	cfg := &config.Config{}
	r := NewRouter(cfg, h)
	server := httptest.NewServer(r)
	defer server.Close()

	t.Run("Recalculate_Success_With_Real_DB", func(t *testing.T) {
		// 1. Создаем реальный проект в БД
		project := models.Project{
			Key:   "HSE",
			Title: "Test Integration Project",
		}

		db.Create(&project)

		// 2. Отправляем запрос на пересчет
		url := fmt.Sprintf("%s/api/v1/projects/%d/analytics/recalculate", server.URL, project.ID)
		resp, err := http.Post(url, "application/json", nil)

		// 3. Проверяем, что хендлер пустил нас дальше (Exists сработал)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusAccepted, resp.StatusCode)

		// 4. Проверяем асинхронную работу (что снимок появился в базе спустя время)
		assert.Eventually(t, func() bool {
			var snapshot models.AnalyticsSnapshot
			err := db.Where("project_id = ?", project.ID).First(&snapshot).Error
			return err == nil
		}, 5*time.Second, 500*time.Millisecond, "Analytics snapshot must be created in Database")
	})

	t.Run("Recalculate_NotFound", func(t *testing.T) {
		// Запрос для ID, которого точно нет
		url := fmt.Sprintf("%s/api/v1/projects/9999/analytics/recalculate", server.URL)
		resp, err := http.Post(url, "application/json", nil)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestRouter_FullIntegration_GetProjects(t *testing.T) {
	db, h := setupTestDeps(t)
	cfg := &config.Config{}
	r := NewRouter(cfg, h)
	server := httptest.NewServer(r)
	defer server.Close()

	db.Create(&models.Project{Key: "P1", Title: "Project 1"})
	db.Create(&models.Project{Key: "P2", Title: "Project 2"})

	resp, err := http.Get(server.URL + "/api/v1/projects?limit=10&page=1")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result struct {
		Projects []models.Project `json:"projects"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	assert.Len(t, result.Projects, 2)
	assert.Equal(t, "P1", result.Projects[0].Key)
}
