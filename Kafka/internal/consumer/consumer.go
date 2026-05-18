package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/microservices-development-hse/kafka/config"
	pb "github.com/microservices-development-hse/kafka/internal/generated/connector"
	"github.com/microservices-development-hse/kafka/internal/logger"
)

type readerIface interface {
	ReadMessage(ctx context.Context) (kafka.Message, error)
	Close() error
}

type connectorClient interface {
	TriggerProjectImport(ctx context.Context, in *pb.ImportRequest, opts ...grpc.CallOption) (*pb.ImportResponse, error)
}

type ImportMessage struct {
	ProjectKey string `json:"project_key"`
}

type Consumer struct {
	reader          readerIface
	connectorClient connectorClient
	grpcConn        *grpc.ClientConn
}

func New(cfg *config.Config) (*Consumer, error) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Kafka.Brokers,
		Topic:          cfg.Kafka.Topic,
		GroupID:        cfg.Kafka.GroupID,
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: time.Second,
		StartOffset:    kafka.LastOffset,
	})

	conn, err := grpc.NewClient(
		cfg.Connector.GRPCAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("grpc dial connector: %w", err)
	}

	return &Consumer{
		reader:          reader,
		connectorClient: pb.NewConnectorServiceClient(conn),
		grpcConn:        conn,
	}, nil
}

func (c *Consumer) Run(ctx context.Context) error {
	logger.Info("consumer: starting, waiting for messages...")

	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}

			logger.Info("consumer: read message error: %v", err)

			continue
		}

		logger.Info("consumer: received message from partition=%d offset=%d: %s",
			msg.Partition, msg.Offset, string(msg.Value))

		if err := c.handleMessage(ctx, msg.Value); err != nil {
			logger.Info("consumer: handle message error: %v", err)
		}
	}
}

func (c *Consumer) handleMessage(ctx context.Context, data []byte) error {
	var msg ImportMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("unmarshal message: %w", err)
	}

	if msg.ProjectKey == "" {
		return fmt.Errorf("empty project_key in message")
	}

	if c.connectorClient == nil {
		return fmt.Errorf("connector client is nil")
	}

	logger.Info("consumer: triggering import for project %q", msg.ProjectKey)

	resp, err := c.connectorClient.TriggerProjectImport(ctx, &pb.ImportRequest{
		ProjectKey: msg.ProjectKey,
	})
	if err != nil {
		return fmt.Errorf("grpc TriggerProjectImport: %w", err)
	}

	if !resp.GetSuccess() {
		return fmt.Errorf("connector import failed: %s", resp.GetMessage())
	}

	logger.Info("consumer: import for project %q completed: %s", msg.ProjectKey, resp.GetMessage())

	return nil
}

func (c *Consumer) Close() error {
	if c.reader != nil {
		if err := c.reader.Close(); err != nil {
			logger.Info("consumer: close reader error: %v", err)
		}
	}

	if c.grpcConn != nil {
		if err := c.grpcConn.Close(); err != nil {
			logger.Info("consumer: close grpc conn error: %v", err)
		}
	}

	return nil
}
