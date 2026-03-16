package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/denvyworking/kafka-redis-orders/internal/ourredis"
	"github.com/denvyworking/kafka-redis-orders/pkg/models"
	"github.com/stretchr/testify/require"
)

// ЭТА ФУНКЦИЯ ЗАМЕНЯЕТ REDIS ДЛЯ ТЕСТОВАНИЯ API СЕРВЕРА
type mockOrderStore struct {
	getOrderFn func(ctx context.Context, orderID string) (*models.Order, error)
}

func (m mockOrderStore) GetOrder(ctx context.Context, orderID string) (*models.Order, error) {
	if m.getOrderFn == nil {
		return nil, errors.New("getOrderFn is not set")
	}
	return m.getOrderFn(ctx, orderID)
}

func TestHandleHealth(t *testing.T) {
	s := NewServerWithStore(mockOrderStore{getOrderFn: func(ctx context.Context, orderID string) (*models.Order, error) {
		return nil, nil
	}})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	s.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	require.JSONEq(t, `{"status":"healthy"}`+"\n", rec.Body.String())
}

func TestMetricsEndpoint(t *testing.T) {
	s := NewServerWithStore(mockOrderStore{getOrderFn: func(ctx context.Context, orderID string) (*models.Order, error) {
		return &models.Order{}, nil
	}})

	s.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/health", nil))

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()

	s.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.True(t, strings.Contains(rec.Header().Get("Content-Type"), "text/plain"))
	require.Contains(t, rec.Body.String(), "http_requests_total")
	require.Contains(t, rec.Body.String(), "http_request_duration_seconds")
}

func TestHandleGetOrderMethodNotAllowed(t *testing.T) {
	s := NewServerWithStore(mockOrderStore{getOrderFn: func(ctx context.Context, orderID string) (*models.Order, error) {
		return &models.Order{}, nil
	}})

	req := httptest.NewRequest(http.MethodPost, "/order/1", nil)
	rec := httptest.NewRecorder()

	s.ServeHTTP(rec, req)

	require.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	require.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
	require.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	require.Contains(t, rec.Body.String(), "Only GET method allowed")
}

func TestHandleGetOrderMissingID(t *testing.T) {
	s := NewServerWithStore(mockOrderStore{getOrderFn: func(ctx context.Context, orderID string) (*models.Order, error) {
		return &models.Order{}, nil
		// указатель тут нужен для хранения, чтобы в памяти хранилась!
	}})

	req := httptest.NewRequest(http.MethodGet, "/order/", nil)
	rec := httptest.NewRecorder()

	s.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Contains(t, rec.Body.String(), "Order ID is required")
}

func TestHandleGetOrderNotFound(t *testing.T) {
	s := NewServerWithStore(mockOrderStore{getOrderFn: func(ctx context.Context, orderID string) (*models.Order, error) {
		return nil, ourredis.ErrOrderNotFound
	}})

	req := httptest.NewRequest(http.MethodGet, "/order/not-exists", nil)
	rec := httptest.NewRecorder()

	s.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
	require.Contains(t, rec.Body.String(), "Order not found")
}

func TestHandleGetOrderInternalError(t *testing.T) {
	s := NewServerWithStore(mockOrderStore{getOrderFn: func(ctx context.Context, orderID string) (*models.Order, error) {
		return nil, errors.New("redis down")
	}})

	req := httptest.NewRequest(http.MethodGet, "/order/42", nil)
	rec := httptest.NewRecorder()

	s.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Contains(t, rec.Body.String(), "Failed to retrieve order")
}

func TestHandleGetOrderSuccess(t *testing.T) {
	expected := &models.Order{
		OrderID: "order-1",
		UserID:  "user-9",
		Total:   99.5,
		Status:  "processed",
	}

	s := NewServerWithStore(mockOrderStore{getOrderFn: func(ctx context.Context, orderID string) (*models.Order, error) {
		require.Equal(t, "order-1", orderID)
		return expected, nil
	}})

	req := httptest.NewRequest(http.MethodGet, "/order/order-1", nil)
	rec := httptest.NewRecorder()

	s.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	require.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))

	var got models.Order
	err := json.Unmarshal(rec.Body.Bytes(), &got)
	require.NoError(t, err)
	require.Equal(t, *expected, got)
}
