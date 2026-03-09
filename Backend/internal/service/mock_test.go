package service

import (
	"context"
	"fmt"

	"github.com/microservices-development-hse/backend/internal/models"
	"github.com/stretchr/testify/mock"
)

type MockAnalyticsRepository struct {
	mock.Mock
}

func (m *MockAnalyticsRepository) wrapError(err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("mock repository error: %w", err)
}

func (m *MockAnalyticsRepository) SaveSnapshot(ctx context.Context, snapshot *models.AnalyticsSnapshot) error {
	args := m.Called(ctx, snapshot)
	return m.wrapError(args.Error(0))
}

func (m *MockAnalyticsRepository) GetLatestSnapshot(ctx context.Context, projectID int, reportType string) (*models.AnalyticsSnapshot, error) {
	args := m.Called(ctx, projectID, reportType)

	var snapshot *models.AnalyticsSnapshot
	if args.Get(0) != nil {
		snapshot = args.Get(0).(*models.AnalyticsSnapshot)
	}

	return snapshot, m.wrapError(args.Error(1))
}

func (m *MockAnalyticsRepository) GetTaskStatusDistribution(ctx context.Context, projectID int) ([]models.DistributionItem, error) {
	args := m.Called(ctx, projectID)
	return args.Get(0).([]models.DistributionItem), m.wrapError(args.Error(1))
}

func (m *MockAnalyticsRepository) GetTaskPriorityDistribution(ctx context.Context, projectID int) ([]models.DistributionItem, error) {
	args := m.Called(ctx, projectID)
	return args.Get(0).([]models.DistributionItem), m.wrapError(args.Error(1))
}

func (m *MockAnalyticsRepository) GetProjectComplexity(ctx context.Context, projectID int) ([]models.TaskComplexity, error) {
	args := m.Called(ctx, projectID)
	return args.Get(0).([]models.TaskComplexity), m.wrapError(args.Error(1))
}

func (m *MockAnalyticsRepository) GetOpenTasksBottlenecks(ctx context.Context, projectID int) ([]models.OpenTaskDuration, error) {
	args := m.Called(ctx, projectID)
	return args.Get(0).([]models.OpenTaskDuration), m.wrapError(args.Error(1))
}

func (m *MockAnalyticsRepository) CalculateTimeInState(ctx context.Context, projectID int) (map[string]float64, error) {
	args := m.Called(ctx, projectID)

	var res map[string]float64
	if args.Get(0) != nil {
		res = args.Get(0).(map[string]float64)
	}

	return res, m.wrapError(args.Error(1))
}
