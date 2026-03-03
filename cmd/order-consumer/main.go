package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/denvyworking/kafka-redis-orders/internal/config"
	ourkfk "github.com/denvyworking/kafka-redis-orders/internal/ourkafka"
	ourrdb "github.com/denvyworking/kafka-redis-orders/internal/ourredis"
	"github.com/denvyworking/kafka-redis-orders/pkg/models"
)

func main() {

	// все консьюмеры с одинаковой groupID будут в одной группе
	// и будут делить между собой партиции топика.
	cfg := config.MustLoad()

	rdb := ourrdb.NewRedisClient(cfg.Redis.Addr)

	ctx := context.Background()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Не удалось подключиться к Redis: %v", err)
	}
	log.Println("✅ Подключено к Redis")

	reader := ourkfk.NewConsumer(cfg.Kafka.Brokers, cfg.Kafka.Topic, cfg.Kafka.GroupID)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// отдельная горутина для обработки сигналов завершения
	go func() {
		sig := <-signalChan
		log.Printf("\nПолучен сигнал остановки: %v", sig)
		log.Println("Начинаем корректное завершение работы...")
		cancel()
	}()

	defer func() {
		if err := reader.Close(); err != nil {
			log.Printf("ошибка закрытия reader: %v", err)
		}
		if err := rdb.Close(); err != nil {
			log.Printf("ошибка закрытия redis: %v", err)
		}
	}()

	log.Println("📬 Запускаем потребитель Kafka...")
	log.Printf("Топик: %s, Группа: %s", cfg.Kafka.Topic, cfg.Kafka.GroupID)

	for {
		select {
		case <-ctx.Done():
			log.Println("Завершение работы потребителя...")
			time.Sleep(500 * time.Millisecond)
			return
		default:
			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("Ошибка чтения сообщения: %v", err)
				continue
			}

			var order models.Order
			if err := json.Unmarshal(msg.Value, &order); err != nil {
				log.Printf("Ошибка парсинга JSON: %v", err)
				continue
			}
			log.Printf("Получен заказ: %+v", order)

			order.Status = "processed"

			value, _ := json.Marshal(order)

			if err := rdb.SetOrder(ctx, order.OrderID, value, cfg.Redis.OrderTTL); err != nil {
				log.Printf("Ошибка записи в Redis: %v", err)
				continue
			}

			log.Printf("✅ Обработан: ID=%s, User=%s, Total=%.2f (сохранено в Redis)",
				order.OrderID,
				order.UserID,
				order.Total)
		}
	}
}
