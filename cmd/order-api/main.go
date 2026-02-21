package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/denvyworking/kafka-redis-orders/internal/ourredis"
)

type Order struct {
	OrderID string  `json:"order_id"`
	UserID  string  `json:"user_id"`
	Total   float64 `json:"total"`
	Status  string  `json:"status"`
}

func HandleGetOrder(redisClient *ourredis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"error": "Only GET method allowed"})
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/order/")
		orderID := path
		if orderID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Order ID is required"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		order, err := redisClient.GetOrder(ctx, orderID)
		if err != nil {
			if order == nil {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{"error": "Order not found"})
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "Failed to retrieve order"})
			}
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(order)
	}
}

func main() {
	redisAddr := os.Getenv("REDDIS_ADDR")
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}
	redisClient := ourredis.NewRedisClient(redisAddr)
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

	http.HandleFunc("/order/", HandleGetOrder(redisClient))

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})
	log.Printf("🚀 Starting server on port %s...", httpPort)
	if err := http.ListenAndServe(":"+httpPort, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
