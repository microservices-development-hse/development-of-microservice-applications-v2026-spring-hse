package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/microservices-development-hse/backend/internal/service"
)

type Handler struct {
	svc *service.AnalyticsService
}

func NewHandler(svc *service.AnalyticsService) *Handler {
	return &Handler{svc: svc}
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *Handler) GetProjects(w http.ResponseWriter, r *http.Request) {
	page := 1
	limit := 20
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			page = v
		}
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			limit = v
		}
	}
	search := r.URL.Query().Get("search")
	resp, err := h.svc.GetAllProjects(page, limit, search)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) AddProject(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Key string `json:"key"`
	}
	_ = json.NewDecoder(r.Body).Decode(&payload)
	res, err := h.svc.AddProjectFromJira(payload.Key)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/projects/")
	id, _ := strconv.Atoi(idStr)
	res, err := h.svc.DeleteProjectByID(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) ProjectStat(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/projects/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "stat" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "bad path"})
		return
	}
	id := parts[0]
	stat, err := h.svc.GetProjectStatByID(id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"data": stat})
}

func (h *Handler) MakeGraph(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Task    string `json:"task"`
		Project string `json:"project"`
	}
	_ = json.NewDecoder(r.Body).Decode(&payload)
	job, err := h.svc.MakeGraph(payload.Task, payload.Project)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, job)
}

func (h *Handler) GetGraph(w http.ResponseWriter, r *http.Request) {
	task := r.URL.Query().Get("task")
	project := r.URL.Query().Get("project")
	g, err := h.svc.GetGraph(task, project)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, g)
}

func (h *Handler) CompareGraphs(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Task     string   `json:"task"`
		Projects []string `json:"projects"`
	}
	_ = json.NewDecoder(r.Body).Decode(&payload)
	res, err := h.svc.CompareGraphs(payload.Task, payload.Projects)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) DeleteGraphs(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Project string `json:"project"`
	}
	_ = json.NewDecoder(r.Body).Decode(&payload)
	res, err := h.svc.DeleteGraphs(payload.Project)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) IsAnalyzed(w http.ResponseWriter, r *http.Request) {
	project := r.URL.Query().Get("project")
	ok, err := h.svc.IsAnalyzed(project)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"analyzed": ok})
}

func (h *Handler) IsEmpty(w http.ResponseWriter, r *http.Request) {
	project := r.URL.Query().Get("project")
	empty, err := h.svc.IsEmpty(project)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"empty": empty})
}
