package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/microservices-development-hse/backend/internal/models"
	"github.com/microservices-development-hse/backend/internal/service"
)

type AnalyticsHandler struct {
	service service.AnalyticsService
	repo    models.AnalyticsRepository
}

func NewAnalyticsHandler(s service.AnalyticsService, r models.AnalyticsRepository) *AnalyticsHandler {
	return &AnalyticsHandler{
		service: s,
		repo:    r,
	}
}

func (h *AnalyticsHandler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	projectID, err := strconv.Atoi(chi.URLParam(r, "projectID"))
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	reportType := r.URL.Query().Get("type")
	if reportType == "" {
		http.Error(w, "Missing type parameter", http.StatusBadRequest)
		return
	}

	validTypes := map[string]bool{
		"status":      true,
		"priority":    true,
		"complexity":  true,
		"bottlenecks": true,
		"life_cycle":  true,
	}

	if !validTypes[reportType] {
		http.Error(w, "Invalid report type. Supported: status, priority, complexity, bottlenecks, life_cycle", http.StatusBadRequest)
		return
	}

	snapshot, err := h.repo.GetLatestSnapshot(r.Context(), projectID, reportType)
	if err != nil {
		http.Error(w, "Analytics not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(snapshot.Data)
}

func (h *AnalyticsHandler) Recalculate(w http.ResponseWriter, r *http.Request) {
	projectID, err := strconv.Atoi(chi.URLParam(r, "projectID"))
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	h.service.RunFullAnalysis(projectID)

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "analysis started"})
}
