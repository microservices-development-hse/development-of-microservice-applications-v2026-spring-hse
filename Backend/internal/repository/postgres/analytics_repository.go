package postgres

import (
	"github.com/microservices-development-hse/backend/internal/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type AnalyticsRepository struct {
	db *gorm.DB
}

func NewAnalyticsRepository(db *gorm.DB) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

func (r *AnalyticsRepository) SaveTaskStateTime(data *models.TaskStateTime) error {
	err := r.db.Save(data).Error
	if err != nil {
		logrus.Errorf("Failed to save state analytics for project %d: %v", data.IDProject, err)
		return err
	}
	return nil
}

func (r *AnalyticsRepository) GetStateAnalytics(projectID int) ([]models.TaskStateTime, error) {
	var results []models.TaskStateTime
	err := r.db.Where("project_id = ?", projectID).Find(&results).Error
	if err != nil {
		logrus.Errorf("Failed to fetch state analytics: %v", err)
		return nil, err
	}
	return results, nil
}
