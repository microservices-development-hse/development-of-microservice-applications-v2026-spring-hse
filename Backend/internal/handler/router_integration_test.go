package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/microservices-development-hse/backend/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestRouter_FullIntegration_Analytics(t *testing.T) {
	env := SetupIntegrationEnv(t)
	r := NewRouter(env.Cfg, env.Handlers)

	server := httptest.NewServer(r)

	defer server.Close()

	t.Run("Recalculate_Success_With_Real_DB", func(t *testing.T) {
		project := models.Project{
			Key:   "HSE",
			Title: "Test Integration Project",
		}

		env.DB.Create(&project)

		url := fmt.Sprintf("%s/api/v1/projects/%d/analytics/recalculate", server.URL, project.ID)
		resp, err := http.Post(url, "application/json", nil)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusAccepted, resp.StatusCode)

		assert.Eventually(t, func() bool {
			var snapshot models.AnalyticsSnapshot

			err := env.DB.Where("project_id = ?", project.ID).Find(&snapshot).Error

			return err == nil
		}, 5*time.Second, 500*time.Millisecond, "Analytics snapshot must be created in Database")
	})

	t.Run("Recalculate_NotFound", func(t *testing.T) {
		url := fmt.Sprintf("%s/api/v1/projects/9999/analytics/recalculate", server.URL)
		resp, err := http.Post(url, "application/json", nil)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestRouter_FullIntegration_GetProjects(t *testing.T) {
	env := SetupIntegrationEnv(t)
	r := NewRouter(env.Cfg, env.Handlers)
	server := httptest.NewServer(r)

	defer server.Close()

	env.DB.Create(&models.Project{Key: "P1", Title: "Project 1"})
	env.DB.Create(&models.Project{Key: "P2", Title: "Project 2"})

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
