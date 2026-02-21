package ourkafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string, topic string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			// TCP это протокол, который используется для связи с брокерами Kafka.
			//  Он обеспечивает надежную передачу данных между клиентами и серверами Kafka.
			Addr:  kafka.TCP(brokers...),
			Topic: topic,
			// LeastBytes - это алгоритм балансировки нагрузки,
			// который выбирает партицию с наименьшим количеством байт
			// для отправки следующего сообщения.
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (p *Producer) Close() error {
	err := p.writer.Close()
	if err != nil {
		return err
	}
	return nil
}

func (p *Producer) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	return p.writer.WriteMessages(ctx, msgs...)
}

func Message(key, value string) kafka.Message {
	return kafka.Message{
		Key:   []byte(key),
		Value: []byte(value),
	}
}
