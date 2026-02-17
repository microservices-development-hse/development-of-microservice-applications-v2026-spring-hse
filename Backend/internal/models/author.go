package models

type Author struct {
	ID   int    `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

type AuthorRepository interface {
	Save(author *Author) error
	GetByID(id int) (*Author, error)
	GetAll() ([]Author, error)
}
