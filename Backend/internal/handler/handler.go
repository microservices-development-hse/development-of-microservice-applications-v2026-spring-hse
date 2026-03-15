package handler

import "github.com/microservices-development-hse/backend/internal/service"

type Handlers struct {
	Project   *ProjectHandler
	Analytics *AnalyticsHandler
	Issue     *IssueHandler
}

func InitializeHandlers(services *service.Services) *Handlers {
	return &Handlers{
		Project:   NewProjectHandler(services.Project),
		Analytics: NewAnalyticsHandler(services.Analytics),
		Issue:     NewIssueHandler(services.Issue),
	}
}
