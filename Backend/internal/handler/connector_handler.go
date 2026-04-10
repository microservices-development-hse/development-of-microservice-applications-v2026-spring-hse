package handler

import (
	"encoding/json"
	"net/http"

	"github.com/microservices-development-hse/backend/internal/service"
)

type ConnectorHandler struct {
	service service.ConnectorService
}

func NewConnectorHandler(s service.ConnectorService) *ConnectorHandler {
	return &ConnectorHandler{service: s}
}

// GetExternalProjects возвращает список проектов из Коннектора
func (h *ConnectorHandler) GetExternalProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := h.service.FetchRemoteProjects()
	if err != nil {
		sendError(w, http.StatusBadGateway, err.Error())
		return
	}

	sendJSON(w, http.StatusOK, projects)
}

// StartImport запускает импорт по ключу проекта
func (h *ConnectorHandler) StartImport(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProjectKey string `json:"project_key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.ProjectKey == "" {
		sendError(w, http.StatusBadRequest, "project_key is required")
		return
	}

	err := h.service.TriggerProjectImport(req.ProjectKey)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendJSON(w, http.StatusOK, map[string]string{"status": "import triggered"})
}
