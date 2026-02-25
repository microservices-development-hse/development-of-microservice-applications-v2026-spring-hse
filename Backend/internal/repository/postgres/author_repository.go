package postgres

import (
	"github.com/microservices-development-hse/backend/internal/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type AuthorRepository struct {
	db *gorm.DB
}

func NewAuthorRepository(db *gorm.DB) *AuthorRepository {
	return &AuthorRepository{db: db}
}

func (r *AuthorRepository) GetAuthorByName(name string) (*models.Author, error) {
	var author models.Author
	err := r.db.Where(models.Author{Name: name}).FirstOrCreate(&author).Error
	if err != nil {
		logrus.Errorf("Failed to get author %s: %v", name, err)
		return nil, err
	}
	return &author, nil
}
