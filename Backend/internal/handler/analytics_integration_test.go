package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/microservices-development-hse/backend/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestAnalytics_Integration_FullFlow(t *testing.T) {
	env := SetupIntegrationEnv(t)
	router := NewRouter(env.Cfg, env.Handlers)
	server := httptest.NewServer(router)
	defer server.Close()

	t.Run("End_To_End_Recalculate_And_Save", func(t *testing.T) {
		project := models.Project{
			Key:   "TEST_PROJ",
			Title: "Test Integration",
		}
		env.DB.Create(&project)

		url := fmt.Sprintf("%s/api/v1/projects/%d/analytics/recalculate", server.URL, project.ID)
		resp, err := http.Post(url, "application/json", nil)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusAccepted, resp.StatusCode)

		assert.Eventually(t, func() bool {
			var snapshot models.AnalyticsSnapshot
			err := env.DB.Where("project_id = ?", project.ID).First(&snapshot).Error
			return err == nil
		}, 5*time.Second, 500*time.Millisecond, "Снимок аналитики должен появиться в БД")
	})
}

func TestConnector_Integration_Import_Validation(t *testing.T) {
	env := SetupIntegrationEnv(t)
	router := NewRouter(env.Cfg, env.Handlers)
	server := httptest.NewServer(router)
	defer server.Close()

	t.Run("Invalid_Project_Key_Returns_Error", func(t *testing.T) {
		requestBody, _ := json.Marshal(map[string]string{
			"project_key": "",
		})

		url := fmt.Sprintf("%s/api/v1/connector/import", server.URL)
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
