package models

import "time"

type StatusChanges struct {
	IssueID    int       `json:"issue_id" db:"issue_id"`
	AuthorID   int       `json:"author_id" db:"author_id"`
	ChangeTime time.Time `json:"change_time" db:"change_time"`
	FromStatus string    `json:"from_status" db:"from_status"`
	ToStatus   string    `json:"to_status" db:"to_status"`
}

type HistoryRepository interface {
	Add(history *StatusChanges) error
	GetByIssueID(issueID int) ([]StatusChanges, error)
	GetByAuthorID(authorID int) ([]StatusChanges, error)
	DeleteByIssueID(issueID int) error
}
