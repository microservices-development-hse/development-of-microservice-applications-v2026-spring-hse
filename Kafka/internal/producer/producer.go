package producer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/microservices-development-hse/kafka/config"
	"github.com/microservices-development-hse/kafka/internal/logger"
)

type ImportMessage struct {
	ProjectKey string `json:"project_key"`
}

type Producer struct {
	writer writer
}

type writer interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

func New(cfg *config.Config) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Kafka.Brokers...),
		Topic:        cfg.Kafka.Topic,
		Balancer:     &kafka.LeastBytes{},
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	return &Producer{writer: writer}
}

func (p *Producer) SendImportRequest(ctx context.Context, projectKey string) error {
	if projectKey == "" {
		return fmt.Errorf("project_key is required")
	}

	msg := ImportMessage{ProjectKey: projectKey}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(projectKey),
		Value: data,
	})
	if err != nil {
		return fmt.Errorf("write message to kafka: %w", err)
	}

	logger.Info("producer: sent import request for project %q", projectKey)

	return nil
}

func (p *Producer) Close() error {
	if err := p.writer.Close(); err != nil {
		return fmt.Errorf("close writer: %w", err)
	}

	return nil
}
