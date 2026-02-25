package postgres

import (
	"github.com/microservices-development-hse/backend/internal/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type IssueRepository struct {
	db *gorm.DB
}

func NewIssueRepository(db *gorm.DB) *IssueRepository {
	return &IssueRepository{db: db}
}

func (r *IssueRepository) CreateIssue(issue *models.Issue) error {
	err := r.db.Create(issue).Error
	if err != nil {
		logrus.Errorf("Failed to create issue %s: %v", issue.Key, err)
		return err
	}
	logrus.Infof("Issue %s created successfully", issue.Key)
	return nil
}

func (r *IssueRepository) UpdateIssue(issue *models.Issue) error {
	err := r.db.Model(&models.Issue{}).Where("key = ?", issue.Key).Updates(issue).Error
	if err != nil {
		logrus.Errorf("Failed to update issue %s: %v", issue.Key, err)
		return err
	}
	return nil
}

func (r *IssueRepository) GetIssueByKey(key string) (*models.Issue, error) {
	var issue models.Issue
	err := r.db.Where("key = ?", key).First(&issue).Error
	if err != nil {
		return nil, err
	}
	return &issue, nil
}
