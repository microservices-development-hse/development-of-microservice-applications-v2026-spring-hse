package service

import (
	"testing"

	"github.com/microservices-development-hse/backend/internal/models"
	"github.com/microservices-development-hse/backend/internal/service/mocks"
	"github.com/stretchr/testify/assert"
)

func TestIssueService_GetIssuesByProject(t *testing.T) {
	mockIssueRepo := mocks.NewIssueRepository(t)
	mockAuthorRepo := mocks.NewAuthorRepository(t)
	svc := NewIssueService(mockIssueRepo, mockAuthorRepo)

	projectID := 1

	t.Run("Success with pagination", func(t *testing.T) {
		limit, page := 5, 2
		expectedOffset := 5

		mockIssues := []models.Issue{
			{Key: "HSE-10", Summary: "Task 10"},
		}

		mockIssueRepo.On("GetIssuesByProjectID", projectID, limit, expectedOffset).
			Return(mockIssues, 100, nil).Once()

		issues, total, err := svc.GetIssuesByProject(projectID, limit, page)

		assert.NoError(t, err)
		assert.Equal(t, 100, total)
		assert.Len(t, issues, 1)
		mockIssueRepo.AssertExpectations(t)
	})

	t.Run("Use default values for invalid page/limit", func(t *testing.T) {
		mockIssueRepo.On("GetIssuesByProjectID", projectID, 10, 0).
			Return([]models.Issue{}, 0, nil).Once()

		_, _, err := svc.GetIssuesByProject(projectID, 0, 0)

		assert.NoError(t, err)
		mockIssueRepo.AssertExpectations(t)
	})
}

func TestIssueService_SyncIssue(t *testing.T) {
	mockIssueRepo := mocks.NewIssueRepository(t)
	mockAuthorRepo := mocks.NewAuthorRepository(t)
	svc := NewIssueService(mockIssueRepo, mockAuthorRepo)

	t.Run("Should create author and issue if they don't exist", func(t *testing.T) {
		authorData := &models.Author{ExternalID: "user-123", Name: "Bob"}
		issueData := &models.Issue{ExternalID: "ext-999", Summary: "New Task"}

		mockAuthorRepo.On("GetAuthorByExternalID", "user-123").Return((*models.Author)(nil), nil).Once()
		mockAuthorRepo.On("CreateAuthor", authorData).Return(nil).Once()

		mockIssueRepo.On("GetIssueByExternalID", "ext-999").Return((*models.Issue)(nil), nil).Once()
		mockIssueRepo.On("CreateIssue", issueData).Return(nil).Once()

		err := svc.SyncIssue(issueData, authorData)

		assert.NoError(t, err)
		mockAuthorRepo.AssertExpectations(t)
		mockIssueRepo.AssertExpectations(t)
	})
}
