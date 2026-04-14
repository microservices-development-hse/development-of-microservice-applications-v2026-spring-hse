package service

import (
	"context"
	"errors"
	"testing"

	"github.com/microservices-development-hse/backend/internal/models"
	"github.com/microservices-development-hse/backend/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAnalyticsService_GetLatestSnapshot(t *testing.T) {
	mockRepo := mocks.NewAnalyticsRepository(t)
	svc := NewAnalyticsService(mockRepo)

	ctx := context.Background()
	projectID := 4
	reportType := "complexity"

	t.Run("Success - Snapshot exists", func(t *testing.T) {
		expectedSnapshot := &models.AnalyticsSnapshot{
			ProjectID: projectID,
			Type:      reportType,
		}

		mockRepo.On("GetLatestSnapshot", ctx, projectID, reportType).Return(expectedSnapshot, nil).Once()

		result, err := svc.GetLatestSnapshot(ctx, projectID, reportType)

		assert.NoError(t, err)
		assert.Equal(t, expectedSnapshot, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failure - Repository error", func(t *testing.T) {
		mockRepo.On("GetLatestSnapshot", ctx, projectID, reportType).
			Return((*models.AnalyticsSnapshot)(nil), errors.New("db error")).Once()

		result, err := svc.GetLatestSnapshot(ctx, projectID, reportType)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestAnalyticsService_ProcessComplexity(t *testing.T) {
	mockRepo := mocks.NewAnalyticsRepository(t)
	svc := &analyticsService{repo: mockRepo}

	ctx := context.Background()
	projectID := 10

	t.Run("Should calculate and save complexity snapshot", func(t *testing.T) {
		mockData := []models.TaskComplexity{
			{IssueKey: "HSE-1", LeadTime: 12.5, MoveCount: 2},
		}

		mockRepo.On("GetProjectComplexity", mock.Anything, projectID).Return(mockData, nil).Once()

		mockRepo.On("SaveSnapshot", mock.Anything, mock.MatchedBy(func(s *models.AnalyticsSnapshot) bool {
			return s.ProjectID == projectID && s.Type == "complexity"
		})).Return(nil).Once()

		svc.processComplexity(ctx, projectID)

		mockRepo.AssertExpectations(t)
	})
}

func TestAnalyticsService_ProcessLifeCycle(t *testing.T) {
	mockRepo := mocks.NewAnalyticsRepository(t)
	svc := &analyticsService{repo: mockRepo}
	ctx := context.Background()
	projectID := 5

	t.Run("Should process life cycle data correctly", func(t *testing.T) {
		mockLifeCycleData := map[string][]float64{
			"In Progress": {2.5, 4.0},
			"Review":      {1.2},
		}

		mockRepo.On("CalculateTimeInState", mock.Anything, projectID).
			Return(mockLifeCycleData, nil).Once()

		mockRepo.On("SaveSnapshot", mock.Anything, mock.MatchedBy(func(s *models.AnalyticsSnapshot) bool {
			return s.Type == "life_cycle"
		})).Return(nil).Once()

		svc.processLifeCycle(ctx, projectID)

		mockRepo.AssertExpectations(t)
	})
}

func TestAnalyticsService_ProcessDistribution(t *testing.T) {
	mockRepo := mocks.NewAnalyticsRepository(t)
	svc := &analyticsService{repo: mockRepo}
	ctx := context.Background()

	t.Run("Should process distribution and save snapshot", func(t *testing.T) {
		mockData := []models.DistributionItem{{Name: "To Do", Value: 5}}

		mockRepo.On("GetTaskStatusDistribution", mock.Anything, 1).Return(mockData, nil).Once()

		mockRepo.On("SaveSnapshot", mock.Anything, mock.MatchedBy(func(s *models.AnalyticsSnapshot) bool {
			return s.ProjectID == 1 && s.Type == "status"
		})).Return(nil).Once()

		svc.processDistribution(ctx, 1, "status")

		mockRepo.AssertExpectations(t)
	})
}

func TestAnalyticsService_ProcessBottlenecks(t *testing.T) {
	mockRepo := mocks.NewAnalyticsRepository(t)
	svc := &analyticsService{repo: mockRepo}
	ctx := context.Background()
	projectID := 7

	t.Run("Should process bottlenecks correctly", func(t *testing.T) {
		mockData := []models.OpenTaskDuration{
			{IssueKey: "HSE-2", CurrentStatus: "In Progress", TimeInStatus: 48.5},
		}

		mockRepo.On("GetOpenTasksBottlenecks", mock.Anything, projectID).Return(mockData, nil).Once()
		mockRepo.On("SaveSnapshot", mock.Anything, mock.MatchedBy(func(s *models.AnalyticsSnapshot) bool {
			return s.Type == "bottlenecks" && s.ProjectID == projectID
		})).Return(nil).Once()

		svc.processBottlenecks(ctx, projectID)
		mockRepo.AssertExpectations(t)
	})
}

func TestAnalyticsService_ProcessComplexity_Error(t *testing.T) {
	mockRepo := mocks.NewAnalyticsRepository(t)
	svc := &analyticsService{repo: mockRepo}

	t.Run("Should log error and return when DB fails", func(t *testing.T) {
		mockRepo.On("GetProjectComplexity", mock.Anything, mock.Anything).
			Return(([]models.TaskComplexity)(nil), errors.New("db disconnect")).Once()

		svc.processComplexity(context.Background(), 1)

		mockRepo.AssertExpectations(t)
	})
}

func TestAnalyticsService_ProcessErrors(t *testing.T) {
	mockRepo := mocks.NewAnalyticsRepository(t)
	svc := &analyticsService{repo: mockRepo}
	ctx := context.Background()

	t.Run("processDistribution - Repo Error", func(t *testing.T) {
		mockRepo.On("GetTaskStatusDistribution", ctx, 1).
			Return([]models.DistributionItem{}, errors.New("fail")).Once()

		svc.processDistribution(ctx, 1, "status")
	})
}
