package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/microservices-development-hse/kafka/config"
	"github.com/microservices-development-hse/kafka/internal/consumer"
	"github.com/microservices-development-hse/kafka/internal/logger"
	"github.com/microservices-development-hse/kafka/internal/producer"
	"github.com/microservices-development-hse/kafka/internal/server"
)

func main() {
	if err := logger.Init(); err != nil {
		panic(err)
	}

	logger.Info("Application starting")

	// ── Config ──
	cfgPath := "config/config.example.yaml"
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		logger.Error("config load failed: %v", err)
	}

	logger.Info("config loaded: kafka=%v topic=%s connector=%s port=%d", cfg.Kafka.Brokers, cfg.Kafka.Topic, cfg.Connector.GRPCAddress, cfg.Server.Port)

	// ── Producer ──
	p := producer.New(cfg)

	defer func() {
		if err := p.Close(); err != nil {
			logger.Info("producer close error: %v", err)
		}
	}()

	// ── Consumer ──
	c, err := consumer.New(cfg)
	if err != nil {
		logger.Error("consumer init failed: %v", err)
	}

	defer func() { _ = c.Close() }()

	// ── HTTP server ──
	srv := server.New(cfg.Server.Port, p)

	// ── Graceful shutdown ──
	ctx, cancel := context.WithCancel(context.Background())
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		logger.Info("shutdown signal received")
		cancel()

		if err := srv.Shutdown(context.Background()); err != nil {
			logger.Error("server shutdown failed: %v", err)
		}
	}()

	// ── Consumer ──
	go func() {
		logger.Info("consumer started")

		if err := c.Run(ctx); err != nil {
			logger.Info("consumer stopped: %v", err)
		}
	}()

	// ── HTTP server ──
	logger.Info("kafka service HTTP listening on :%d", cfg.Server.Port)

	if err := srv.Start(); err != nil {
		logger.Info("HTTP server stopped: %v", err)
	}

	logger.Info("kafka service stopped")
}
