package ourredis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/denvyworking/kafka-redis-orders/pkg/models"
)

type Client struct {
	*redis.Client
}

func NewRedisClient(addr string) *Client {
	return &Client{
		Client: redis.NewClient(&redis.Options{
			Addr: addr,
		}),
	}
}

func (c *Client) SetOrder(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return c.Set(ctx, key, value, ttl).Err()
}

func (c *Client) Close() error {
	return c.Client.Close()
}

func (r *Client) GetOrder(ctx context.Context, orderID string) (*models.Order, error) {
	key := fmt.Sprintf("order:%s", orderID)
	data, err := r.Client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("failed to get order from Redis: %w", err)
	}

	var order models.Order
	if err := json.Unmarshal(data, &order); err != nil {
		return nil, fmt.Errorf("failed to unmarshal order data: %w", err)
	}

	return &order, nil
}
