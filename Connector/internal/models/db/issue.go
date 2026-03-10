package db

// import "time"

type Issue struct {
	ID        int         `db:"id"`
	ProjectID int         `db:"project_id"`
	Key       string      `db:"key"`
	Summary   string      `db:"summary"`
	Status    string      `db:"status"`
	Created   string      `db:"created"` //
	Updated   string      `db:"updated"` //
	Changelog interface{} `db:"changelog"`
}
