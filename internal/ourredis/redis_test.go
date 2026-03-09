package ourredis

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/denvyworking/kafka-redis-orders/pkg/models"
	"github.com/stretchr/testify/require"
)

func TestClient_SetOrder_Success(t *testing.T) {
	// тестируем с помощью miniredis - это легковесный
	// in-memory Redis, который идеально подходит для юнит-тестов
	mr := miniredis.RunT(t)
	// адрес miniredis передаем в наш клиент, чтобы он подключился к нему
	redisClient := NewRedisClient(mr.Addr())
	t.Cleanup(func() {
		require.NoError(t, redisClient.Close())
	})

	payload := []byte(`{"order_id":"123","user_id":"u1","total":11.5,"status":"processed"}`)
	err := redisClient.SetOrder(context.Background(), "123", payload, 1*time.Hour)
	require.NoError(t, err)

	stored, err := mr.Get("order:123")
	require.NoError(t, err)
	require.Equal(t, string(payload), stored)
	require.True(t, mr.Exists("order:123"))
}

func TestClient_GetOrder_Success(t *testing.T) {
	mr := miniredis.RunT(t)
	redisClient := NewRedisClient(mr.Addr())
	t.Cleanup(func() {
		require.NoError(t, redisClient.Close())
	})

	expected := models.Order{
		OrderID: "123",
		UserID:  "u1",
		Total:   11.5,
		Status:  "processed",
	}

	raw, err := json.Marshal(expected)
	require.NoError(t, err)
	require.NoError(t, mr.Set("order:123", string(raw)))

	order, err := redisClient.GetOrder(context.Background(), "123")
	require.NoError(t, err)
	require.Equal(t, expected, *order)
}

func TestClient_GetOrder_NotFound(t *testing.T) {
	mr := miniredis.RunT(t)
	redisClient := NewRedisClient(mr.Addr())
	t.Cleanup(func() {
		require.NoError(t, redisClient.Close())
	})

	_, err := redisClient.GetOrder(context.Background(), "nonexistent")
	require.ErrorIs(t, err, ErrOrderNotFound)
}

func TestClient_GetOrder_InvalidJSON(t *testing.T) {
	mr := miniredis.RunT(t)
	redisClient := NewRedisClient(mr.Addr())
	t.Cleanup(func() {
		require.NoError(t, redisClient.Close())
	})

	require.NoError(t, mr.Set("order:123", "not-json"))

	_, err := redisClient.GetOrder(context.Background(), "123")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to unmarshal order data")
}

func TestClient_Close(t *testing.T) {
	mr := miniredis.RunT(t)
	redisClient := NewRedisClient(mr.Addr())

	err := redisClient.Close()
	require.NoError(t, err)
}

func TestClient_SetOrder_TTL(t *testing.T) {
	mr := miniredis.RunT(t)
	redisClient := NewRedisClient(mr.Addr())
	t.Cleanup(func() {
		require.NoError(t, redisClient.Close())
	})

	err := redisClient.SetOrder(context.Background(), "123", []byte(`{"order_id":"123"}`), 1*time.Second)
	require.NoError(t, err)
	require.Less(t, mr.TTL("order:123"), 2*time.Second)
	require.Greater(t, mr.TTL("order:123"), 0*time.Second)
}

func TestClient_SetOrder_ContextCanceled(t *testing.T) {
	mr := miniredis.RunT(t)
	redisClient := NewRedisClient(mr.Addr())
	t.Cleanup(func() {
		require.NoError(t, redisClient.Close())
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // сразу отменяем контекст

	err := redisClient.SetOrder(ctx, "123", []byte(`{"order_id":"123"}`), 1*time.Second)
	require.ErrorIs(t, err, context.Canceled)
}
