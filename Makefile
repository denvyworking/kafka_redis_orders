# 📦 Переменные (чтобы не дублировать пути)
PRODUCER_CMD = cmd/order-producer/main.go
CONSUMER_CMD = cmd/order-consumer/main.go
API_CMD = cmd/order-api/main.go

# 🎯 Цели (commands)
.PHONY: producer
producer:
	@echo "🚀 Запускаем Producer..."
	go run $(PRODUCER_CMD)


.PHONY: consumer
consumer:
	go run $(CONSUMER_CMD)


.PHONY: api
api:
	go run $(API_CMD)




.PHONY: build
build:
	@echo "🔨 Собираем бинарники..."
	go build -o bin/order-producer $(PRODUCER_CMD)
	go build -o bin/order-consumer $(CONSUMER_CMD)
	go build -o bin/order-api $(API_CMD)
	@echo "✅ Бинарники собраны в папке bin/"


.PHONY: test
test:
	@echo "🧪 Запускаем тесты..."
	go test -v ./...


.PHONY: fmt
fmt:
	@echo "✨ Форматируем код..."
	go fmt ./...


.PHONY: lint
lint:
	golangci-lint run ./...


.PHONY: clean
clean:
	rm -rf bin/

.PHONY: apiHealth
api-health:
	curl -s http://localhost:8080/health

.PHONY: orderHandler
orderHandler:
	curl -s http://localhost:8080/order/non-existent-id


# Справка (список доступных команд)
.PHONY: help
help:
	@echo "📋 Available commands:"
	@echo "  make producer     — start the Kafka producer"
	@echo "  make consumer     — start the Kafka consumer"
	@echo "  make api          — start the HTTP API server"
	@echo "  make build        — compile all binaries to bin/"
	@echo "  make test         — run all tests"
	@echo "  make fmt          — format Go code"
	@echo "  make lint         — run linter (requires golangci-lint)"
	@echo "  make clean        — remove built binaries"
	@echo "  make api-health   — check API health endpoint"
	@echo "  make help         — show this help message"
	@echo "make orderHandler — test API order handler with a non-existent ID"