package models

type Author struct {
	ID   int    `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

type AuthorRepository interface {
	GetAuthorByName(name string) (*Author, error)
	//Save(author *Author) error
	//GetByID(id int) (*Author, error)
	//GetAll() ([]Author, error)
}

func (Author) TableName() string {
	return "Author"
}
