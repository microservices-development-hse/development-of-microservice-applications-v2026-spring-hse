package service

import (
	"errors"
	"testing"

	"github.com/microservices-development-hse/backend/internal/models"
	"github.com/microservices-development-hse/backend/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProjectService_GetProjectsList(t *testing.T) {
	mockRepo := mocks.NewProjectRepository(t)
	svc := NewProjectService(mockRepo)

	t.Run("Success retrieval", func(t *testing.T) {
		mockData := []models.Project{{ID: 1, Title: "HSE Project"}}
		mockRepo.On("GetAllProjects", 10, 0).Return(mockData, 1, nil).Once()

		projects, total, err := svc.GetProjectsList(10, 1)

		assert.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Equal(t, "HSE Project", projects[0].Title)
		mockRepo.AssertExpectations(t)
	})
}

func TestProjectService_FullCycle(t *testing.T) {
	mockRepo := mocks.NewProjectRepository(t)
	svc := NewProjectService(mockRepo)

	t.Run("Create Success", func(t *testing.T) {
		key, title, url := "HSE", "Project", "http://hse.ru"

		mockRepo.On("CreateProject", mock.MatchedBy(func(p *models.Project) bool {
			return p.Key == key && p.Title == title && p.URL == url
		})).Return(nil).Once()

		res, err := svc.CreateProject(key, title, url)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, key, res.Key)
	})

	t.Run("Create - Repo Error", func(t *testing.T) {
		mockRepo.On("CreateProject", mock.Anything).Return(errors.New("db fail")).Once()

		res, err := svc.CreateProject("K", "T", "U")

		assert.Error(t, err)
		assert.Nil(t, res)
	})
}

func TestProjectService_UpdateProject(t *testing.T) {
	mockRepo := mocks.NewProjectRepository(t)
	svc := NewProjectService(mockRepo)

	t.Run("Update - Success", func(t *testing.T) {
		existing := &models.Project{ID: 1, Key: "OLD", Title: "Old Title"}

		mockRepo.On("GetProjectByID", 1).Return(existing, nil).Once()
		mockRepo.On("UpdateProject", mock.MatchedBy(func(p *models.Project) bool {
			return p.ID == 1 && p.Key == "NEW"
		})).Return(nil).Once()

		res, err := svc.UpdateProject(1, "NEW", "New Title")
		assert.NoError(t, err)
		assert.Equal(t, "NEW", res.Key)
	})

	t.Run("Update - Project Not Found", func(t *testing.T) {
		mockRepo.On("GetProjectByID", 1).Return((*models.Project)(nil), nil).Once()

		res, err := svc.UpdateProject(1, "NEW", "New Title")

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("Update - Repository Error on Get", func(t *testing.T) {
		mockRepo.On("GetProjectByID", 1).Return((*models.Project)(nil), errors.New("db error")).Once()

		_, err := svc.UpdateProject(1, "NEW", "New Title")
		assert.Error(t, err)
	})
}

func TestProjectService_DeleteProject(t *testing.T) {
	mockRepo := mocks.NewProjectRepository(t)
	svc := NewProjectService(mockRepo)

	t.Run("Delete - Success", func(t *testing.T) {
		mockRepo.On("DeleteProject", 10).Return(nil).Once()
		err := svc.DeleteProject(10)
		assert.NoError(t, err)
	})

	t.Run("Delete - Error", func(t *testing.T) {
		mockRepo.On("DeleteProject", 10).Return(errors.New("constraint violation")).Once()
		err := svc.DeleteProject(10)
		assert.Error(t, err)
	})
}

func TestProjectService_GetProjectDetails(t *testing.T) {
	mockRepo := mocks.NewProjectRepository(t)
	svc := NewProjectService(mockRepo)

	t.Run("Success_with_details", func(t *testing.T) {
		id := 1
		mockProject := &models.Project{ID: id, Title: "Detail Test"}

		mockRepo.On("GetProjectByID", id).Return(mockProject, nil).Once()

		mockStats := map[string]interface{}{"completed_tasks": 5}
		mockRepo.On("GetDryStatistics", id).Return(mockStats, nil).Once()

		project, details, err := svc.GetProjectDetails(id)

		assert.NoError(t, err)
		assert.Equal(t, "Detail Test", project.Title)
		assert.NotNil(t, details)
		assert.Equal(t, 5, details["completed_tasks"])
	})
}

func TestProjectService_Create(t *testing.T) {
	mockRepo := mocks.NewProjectRepository(t)
	svc := NewProjectService(mockRepo)

	t.Run("Create Success", func(t *testing.T) {
		expectedProject := &models.Project{
			Key:   "HSE",
			Title: "New Project",
			URL:   "http://hse.ru",
		}

		mockRepo.On("CreateProject", expectedProject).Return(nil).Once()

		res, err := svc.CreateProject(expectedProject.Key, expectedProject.Title, expectedProject.URL)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, expectedProject.Key, res.Key)
	})
}

func TestProjectService_EdgeCases(t *testing.T) {
	mockRepo := mocks.NewProjectRepository(t)
	svc := NewProjectService(mockRepo)

	t.Run("Update - Repository Error on Final Save", func(t *testing.T) {
		existing := &models.Project{ID: 1, Key: "OLD"}
		mockRepo.On("GetProjectByID", 1).Return(existing, nil).Once()

		mockRepo.On("UpdateProject", mock.Anything).Return(errors.New("db write error")).Once()

		_, err := svc.UpdateProject(1, "NEW", "New Title")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "repository error")
	})

	t.Run("GetDetails - Stats Failure (Graceful)", func(t *testing.T) {
		id := 5
		mockProject := &models.Project{ID: id, Title: "Resilient Project"}

		mockRepo.On("GetProjectByID", id).Return(mockProject, nil).Once()

		mockRepo.On("GetDryStatistics", id).Return(nil, errors.New("stats timeout")).Once()

		project, details, err := svc.GetProjectDetails(id)

		assert.NoError(t, err)
		assert.NotNil(t, project)
		assert.Empty(t, details)
	})
}
