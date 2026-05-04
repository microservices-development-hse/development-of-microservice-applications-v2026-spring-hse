package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/microservices-development-hse/backend/internal/config"
	"github.com/microservices-development-hse/backend/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestInitializeHandlers(t *testing.T) {
	services := &service.Services{}
	h := InitializeHandlers(services)

	assert.NotNil(t, h.Project)
	assert.NotNil(t, h.Analytics)
	assert.NotNil(t, h.Issue)
	assert.NotNil(t, h.Connector)
}

func TestSendJSON_EncodingError(t *testing.T) {
	rr := httptest.NewRecorder()

	invalidData := make(chan int)

	sendJSON(rr, http.StatusOK, invalidData)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRouter_Coverage(t *testing.T) {
	cfg := &config.Config{}
	h := &Handlers{
		Project: NewProjectHandler(nil),
	}

	r := NewRouter(cfg, h)
	assert.NotNil(t, r)
}
