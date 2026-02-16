// issue-4

package db

import "time"

type Issue struct {
	ID        int         `db:"id"`
	ProjectID int         `db:"project_id"`
	Key       string      `db:"key"`
	Summary   string      `db:"summary"`
	Status    string      `db:"status"`
	Created   time.Time   `db:"created"`
	Updated   time.Time   `db:"updated"`
	Changelog interface{} `db:"changelog"`
}
