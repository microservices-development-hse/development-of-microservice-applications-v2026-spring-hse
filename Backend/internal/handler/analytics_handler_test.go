package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/microservices-development-hse/backend/internal/handler/mocks"
	"github.com/microservices-development-hse/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func withChiContext(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestAnalyticsHandler_GetAnalytics(t *testing.T) {
	mockSvc := new(mocks.AnalyticsService)
	h := NewAnalyticsHandler(mockSvc)

	t.Run("Success", func(t *testing.T) {
		projectID := 1
		reportType := "status"
		mockData := `{"open": 5, "done": 10}`

		mockSnapshot := &models.AnalyticsSnapshot{
			Data: json.RawMessage(mockData),
		}

		mockSvc.On("GetLatestSnapshot", mock.Anything, projectID, reportType).
			Return(mockSnapshot, nil).Once()

		req := httptest.NewRequest("GET", "/api/v1/projects/1/analytics?type=status", nil)
		req = withChiContext(req, "id", "1")

		rr := httptest.NewRecorder()
		h.GetAnalytics(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.JSONEq(t, mockData, rr.Body.String())
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	})

	t.Run("Missing Type Parameter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/projects/1/analytics", nil)
		req = withChiContext(req, "id", "1")

		rr := httptest.NewRecorder()
		h.GetAnalytics(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Missing type parameter")
	})

	t.Run("Invalid Report Type", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/projects/1/analytics?type=unknown", nil)
		req = withChiContext(req, "id", "1")

		rr := httptest.NewRecorder()
		h.GetAnalytics(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid report type")
	})

	t.Run("Not Found in DB", func(t *testing.T) {
		mockSvc.On("GetLatestSnapshot", mock.Anything, 1, "complexity").
			Return(nil, errors.New("not found")).Once()

		req := httptest.NewRequest("GET", "/api/v1/projects/1/analytics?type=complexity", nil)
		req = withChiContext(req, "id", "1")

		rr := httptest.NewRecorder()
		h.GetAnalytics(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestAnalyticsHandler_Recalculate(t *testing.T) {
	mockSvc := new(mocks.AnalyticsService)
	h := NewAnalyticsHandler(mockSvc)

	t.Run("Success Trigger", func(t *testing.T) {
		projectID := 1

		mockSvc.On("RunFullAnalysis", projectID).Return().Once()

		req := httptest.NewRequest("POST", "/api/v1/projects/1/analytics/recalculate", nil)
		req = withChiContext(req, "id", "1")

		rr := httptest.NewRecorder()
		h.Recalculate(rr, req)

		assert.Equal(t, http.StatusAccepted, rr.Code)
		assert.Contains(t, rr.Body.String(), "analysis started")
	})

	t.Run("Invalid Project ID", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/projects/abc/analytics/recalculate", nil)
		req = withChiContext(req, "id", "abc")

		rr := httptest.NewRecorder()
		h.Recalculate(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestAnalyticsHandler_Recalculate_Error(t *testing.T) {
	mockSvc := new(mocks.AnalyticsService)
	h := NewAnalyticsHandler(mockSvc)

	t.Run("Invalid ID", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/recalculate", nil)
		req = withChiContext(req, "id", "not-a-number")
		rr := httptest.NewRecorder()

		h.Recalculate(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}
