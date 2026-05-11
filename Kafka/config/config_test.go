package config

import (
	"os"
	"strings"
	"testing"
)

func TestLoad_Success(t *testing.T) {
	content := `
Kafka:
  brokers: ["localhost:9092"]
  topic: "test-topic"
  groupId: "group"
Connector:
  grpcAddress: "localhost:50051"
Server:
  port: 8080
`

	tmp, err := os.CreateTemp("", "cfg-*.yaml")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.Remove(tmp.Name()) }()

	_, _ = tmp.Write([]byte(content))
	_ = tmp.Close()

	cfg, err := Load(tmp.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("wrong port")
	}
}

func TestLoad_FileError(t *testing.T) {
	_, err := Load("not-exists.yaml")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	file, err := os.CreateTemp("", "bad_config_*.yaml")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.Remove(file.Name()) }()

	_, err = file.WriteString("Kafka: [:::]")
	if err != nil {
		t.Fatal(err)
	}

	_ = file.Close()

	_, err = Load(file.Name())
	if err == nil {
		t.Fatal("expected yaml parse error")
	}

	if !strings.Contains(err.Error(), "parse yaml") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLoad_InvalidValidate(t *testing.T) {
	tmp, _ := os.CreateTemp("", "cfg-*.yaml")

	defer func() { _ = os.Remove(tmp.Name()) }()

	_, _ = tmp.Write([]byte("::::invalid yaml::::"))
	_ = tmp.Close()

	_, err := Load(tmp.Name())
	if err == nil {
		t.Fatal("expected yaml error")
	}
}

func TestValidate_AllErrors(t *testing.T) {
	tests := []Config{
		{},
		{Kafka: KafkaConfig{Brokers: []string{"b"}}},
		{Kafka: KafkaConfig{Brokers: []string{"b"}, Topic: "t"}},
		{Kafka: KafkaConfig{Brokers: []string{"b"}, Topic: "t", GroupID: "g"}},
		{
			Kafka: KafkaConfig{
				Brokers: []string{"b"},
				Topic:   "t",
				GroupID: "g",
			},
			Connector: ConnectorConfig{GRPCAddress: "x"},
			Server:    ServerConfig{Port: 0},
		},
	}

	for i, cfg := range tests {
		if cfg.Validate() == nil {
			t.Errorf("case %d: expected error", i)
		}
	}
}

func TestValidate_Success(t *testing.T) {
	cfg := Config{
		Kafka: KafkaConfig{
			Brokers: []string{"b"},
			Topic:   "t",
			GroupID: "g",
		},
		Connector: ConnectorConfig{GRPCAddress: "x"},
		Server:    ServerConfig{Port: 8080},
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
