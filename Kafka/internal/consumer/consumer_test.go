package consumer

import (
	"context"
	"errors"
	"testing"

	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"

	"github.com/microservices-development-hse/kafka/config"
	pb "github.com/microservices-development-hse/kafka/internal/generated/connector"
)

type mockConnector struct {
	resp *pb.ImportResponse
	err  error
}

func (m *mockConnector) TriggerProjectImport(ctx context.Context, in *pb.ImportRequest, opts ...grpc.CallOption) (*pb.ImportResponse, error) {
	return m.resp, m.err
}

type mockReader struct {
	msgs     []kafka.Message
	err      error
	closeErr error
}

func (m *mockReader) ReadMessage(ctx context.Context) (kafka.Message, error) {
	if m.err != nil {
		return kafka.Message{}, m.err
	}

	if len(m.msgs) == 0 {
		return kafka.Message{}, errors.New("no messages")
	}

	msg := m.msgs[0]
	m.msgs = m.msgs[1:]

	return msg, nil
}

func (m *mockReader) Close() error {
	return m.closeErr
}

func TestNew_Success(t *testing.T) {
	cfg := &config.Config{
		Kafka: config.KafkaConfig{
			Brokers: []string{"localhost:9092"},
			Topic:   "test",
			GroupID: "group",
		},
		Connector: config.ConnectorConfig{
			GRPCAddress: "localhost:12345",
		},
	}

	c, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c == nil {
		t.Fatal("expected consumer")
	}

	_ = c.Close()
}

func TestHandleMessage_Success(t *testing.T) {
	c := &Consumer{
		connectorClient: &mockConnector{
			resp: &pb.ImportResponse{Success: true, Message: "ok"},
		},
	}

	err := c.handleMessage(context.Background(), []byte(`{"project_key":"TEST"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHandleMessage_InvalidJSON(t *testing.T) {
	c := &Consumer{}
	if err := c.handleMessage(context.Background(), []byte("bad")); err == nil {
		t.Fatal("expected error")
	}
}

func TestHandleMessage_EmptyKey(t *testing.T) {
	c := &Consumer{}
	if err := c.handleMessage(context.Background(), []byte(`{"project_key":""}`)); err == nil {
		t.Fatal("expected error")
	}
}

func TestHandleMessage_NilClient(t *testing.T) {
	c := &Consumer{}
	if err := c.handleMessage(context.Background(), []byte(`{"project_key":"TEST"}`)); err == nil {
		t.Fatal("expected error")
	}
}

func TestHandleMessage_GRPCError(t *testing.T) {
	c := &Consumer{
		connectorClient: &mockConnector{
			err: errors.New("grpc fail"),
		},
	}

	if err := c.handleMessage(context.Background(), []byte(`{"project_key":"TEST"}`)); err == nil {
		t.Fatal("expected error")
	}
}

func TestHandleMessage_NotSuccess(t *testing.T) {
	c := &Consumer{
		connectorClient: &mockConnector{
			resp: &pb.ImportResponse{Success: false, Message: "fail"},
		},
	}

	if err := c.handleMessage(context.Background(), []byte(`{"project_key":"TEST"}`)); err == nil {
		t.Fatal("expected error")
	}
}

func TestRun_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	c := &Consumer{
		reader: &mockReader{},
	}

	if err := c.Run(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRun_ReadError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	c := &Consumer{
		reader: &mockReader{
			err: errors.New("read error"),
		},
	}

	go func() {
		cancel()
	}()

	_ = c.Run(ctx)
}

func TestRun_HandleMessageError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	c := &Consumer{
		reader: &mockReader{
			msgs: []kafka.Message{
				{Value: []byte(`{"project_key":""}`)},
			},
		},
		connectorClient: &mockConnector{},
	}

	go func() {
		cancel()
	}()

	_ = c.Run(ctx)
}

func TestClose_ReaderError(t *testing.T) {
	c := &Consumer{
		reader: &mockReader{
			closeErr: errors.New("reader close error"),
		},
	}

	if err := c.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
