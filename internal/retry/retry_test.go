package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/denvyworking/kafka-redis-orders/internal/config"
	"github.com/stretchr/testify/require"
)

func fastRetryConfig() Config {
	return Config{
		InitialInterval: 1 * time.Millisecond,
		MaxInterval:     2 * time.Millisecond,
		MaxElapsedTime:  12 * time.Millisecond,
		Multiplier:      1.2,
	}
}

func TestNewConfig(t *testing.T) {
	in := config.RetryConfig{
		InitialInterval: 200 * time.Millisecond,
		MaxInterval:     5 * time.Second,
		MaxElapsedTime:  30 * time.Second,
		Multiplier:      2, // сколько раз пробуем повторить после
		// первой неудачной попытки (2 = 100% увеличение интервала)
	}

	got := NewConfig(in)

	require.Equal(t, in.InitialInterval, got.InitialInterval)
	require.Equal(t, in.MaxInterval, got.MaxInterval)
	require.Equal(t, in.MaxElapsedTime, got.MaxElapsedTime)
	require.Equal(t, in.Multiplier, got.Multiplier)
}

func TestDoSuccessFirstAttempt(t *testing.T) {
	cfg := fastRetryConfig()
	ctx := context.Background()

	attempts := 0
	err := cfg.Do(ctx, "success-first", func() error {
		attempts++
		return nil
	})

	require.NoError(t, err)
	require.Equal(t, 1, attempts)
}

func TestDoRetriesThenSuccess(t *testing.T) {
	cfg := fastRetryConfig()
	ctx := context.Background()

	attempts := 0
	err := cfg.Do(ctx, "retries-then-success", func() error {
		attempts++
		if attempts < 3 {
			return errors.New("temporary failure")
		}
		return nil
	})

	require.NoError(t, err)
	require.Equal(t, 3, attempts)
}

func TestDoStopsOnContextCanceledError(t *testing.T) {
	cfg := fastRetryConfig()
	ctx := context.Background()

	attempts := 0
	err := cfg.Do(ctx, "ctx-canceled", func() error {
		attempts++
		return context.Canceled
	})

	require.ErrorIs(t, err, context.Canceled)
	require.Equal(t, 1, attempts)
}

func TestDoStopsOnContextDeadlineExceededError(t *testing.T) {
	cfg := fastRetryConfig()
	ctx := context.Background()

	attempts := 0
	err := cfg.Do(ctx, "ctx-deadline", func() error {
		attempts++
		return context.DeadlineExceeded
	})

	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.Equal(t, 1, attempts)
}

func TestDoExhaustsRetriesAndReturnsLastError(t *testing.T) {
	cfg := Config{
		InitialInterval: 1 * time.Millisecond,
		MaxInterval:     1 * time.Millisecond,
		MaxElapsedTime:  5 * time.Millisecond,
		Multiplier:      1,
	}
	ctx := context.Background()

	permanentErr := errors.New("always failing")
	attempts := 0
	err := cfg.Do(ctx, "exhausted", func() error {
		attempts++
		return permanentErr
	})

	require.ErrorIs(t, err, permanentErr)
	require.GreaterOrEqual(t, attempts, 2)
}
