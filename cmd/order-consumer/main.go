package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	ourkfk "github.com/yourname/go-kafka-redis-playground/internal/ourkafka"
	ourrdb "github.com/yourname/go-kafka-redis-playground/internal/ourredis"
)

type Order struct {
	OrderID string  `json:"order_id"`
	UserID  string  `json:"user_id"`
	Total   float64 `json:"total"`
	Status  string  `json:"status"`
}

func main() {
	brokers := []string{"localhost:9092"}

	topic := "orders"

	// все консьюмеры с одинаковой groupID будут в одной группе
	// и будут делить между собой партиции топика.
	groupID := "order-consumer-group"

	redisAddr := os.Getenv("REDDIS_ADDR")
	rdb := ourrdb.NewRedisClient(redisAddr)

	ctx := context.Background()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Не удалось подключиться к Redis: %v", err)
	}
	log.Println("✅ Подключено к Redis")

	reader := ourkfk.NewConsumer(brokers, groupID, topic)

	defer func() {
		if err := reader.Close(); err != nil {
			log.Printf("ошибка закрытия reader: %v", err)
		}
		if err := rdb.Close(); err != nil {
			log.Printf("ошибка закрытия redis: %v", err)
		}
	}()

	log.Println("📬 Запускаем потребитель Kafka...")
	log.Printf("Топик: %s, Группа: %s", topic, groupID)

	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("Ошибка чтения сообщения: %v", err)
			continue
		}

		var order Order
		if err := json.Unmarshal(msg.Value, &order); err != nil {
			log.Printf("Ошибка парсинга JSON: %v", err)
			continue
		}
		log.Printf("Получен заказ: %+v", order)

		order.Status = "processed"

		key := fmt.Sprintf("order:%s", order.OrderID)
		value, _ := json.Marshal(order)

		if err := rdb.Set(ctx, key, value, 1*time.Hour).Err(); err != nil {
			log.Printf("⚠️ Ошибка записи в Redis: %v", err)
			continue
		}

		log.Printf("✅ Обработан: ID=%s, User=%s, Total=%.2f (сохранено в Redis)",
			order.OrderID,
			order.UserID,
			order.Total)
	}
}
