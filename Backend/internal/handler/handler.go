package handler

import (
	"encoding/json"
	"net/http"

	"github.com/microservices-development-hse/backend/internal/service"
	"github.com/sirupsen/logrus"
)

type Handlers struct {
	Project   *ProjectHandler
	Analytics *AnalyticsHandler
	Issue     *IssueHandler
	Connector *ConnectorHandler
}

func InitializeHandlers(services *service.Services) *Handlers {
	return &Handlers{
		Project:   NewProjectHandler(services.Project),
		Analytics: NewAnalyticsHandler(services.Analytics),
		Issue:     NewIssueHandler(services.Issue),
		Connector: NewConnectorHandler(services.Connector),
	}
}

func sendJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		logrus.Errorf("Handler: failed to encode JSON response: %v", err)
	}
}

func sendError(w http.ResponseWriter, status int, message string) {
	sendJSON(w, status, map[string]string{"error": message})
}
