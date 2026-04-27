package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/microservices-development-hse/backend/internal/config"
	"github.com/microservices-development-hse/backend/internal/handler/mocks"
	"github.com/stretchr/testify/assert"
)

func TestConnector_Integration_StartImport(t *testing.T) {
	cfg := &config.Config{}
	mockConnectorSvc := new(mocks.ConnectorService)

	mockConnectorSvc.On("TriggerProjectImport", "ALOIS").Return(nil)

	h := &Handlers{
		Connector: NewConnectorHandler(mockConnectorSvc),
	}

	r := NewRouter(cfg, h)
	server := httptest.NewServer(r)
	defer server.Close()

	requestBody, _ := json.Marshal(map[string]string{
		"project_key": "ALOIS",
	})

	resp, err := http.Post(
		server.URL+"/api/v1/connector/import",
		"application/json",
		bytes.NewBuffer(requestBody),
	)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]string
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)

	assert.Equal(t, "import triggered", result["status"])

	mockConnectorSvc.AssertExpectations(t)
}

func TestConnector_Integration_ImportInvalidBody(t *testing.T) {
	cfg := &config.Config{}
	mockConnectorSvc := new(mocks.ConnectorService)
	h := &Handlers{
		Connector: NewConnectorHandler(mockConnectorSvc),
	}
	r := NewRouter(cfg, h)
	server := httptest.NewServer(r)
	defer server.Close()

	requestBody, _ := json.Marshal(map[string]string{
		"project_key": "",
	})

	resp, err := http.Post(
		server.URL+"/api/v1/connector/import",
		"application/json",
		bytes.NewBuffer(requestBody),
	)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
