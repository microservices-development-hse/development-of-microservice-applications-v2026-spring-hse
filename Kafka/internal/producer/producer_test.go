package producer

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/microservices-development-hse/kafka/config"
	"github.com/segmentio/kafka-go"
)

type mockWriter struct {
	writeErr error
	closeErr error
	lastMsg  kafka.Message
}

func (m *mockWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	if len(msgs) > 0 {
		m.lastMsg = msgs[0]
	}

	return m.writeErr
}

func (m *mockWriter) Close() error {
	return m.closeErr
}

func TestSendImportRequest_EmptyKey(t *testing.T) {
	p := &Producer{}

	err := p.SendImportRequest(context.Background(), "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSendImportRequest_Success(t *testing.T) {
	mock := &mockWriter{}
	p := &Producer{writer: mock}

	err := p.SendImportRequest(context.Background(), "TEST")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(mock.lastMsg.Key) != "TEST" {
		t.Errorf("expected key TEST, got %s", string(mock.lastMsg.Key))
	}

	if len(mock.lastMsg.Value) == 0 {
		t.Error("expected non-empty value")
	}
}

func TestSendImportRequest_WriteError(t *testing.T) {
	mock := &mockWriter{writeErr: errors.New("kafka fail")}
	p := &Producer{writer: mock}

	err := p.SendImportRequest(context.Background(), "TEST")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestClose_Success(t *testing.T) {
	mock := &mockWriter{}
	p := &Producer{writer: mock}

	err := p.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClose_Error(t *testing.T) {
	mock := &mockWriter{closeErr: errors.New("close fail")}
	p := &Producer{writer: mock}

	err := p.Close()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNew(t *testing.T) {
	cfg := &config.Config{
		Kafka: config.KafkaConfig{
			Brokers: []string{"localhost:9092"},
			Topic:   "test-topic",
		},
	}

	producer := New(cfg)
	if producer == nil {
		t.Fatal("expected non-nil producer")
	}

	if producer.writer == nil {
		t.Fatal("expected non-nil writer")
	}

	writer, ok := producer.writer.(*kafka.Writer)
	if !ok {
		t.Fatal("expected *kafka.Writer type")
	}

	if writer.Topic != cfg.Kafka.Topic {
		t.Errorf("expected topic %s, got %s", cfg.Kafka.Topic, writer.Topic)
	}
}

func TestSendImportRequest_ContextCanceled(t *testing.T) {
	mock := &mockWriter{writeErr: context.Canceled}
	p := &Producer{writer: mock}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := p.SendImportRequest(ctx, "TEST")
	if err == nil {
		t.Fatal("expected error from canceled context")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled error, got %v", err)
	}
}

func TestSendImportRequest_ContextDeadlineExceeded(t *testing.T) {
	mock := &mockWriter{writeErr: context.DeadlineExceeded}
	p := &Producer{writer: mock}

	ctx, cancel := context.WithTimeout(context.Background(), -1*time.Second)
	defer cancel()

	err := p.SendImportRequest(ctx, "TEST")
	if err == nil {
		t.Fatal("expected error from deadline exceeded")
	}

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded error, got %v", err)
	}
}

func TestSendImportRequest_MultipleMessages(t *testing.T) {
	mock := &mockWriter{}
	p := &Producer{writer: mock}

	keys := []string{"PROJECT1", "PROJECT2", "PROJECT3"}

	for _, key := range keys {
		err := p.SendImportRequest(context.Background(), key)
		if err != nil {
			t.Fatalf("unexpected error for key %s: %v", key, err)
		}

		if string(mock.lastMsg.Key) != key {
			t.Errorf("expected key %s, got %s", key, string(mock.lastMsg.Key))
		}
	}
}

func TestSendImportRequest_ValidJSONFormat(t *testing.T) {
	mock := &mockWriter{}
	p := &Producer{writer: mock}

	err := p.SendImportRequest(context.Background(), "TEST_PROJECT")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var msg ImportMessage

	err = json.Unmarshal(mock.lastMsg.Value, &msg)
	if err != nil {
		t.Fatalf("value is not valid JSON: %v", err)
	}

	if msg.ProjectKey != "TEST_PROJECT" {
		t.Errorf("expected ProjectKey TEST_PROJECT, got %s", msg.ProjectKey)
	}
}
