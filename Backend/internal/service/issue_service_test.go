package service

import (
	"errors"
	"testing"

	"github.com/microservices-development-hse/backend/internal/models"
	"github.com/microservices-development-hse/backend/internal/service/mocks"
	"github.com/stretchr/testify/assert"
)

func TestIssueService_GetIssuesByProject(t *testing.T) {
	mockIssueRepo := mocks.NewIssueRepository(t)
	svc := NewIssueService(mockIssueRepo)

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

func TestIssueService_GetIssueDetails(t *testing.T) {
	mockIssueRepo := mocks.NewIssueRepository(t)
	svc := NewIssueService(mockIssueRepo)

	t.Run("Success - Issue found", func(t *testing.T) {
		issueKey := "HSE-101"
		expectedIssue := &models.Issue{
			ExternalID: issueKey,
			Summary:    "Fix assembly bug",
		}

		mockIssueRepo.On("GetIssueByKey", issueKey).Return(expectedIssue, nil).Once()

		result, err := svc.GetIssueDetails(issueKey)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Fix assembly bug", result.Summary)
	})

	t.Run("Failure - Repository error", func(t *testing.T) {
		issueKey := "HSE-ERR"

		mockIssueRepo.On("GetIssueByKey", issueKey).
			Return((*models.Issue)(nil), errors.New("db connection lost")).Once()

		result, err := svc.GetIssueDetails(issueKey)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "db connection lost")
	})
}
