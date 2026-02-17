package db

type User struct {
	ID          int    `db:"id"`
	Username    string `db:"username"`
	DisplayName string `db:"display_name"`
}
