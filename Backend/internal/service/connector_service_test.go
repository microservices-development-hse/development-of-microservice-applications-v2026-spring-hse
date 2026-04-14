package service

import (
	"fmt"
	"testing"

	pb "github.com/microservices-development-hse/backend/internal/generated/connector"
	"github.com/microservices-development-hse/backend/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestConnectorService(t *testing.T) {
	mockClient := mocks.NewConnectorServiceClient(t)
	svc := NewConnectorService(mockClient)

	t.Run("FetchRemoteProjects - Success", func(t *testing.T) {
		mockRPCResp := &pb.ProjectList{
			Projects: []*pb.Project{
				{Key: "JIRA-1", Title: "Remote Task"},
			},
		}

		mockClient.On("FetchRemoteProjects", mock.Anything, mock.Anything, mock.Anything).
			Return(mockRPCResp, nil).Once()

		projects, err := svc.FetchRemoteProjects()

		assert.NoError(t, err)
		assert.Len(t, projects, 1)
		assert.Equal(t, "JIRA-1", projects[0].Key)
	})

	t.Run("TriggerProjectImport - Success", func(t *testing.T) {
		projectKey := "HSE-CODE"

		mockResp := &pb.ImportResponse{Success: true}

		mockClient.On("TriggerProjectImport", mock.Anything, &pb.ImportRequest{
			ProjectKey: projectKey,
		}, mock.Anything).Return(mockResp, nil).Once()

		err := svc.TriggerProjectImport(projectKey)

		assert.NoError(t, err)
	})

	t.Run("TriggerProjectImport - Remote Error", func(t *testing.T) {
		mockResp := &pb.ImportResponse{Success: false}

		mockClient.On("TriggerProjectImport", mock.Anything, mock.Anything, mock.Anything).
			Return(mockResp, nil).Once()

		err := svc.TriggerProjectImport("BAD-KEY")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "connector failed to start import")
	})
}

func TestConnectorService_Errors(t *testing.T) {
	mockClient := mocks.NewConnectorServiceClient(t)
	svc := NewConnectorService(mockClient)

	t.Run("FetchRemoteProjects - gRPC error", func(t *testing.T) {
		mockClient.On("FetchRemoteProjects", mock.Anything, mock.Anything, mock.Anything).
			Return((*pb.ProjectList)(nil), fmt.Errorf("connection refused")).Once()

		projects, err := svc.FetchRemoteProjects()

		assert.Error(t, err)
		assert.Nil(t, projects)
		assert.Contains(t, err.Error(), "connection refused")
	})

	t.Run("TriggerProjectImport - gRPC error", func(t *testing.T) {
		mockClient.On("TriggerProjectImport", mock.Anything, mock.Anything, mock.Anything).
			Return((*pb.ImportResponse)(nil), fmt.Errorf("timeout")).Once()

		err := svc.TriggerProjectImport("KEY")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timeout")
	})
}
