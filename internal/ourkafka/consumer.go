package ourkafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(brokers []string, groupID, topic string) *Consumer {
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

func (p *Consumer) Close() error {
	err := p.reader.Close()
	if err != nil {
		return err
	}
	return nil
}

func (p *Consumer) ReadMessage(ctx context.Context) (kafka.Message, error) {
	return p.reader.ReadMessage(ctx)
}
