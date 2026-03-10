package service

import (
	"fmt"

	"github.com/microservices-development-hse/backend/internal/models"
	"github.com/sirupsen/logrus"
)

type ProjectService interface {
	GetProjectsList(limit, page int) ([]models.Project, int, error)
	GetProjectDetails(id int) (*models.Project, map[string]interface{}, error)
	CreateProject(key, title string) (*models.Project, error)
	UpdateProject(id int, key, title string) (*models.Project, error)
	DeleteProject(id int) error
}

type projectService struct {
	repo models.ProjectRepository
}

func NewProjectService(repo models.ProjectRepository) ProjectService {
	return &projectService{repo: repo}
}

func (s *projectService) CreateProject(key, title string) (*models.Project, error) {
	project := &models.Project{
		Key:   key,
		Title: title,
	}

	if err := s.repo.CreateProject(project); err != nil {
		logrus.Errorf("Service: failed to create project: %v", err)
		return nil, fmt.Errorf("repository error: %w", err)
	}

	return project, nil
}

func (s *projectService) UpdateProject(id int, key, title string) (*models.Project, error) {
	project, err := s.repo.GetProjectByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to find project for update: %w", err)
	}

	if project == nil {
		return nil, fmt.Errorf("project with ID %d not found", id)
	}

	project.Key = key
	project.Title = title

	if err := s.repo.UpdateProject(project); err != nil {
		logrus.Errorf("Service: failed to update project %d: %v", id, err)
		return nil, fmt.Errorf("repository error: %w", err)
	}

	return project, nil
}

func (s *projectService) DeleteProject(id int) error {
	if err := s.repo.DeleteProject(id); err != nil {
		logrus.Errorf("Service: failed to delete project %d: %v", id, err)
		return fmt.Errorf("repository error: %w", err)
	}

	return nil
}

func (s *projectService) GetProjectsList(limit, page int) ([]models.Project, int, error) {
	offset := (page - 1) * limit

	projects, totalCount, err := s.repo.GetAllProjects(limit, offset)
	if err != nil {
		logrus.Errorf("Service: could not retrieve projects list: %v", err)
		return nil, 0, fmt.Errorf("repository error: %w", err)
	}

	return projects, totalCount, nil
}

func (s *projectService) GetProjectDetails(id int) (*models.Project, map[string]interface{}, error) {
	project, err := s.repo.GetProjectByID(id)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get project: %w", err)
	}

	if project == nil {
		return nil, nil, fmt.Errorf("project with ID %d not found", id)
	}

	stats, err := s.repo.GetBasicStats(id)
	if err != nil {
		logrus.Warnf("Service: statistics for project %d are incomplete: %v", id, err)
	}

	return project, stats, nil
}
