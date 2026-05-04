package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/microservices-development-hse/backend/internal/handler/mocks"
	"github.com/microservices-development-hse/backend/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestProjectHandler_CreateProject(t *testing.T) {
	mockSvc := new(mocks.ProjectService)
	h := NewProjectHandler(mockSvc)

	t.Run("Success", func(t *testing.T) {
		reqBody := `{"key": "HSE", "title": "New Project", "url": "http://hse.ru"}`
		expectedProject := &models.Project{ID: 1, Key: "HSE", Title: "New Project"}

		mockSvc.On("CreateProject", "HSE", "New Project", "http://hse.ru").
			Return(expectedProject, nil).Once()

		req := httptest.NewRequest("POST", "/api/v1/projects", strings.NewReader(reqBody))
		rr := httptest.NewRecorder()

		h.CreateProject(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		assert.Contains(t, rr.Body.String(), "New Project")
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/projects", strings.NewReader(`{invalid-json}`))
		rr := httptest.NewRecorder()

		h.CreateProject(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestProjectHandler_GetProjectByID(t *testing.T) {
	mockSvc := new(mocks.ProjectService)
	h := NewProjectHandler(mockSvc)

	t.Run("Success", func(t *testing.T) {
		projectID := 1
		mockProject := &models.Project{ID: projectID, Title: "HSE Test"}

		mockSvc.On("GetProjectDetails", projectID).Return(mockProject, map[string]interface{}{}, nil).Once()

		req := httptest.NewRequest("GET", "/api/v1/projects/1", nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()

		h.GetProjectByID(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "HSE Test")
		mockSvc.AssertExpectations(t)
	})

	t.Run("Invalid ID Format", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/projects/abc", nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "abc")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()

		h.GetProjectByID(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid project ID")
	})
}

func TestProjectHandler_GetAllProjects(t *testing.T) {
	mockSvc := new(mocks.ProjectService)
	h := NewProjectHandler(mockSvc)

	t.Run("Success with Pagination", func(t *testing.T) {
		mockProjects := []models.Project{{ID: 1, Title: "P1"}, {ID: 2, Title: "P2"}}
		mockSvc.On("GetProjectsList", 10, 1).Return(mockProjects, 2, nil).Once()

		req := httptest.NewRequest("GET", "/api/v1/projects?limit=10&page=1", nil)
		rr := httptest.NewRecorder()

		h.GetAllProjects(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "pageInfo")
		assert.Contains(t, rr.Body.String(), "pagesCount")
	})
}

func TestProjectHandler_UpdateProject(t *testing.T) {
	mockSvc := new(mocks.ProjectService)
	h := NewProjectHandler(mockSvc)

	t.Run("Success Update", func(t *testing.T) {
		projectID := 5
		reqBody := `{"key": "UPD", "title": "Updated Title"}`
		updatedProject := &models.Project{ID: projectID, Key: "UPD", Title: "Updated Title"}

		mockSvc.On("UpdateProject", projectID, "UPD", "Updated Title").
			Return(updatedProject, nil).Once()

		req := httptest.NewRequest("PUT", "/api/v1/projects/5", strings.NewReader(reqBody))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "5")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.UpdateProject(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "Updated Title")
	})
}

func TestProjectHandler_DeleteProject(t *testing.T) {
	mockSvc := new(mocks.ProjectService)
	h := NewProjectHandler(mockSvc)

	t.Run("Success Delete", func(t *testing.T) {
		projectID := 10
		mockSvc.On("DeleteProject", projectID).Return(nil).Once()

		req := httptest.NewRequest("DELETE", "/api/v1/projects/10", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "10")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.DeleteProject(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
	})
}

func TestProjectHandler_ErrorPaths(t *testing.T) {
	mockSvc := new(mocks.ProjectService)
	h := NewProjectHandler(mockSvc)

	t.Run("CreateProject_ServiceError", func(t *testing.T) {
		mockSvc.On("CreateProject", "FAIL", "Title", "http://err.com").
			Return(nil, errors.New("database is down")).Once()

		reqBody := `{"key": "FAIL", "title": "Title", "url": "http://err.com"}`
		req := httptest.NewRequest("POST", "/api/v1/projects", strings.NewReader(reqBody))
		rr := httptest.NewRecorder()

		h.CreateProject(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "Could not create project")
	})

	t.Run("GetAllProjects_ServiceError", func(t *testing.T) {
		mockSvc.On("GetProjectsList", 10, 1).
			Return(nil, 0, errors.New("query timeout")).Once()

		req := httptest.NewRequest("GET", "/api/v1/projects", nil)
		rr := httptest.NewRecorder()

		h.GetAllProjects(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "Could not fetch projects")
	})

	t.Run("DeleteProject_ServiceError", func(t *testing.T) {
		projectID := 99
		mockSvc.On("DeleteProject", projectID).
			Return(errors.New("permission denied")).Once()

		req := httptest.NewRequest("DELETE", "/api/v1/projects/99", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "99")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.DeleteProject(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestSendJSON_Error(t *testing.T) {
	rr := httptest.NewRecorder()
	invalidData := make(chan int)

	sendJSON(rr, http.StatusOK, invalidData)

	assert.Equal(t, http.StatusOK, rr.Code)
}
