package config

import "os"

type RedisConfig struct {
	Addr string
}

type HTTPConfig struct {
	Port string
}

func GetRedisConfig() RedisConfig {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	return RedisConfig{Addr: addr}
}

func GetHTTPConfig() HTTPConfig {
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}
	return HTTPConfig{Port: port}
}
