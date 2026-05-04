package service

import (
	"fmt"

	"github.com/microservices-development-hse/backend/internal/models"
	"github.com/sirupsen/logrus"
)

type IssueService interface {
	GetIssuesByProject(projectID, limit, page int) ([]models.Issue, int, error)
	GetIssueDetails(key string) (*models.Issue, error)
}

type issueService struct {
	issueRepo models.IssueRepository
}

func NewIssueService(ir models.IssueRepository) IssueService {
	return &issueService{
		issueRepo: ir,
	}
}

func (s *issueService) GetIssuesByProject(projectID, limit, page int) ([]models.Issue, int, error) {
	if limit <= 0 {
		limit = 10
	}

	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * limit

	issues, totalCount, err := s.issueRepo.GetIssuesByProjectID(projectID, limit, offset)
	if err != nil {
		logrus.Errorf("Service: failed to fetch issues for project %d: %v", projectID, err)
		return nil, 0, fmt.Errorf("repository error: %w", err)
	}

	return issues, totalCount, nil
}

func (s *issueService) GetIssueDetails(key string) (*models.Issue, error) {
	issue, err := s.issueRepo.GetIssueByKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to find issue %s: %w", key, err)
	}

	if issue == nil {
		return nil, fmt.Errorf("issue %s not found", key)
	}

	return issue, nil
}
