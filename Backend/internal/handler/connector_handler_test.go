package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/microservices-development-hse/backend/internal/handler/mocks"
	"github.com/microservices-development-hse/backend/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestConnectorHandler_GetExternalProjects(t *testing.T) {
	mockSvc := new(mocks.ConnectorService)
	h := NewConnectorHandler(mockSvc)

	t.Run("Success", func(t *testing.T) {
		mockProjects := []models.Project{
			{Key: "JIRA-1", Title: "External Proj 1"},
		}
		mockSvc.On("FetchRemoteProjects").Return(mockProjects, nil).Once()

		req := httptest.NewRequest("GET", "/api/v1/external/projects", nil)
		rr := httptest.NewRecorder()

		h.GetExternalProjects(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "JIRA-1")
		mockSvc.AssertExpectations(t)
	})

	t.Run("Gateway Error", func(t *testing.T) {
		mockSvc.On("FetchRemoteProjects").Return(nil, errors.New("grpc connection failed")).Once()

		req := httptest.NewRequest("GET", "/api/v1/external/projects", nil)
		rr := httptest.NewRecorder()

		h.GetExternalProjects(rr, req)

		assert.Equal(t, http.StatusBadGateway, rr.Code)
		assert.Contains(t, rr.Body.String(), "grpc connection failed")
	})
}

func TestConnectorHandler_StartImport(t *testing.T) {
	mockSvc := new(mocks.ConnectorService)
	h := NewConnectorHandler(mockSvc)

	t.Run("Success Trigger", func(t *testing.T) {
		projectKey := "HSE-PROJ"
		mockSvc.On("TriggerProjectImport", projectKey).Return(nil).Once()

		body := map[string]string{"project_key": projectKey}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/external/import", bytes.NewBuffer(jsonBody))
		rr := httptest.NewRecorder()

		h.StartImport(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "import triggered")
	})

	t.Run("Missing Key", func(t *testing.T) {
		body := map[string]string{"project_key": ""}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/external/import", bytes.NewBuffer(jsonBody))
		rr := httptest.NewRecorder()

		h.StartImport(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "project_key is required")
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/external/import", bytes.NewBufferString("{invalid-json}"))
		rr := httptest.NewRecorder()

		h.StartImport(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid request body")
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		mockSvc.On("TriggerProjectImport", "ERR").Return(errors.New("sync error")).Once()

		body := map[string]string{"project_key": "ERR"}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/v1/external/import", bytes.NewBuffer(jsonBody))
		rr := httptest.NewRecorder()

		h.StartImport(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "sync error")
	})
}
