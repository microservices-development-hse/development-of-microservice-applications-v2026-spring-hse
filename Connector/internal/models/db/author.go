package db

type Author struct {
	ID         int    `db:"id"`
	ExternalID string `db:"external_id"`
	Username   string `db:"name"`
}
