package ourredis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
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
