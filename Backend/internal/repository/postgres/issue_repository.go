package postgres

import (
	"errors"
	"fmt"

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
		return fmt.Errorf("repository error: %w", err)
	}

	logrus.Infof("Issue %s created successfully", issue.Key)

	return nil
}

func (r *IssueRepository) UpdateIssue(issue *models.Issue) error {
	err := r.db.Save(issue).Error
	if err != nil {
		logrus.Errorf("Failed to update issue %s: %v", issue.Key, err)
		return fmt.Errorf("repository error: %w", err)
	}

	logrus.Infof("Issue %s updated successfully", issue.Key)

	return nil
}

func (r *IssueRepository) GetIssueByKey(key string) (*models.Issue, error) {
	var issue models.Issue

	err := r.db.Where("key = ?", key).First(&issue).Error
	if err != nil {
		return nil, fmt.Errorf("repository error: %w", err)
	}

	return &issue, nil
}

func (r *IssueRepository) GetIssuesByProjectID(projectID int, limit, offset int) ([]models.Issue, int, error) {
	var (
		issues     []models.Issue
		totalCount int64
	)

	countQuery := `SELECT count(*) FROM issues WHERE project_id = ?`
	if err := r.db.Raw(countQuery, projectID).Scan(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count issues: %w", err)
	}

	dataQuery := `SELECT * FROM issues WHERE project_id = ? LIMIT ? OFFSET ?`
	if err := r.db.Raw(dataQuery, projectID, limit, offset).Scan(&issues).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch issues: %w", err)
	}

	return issues, int(totalCount), nil
}

func (r *IssueRepository) GetIssueByExternalID(externalID string) (*models.Issue, error) {
	var issue models.Issue

	err := r.db.Where("external_id = ?", externalID).First(&issue).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		logrus.Errorf("Failed to find issue by external ID %s: %v", externalID, err)

		return nil, fmt.Errorf("repository error: %w", err)
	}

	return &issue, nil
}

func (r *IssueRepository) DeleteIssue(id int) error {
	err := r.db.Delete(&models.Issue{}, id).Error
	if err != nil {
		logrus.Errorf("Failed to delete issue ID %d: %v", id, err)
		return fmt.Errorf("repository error: %w", err)
	}

	logrus.Infof("Issue ID %d deleted successfully", id)

	return nil
}

func (r *IssueRepository) UpdateIssueWithHistory(issue *models.Issue, fromStatus string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(issue).Error; err != nil {
			return err
		}

		if fromStatus != "" && fromStatus != issue.Status {
			change := models.StatusChanges{
				IssueID:    issue.ID,
				FromStatus: fromStatus,
				ToStatus:   issue.Status,
			}

			if err := tx.Create(&change).Error; err != nil {
				return fmt.Errorf("failed to record status change: %w", err)
			}

			logrus.Infof("History: status change recorded for issue %s (%s -> %s)", issue.Key, fromStatus, issue.Status)
		}

		return nil
	})
}
