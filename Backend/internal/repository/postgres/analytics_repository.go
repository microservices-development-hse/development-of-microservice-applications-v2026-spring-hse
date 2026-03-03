package postgres

import (
	"context"

	"github.com/microservices-development-hse/backend/internal/models"
	"gorm.io/gorm"
)

type AnalyticsRepository struct {
	db *gorm.DB
}

func NewAnalyticsRepository(db *gorm.DB) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

func (r *AnalyticsRepository) SaveSnapshot(ctx context.Context, snapshot *models.AnalyticsSnapshot) error {
	return r.db.WithContext(ctx).Create(snapshot).Error
}

func (r *AnalyticsRepository) GetLatestSnapshot(ctx context.Context, projectID int, reportType string) (*models.AnalyticsSnapshot, error) {
	var snapshot models.AnalyticsSnapshot

	err := r.db.Where("project_id = ? AND type = ?", projectID, reportType).
		Order("creation_time DESC").
		First(&snapshot).Error
	if err != nil {
		return nil, err
	}

	return &snapshot, nil
}

func (r *AnalyticsRepository) GetTaskStatusDistribution(ctx context.Context, projectID int) ([]models.DistributionItem, error) {
	var results []models.DistributionItem

	err := r.db.WithContext(ctx).Table("Issue").
		Select("status as name, count(*) as value").
		Where("project_id = ?", projectID).
		Group("status").
		Scan(&results).Error

	return results, err
}

func (r *AnalyticsRepository) GetTaskPriorityDistribution(ctx context.Context, projectID int) ([]models.DistributionItem, error) {
	var results []models.DistributionItem

	err := r.db.WithContext(ctx).Table("Issue").
		Select("priority as name, count(*) as value").
		Where("project_id = ?", projectID).
		Group("priority").
		Scan(&results).Error

	return results, err
}

func (r *AnalyticsRepository) GetProjectComplexity(ctx context.Context, projectID int) ([]models.TaskComplexity, error) {
	var results []models.TaskComplexity

	err := r.db.WithContext(ctx).Table("Issue").
		Select("key as issue_key, "+
			"EXTRACT(EPOCH FROM (closed_time - created_time))/3600 as lead_time, "+
			"COUNT(StatusChanges.issue_id) as move_count").
		Joins("LEFT JOIN StatusChanges ON StatusChanges.issue_id = Issue.id").
		Where("Issue.project_id = ? AND Issue.closed_time IS NOT NULL", projectID).
		Group("Issue.id, Issue.key, Issue.closed_time, Issue.created_time").
		Scan(&results).Error

	return results, err
}

func (r *AnalyticsRepository) GetOpenTasksBottlenecks(ctx context.Context, projectID int) ([]models.OpenTaskDuration, error) {
	var results []models.OpenTaskDuration

	query := `
		SELECT 
			i.key as issue_key, 
			i.status as current_status,
			EXTRACT(EPOCH FROM (NOW() - COALESCE(MAX(sc.change_time), i.created_time)))/3600 as time_in_status
		FROM "Issue" i
		LEFT JOIN "StatusChanges" sc ON sc.issue_id = i.id
		WHERE i.project_id = ? AND i.closed_time IS NULL
		GROUP BY i.id, i.key, i.status, i.created_time
	`
	err := r.db.WithContext(ctx).Raw(query, projectID).Scan(&results).Error

	return results, err
}

func (r *AnalyticsRepository) CalculateTimeInState(ctx context.Context, projectID int) (map[string]float64, error) {
	var changes []models.StatusChanges

	err := r.db.Table("StatusChanges").
		Joins("JOIN Issue ON Issue.id = StatusChanges.issue_id").
		Where("Issue.project_id = ?", projectID).
		Order("issue_id, change_time ASC").
		Scan(&changes).Error
	if err != nil {
		return nil, err
	}

	stateDurations := make(map[string]float64)

	for i := 1; i < len(changes); i++ {
		if changes[i].IssueID == changes[i-1].IssueID {
			duration := changes[i].ChangeTime.Sub(changes[i-1].ChangeTime).Hours()
			stateDurations[changes[i-1].ToStatus] += duration
		}
	}

	return stateDurations, nil
}
