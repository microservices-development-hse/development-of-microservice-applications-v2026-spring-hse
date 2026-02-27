package models

import (
	"encoding/json"
	"time"
)

type DistributionItem struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type OpenTaskTime struct {
	IDProject    int             `json:"project_id"`
	CreationTime time.Time       `json:"creation_time"`
	Data         json.RawMessage `json:"data"`
}

type TaskPriorityCount struct {
	IDProject    int             `json:"project_id"`
	CreationTime time.Time       `json:"creation_time"`
	State        string          `json:"state"`
	Data         json.RawMessage `json:"data"`
}

type TaskStateTime struct {
	IDProject    int             `json:"project_id"`
	CreationTime time.Time       `json:"creation_time"`
	Data         json.RawMessage `json:"data"`
	State        string          `json:"state"`
}

type ComplexityTaskTime struct {
	IDProject    int             `json:"project_id"`
	CreationTime time.Time       `json:"creation_time"`
	Data         json.RawMessage `json:"data"`
}

type ActivityByTask struct {
	IDProject    int             `json:"project_id"`
	CreationTime time.Time       `json:"creation_time"`
	State        string          `json:"state"`
	Data         json.RawMessage `json:"data"`
}

type TaskStatusDuration struct {
	Status   string
	Duration float64
}

type TaskComplexity struct {
	IssueKey  string
	LeadTime  float64
	MoveCount int
}

type OpenTaskDuration struct {
	IssueKey      string
	CurrentStatus string
	TimeInStatus  float64
}

func (OpenTaskTime) TableName() string {
	return "OpenTaskTime"
}

func (TaskStateTime) TableName() string {
	return "TaskStateTime"
}

func (ComplexityTaskTime) TableName() string {
	return "ComplexityTaskTime"
}

func (TaskPriorityCount) TableName() string {
	return "TaskPriorityCount"
}

func (ActivityByTask) TableName() string {
	return "ActivityByTask"
}

type AnalyticsRepository interface {
	SaveTaskStateTime(data *TaskStateTime) error
	GetStateAnalytics(projectID int) ([]TaskStateTime, error)
	GetTaskPriorityDistribution(projectID int) ([]DistributionItem, error)
	GetTaskStatusDistribution(projectID int) ([]DistributionItem, error)
	CalculateTimeInState(projectID int) (map[string]float64, error)
	GetProjectComplexity(projectID int) ([]TaskComplexity, error)
	GetOpenTasksBottlenecks(projectID int) ([]OpenTaskDuration, error)
}
