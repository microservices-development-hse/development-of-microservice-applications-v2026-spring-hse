package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"fmt"

	"github.com/microservices-development-hse/connector/config"
	"github.com/microservices-development-hse/connector/internal/database"
	etlprocess "github.com/microservices-development-hse/connector/internal/etl-process"
	pb "github.com/microservices-development-hse/connector/internal/generated/connector"
	grpcserver "github.com/microservices-development-hse/connector/internal/grpc"
	jiraclient "github.com/microservices-development-hse/connector/internal/jira"
	"github.com/microservices-development-hse/connector/internal/logger"
	"github.com/microservices-development-hse/connector/internal/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// -------------------- LOGGER --------------------
	if err := logger.Init(); err != nil {
		panic(err)
	}

	logger.Info("Application starting")

	// -------------------- CONFIG --------------------
	cfgPath := "config/config.example.yaml"
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		logger.Error("config load failed: %v", err)
		return
	}

	logger.Info("Config loaded successfully")

	// -------------------- DATABASE --------------------
	if err := database.Init(cfg.DBSettings); err != nil {
		logger.Error("database init failed: %v", err)
		return
	}

	if err := database.InitStatements(); err != nil {
		logger.Error("database statements init failed: %v", err)
		return
	}

	logger.Info("Database initialized")

	defer database.CloseStatements()
	defer func() {
		if err := database.Close(); err != nil {
			logger.Error("database close failed: %v", err)
		}
	}()

	// -------------------- JIRA CLIENT --------------------
	retryCfg := jiraclient.RetryConfig{
		MinTimeSleep: cfg.ProgramSettings.MinTimeSleep,
		MaxTimeSleep: cfg.ProgramSettings.MaxTimeSleep,
	}
	client := jiraclient.NewClient(cfg.ProgramSettings.JiraURL)

	// -------------------- HTTP SERVER --------------------
	srv := server.New(
		cfg.ProgramSettings.Port,
		client,
		retryCfg,
		cfg.ProgramSettings.IssueInOneRequest,
		database.GetDB(),
		cfg.ProgramSettings.ThreadCount,
	)

	go func() {
		if err := srv.Start(); err != nil {
			logger.Error("server stopped with error: %v", err)
		}
	}()

	// -------------------- gRPC SERVER --------------------
	extractor := etlprocess.NewExtractor(
		client,
		retryCfg,
		cfg.ProgramSettings.IssueInOneRequest,
		cfg.ProgramSettings.ThreadCount,
	)
	loader := etlprocess.NewLoader(
		database.GetDB(),
		database.StmtUpsertProject,
		database.StmtUpsertIssue,
		database.StmtInsertAuthor,
		database.StmtInsertStatusChange,
	)

	grpcSrv := grpc.NewServer()
	pb.RegisterConnectorServiceServer(grpcSrv, grpcserver.NewConnectorGRPCServer(extractor, loader))
	reflection.Register(grpcSrv)

	grpcAddr := fmt.Sprintf(":%d", cfg.ProgramSettings.GRPCPort)

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Error("grpc: failed to listen on %s: %v", grpcAddr, err)
		return
	}

	go func() {
		logger.Info("grpc: listening on %s", grpcAddr)

		if err := grpcSrv.Serve(lis); err != nil {
			logger.Error("grpc: server error: %v", err)
		}
	}()

	// -------------------- GRACEFUL SHUTDOWN --------------------
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info("Shutdown signal received")

	grpcSrv.GracefulStop()
	logger.Info("gRPC server stopped")

	if err := srv.Shutdown(); err != nil {
		logger.Error("server shutdown failed: %v", err)
	}

	logger.Info("Application stopped")
}
