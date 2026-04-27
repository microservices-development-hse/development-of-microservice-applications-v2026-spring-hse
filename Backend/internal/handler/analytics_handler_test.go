package handler

import (
	"context"
	"encoding/json"
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
	t.Run("Success", func(t *testing.T) {
		mAn := new(mocks.AnalyticsService)
		mPr := new(mocks.ProjectService)
		h := NewAnalyticsHandler(mAn, mPr)

		projectID := 1
		mPr.On("Exists", projectID).Return(true, nil).Once()
		mAn.On("GetLatestSnapshot", mock.Anything, projectID, "status").
			Return(&models.AnalyticsSnapshot{Data: json.RawMessage(`{"status":"ok"}`)}, nil).Once()

		req := httptest.NewRequest("GET", "/api/v1/projects/1/analytics?type=status", nil)
		req = withChiContext(req, "id", "1")
		rr := httptest.NewRecorder()

		h.GetAnalytics(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mPr.AssertExpectations(t)
		mAn.AssertExpectations(t)
	})

	t.Run("Project_Not_Found", func(t *testing.T) {
		mAn := new(mocks.AnalyticsService)
		mPr := new(mocks.ProjectService)
		h := NewAnalyticsHandler(mAn, mPr)

		projectID := 404
		mPr.On("Exists", projectID).Return(false, nil).Once()

		req := httptest.NewRequest("GET", "/api/v1/projects/404/analytics?type=status", nil)
		req = withChiContext(req, "id", "404")
		rr := httptest.NewRecorder()

		h.GetAnalytics(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		mPr.AssertExpectations(t)
		mAn.AssertNotCalled(t, "GetLatestSnapshot", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("Missing_Type", func(t *testing.T) {
		mAn := new(mocks.AnalyticsService)
		mPr := new(mocks.ProjectService)
		h := NewAnalyticsHandler(mAn, mPr)

		projectID := 1
		// Твой хендлер вызывает Exists() ПЕРЕД проверкой типа, поэтому мок ДОЛЖЕН знать, что ответить
		mPr.On("Exists", projectID).Return(true, nil).Once()

		req := httptest.NewRequest("GET", "/api/v1/projects/1/analytics", nil) // Нет параметра ?type=
		req = withChiContext(req, "id", "1")
		rr := httptest.NewRecorder()

		h.GetAnalytics(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		mPr.AssertExpectations(t)
	})
}

func TestAnalyticsHandler_Recalculate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mAn := new(mocks.AnalyticsService)
		mPr := new(mocks.ProjectService)
		h := NewAnalyticsHandler(mAn, mPr)

		projectID := 1
		mPr.On("Exists", projectID).Return(true, nil).Once()
		mAn.On("RunFullAnalysis", projectID).Return().Once()

		req := httptest.NewRequest("POST", "/api/v1/projects/1/analytics/recalculate", nil)
		req = withChiContext(req, "id", "1")
		rr := httptest.NewRecorder()

		h.Recalculate(rr, req)

		assert.Equal(t, http.StatusAccepted, rr.Code)
		mPr.AssertExpectations(t)
		mAn.AssertExpectations(t)
	})

	t.Run("Project_Not_Found", func(t *testing.T) {
		mAn := new(mocks.AnalyticsService)
		mPr := new(mocks.ProjectService)
		h := NewAnalyticsHandler(mAn, mPr)

		projectID := 999
		mPr.On("Exists", projectID).Return(false, nil).Once()

		req := httptest.NewRequest("POST", "/api/v1/projects/999/analytics/recalculate", nil)
		req = withChiContext(req, "id", "999")
		rr := httptest.NewRecorder()

		h.Recalculate(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		mPr.AssertExpectations(t)
		mAn.AssertNotCalled(t, "RunFullAnalysis", mock.Anything)
	})

	t.Run("Invalid_ID", func(t *testing.T) {
		h := NewAnalyticsHandler(nil, nil)
		req := httptest.NewRequest("POST", "/api/v1/projects/abc/analytics/recalculate", nil)
		req = withChiContext(req, "id", "abc")
		rr := httptest.NewRecorder()

		h.Recalculate(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}
