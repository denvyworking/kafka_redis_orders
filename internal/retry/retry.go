package retry

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/denvyworking/kafka-redis-orders/internal/config"
)

type Config struct {
	InitialInterval time.Duration
	MaxInterval     time.Duration
	MaxElapsedTime  time.Duration
	Multiplier      float64
}

func NewConfig(c config.RetryConfig) Config {
	return Config{
		InitialInterval: c.InitialInterval,
		MaxInterval:     c.MaxInterval,
		MaxElapsedTime:  c.MaxElapsedTime,
		Multiplier:      c.Multiplier,
	}
}

func (c Config) Do(ctx context.Context, op string, fn func() error) error {
	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = c.InitialInterval
	bo.MaxInterval = c.MaxInterval
	bo.MaxElapsedTime = c.MaxElapsedTime
	bo.Multiplier = c.Multiplier

	attempt := 0

	err := backoff.RetryNotify(
		func() error {
			attempt++
			err := fn()
			if err == nil {
				return nil
			}

			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return backoff.Permanent(err)
			}

			return err
		},
		backoff.WithContext(bo, ctx),
		func(err error, nextDelay time.Duration) {
			log.Printf("retry op=%s attempt=%d err=%v next_in=%s", op, attempt, err, nextDelay)
		},
	)

	if err != nil {
		log.Printf("retry exhausted op=%s attempts=%d err=%v", op, attempt, err)
	}

	return err
}
