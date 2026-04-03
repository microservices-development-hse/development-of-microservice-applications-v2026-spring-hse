package models

import (
	"time"
)

type Issue struct {
	ID          int        `json:"id" db:"id"`
	ExternalID  string     `json:"external_id" db:"external_id"`
	ProjectID   int        `json:"project_id" db:"project_id"`
	AuthorID    int        `json:"author_id" db:"author_id"`
	AssigneeID  int        `json:"assignee_id" db:"assignee_id"`
	Key         string     `json:"key" db:"key"`
	Summary     string     `json:"summary" db:"summary"`
	Priority    string     `json:"priority" db:"priority"`
	Status      string     `json:"status" db:"status"`
	CreatedTime time.Time  `json:"created_time" db:"created_time"`
	ClosedTime  *time.Time `json:"closed_time,omitempty" db:"closed_time"`
	UpdatedTime time.Time  `json:"updated_time" db:"updated_time"`
	TimeSpent   int        `json:"time_spent" db:"time_spent"`
}

type IssueRepository interface {
	CreateIssue(issue *Issue) error
	UpdateIssue(issue *Issue) error
	GetIssuesByProjectID(projectID int, limit, offset int) ([]Issue, int, error)
	GetIssueByExternalID(externalID string) (*Issue, error)
	GetIssueByKey(key string) (*Issue, error)
	DeleteIssue(id int) error
}

func (Issue) TableName() string {
	return "issues"
}
