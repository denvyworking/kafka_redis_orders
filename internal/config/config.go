package config

import (
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

const defaultConfigPath = "configs/local.yaml"

type Config struct {
	Kafka KafkaConfig `yaml:"kafka"`
	Redis RedisConfig `yaml:"redis"`
	HTTP  HTTPConfig  `yaml:"http"`
	Retry RetryConfig `yaml:"retry"`
}

type RetryConfig struct {
	InitialInterval time.Duration `yaml:"initial_interval" env:"RETRY_INITIAL_INTERVAL" env-default:"200ms"`
	MaxInterval     time.Duration `yaml:"max_interval" env:"RETRY_MAX_INTERVAL" env-default:"5s"`
	MaxElapsedTime  time.Duration `yaml:"max_elapsed_time" env:"RETRY_MAX_ELAPSED_TIME" env-default:"30s"`
	Multiplier      float64       `yaml:"multiplier" env:"RETRY_MULTIPLIER" env-default:"2.0"`
}

type KafkaConfig struct {
	Brokers []string `yaml:"brokers" env:"KAFKA_BROKERS" env-separator:","`
	Topic   string   `yaml:"topic"   env:"KAFKA_TOPIC"`
	GroupID string   `yaml:"group_id" env:"KAFKA_GROUP_ID"`
}

type RedisConfig struct {
	Addr     string        `yaml:"addr"      env:"REDIS_ADDR"`
	OrderTTL time.Duration `yaml:"order_ttl" env:"REDIS_ORDER_TTL" env-default:"1h"`
}

type HTTPConfig struct {
	Port string `yaml:"port" env:"HTTP_PORT" env-default:"8080"`
}

func Load(path string) (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// MustLoad читает конфиг из пути CONFIG_PATH (env),
// если переменная не задана - использует configs/local.yaml.
func MustLoad() *Config {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = defaultConfigPath
	}
	cfg, err := Load(path)
	if err != nil {
		panic("failed to load config from " + path + ": " + err.Error())
	}
	return cfg
}
