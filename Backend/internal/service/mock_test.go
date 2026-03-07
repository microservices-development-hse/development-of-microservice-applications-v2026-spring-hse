package service

import (
	"context"

	"github.com/microservices-development-hse/backend/internal/models"
	"github.com/stretchr/testify/mock"
)

// MockAnalyticsRepository — мок-структура для подмены реального репозитория в тестах сервиса.
// Она реализует интерфейс AnalyticsRepository из файла models/analytics.go.
type MockAnalyticsRepository struct {
	mock.Mock
}

func (m *MockAnalyticsRepository) SaveSnapshot(ctx context.Context, snapshot *models.AnalyticsSnapshot) error {
	args := m.Called(ctx, snapshot)
	return args.Error(0)
}

func (m *MockAnalyticsRepository) GetLatestSnapshot(ctx context.Context, projectID int, reportType string) (*models.AnalyticsSnapshot, error) {
	args := m.Called(ctx, projectID, reportType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*models.AnalyticsSnapshot), args.Error(1)
}

func (m *MockAnalyticsRepository) GetTaskStatusDistribution(ctx context.Context, projectID int) ([]models.DistributionItem, error) {
	args := m.Called(ctx, projectID)
	return args.Get(0).([]models.DistributionItem), args.Error(1)
}

func (m *MockAnalyticsRepository) GetTaskPriorityDistribution(ctx context.Context, projectID int) ([]models.DistributionItem, error) {
	args := m.Called(ctx, projectID)
	return args.Get(0).([]models.DistributionItem), args.Error(1)
}

func (m *MockAnalyticsRepository) GetProjectComplexity(ctx context.Context, projectID int) ([]models.TaskComplexity, error) {
	args := m.Called(ctx, projectID)
	return args.Get(0).([]models.TaskComplexity), args.Error(1)
}

func (m *MockAnalyticsRepository) GetOpenTasksBottlenecks(ctx context.Context, projectID int) ([]models.OpenTaskDuration, error) {
	args := m.Called(ctx, projectID)
	return args.Get(0).([]models.OpenTaskDuration), args.Error(1)
}

func (m *MockAnalyticsRepository) CalculateTimeInState(ctx context.Context, projectID int) (map[string]float64, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(map[string]float64), args.Error(1)
}
