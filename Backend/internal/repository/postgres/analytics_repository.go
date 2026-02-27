package postgres

import (
	"fmt"
	"time"

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

func (r *AnalyticsRepository) GetTaskPriorityDistribution(projectID int) ([]models.DistributionItem, error) {
	var results []models.DistributionItem

	err := r.db.Model(&models.Issue{}).
		Select("priority as name, count(*) as value").
		Where("project_id = ?", projectID).
		Group("priority").
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to calculate priority distribution: %w", err)
	}

	return results, nil
}

func (r *AnalyticsRepository) GetTaskStatusDistribution(projectID int) ([]models.DistributionItem, error) {
	var results []models.DistributionItem

	err := r.db.Model(&models.Issue{}).
		Select("status as name, count(*) as value").
		Where("project_id = ?", projectID).
		Group("status").
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to calculate status distribution: %w", err)
	}

	return results, nil
}

func (r *AnalyticsRepository) CalculateTimeInState(projectID int) (map[string]float64, error) {
	var changes []struct {
		IssueID    int
		ToStatus   string
		ChangeTime time.Time
	}

	err := r.db.Table("status_changes").
		Select("status_changes.issue_id, status_changes.to_status, status_changes.change_time").
		Joins("JOIN issues ON issues.id = status_changes.issue_id").
		Where("issues.project_id = ?", projectID).
		Order("status_changes.issue_id, status_changes.change_time ASC").
		Scan(&changes).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch status changes: %w", err)
	}

	stateDurations := make(map[string]float64)
	now := time.Now()

	if len(changes) == 0 {
		return stateDurations, nil
	}

	for i := 0; i < len(changes); i++ {
		var duration float64

		if i+1 < len(changes) && changes[i].IssueID == changes[i+1].IssueID {
			duration = changes[i+1].ChangeTime.Sub(changes[i].ChangeTime).Hours()
		} else {
			var issue models.Issue
			if err := r.db.First(&issue, changes[i].IssueID).Error; err == nil {
				if issue.Status != "Done" {
					duration = now.Sub(changes[i].ChangeTime).Hours()
				}
			}
		}
		stateDurations[changes[i].ToStatus] += duration
	}

	return stateDurations, nil
}

func (r *AnalyticsRepository) GetProjectComplexity(projectID int) ([]models.TaskComplexity, error) {
	var results []models.TaskComplexity

	err := r.db.Table("issues").
		Select("issues.key as issue_key, "+
			"EXTRACT(EPOCH FROM (issues.closed_time - issues.created_time))/3600 as lead_time, "+
			"count(status_changes.id) as move_count").
		Joins("LEFT JOIN status_changes ON status_changes.issue_id = issues.id").
		Where("issues.project_id = ? AND issues.closed_time IS NOT NULL", projectID).
		Group("issues.id, issues.key, issues.closed_time, issues.created_time").
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to calculate complexity metrics: %w", err)
	}

	return results, nil
}

func (r *AnalyticsRepository) GetOpenTasksBottlenecks(projectID int) ([]models.OpenTaskDuration, error) {
	var results []models.OpenTaskDuration

	err := r.db.Table("issues").
		Select("issues.key as issue_key, "+
			"issues.status as current_status, "+
			"EXTRACT(EPOCH FROM (NOW() - MAX(status_changes.change_time)))/3600 as time_in_status").
		Joins("JOIN status_changes ON status_changes.issue_id = issues.id").
		Where("issues.project_id = ? AND issues.closed_time IS NULL", projectID).
		Group("issues.id, issues.key, issues.status").
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to detect bottlenecks: %w", err)
	}

	return results, nil
}
