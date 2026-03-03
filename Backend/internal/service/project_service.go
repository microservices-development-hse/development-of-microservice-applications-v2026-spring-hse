package service

import (
	"fmt"

	"github.com/microservices-development-hse/backend/internal/domain"
)

type AnalyticsService struct {
	repo domain.AnalyticsRepository
}

func NewAnalyticsService(r domain.AnalyticsRepository) *AnalyticsService {
	return &AnalyticsService{repo: r}
}

func (s *AnalyticsService) GetAllProjects(page, limit int, search string) (domain.ProjectsResponse, error) {
	res, err := s.repo.GetAllProjects(page, limit, search)
	if err != nil {
		return domain.ProjectsResponse{}, fmt.Errorf("AnalyticsService.GetAllProjects: %w", err)
	}

	return res, nil
}

func (s *AnalyticsService) AddProjectFromJira(key string) (domain.OperationResult, error) {
	res, err := s.repo.AddProjectFromJira(key)
	if err != nil {
		return domain.OperationResult{}, fmt.Errorf("AnalyticsService.AddProjectFromJira: %w", err)
	}

	return res, nil
}

func (s *AnalyticsService) DeleteProjectByID(id int) (domain.OperationResult, error) {
	res, err := s.repo.DeleteProjectByID(id)
	if err != nil {
		return domain.OperationResult{}, fmt.Errorf("AnalyticsService.DeleteProjectByID: %w", err)
	}

	return res, nil
}

func (s *AnalyticsService) GetProjectStatByID(id string) (domain.ProjectStat, error) {
	res, err := s.repo.GetProjectStatByID(id)
	if err != nil {
		return domain.ProjectStat{}, fmt.Errorf("AnalyticsService.GetProjectStatByID: %w", err)
	}

	return res, nil
}

func (s *AnalyticsService) MakeGraph(task, project string) (domain.GraphJob, error) {
	res, err := s.repo.MakeGraph(task, project)
	if err != nil {
		return domain.GraphJob{}, fmt.Errorf("AnalyticsService.MakeGraph: %w", err)
	}

	return res, nil
}

func (s *AnalyticsService) GetGraph(task, project string) (domain.GraphResponse, error) {
	res, err := s.repo.GetGraph(task, project)
	if err != nil {
		return domain.GraphResponse{}, fmt.Errorf("AnalyticsService.GetGraph: %w", err)
	}

	return res, nil
}

func (s *AnalyticsService) CompareGraphs(task string, projects []string) (domain.GraphResponse, error) {
	res, err := s.repo.CompareGraphs(task, projects)
	if err != nil {
		return domain.GraphResponse{}, fmt.Errorf("AnalyticsService.CompareGraphs: %w", err)
	}

	return res, nil
}

func (s *AnalyticsService) DeleteGraphs(project string) (domain.OperationResult, error) {
	res, err := s.repo.DeleteGraphs(project)
	if err != nil {
		return domain.OperationResult{}, fmt.Errorf("AnalyticsService.DeleteGraphs: %w", err)
	}

	return res, nil
}

func (s *AnalyticsService) IsAnalyzed(project string) (bool, error) {
	res, err := s.repo.IsAnalyzed(project)
	if err != nil {
		return false, fmt.Errorf("AnalyticsService.IsAnalyzed: %w", err)
	}

	return res, nil
}

func (s *AnalyticsService) IsEmpty(project string) (bool, error) {
	res, err := s.repo.IsEmpty(project)
	if err != nil {
		return false, fmt.Errorf("AnalyticsService.IsEmpty: %w", err)
	}

	return res, nil
}
