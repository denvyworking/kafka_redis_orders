package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLoad_ValidYaml(t *testing.T) {
	cfg, err := Load("../../configs/local.yaml")
	require.NoError(t, err)
	require.Equal(t, []string{"localhost:9092"}, cfg.Kafka.Brokers)
	require.Equal(t, "orders", cfg.Kafka.Topic)
	require.Equal(t, "order-consumer-group", cfg.Kafka.GroupID)
	require.Equal(t, "localhost:6379", cfg.Redis.Addr)
	require.Equal(t, 3600*time.Second, cfg.Redis.OrderTTL)
	require.Equal(t, "8080", cfg.HTTP.Port)
	require.Equal(t, 200*time.Millisecond, cfg.Retry.InitialInterval)
	require.Equal(t, 5*time.Second, cfg.Retry.MaxInterval)
	require.Equal(t, 30*time.Second, cfg.Retry.MaxElapsedTime)
	require.Equal(t, 2.0, cfg.Retry.Multiplier)
}

func TestLoad_ProdYamlExact(t *testing.T) {
	cfg, err := Load("../../configs/prod.yaml")
	require.NoError(t, err)

	require.Equal(t, []string{"kafka:9092"}, cfg.Kafka.Brokers)
	require.Equal(t, "orders", cfg.Kafka.Topic)
	require.Equal(t, "order-consumer-group", cfg.Kafka.GroupID)

	require.Equal(t, "redis:6379", cfg.Redis.Addr)
	require.Equal(t, 24*time.Hour, cfg.Redis.OrderTTL)

	require.Equal(t, "8080", cfg.HTTP.Port)

	require.Equal(t, 200*time.Millisecond, cfg.Retry.InitialInterval)
	require.Equal(t, 5*time.Second, cfg.Retry.MaxInterval)
	require.Equal(t, 30*time.Second, cfg.Retry.MaxElapsedTime)
	require.Equal(t, 2.0, cfg.Retry.Multiplier)
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/bad.yaml"

	// Заведомо сломанный YAML
	err := os.WriteFile(path, []byte("kafka:\n  brokers: [\n"), 0o600)
	require.NoError(t, err)

	_, err = Load(path)
	require.Error(t, err)
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load("../../configs/missing.yaml")
	require.Error(t, err)
}

func TestMustLoad_UsesConfigPathEnv(t *testing.T) {
	t.Setenv("CONFIG_PATH", "../../configs/prod.yaml")

	cfg := MustLoad()

	require.Equal(t, []string{"kafka:9092"}, cfg.Kafka.Brokers)
	require.Equal(t, "redis:6379", cfg.Redis.Addr)
	require.Equal(t, 24*time.Hour, cfg.Redis.OrderTTL)
}

func TestMustLoad_PanicsOnBadPath(t *testing.T) {
	t.Setenv("CONFIG_PATH", "../../configs/not-found.yaml")

	require.Panics(t, func() {
		_ = MustLoad()
	})
}

// env > yaml
func TestLoad_EnvOverridesYaml(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "config.yaml")
	yamlContent := `
kafka:
  brokers: ["yaml-broker:9092"]
  topic: "yaml-topic"
  group_id: "g"
  
redis:
  addr: "localhost:6379"

http:
  port: "8080"

retry:
  initial_interval: 200ms
  max_interval: 5s
`
	err := os.WriteFile(yamlPath, []byte(yamlContent), 0o600)
	require.NoError(t, err)

	t.Setenv("CONFIG_PATH", yamlPath)
	t.Setenv("KAFKA_BROKERS", "env-broker:9092")
	t.Setenv("KAFKA_TOPIC", "env-topic")

	cfg := MustLoad()

	require.Equal(t, []string{"env-broker:9092"}, cfg.Kafka.Brokers)
	require.Equal(t, "env-topic", cfg.Kafka.Topic)
}

func TestLoad_MissingRequiredField(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "config.yaml")

	// Отсутствует обязательное поле 'topic' для KafkaConfig'
	yamlContent := `
kafka:
  topic: "orders"
  group_id: "g"

redis:
  addr: "localhost:6379"

http:
  port: "8080"

retry:
  initial_interval: 200ms
  max_interval: 5s
`
	err := os.WriteFile(yamlPath, []byte(yamlContent), 0o600)
	require.NoError(t, err)

	t.Setenv("CONFIG_PATH", yamlPath)

	_, err = Load(yamlPath)
	require.Error(t, err) // cleanenv с тегом env-required вернёт ошибку
}
