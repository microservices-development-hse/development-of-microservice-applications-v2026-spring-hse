package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

var errReadConfig = errors.New("failed to read config file")
var errParseConfig = errors.New("failed to parse config file")

type Config struct {
	DBSettings      DBConfig      `yaml:"DBSettings" validate:"required"`
	ProgramSettings ProgramConfig `yaml:"ProgramSettings" validate:"required"`
}

type DBConfig struct {
	DBUser     string `yaml:"dbUser" validate:"required"`
	DBPassword string `yaml:"dbPassword" validate:"required"`
	DBHost     string `yaml:"dbHost" validate:"required,hostname_rfc1123|ip"`
	DBPort     int    `yaml:"dbPort" validate:"required,gt=0,lte=65535"`
	DBName     string `yaml:"dbName" validate:"required"`
}

type ProgramConfig struct {
	BindAddress      string `yaml:"bindAddress" validate:"required,ip"`
	BindPort         int    `yaml:"bindPort" validate:"required,gt=0,lte=65535"`
	ResourceTimeout  int    `yaml:"resourceTimeout" validate:"required,gt=0"`
	AnalyticsTimeout int    `yaml:"analyticsTimeout" validate:"required,gt=0"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errReadConfig
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, errParseConfig
	}

	validate := validator.New()
	if err := validate.Struct(config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}
