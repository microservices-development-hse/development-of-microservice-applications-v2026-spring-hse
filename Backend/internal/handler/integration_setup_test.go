package handler

import (
	"testing"

	"github.com/microservices-development-hse/backend/internal/config"
	pb "github.com/microservices-development-hse/backend/internal/generated/connector"
	"github.com/microservices-development-hse/backend/internal/repository/postgres"
	"github.com/microservices-development-hse/backend/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	gormpg "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type SharedTestEnv struct {
	Cfg        *config.Config
	DB         *gorm.DB
	Handlers   *Handlers
	GrpcClient pb.ConnectorServiceClient
}

func SetupIntegrationEnv(t *testing.T) *SharedTestEnv {
	cfg, err := config.LoadConfig("../../configs/config.yaml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Подключение к БД для тестов
	dsn := "host=localhost user=pguser password=pgpassword dbname=testdb port=5432 sslmode=disable"
	db, err := gorm.Open(gormpg.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test DB: %v", err)
	}

	// Очистка таблиц перед тестом
	db.Exec("TRUNCATE TABLE projects, analytics_snapshots RESTART IDENTITY CASCADE")

	// Настройка gRPC клиента
	conn, err := grpc.NewClient("localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect to gRPC: %v", err)
	}
	grpcClient := pb.NewConnectorServiceClient(conn)

	repos := postgres.NewRepositories(db)

	projectSvc := service.NewProjectService(repos.Project)

	analyticsSvc := service.NewAnalyticsService(repos.Analytics)

	connectorSvc := service.NewConnectorService(grpcClient)

	h := &Handlers{
		Project:   NewProjectHandler(projectSvc),
		Analytics: NewAnalyticsHandler(analyticsSvc, projectSvc),
		Connector: NewConnectorHandler(connectorSvc),
	}

	return &SharedTestEnv{
		Cfg:        cfg,
		DB:         db,
		Handlers:   h,
		GrpcClient: grpcClient,
	}
}
