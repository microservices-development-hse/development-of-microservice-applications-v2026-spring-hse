package postgres

import (
	"errors"
	"fmt"

	"github.com/microservices-development-hse/backend/internal/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{
		db: db,
	}
}

func (r *ProjectRepository) CreateProject(project *models.Project) error {
	err := r.db.Create(project).Error
	if err != nil {
		logrus.Errorf("Failed to create project %s: %v", project.Title, err)
		return err
	}

	logrus.Infof("Project created: %s (ID: %d)", project.Title, project.ID)

	return nil
}

func (r *ProjectRepository) GetAllProjects(limit, offset int) ([]models.Project, int, error) {
	var projects []models.Project

	var totalCount int64

	countQuery := `SELECT count(*) FROM "Project"`
	if err := r.db.Raw(countQuery).Scan(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count projects: %w", err)
	}

	dataQuery := `SELECT * FROM "Project" LIMIT ? OFFSET ?`
	if err := r.db.Raw(dataQuery, limit, offset).Scan(&projects).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch projects page: %w", err)
	}

	return projects, int(totalCount), nil
}

func (r *ProjectRepository) GetProjectByID(id int) (*models.Project, error) {
	var project models.Project

	err := r.db.First(&project, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logrus.Infof("Project with ID %d not found", id)
			return nil, nil
		}

		logrus.Errorf("Error fetching project by ID %d: %v", id, err)

		return nil, err
	}

	return &project, nil
}

func (r *ProjectRepository) GetProjectByKey(key string) (*models.Project, error) {
	var project models.Project

	err := r.db.Where("key = ?", key).First(&project).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logrus.Infof("Project with key %s not found", key)
			return nil, nil
		}

		logrus.Errorf("Error fetching project by key %s: %v", key, err)

		return nil, err
	}

	return &project, nil
}

func (r *ProjectRepository) GetBasicStats(projectID int) (map[string]interface{}, error) {
	var total, closed int64

	if err := r.db.Table("Issue").Where("project_id = ?", projectID).Count(&total).Error; err != nil {
		logrus.Errorf("Repository: stats error (total) for project %d: %v", projectID, err)
		return nil, err
	}

	if err := r.db.Table("Issue").Where("project_id = ? AND closed_time IS NOT NULL", projectID).Count(&closed).Error; err != nil {
		logrus.Errorf("Repository: stats error (closed) for project %d: %v", projectID, err)
		return nil, err
	}

	return map[string]interface{}{
		"total_tasks":  total,
		"closed_tasks": closed,
		"open_tasks":   total - closed,
	}, nil
}

func (r *ProjectRepository) UpdateProject(project *models.Project) error {
	err := r.db.Model(project).Updates(project).Error
	if err != nil {
		logrus.Errorf("Failed to update project ID %d: %v", project.ID, err)
		return err
	}

	logrus.Infof("Project ID %d updated successfully", project.ID)

	return nil
}

func (r *ProjectRepository) DeleteProject(id int) error {
	err := r.db.Delete(&models.Project{}, id).Error
	if err != nil {
		logrus.Errorf("Failed to delete project ID %d: %v", id, err)
		return err
	}

	logrus.Infof("Project ID %d deleted", id)

	return nil
}
