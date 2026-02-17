package models

import (
	"encoding/json"
	"time"
)

type OpenTaskTimeData struct {
	IDProject    int             `json:"project_id"`
	CreationTime time.Time       `json:"creation_time"`
	Data         json.RawMessage `json:"data"`
}

type TaskPriorityCountData struct {
	IDProject    int             `json:"project_id"`
	CreationTime time.Time       `json:"creation_time"`
	State        string          `json:"state"`
	Data         json.RawMessage `json:"data"`
}

type TaskStateTimeData struct {
	IDProject    int             `json:"project_id"`
	CreationTime time.Time       `json:"creation_time"`
	Data         json.RawMessage `json:"data"`
	State        string          `json:"state"`
}

type ComplexityTaskTimeData struct {
	IDProject    int             `json:"project_id"`
	CreationTime time.Time       `json:"creation_time"`
	Data         json.RawMessage `json:"data"`
}

type ActivityByTaskData struct {
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
