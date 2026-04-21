package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/microservices-development-hse/backend/internal/handler/mocks"
	"github.com/microservices-development-hse/backend/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestIssueHandler_GetProjectIssues(t *testing.T) {
	mockSvc := new(mocks.IssueService)
	h := NewIssueHandler(mockSvc)

	t.Run("Success Pagination", func(t *testing.T) {
		projectID := 1
		limit := 10
		page := 1
		mockIssues := []models.Issue{
			{ID: 101, Key: "TASK-1", Summary: "First Task"},
			{ID: 102, Key: "TASK-2", Summary: "Second Task"},
		}

		mockSvc.On("GetIssuesByProject", projectID, limit, page).
			Return(mockIssues, 2, nil).Once()

		req := httptest.NewRequest("GET", "/api/v1/projects/1/issues?limit=10&page=1", nil)
		req = withChiContext(req, "id", "1")
		rr := httptest.NewRecorder()

		h.GetProjectIssues(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "TASK-1")
		assert.Contains(t, rr.Body.String(), "pageInfo")
	})

	t.Run("Invalid Project ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/projects/0/issues", nil)
		req = withChiContext(req, "id", "0")
		rr := httptest.NewRecorder()

		h.GetProjectIssues(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid project ID")
	})

	t.Run("Service Error (logrus.Errorf coverage)", func(t *testing.T) {
		mockSvc.On("GetIssuesByProject", 1, 10, 1).
			Return(nil, 0, errors.New("db connection lost")).Once()

		req := httptest.NewRequest("GET", "/api/v1/projects/1/issues", nil)
		req = withChiContext(req, "id", "1")
		rr := httptest.NewRecorder()

		h.GetProjectIssues(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "Internal server error")
	})
}

func TestIssueHandler_GetIssueByKey(t *testing.T) {
	mockSvc := new(mocks.IssueService)
	h := NewIssueHandler(mockSvc)

	t.Run("Success Detail", func(t *testing.T) {
		issueKey := "HSE-123"
		mockIssue := &models.Issue{ID: 1, Key: issueKey, Summary: "Fix analytical bug"}

		mockSvc.On("GetIssueDetails", issueKey).Return(mockIssue, nil).Once()

		req := httptest.NewRequest("GET", "/api/v1/issues/HSE-123", nil)
		req = withChiContext(req, "key", issueKey)
		rr := httptest.NewRecorder()

		h.GetIssueByKey(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "Fix analytical bug")
	})

	t.Run("Issue Not Found", func(t *testing.T) {
		mockSvc.On("GetIssueDetails", "UNKNOWN").
			Return(nil, errors.New("not found")).Once()

		req := httptest.NewRequest("GET", "/api/v1/issues/UNKNOWN", nil)
		req = withChiContext(req, "key", "UNKNOWN")
		rr := httptest.NewRecorder()

		h.GetIssueByKey(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "Issue not found")
	})
}
