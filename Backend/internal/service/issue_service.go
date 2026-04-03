package service

import (
	"fmt"

	"github.com/microservices-development-hse/backend/internal/models"
	"github.com/sirupsen/logrus"
)

type IssueService interface {
	GetIssuesByProject(projectID, limit, page int) ([]models.Issue, int, error)
	GetIssueDetails(key string) (*models.Issue, error)
	SyncIssue(issue *models.Issue, authorData *models.Author) error
}

type issueService struct {
	issueRepo  models.IssueRepository
	authorRepo models.AuthorRepository
}

func NewIssueService(ir models.IssueRepository, ar models.AuthorRepository) IssueService {
	return &issueService{
		issueRepo:  ir,
		authorRepo: ar,
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

func (s *issueService) SyncIssue(issue *models.Issue, authorData *models.Author) error {
	if authorData != nil {
		author, err := s.authorRepo.GetAuthorByExternalID(authorData.ExternalID)
		if err != nil {
			return fmt.Errorf("failed to sync author during issue sync: %w", err)
		}

		if author == nil {
			if err := s.authorRepo.CreateAuthor(authorData); err != nil {
				return fmt.Errorf("failed to create author: %w", err)
			}

			issue.AuthorID = authorData.ID
		} else {
			issue.AuthorID = author.ID
		}
	}

	existingIssue, err := s.issueRepo.GetIssueByExternalID(issue.ExternalID)
	if err != nil {
		return fmt.Errorf("failed to check existing issue: %w", err)
	}

	if existingIssue == nil {
		if err := s.issueRepo.CreateIssue(issue); err != nil {
			return fmt.Errorf("failed to create issue: %w", err)
		}

		logrus.Infof("Service: issue %s created", issue.Key)
	} else {
		oldStatus := existingIssue.Status
		issue.ID = existingIssue.ID

		if err := s.issueRepo.UpdateIssueWithHistory(issue, oldStatus); err != nil {
			return fmt.Errorf("failed to update issue with history: %w", err)
		}

		logrus.Infof("Service: issue %s updated (status sync)", issue.Key)
	}

	return nil
}
