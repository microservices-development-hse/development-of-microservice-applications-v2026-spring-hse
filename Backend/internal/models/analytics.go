package models

import (
	"encoding/json"
	"time"
)

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

type AnalyticsRepository interface {
	GetAnalyticsData(projectKey string, taskNumber int) (interface{}, error)
	RunAnalysis(projectKey string, taskNumber int) error
	CheckIfAnalyzed(projectKey string) (bool, error)
}
