package service

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/microservices-development-hse/backend/internal/models"
	"github.com/sirupsen/logrus"
)

type AnalyticsService interface {
	RunFullAnalysis(projectID int)
}

type analyticsService struct {
	repo models.AnalyticsRepository
}

func NewAnalyticsService(repo models.AnalyticsRepository) AnalyticsService {
	return &analyticsService{repo: repo}
}

func (s *analyticsService) RunFullAnalysis(projectID int) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		logrus.Infof("Starting comprehensive analysis for project %d", projectID)

		var wg sync.WaitGroup

		analysisTasks := []struct {
			name string
			fn   func(context.Context, int)
		}{
			{"status_distribution", func(c context.Context, id int) { s.processDistribution(c, id, "status") }},
			{"priority_distribution", func(c context.Context, id int) { s.processDistribution(c, id, "priority") }},
			{"complexity", func(c context.Context, id int) { s.processComplexity(c, id) }},
			{"bottlenecks", func(c context.Context, id int) { s.processBottlenecks(c, id) }},
			{"time_in_state", func(c context.Context, id int) { s.processLifeCycle(c, id) }},
		}

		for _, task := range analysisTasks {
			wg.Add(1)

			go func(t struct {
				name string
				fn   func(context.Context, int)
			}) {
				defer wg.Done()

				logrus.Debugf("Executing sub-task: %s", t.name)
				t.fn(ctx, projectID)
			}(task)
		}

		wg.Wait()

		logrus.Infof("All analytical snapshots for project %d have been synchronized", projectID)
	}()
}

func (s *analyticsService) processDistribution(ctx context.Context, projectID int, distType string) {
	var data []models.DistributionItem

	var err error

	switch distType {
	case "status":
		data, err = s.repo.GetTaskStatusDistribution(ctx, projectID)
	case "priority":
		data, err = s.repo.GetTaskPriorityDistribution(ctx, projectID)
	default:
		logrus.Errorf("Service: unknown distribution type '%s'", distType)
		return
	}

	if err != nil {
		logrus.Errorf("Service [%s]: failed to get distribution: %v", distType, err)
		return
	}

	s.saveSnapshot(ctx, projectID, distType, data)
}

func (s *analyticsService) processComplexity(ctx context.Context, projectID int) {
	data, err := s.repo.GetProjectComplexity(ctx, projectID)
	if err != nil {
		logrus.Errorf("Service [complexity]: %v", err)
		return
	}

	s.saveSnapshot(ctx, projectID, "complexity", data)
}

func (s *analyticsService) processBottlenecks(ctx context.Context, projectID int) {
	data, err := s.repo.GetOpenTasksBottlenecks(ctx, projectID)
	if err != nil {
		logrus.Errorf("Service [bottlenecks]: %v", err)
		return
	}

	s.saveSnapshot(ctx, projectID, "bottlenecks", data)
}

func (s *analyticsService) processLifeCycle(ctx context.Context, projectID int) {
	data, err := s.repo.CalculateTimeInState(ctx, projectID)
	if err != nil {
		logrus.Errorf("Service: failed to calculate life cycle: %v", err)
		return
	}

	s.saveSnapshot(ctx, projectID, "life_cycle", data)
}

func (s *analyticsService) saveSnapshot(ctx context.Context, projectID int, t string, data interface{}) {
	jsonData, _ := json.Marshal(data)
	snapshot := &models.AnalyticsSnapshot{
		ProjectID:    projectID,
		Type:         t,
		CreationTime: time.Now(),
		Data:         json.RawMessage(jsonData),
	}

	if err := s.repo.SaveSnapshot(ctx, snapshot); err != nil {
		logrus.Errorf("Service: failed to save snapshot %s: %v", t, err)
	}
}
