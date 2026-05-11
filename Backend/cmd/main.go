package main

import (
	"fmt"
	"net/http"

	"github.com/microservices-development-hse/backend/internal/config"
	pb "github.com/microservices-development-hse/backend/internal/generated/connector"
	"github.com/microservices-development-hse/backend/internal/handler"
	"github.com/microservices-development-hse/backend/internal/logger"
	"github.com/microservices-development-hse/backend/internal/repository/postgres"
	"github.com/microservices-development-hse/backend/internal/service"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	if err := logger.InitLogger(); err != nil {
		fmt.Printf("Failed to init logger: %v\n", err)
		return
	}

	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		logrus.Fatalf("Config error: %v", err)
	}

	conn, err := grpc.NewClient("dev_connector:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logrus.Fatalf("did not connect: %v", err)
	}

	defer func() {
		if err := conn.Close(); err != nil {
			logrus.Fatalf("failed to close connection: %v", err)
		}
	}()

	grpcClient := pb.NewConnectorServiceClient(conn)

	repos, err := postgres.InitializeRepositories(cfg)
	if err != nil {
		logrus.Fatal(err)
	}

	services := service.InitializeServices(repos, grpcClient)
	handlers := handler.InitializeHandlers(services)
	r := handler.NewRouter(cfg, handlers)
	addr := fmt.Sprintf("%s:%d", cfg.ProgramSettings.BindAddress, cfg.ProgramSettings.BindPort)

	logrus.Infof("Server is starting at %s", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		logrus.Fatalf("Server failed: %v", err)
	}
}
