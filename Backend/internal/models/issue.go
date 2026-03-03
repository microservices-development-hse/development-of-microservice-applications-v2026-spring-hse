package models

import (
	"time"
)

type Issue struct {
	ID          int        `json:"id" db:"id"`
	ProjectID   int        `json:"project_id" db:"project_id"`
	AuthorID    int        `json:"author_id" db:"author_id"`
	AssigneeID  int        `json:"assignee_id" db:"assignee_id"`
	Key         string     `json:"key" db:"key"`
	Summary     string     `json:"summary" db:"summary"`
	Description string     `json:"description,omitempty" db:"description"`
	Type        string     `json:"type" db:"type"`
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
	//GetIssueByProjectID(projectID int) ([]Issue, error)
	GetIssueByKey(key string) (*Issue, error)
	//UpdateStatus(id int, newStatus string) error
	//GetStatsByProject(projectID int) (map[string]int, error)
	//DeleteByProject(projectID int) error
}

func (Issue) TableName() string {
	return "Issue"
}
