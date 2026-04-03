package postgres

import (
	"errors"
	"fmt"
	"math"

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
	err := r.db.Table("projects").Create(project).Error
	if err != nil {
		logrus.Errorf("Failed to create project %s: %v", project.Title, err)
		return err
	}

	logrus.Infof("Project created: %s (ID: %d)", project.Title, project.ID)

	return nil
}

func (r *ProjectRepository) GetAllProjects(limit, offset int) ([]models.Project, int, error) {
	var (
		projects   []models.Project
		totalCount int64
	)

	if err := r.db.Table("projects").Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count projects: %w", err)
	}

	if err := r.db.Table("projects").Limit(limit).Offset(offset).Find(&projects).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch projects page: %w", err)
	}

	return projects, int(totalCount), nil
}

func (r *ProjectRepository) GetProjectByID(id int) (*models.Project, error) {
	var project models.Project

	err := r.db.Table("projects").Where("id = ?", id).First(&project).Error
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

	err := r.db.Table("projects").Where("key = ?", key).First(&project).Error
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

	if err := r.db.Table("issues").Where("project_id = ?", projectID).Count(&total).Error; err != nil {
		logrus.Errorf("Repository: stats error (total) for project %d: %v", projectID, err)
		return nil, err
	}

	if err := r.db.Table("issues").Where("project_id = ? AND closed_time IS NOT NULL", projectID).Count(&closed).Error; err != nil {
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
	err := r.db.Table("projects").Where("id = ?", project.ID).Updates(project).Error
	if err != nil {
		logrus.Errorf("Failed to update project ID %d: %v", project.ID, err)
		return err
	}

	logrus.Infof("Project ID %d updated successfully", project.ID)

	return nil
}

func (r *ProjectRepository) DeleteProject(id int) error {
	err := r.db.Table("projects").Where("id = ?", id).Delete(nil).Error
	if err != nil {
		logrus.Errorf("Failed to delete project ID %d: %v", id, err)
		return err
	}

	logrus.Infof("Project ID %d deleted", id)

	return nil
}

func (r *ProjectRepository) GetDryStatistics(projectID int) (map[string]interface{}, error) {
	var stats struct {
		Total      int64
		Open       int64
		Closed     int64
		Reopened   int64
		Resolved   int64
		InProgress int64
		AvgLead    *float64 // Указатель на случай NULL
	}

	err := r.db.Table("issues").
		Select(`
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE closed_time IS NULL) as open,
			COUNT(*) FILTER (WHERE closed_time IS NOT NULL) as closed,
			COUNT(*) FILTER (WHERE status = 'Reopened') as reopened,
			COUNT(*) FILTER (WHERE status = 'Resolved') as resolved,
			COUNT(*) FILTER (WHERE status = 'In Progress') as in_progress,
			AVG(EXTRACT(EPOCH FROM (closed_time - created_time)) / 3600) FILTER (WHERE closed_time IS NOT NULL) as avg_lead
		`).
		Where("project_id = ?", projectID).
		Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	// Среднее количество задач в день за последнюю неделю
	var weeklyCount int64
	err = r.db.Table("issues").
		Where("project_id = ? AND created_time > NOW() - INTERVAL '7 days'", projectID).
		Count(&weeklyCount).Error

	if err != nil {
		return nil, err
	}

	avgLeadValue := 0.0
	if stats.AvgLead != nil {
		avgLeadValue = *stats.AvgLead
	}

	return map[string]interface{}{
		"total_tasks":       stats.Total,
		"open_tasks":        stats.Open,
		"closed_tasks":      stats.Closed,
		"reopened_tasks":    stats.Reopened,
		"resolved_tasks":    stats.Resolved,
		"in_progress_tasks": stats.InProgress,
		"avg_lead_time_h":   avgLeadValue,
		"avg_daily_weekly":  math.Round(float64(weeklyCount)/7.0*100) / 100,
	}, nil
}
