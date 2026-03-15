package service

import "github.com/microservices-development-hse/backend/internal/repository/postgres"

type Services struct {
	Project   ProjectService
	Analytics AnalyticsService
	Issue     IssueService
}

func InitializeServices(repos *postgres.Repositories) *Services {
	return &Services{
		Project:   NewProjectService(repos.Project),
		Analytics: NewAnalyticsService(repos.Analytics),
		Issue:     NewIssueService(repos.Issue, repos.Author),
	}
}
