package ourredis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/denvyworking/kafka-redis-orders/pkg/models"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	*redis.Client
}

var ErrOrderNotFound = errors.New("order not found")

func NewRedisClient(addr string) *Client {
	return &Client{
		Client: redis.NewClient(&redis.Options{
			Addr: addr,
		}),
	}
}

// 2
func (c *Client) SetOrder(ctx context.Context, orderID string, value []byte, ttl time.Duration) error {
	key := fmt.Sprintf("order:%s", orderID)
	return c.Set(ctx, key, value, ttl).Err()
}

func (c *Client) Close() error {
	return c.Client.Close()
}

// 3
func (r *Client) GetOrder(ctx context.Context, orderID string) (*models.Order, error) {
	key := fmt.Sprintf("order:%s", orderID)
	data, err := r.Get(ctx, key).Bytes()
	if err != nil {
		//4
		if err == redis.Nil {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get order from Redis: %w", err)
	}

	//5
	var order models.Order
	if err := json.Unmarshal(data, &order); err != nil {
		return nil, fmt.Errorf("failed to unmarshal order data: %w", err)
	}

	return &order, nil
}
