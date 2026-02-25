package postgres

import (
	"github.com/microservices-development-hse/backend/internal/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type HistoryRepository struct {
	db *gorm.DB
}

func NewHistoryRepository(db *gorm.DB) *HistoryRepository {
	return &HistoryRepository{db: db}
}

func (r *HistoryRepository) AddStatusChange(change *models.StatusChanges) error {
	err := r.db.Create(change).Error
	if err != nil {
		logrus.Errorf("Failed to record status change for issue %d: %v", change.IssueID, err)
		return err
	}
	return nil
}

func (r *HistoryRepository) GetHistoryByIssueID(issueID int) ([]models.StatusChanges, error) {
	var changes []models.StatusChanges
	err := r.db.Where("issue_id = ?", issueID).Order("change_time ASC").Find(&changes).Error
	if err != nil {
		logrus.Errorf("Failed to fetch history for issue %d: %v", issueID, err)
		return nil, err
	}
	return changes, nil
}
