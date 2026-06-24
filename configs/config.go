package configs

import (
	"fmt"
	"time"
)

// yaml format
var defaultConfig = []byte(`
env: dev
grpc_port: 50050
http_port: 50051
database_postgres:
  host: localhost
  port: 5434
  database: ecommerce
  schema: user
  username: root
  password: Abc12345
  logging_enabled: true
allowed_origins: "http://localhost:5173,http://172.93.163.153"
http_client:
  timeout: 30s
  max_idle_conns: 100
  max_idle_conns_per_host: 10
  max_conns_per_host: 100
  max_retries: 3
  retry_initial_delay: 100ms
  retry_max_delay: 10s
  retry_multiplier: 2.0
  max_response_body_size: 10485760

`)

type (
	Config struct {
		Env              string            `mapstructure:"env"`
		GrpcPort         string            `mapstructure:"grpc_port"`
		HttpPort         string            `mapstructure:"http_port"`
		DatabasePostgres SqlDatabaseConfig `mapstructure:"database_postgres"`
		AllowedOrigins   []string          `mapstructure:"allowed_origins"`
		HttpClientConfig HttpClientConfig  `mapstructure:"http_client"`
	}

	SqlDatabaseConfig struct {
		Host           string `mapstructure:"host"`
		Port           uint32 `mapstructure:"port"`
		Database       string `mapstructure:"database"`
		Schema         string `mapstructure:"schema"`
		Username       string `mapstructure:"username"`
		Password       string `mapstructure:"password"`
		LoggingEnabled bool   `mapstructure:"logging_enabled"`
		AutoMigration  bool   `mapstructure:"auto_migration"`
	}

	HttpClientConfig struct {
		Timeout             time.Duration `mapstructure:"timeout"`
		MaxIdleConns        int           `mapstructure:"max_idle_conns"`
		MaxIdleConnsPerHost int           `mapstructure:"max_idle_conns_per_host"`
		MaxConnsPerHost     int           `mapstructure:"max_conns_per_host"`
		MaxRetries          int           `mapstructure:"max_retries"`
		RetryInitialDelay   time.Duration `mapstructure:"retry_initial_delay"`
		RetryMaxDelay       time.Duration `mapstructure:"retry_max_delay"`
		RetryMultiplier     float64       `mapstructure:"retry_multiplier"`
		MaxResponseBodySize int64         `mapstructure:"max_response_body_size"`
	}
)

func LoadConfig() (*Config, error) {
	configMap := &Config{}

	err := Load(configMap, defaultConfig)
	if err != nil {
		fmt.Printf("Error while load config, err: %v", err)
		return nil, err
	}

	return configMap, nil
}
