package service

import "github.com/microservices-development-hse/backend/internal/domain"

type AnalyticsService struct {
	repo domain.AnalyticsRepository
}

func NewAnalyticsService(r domain.AnalyticsRepository) *AnalyticsService {
	return &AnalyticsService{repo: r}
}

func (s *AnalyticsService) GetAllProjects(page, limit int, search string) (domain.ProjectsResponse, error) {
	return s.repo.GetAllProjects(page, limit, search)
}
func (s *AnalyticsService) AddProjectFromJira(key string) (domain.OperationResult, error) {
	return s.repo.AddProjectFromJira(key)
}
func (s *AnalyticsService) DeleteProjectByID(id int) (domain.OperationResult, error) {
	return s.repo.DeleteProjectByID(id)
}
func (s *AnalyticsService) GetProjectStatByID(id string) (domain.ProjectStat, error) {
	return s.repo.GetProjectStatByID(id)
}
func (s *AnalyticsService) MakeGraph(task, project string) (domain.GraphJob, error) {
	return s.repo.MakeGraph(task, project)
}
func (s *AnalyticsService) GetGraph(task, project string) (domain.GraphResponse, error) {
	return s.repo.GetGraph(task, project)
}
func (s *AnalyticsService) CompareGraphs(task string, projects []string) (domain.GraphResponse, error) {
	return s.repo.CompareGraphs(task, projects)
}
func (s *AnalyticsService) DeleteGraphs(project string) (domain.OperationResult, error) {
	return s.repo.DeleteGraphs(project)
}
func (s *AnalyticsService) IsAnalyzed(project string) (bool, error) {
	return s.repo.IsAnalyzed(project)
}
func (s *AnalyticsService) IsEmpty(project string) (bool, error) {
	return s.repo.IsEmpty(project)
}
