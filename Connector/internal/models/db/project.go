package db

type Project struct {
	ID    int    `db:"id"`
	Key   string `db:"key"`
	Title string `db:"title"`
	URL   string `db:"url"`
}
