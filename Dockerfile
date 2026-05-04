FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install dependencies before copying application source files.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/order-api ./cmd/order-api
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/order-consumer ./cmd/order-consumer
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/order-producer ./cmd/order-producer

FROM alpine:3.20

RUN adduser -D appuser
WORKDIR /app

COPY --from=builder /bin/order-api /usr/local/bin/order-api
COPY --from=builder /bin/order-consumer /usr/local/bin/order-consumer
COPY --from=builder /bin/order-producer /usr/local/bin/order-producer
COPY --from=builder /app/configs ./configs

USER appuser

CMD ["order-api"]