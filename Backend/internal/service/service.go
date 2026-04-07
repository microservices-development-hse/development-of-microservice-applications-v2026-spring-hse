package service

import "github.com/microservices-development-hse/backend/internal/repository/postgres"

type Services struct {
	Project   ProjectService
	Analytics AnalyticsService
	Issue     IssueService
	Connector ConnectorService
}

func InitializeServices(repos *postgres.Repositories, connectorURL string) *Services {
	return &Services{
		Project:   NewProjectService(repos.Project),
		Analytics: NewAnalyticsService(repos.Analytics),
		Issue:     NewIssueService(repos.Issue, repos.Author),
		Connector: NewConnectorService(connectorURL),
	}
}
