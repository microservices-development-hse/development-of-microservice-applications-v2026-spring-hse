package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/microservices-development-hse/backend/internal/models"
	"github.com/microservices-development-hse/backend/internal/service"
	"github.com/sirupsen/logrus"
)

type IssueHandler struct {
	service service.IssueService
}

func NewIssueHandler(s service.IssueService) *IssueHandler {
	return &IssueHandler{service: s}
}

func (h *IssueHandler) GetProjectIssues(w http.ResponseWriter, r *http.Request) {
	projectID, _ := strconv.Atoi(chi.URLParam(r, "projectID"))
	if projectID <= 0 {
		h.sendError(w, http.StatusBadRequest, "Invalid project ID")

		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 10
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}

	issues, totalCount, err := h.service.GetIssuesByProject(projectID, limit, page)
	if err != nil {
		logrus.Errorf("Handler: failed to get issues: %v", err)
		h.sendError(w, http.StatusInternalServerError, "Internal server error")

		return
	}

	response := map[string]interface{}{
		"issues": issues,
		"pageInfo": map[string]interface{}{
			"currentPage":   page,
			"projectsCount": totalCount,
			"pagesCount":    (totalCount + limit - 1) / limit,
		},
	}

	h.sendJSON(w, http.StatusOK, response)
}

func (h *IssueHandler) GetIssueByKey(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if key == "" {
		h.sendError(w, http.StatusBadRequest, "Issue key is required")

		return
	}

	issue, err := h.service.GetIssueDetails(key)
	if err != nil {
		h.sendError(w, http.StatusNotFound, "Issue not found")

		return
	}

	h.sendJSON(w, http.StatusOK, issue)
}

func (h *IssueHandler) SyncIssue(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Issue  models.Issue  `json:"issue"`
		Author models.Author `json:"author"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.SyncIssue(&payload.Issue, &payload.Author); err != nil {
		logrus.Errorf("Handler: sync failed: %v", err)
		h.sendError(w, http.StatusInternalServerError, "Sync failed")

		return
	}

	h.sendJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *IssueHandler) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logrus.Errorf("failed to encode response: %v", err)
	}
}

func (h *IssueHandler) sendError(w http.ResponseWriter, status int, message string) {
	h.sendJSON(w, status, map[string]string{"error": message})
}
