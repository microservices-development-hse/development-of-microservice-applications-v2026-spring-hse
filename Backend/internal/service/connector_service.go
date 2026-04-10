package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/microservices-development-hse/backend/internal/models"
)

type ConnectorService interface {
	FetchRemoteProjects() ([]models.Project, error)
	TriggerProjectImport(projectKey string) error
}

type connectorService struct {
	connectorURL string
}

func NewConnectorService(url string) ConnectorService {
	return &connectorService{
		connectorURL: url,
	}
}

func (s *connectorService) FetchRemoteProjects() ([]models.Project, error) {
	resp, err := http.Get(fmt.Sprintf("%s/projects", s.connectorURL))
	if err != nil {
		return nil, fmt.Errorf("connector unavailable: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	var data struct {
		Projects []models.Project `json:"projects"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode projects: %w", err)
	}

	return data.Projects, nil
}

func (s *connectorService) TriggerProjectImport(projectKey string) error {
	url := fmt.Sprintf("%s/updateProject?project=%s", s.connectorURL, projectKey)

	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return fmt.Errorf("failed to trigger import: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("connector returned error: %d", resp.StatusCode)
	}

	return nil
}
