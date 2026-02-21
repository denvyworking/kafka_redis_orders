package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/denvyworking/kafka-redis-orders/internal/api"
	"github.com/denvyworking/kafka-redis-orders/internal/config"
	"github.com/denvyworking/kafka-redis-orders/internal/ourredis"
)

func main() {
	redisCfg := config.GetRedisConfig()
	httpCfg := config.GetHTTPConfig()

	redisClient := ourredis.NewRedisClient(redisCfg.Addr)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		fmt.Printf("Error connecting to Redis: %v\n", err)
		return
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			fmt.Printf("Error closing Redis client: %v\n", err)
		}
	}()

	log.Println("✅ Подключено к Redis")

	server := api.NewServer(redisClient)

	log.Printf("🚀 Starting server on port %s...", httpCfg.Port)
	if err := http.ListenAndServe(":"+httpCfg.Port, server); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
