package db

import "time"

type StatusChange struct {
	IssueID    int       `db:"issue_id"`
	AuthorID   int       `db:"author_id"`
	ChangeTime time.Time `db:"change_time"`
	FromStatus string    `db:"from_status"`
	ToStatus   string    `db:"to_status"`
}
