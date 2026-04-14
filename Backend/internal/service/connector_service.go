package service

import (
	"context"
	"fmt"

	pb "github.com/microservices-development-hse/backend/internal/generated/connector"
	"github.com/microservices-development-hse/backend/internal/models"
)

type ConnectorService interface {
	FetchRemoteProjects() ([]models.Project, error)
	TriggerProjectImport(projectKey string) error
}

type connectorService struct {
	client pb.ConnectorServiceClient
}

func NewConnectorService(client pb.ConnectorServiceClient) ConnectorService {
	return &connectorService{
		client: client,
	}
}

func (s *connectorService) FetchRemoteProjects() ([]models.Project, error) {
	ctx := context.Background()

	resp, err := s.client.FetchRemoteProjects(ctx, &pb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("gRPC call failed: %w", err)
	}

	var projects []models.Project

	for _, p := range resp.Projects {
		projects = append(projects, models.Project{Key: p.Key, Title: p.Title})
	}

	return projects, nil
}

func (s *connectorService) TriggerProjectImport(projectKey string) error {
	ctx := context.Background()

	resp, err := s.client.TriggerProjectImport(ctx, &pb.ImportRequest{
		ProjectKey: projectKey,
	})

	if err != nil {
		return fmt.Errorf("gRPC import call failed: %w", err)
	}

	if !resp.GetSuccess() {
		return fmt.Errorf("connector failed to start import: %s", resp.GetMessage())
	}

	return nil
}
