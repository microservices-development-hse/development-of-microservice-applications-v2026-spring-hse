package db

import "time"

type Issue struct {
	ID          int        `db:"id"`
	ExternalID  string     `db:"external_id"`
	ProjectID   int        `db:"project_id"`
	AuthorID    *int       `db:"author_id"`
	AssigneeID  *int       `db:"assignee_id"`
	Key         string     `db:"key"`
	Summary     string     `db:"summary"`
	Priority    string     `db:"priority"`
	Status      string     `db:"status"`
	CreatedTime time.Time  `db:"created_time"`
	ClosedTime  *time.Time `db:"closed_time"`
	UpdatedTime time.Time  `db:"updated_time"`
	TimeSpent   int        `db:"time_spent"`
}
