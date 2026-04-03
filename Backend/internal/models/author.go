package models

type Author struct {
	ID         int    `json:"id" gorm:"primaryKey"`
	ExternalID string `json:"external_id" gorm:"uniqueIndex"`
	Name       string `json:"name" db:"name"`
}

type AuthorRepository interface {
	GetAuthorByExternalID(externalID string) (*Author, error)
	CreateAuthor(author *Author) error
	UpdateAuthor(author *Author) error
	GetAuthorByName(name string) (*Author, error)
}

func (Author) TableName() string {
	return "authors"
}
