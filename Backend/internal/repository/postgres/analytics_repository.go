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

	err := r.db.WithContext(ctx).
		Model(&models.AnalyticsSnapshot{}).Where("project_id = ? AND type = ?", projectID, reportType).
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

	leadTimeExpr := "EXTRACT(EPOCH FROM (issues.closed_time - issues.created_time)) / 3600"

	err := r.db.WithContext(ctx).
		Table("issues").
		Select(
			"issues.key AS issue_key",
			leadTimeExpr+" AS lead_time",
			"COUNT(sc.issue_id) AS move_count",
		).
		Joins("LEFT JOIN status_changes sc ON sc.issue_id = issues.id").
		Where("issues.project_id = ? AND issues.closed_time IS NOT NULL", projectID).
		Group("issues.id, issues.key, issues.closed_time, issues.created_time").
		Scan(&results).Error

	return results, err
}

func (r *AnalyticsRepository) GetOpenTasksBottlenecks(ctx context.Context, projectID int) ([]models.OpenTaskDuration, error) {
	var results []models.OpenTaskDuration

	timeInStatusExpr := "EXTRACT(EPOCH FROM (NOW() - COALESCE(MAX(sc.change_time), issues.created_time))) / 3600"

	err := r.db.WithContext(ctx).
		Table("issues").
		Select("issues.key AS issue_key",
			"issues.status AS current_status",
			timeInStatusExpr+" AS time_in_status").
		Joins("LEFT JOIN status_changes sc ON sc.issue_id = issues.id").
		Where("issues.project_id = ? AND issues.closed_time IS NULL", projectID).
		Group("issues.id, issues.key, issues.status, issues.created_time").
		Scan(&results).Error

	return results, err
}

func (r *AnalyticsRepository) CalculateTimeInState(ctx context.Context, projectID int) (map[string]float64, error) {
	var changes []models.StatusChanges

	err := r.db.WithContext(ctx).
		Table("status_changes sc").
		Select("sc.*").
		Joins("JOIN issues i ON i.id = sc.issue_id").
		Where("i.project_id = ?", projectID).
		Order("sc.issue_id, sc.change_time ASC").
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
