package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Kafka     KafkaConfig     `yaml:"Kafka"`
	Connector ConnectorConfig `yaml:"Connector"`
	Server    ServerConfig    `yaml:"Server"`
}

type KafkaConfig struct {
	Brokers []string `yaml:"brokers"`
	Topic   string   `yaml:"topic"`
	GroupID string   `yaml:"groupId"`
}

type ConnectorConfig struct {
	GRPCAddress string `yaml:"grpcAddress"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	if len(c.Kafka.Brokers) == 0 {
		return fmt.Errorf("kafka brokers are required")
	}

	if c.Kafka.Topic == "" {
		return fmt.Errorf("kafka topic is required")
	}

	if c.Kafka.GroupID == "" {
		return fmt.Errorf("kafka groupId is required")
	}

	if c.Connector.GRPCAddress == "" {
		return fmt.Errorf("connector grpcAddress is required")
	}

	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("server port must be between 1 and 65535")
	}

	return nil
}
