package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/denvyworking/kafka-redis-orders/internal/ourredis"
	"github.com/denvyworking/kafka-redis-orders/pkg/models"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type orderStore interface {
	GetOrder(ctx context.Context, orderID string) (*models.Order, error)
}

type Server struct {
	redis           orderStore
	mux             *http.ServeMux
	registry        *prometheus.Registry
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
}

func NewServer(redisClient *ourredis.Client) *Server {
	return NewServerWithStore(redisClient)
}

func NewServerWithStore(store orderStore) *Server {
	registry := prometheus.NewRegistry()
	requestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total amount of HTTP requests grouped by route, method and status code.",
		},
		[]string{"route", "method", "status"},
	)

	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds grouped by route and method.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"route", "method"},
	)

	registry.MustRegister(requestsTotal, requestDuration)

	s := &Server{
		redis:           store,
		mux:             http.NewServeMux(),
		registry:        registry,
		requestsTotal:   requestsTotal,
		requestDuration: requestDuration,
	}

	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.Handle("/metrics", promhttp.HandlerFor(s.registry, promhttp.HandlerOpts{}))
	s.mux.HandleFunc("/order/", s.instrument("/order/", s.handleGetOrder))
	s.mux.HandleFunc("/health", s.instrument("/health", s.handleHealth))
}

func (s *Server) instrument(route string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next(recorder, r)

		status := strconv.Itoa(recorder.status)
		s.requestsTotal.WithLabelValues(route, r.Method, status).Inc()
		s.requestDuration.WithLabelValues(route, r.Method).Observe(time.Since(startedAt).Seconds())
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (sr *statusRecorder) WriteHeader(statusCode int) {
	sr.status = statusCode
	sr.ResponseWriter.WriteHeader(statusCode)
}

// ServeHTTP реализует интерфейс http.Handler,
// позволяя Server обрабатывать HTTP-запросы
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) handleGetOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Only GET method allowed"})
		return
	}

	orderID := strings.TrimPrefix(r.URL.Path, "/order/")
	if orderID == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Order ID is required"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	order, err := s.redis.GetOrder(ctx, orderID)
	if err != nil {
		if errors.Is(err, ourredis.ErrOrderNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Order not found"})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to retrieve order"})
		}
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(order)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}
