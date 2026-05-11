package service

import (
	pb "github.com/microservices-development-hse/backend/internal/generated/connector"
	"github.com/microservices-development-hse/backend/internal/repository/postgres"
)

type Services struct {
	Project   ProjectService
	Analytics AnalyticsService
	Issue     IssueService
	Connector ConnectorService
}

func InitializeServices(repos *postgres.Repositories, connectorClient pb.ConnectorServiceClient, kafkaURL string) *Services {
	return &Services{
		Project:   NewProjectService(repos.Project),
		Analytics: NewAnalyticsService(repos.Analytics),
		Issue:     NewIssueService(repos.Issue),
		Connector: NewConnectorService(connectorClient, kafkaURL),
	}
}
