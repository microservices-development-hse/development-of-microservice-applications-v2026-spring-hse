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
	projectID, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if projectID <= 0 {
		sendError(w, http.StatusBadRequest, "Invalid project ID")

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
		sendError(w, http.StatusInternalServerError, "Internal server error")

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

	sendJSON(w, http.StatusOK, response)
}

func (h *IssueHandler) GetIssueByKey(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if key == "" {
		sendError(w, http.StatusBadRequest, "Issue key is required")

		return
	}

	issue, err := h.service.GetIssueDetails(key)
	if err != nil {
		sendError(w, http.StatusNotFound, "Issue not found")

		return
	}

	sendJSON(w, http.StatusOK, issue)
}

func (h *IssueHandler) SyncIssue(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Issue  models.Issue  `json:"issue"`
		Author models.Author `json:"author"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.SyncIssue(&payload.Issue, &payload.Author); err != nil {
		logrus.Errorf("Handler: sync failed: %v", err)
		sendError(w, http.StatusInternalServerError, "Sync failed")

		return
	}

	sendJSON(w, http.StatusOK, map[string]string{"status": "success"})
}
