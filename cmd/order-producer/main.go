package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"

	"time"

	ourkfk "github.com/yourname/go-kafka-redis-playground/internal/ourkafka"
)

type Order struct {
	OrderID string  `json:"order_id"`
	UserID  string  `json:"user_id"`
	Total   float64 `json:"total"`
}

func main() {
	// Продюсер - это сервер или приложение, которое отправляет сообщения в Kafka.
	// Он отвечает за создание и отправку сообщений в определенные топики Kafka.
	// В нашем примере это localhost:9092, который является адресом нашего Kafka брокера,
	// и топик "orders", куда мы будем отправлять наши заказы.
	brokers := []string{"localhost:9092"}
	topic := "orders"

	producer := ourkfk.NewProducer(brokers, topic)
	defer func() {
		if err := producer.Close(); err != nil {
			log.Fatalf("failed to close producer: %v", err)
		}
	}()

	rand.Seed(time.Now().UnixNano())
	ctx := context.Background()

	log.Println("📬 Начинаем отправку заказов...")

	for i := 0; i < 10; i++ {
		order := Order{
			OrderID: fmt.Sprintf("order-%d-%d", i, time.Now().Unix()),
			UserID:  fmt.Sprintf("user-%d", rand.Intn(3)),
			Total:   10 + rand.Float64()*90,
		}

		payload, err := json.Marshal(order)
		if err != nil {
			log.Fatalf("ошибка JSON: %v", err)
		}

		msg := ourkfk.Message(order.UserID, string(payload))
		// WriteMessages - это метод, который отправляет одно или
		// несколько сообщений в Kafka.
		if err := producer.WriteMessages(ctx, msg); err != nil {
			log.Fatalf("ошибка отправки: %v", err)
		}

		log.Printf("Отправлен: %+v", order)
		time.Sleep(500 * time.Millisecond)
	}

	log.Println("Готово!")
}
