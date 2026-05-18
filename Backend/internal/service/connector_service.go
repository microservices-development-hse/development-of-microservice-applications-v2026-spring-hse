package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	pb "github.com/microservices-development-hse/backend/internal/generated/connector"
	"github.com/microservices-development-hse/backend/internal/models"
)

type ConnectorService interface {
	FetchRemoteProjects() ([]models.Project, error)
	TriggerProjectImport(projectKey string) error
}

type connectorService struct {
	grpcClient pb.ConnectorServiceClient
	kafkaURL   string
	httpClient *http.Client
}

func NewConnectorService(client pb.ConnectorServiceClient, kafkaURL string) ConnectorService {
	return &connectorService{
		grpcClient: client,
		kafkaURL:   kafkaURL,
		httpClient: &http.Client{},
	}
}

func (s *connectorService) FetchRemoteProjects() ([]models.Project, error) {
	ctx := context.Background()

	resp, err := s.grpcClient.FetchRemoteProjects(ctx, &pb.Empty{})
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
	body, _ := json.Marshal(map[string]string{"project_key": projectKey})

	resp, err := s.httpClient.Post(
		s.kafkaURL+"/import",
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return fmt.Errorf("kafka service unavailable: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("kafka service returned: %d", resp.StatusCode)
	}

	return nil
}
