package service

import (
	"bytes"
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
	body, err := json.Marshal(map[string]string{
		"project_key": projectKey,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal import request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/import", s.connectorURL), bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create import request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to trigger import: %w", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("error closing body:", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("connector returned error: %d", resp.StatusCode)
	}

	return nil
}
