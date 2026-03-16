package ourkafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type messageReader interface {
	ReadMessage(ctx context.Context) (kafka.Message, error)
	Close() error
}

type Consumer struct {
	reader messageReader
}

func NewConsumer(brokers []string, topic, groupID string) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			GroupID:  groupID,
			Topic:    topic,
			MinBytes: 10e3,
			MaxBytes: 10e6,
		}),
	}
}

func newConsumerWithReader(reader messageReader) *Consumer {
	return &Consumer{reader: reader}
}

func (p *Consumer) Close() error {
	return p.reader.Close()
}

func (p *Consumer) ReadMessage(ctx context.Context) (kafka.Message, error) {
	return p.reader.ReadMessage(ctx)
}
