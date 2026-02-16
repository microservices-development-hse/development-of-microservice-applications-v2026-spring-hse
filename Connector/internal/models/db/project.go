// issue-4

package db

type Project struct {
	ID   int    `db:"id"`
	Key  string `db:"key"`
	Name string `db:"name"`
	URL  string `db:"url"`
}
