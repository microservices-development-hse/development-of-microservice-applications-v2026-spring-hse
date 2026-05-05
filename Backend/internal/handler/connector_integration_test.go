package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/microservices-development-hse/backend/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestConnector_Integration_RealFlow(t *testing.T) {
	env := SetupIntegrationEnv(t)
	r := NewRouter(env.Cfg, env.Handlers)

	server := httptest.NewServer(r)
	defer server.Close()

	t.Run("StartImport_ActualCall", func(t *testing.T) {
		project := models.Project{
			Key:   "ALOIS",
			Title: "Test Connector Project",
		}
		env.DB.Create(&project)

		requestBody, _ := json.Marshal(map[string]string{
			"project_key": project.Key,
		})

		resp, err := http.Post(
			server.URL+"/api/v1/connector/import",
			"application/json",
			bytes.NewBuffer(requestBody),
		)

		assert.NoError(t, err)

		if resp.StatusCode == http.StatusInternalServerError {
			t.Log("Warning: Connector service might not be running (gRPC error)")
		} else {
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var result map[string]string

			err = json.NewDecoder(resp.Body).Decode(&result)
			assert.NoError(t, err)
			assert.Equal(t, "import triggered", result["status"])
		}
	})

	t.Run("Import_EmptyKey_BadRequest", func(t *testing.T) {
		requestBody, _ := json.Marshal(map[string]string{
			"project_key": "",
		})

		resp, err := http.Post(
			server.URL+"/api/v1/connector/import",
			"application/json",
			bytes.NewBuffer(requestBody),
		)

		assert.NoError(t, err)
		// Проверяем валидацию на пустой ключ
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
