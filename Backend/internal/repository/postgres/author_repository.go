package postgres

import (
	"fmt"

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

func (r *AuthorRepository) GetAuthorByExternalID(externalID string) (*models.Author, error) {
	var author models.Author
	err := r.db.Where("external_id = ?", externalID).First(&author).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		logrus.Errorf("Failed to find author by external ID %s: %v", externalID, err)

		return nil, fmt.Errorf("repository error: %w", err)
	}

	return &author, nil
}

func (r *AuthorRepository) CreateAuthor(author *models.Author) error {
	err := r.db.Create(author).Error
	if err != nil {
		logrus.Errorf("Failed to create author %s: %v", author.Name, err)
		return fmt.Errorf("repository error: %w", err)
	}

	logrus.Infof("Author %s (ExternalID: %s) created successfully", author.Name, author.ExternalID)

	return nil
}

func (r *AuthorRepository) UpdateAuthor(author *models.Author) error {
	err := r.db.Save(author).Error
	if err != nil {
		logrus.Errorf("Failed to update author %s: %v", author.Name, err)
		return fmt.Errorf("repository error: %w", err)
	}

	logrus.Infof("Author %s updated successfully", author.Name)

	return nil
}

func (r *AuthorRepository) GetAuthorByName(name string) (*models.Author, error) {
	var author models.Author
	err := r.db.Where("name = ?", name).First(&author).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		logrus.Errorf("Failed to find author by name %s: %v", name, err)

		return nil, fmt.Errorf("repository error: %w", err)
	}

	return &author, nil
}
