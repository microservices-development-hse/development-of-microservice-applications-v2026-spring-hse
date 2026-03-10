package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	etlprocess "github.com/microservices-development-hse/connector/internal/etl-process"
	jiraclient "github.com/microservices-development-hse/connector/internal/jira"
	"github.com/microservices-development-hse/connector/internal/logger"
)

type PageInfo struct {
	CurrentPage   int `json:"currentPage"`
	PageCount     int `json:"pageCount"`
	ProjectsCount int `json:"projectsCount"`
}

type ProjectItem struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type ProjectsResponse struct {
	Projects []ProjectItem `json:"projects"`
	PageInfo PageInfo      `json:"pageInfo"`
}

type ProjectsHandler struct {
	extractor *etlprocess.Extractor
}

func NewProjectsHandler(client *jiraclient.Client, retryConfig jiraclient.RetryConfig, maxResults int) *ProjectsHandler {
	return &ProjectsHandler{
		extractor: etlprocess.NewExtractor(client, retryConfig, maxResults),
	}
}

func (h *ProjectsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()

	limit := 20

	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}

	page := 1

	if v := q.Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			page = n
		}
	}

	search := strings.ToLower(q.Get("search"))

	jiraProjects, err := h.extractor.GetProjects()
	if err != nil {
		logger.Error("projects handler: %v", err)
		http.Error(w, "failed to fetch projects from Jira", http.StatusBadGateway)

		return
	}

	var filtered []ProjectItem

	for _, p := range jiraProjects {
		if search != "" {
			keyMatch := strings.Contains(strings.ToLower(p.Key), search)
			nameMatch := strings.Contains(strings.ToLower(p.Name), search)

			if !keyMatch && !nameMatch {
				continue
			}
		}

		filtered = append(filtered, ProjectItem{
			Key:  p.Key,
			Name: p.Name,
			URL:  p.Self,
		})
	}

	total := len(filtered)

	pageCount := (total + limit - 1) / limit
	if pageCount == 0 {
		pageCount = 1
	}

	start := (page - 1) * limit
	end := start + limit

	if start > total {
		start = total
	}

	if end > total {
		end = total
	}

	resp := ProjectsResponse{
		Projects: filtered[start:end],
		PageInfo: PageInfo{
			CurrentPage:   page,
			PageCount:     pageCount,
			ProjectsCount: total,
		},
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Error("projects handler: encode response: %v", err)
	}
}
