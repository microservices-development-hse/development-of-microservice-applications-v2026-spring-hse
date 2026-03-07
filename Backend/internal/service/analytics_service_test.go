package service

import (
	"testing"
	"time"

	"github.com/microservices-development-hse/backend/internal/models"
	"github.com/stretchr/testify/mock"
)

func TestRunFullAnalysis(t *testing.T) {
	mockRepo := new(MockAnalyticsRepository)
	serv := NewAnalyticsService(mockRepo)
	projectID := 1

	mockRepo.On("GetTaskStatusDistribution", mock.Anything, projectID).
		Return([]models.DistributionItem{{Name: "To Do", Value: 5}}, nil)

	mockRepo.On("GetTaskPriorityDistribution", mock.Anything, projectID).
		Return([]models.DistributionItem{{Name: "High", Value: 2}}, nil)

	mockRepo.On("GetProjectComplexity", mock.Anything, projectID).
		Return([]models.TaskComplexity{{IssueKey: "TEST-1", LeadTime: 10.5}}, nil)

	mockRepo.On("GetOpenTasksBottlenecks", mock.Anything, projectID).
		Return([]models.OpenTaskDuration{{IssueKey: "TEST-2", TimeInStatus: 48.0}}, nil)

	mockRepo.On("CalculateTimeInState", mock.Anything, projectID).
		Return(map[string]float64{"In Progress": 120.5}, nil)

	mockRepo.On("SaveSnapshot", mock.Anything, mock.MatchedBy(func(s *models.AnalyticsSnapshot) bool {
		return s.ProjectID == projectID
	})).Return(nil).Times(5)

	serv.RunFullAnalysis(projectID)

	time.Sleep(100 * time.Millisecond)

	mockRepo.AssertExpectations(t)
}
