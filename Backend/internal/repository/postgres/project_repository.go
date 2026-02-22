package postgres

import (
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

func (r *ProjectRepository) GetAllProjects() ([]models.Project, error) {
	var projects []models.Project
	err := r.db.Find(&projects).Error
	if err != nil {
		logrus.Errorf("Failed to get all projects: %v", err)
		return nil, err
	}

	return projects, nil
}

func (r *ProjectRepository) GetProjectByID(id int) (*models.Project, error) {
	var project models.Project
	err := r.db.First(&project, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		logrus.Errorf("Error fetching project %d: %v", id, err)

		return nil, err
	}

	return &project, nil
}

func (r *ProjectRepository) GetProjectByKey(key string) (*models.Project, error) {
	var project models.Project

	err := r.db.Where("key = ?", key).First(&project).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logrus.Infof("Project with key %s not found", key)
			return nil, nil
		}

		logrus.Errorf("Error fetching project by key %s: %v", key, err)

		return nil, err
	}

	return &project, nil
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
