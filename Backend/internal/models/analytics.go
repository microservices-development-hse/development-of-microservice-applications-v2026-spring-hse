package models

import (
	"context"
	"encoding/json"
	"time"
)

type AnalyticsSnapshot struct {
	ID           uint            `gorm:"primaryKey" json:"id"`
	ProjectID    int             `gorm:"index;not null" json:"project_id"`
	Type         string          `gorm:"type:varchar(50);index;not null" json:"type"`
	CreationTime time.Time       `gorm:"default:now()" json:"creation_time"`
	Data         json.RawMessage `gorm:"type:jsonb;not null" json:"data"`
}

func (AnalyticsSnapshot) TableName() string {
	return "AnalyticsSnapshot"
}

type TaskComplexity struct {
	IssueKey  string  `json:"issue_key"`
	LeadTime  float64 `json:"lead_time"` // в часах
	MoveCount int     `json:"move_count"`
}

type OpenTaskDuration struct {
	IssueKey      string  `json:"issue_key"`
	CurrentStatus string  `json:"current_status"`
	TimeInStatus  float64 `json:"time_in_status"` // в часах
}

type DistributionItem struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type AnalyticsRepository interface {
	SaveSnapshot(ctx context.Context, snapshot *AnalyticsSnapshot) error
	GetLatestSnapshot(ctx context.Context, projectID int, reportType string) (*AnalyticsSnapshot, error)

	GetTaskStatusDistribution(ctx context.Context, projectID int) ([]DistributionItem, error)
	GetTaskPriorityDistribution(ctx context.Context, projectID int) ([]DistributionItem, error)
	GetProjectComplexity(ctx context.Context, projectID int) ([]TaskComplexity, error)
	GetOpenTasksBottlenecks(ctx context.Context, projectID int) ([]OpenTaskDuration, error)
	CalculateTimeInState(ctx context.Context, projectID int) (map[string]float64, error)
}
