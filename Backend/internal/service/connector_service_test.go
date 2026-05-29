package service

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	pb "github.com/microservices-development-hse/backend/internal/generated/connector"
	"github.com/microservices-development-hse/backend/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestConnectorService(t *testing.T) {
	mockClient := mocks.NewConnectorServiceClient(t)
	svc := NewConnectorService(mockClient, "http://localhost:8082")

	t.Run("FetchRemoteProjects - Success", func(t *testing.T) {
		mockRPCResp := &pb.ProjectList{
			Projects: []*pb.Project{
				{Key: "JIRA-1", Title: "Remote Task"}, // ✅ убрали пробелы
			},
		}

		// ✅ Исправленный сигнатур мока (3 аргумента: ctx, req, opts)
		mockClient.On("FetchRemoteProjects", mock.Anything, mock.Anything, mock.Anything).
			Return(mockRPCResp, nil).Once()

		projects, err := svc.FetchRemoteProjects()

		assert.NoError(t, err)
		assert.Len(t, projects, 1)
		assert.Equal(t, "JIRA-1", projects[0].Key)
		mockClient.AssertExpectations(t)
	})

	t.Run("TriggerProjectImport - Success", func(t *testing.T) {
		projectKey := "HSE-CODE"

		// ✅ Мокаем HTTP-эндпоинт через httptest
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/import", r.URL.Path)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		}))
		defer server.Close()

		svc := NewConnectorService(mockClient, server.URL)

		err := svc.TriggerProjectImport(projectKey)

		assert.NoError(t, err)
	})

	t.Run("TriggerProjectImport - Remote Error", func(t *testing.T) {
		// ✅ Мокаем ошибку от HTTP-сервера
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"import failed"}`))
		}))
		defer server.Close()

		svc := NewConnectorService(mockClient, server.URL)

		err := svc.TriggerProjectImport("BAD-KEY")

		assert.Error(t, err)
		if err != nil { // ✅ Защита от nil pointer
			assert.Contains(t, err.Error(), "kafka service returned: 500")
		}
	})
}

func TestConnectorService_Errors(t *testing.T) {
	mockClient := mocks.NewConnectorServiceClient(t)
	svc := NewConnectorService(mockClient, "http://localhost:8082")

	t.Run("FetchRemoteProjects - gRPC error", func(t *testing.T) {
		mockClient.On("FetchRemoteProjects", mock.Anything, mock.Anything, mock.Anything).
			Return((*pb.ProjectList)(nil), fmt.Errorf("connection refused")).Once()

		projects, err := svc.FetchRemoteProjects()

		assert.Error(t, err)
		assert.Nil(t, projects)
		assert.Contains(t, err.Error(), "connection refused")
		mockClient.AssertExpectations(t)
	})

	t.Run("TriggerProjectImport - HTTP error", func(t *testing.T) {
		// ✅ Мокаем недоступный хост для ошибки сетевого уровня
		svc := NewConnectorService(mockClient, "http://127.0.0.1:1")

		err := svc.TriggerProjectImport("KEY")

		assert.Error(t, err)
		if err != nil { // ✅ Защита от nil pointer
			assert.Contains(t, err.Error(), "kafka service unavailable")
		}
	})
}
