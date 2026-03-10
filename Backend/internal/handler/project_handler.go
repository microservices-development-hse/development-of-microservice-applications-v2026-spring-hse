package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/microservices-development-hse/backend/internal/service"
	"github.com/sirupsen/logrus"
)

type ProjectHandler struct {
	service service.ProjectService
}

func NewProjectHandler(service service.ProjectService) *ProjectHandler {
	return &ProjectHandler{service: service}
}

type ProjectRequest struct {
	Key   string `json:"key"`
	Title string `json:"title"`
}

func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var req ProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	project, err := h.service.CreateProject(req.Key, req.Title)
	if err != nil {
		logrus.Errorf("Handler: failed to create project: %v", err)
		h.sendError(w, http.StatusInternalServerError, "Could not create project")
		return
	}

	h.sendJSON(w, http.StatusCreated, project)
}

func (h *ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	var req ProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	project, err := h.service.UpdateProject(id, req.Key, req.Title)
	if err != nil {
		logrus.Errorf("Handler: failed to update project %d: %v", id, err)
		h.sendError(w, http.StatusInternalServerError, "Could not update project")
		return
	}

	h.sendJSON(w, http.StatusOK, project)
}

func (h *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	if err := h.service.DeleteProject(id); err != nil {
		logrus.Errorf("Handler: failed to delete project %d: %v", id, err)
		h.sendError(w, http.StatusInternalServerError, "Could not delete project")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProjectHandler) GetAllProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := h.service.GetProjectsList()
	if err != nil {
		logrus.Errorf("Handler: failed to get projects list: %v", err)
		h.sendError(w, http.StatusInternalServerError, "Could not fetch projects")
		return
	}

	h.sendJSON(w, http.StatusOK, projects)
}

func (h *ProjectHandler) GetProjectByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	project, stats, err := h.service.GetProjectDetails(id)
	if err != nil {
		logrus.Errorf("Handler: failed to get project %d: %v", id, err)
		h.sendError(w, http.StatusNotFound, "Project not found")
		return
	}

	response := map[string]interface{}{
		"project": project,
		"stats":   stats,
	}

	h.sendJSON(w, http.StatusOK, response)
}

func (h *ProjectHandler) sendJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		logrus.Errorf("Handler: failed to encode JSON response: %v", err)
	}
}

func (h *ProjectHandler) sendError(w http.ResponseWriter, status int, message string) {
	h.sendJSON(w, status, map[string]string{"error": message})
}
