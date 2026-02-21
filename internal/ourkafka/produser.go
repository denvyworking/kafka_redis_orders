package ourkafka

import "github.com/segmentio/kafka-go"

func NewProducer(brokers []string, topic string) *kafka.Writer {
	return &kafka.Writer{
		// TCP это протокол, который используется для связи с брокерами Kafka.
		//  Он обеспечивает надежную передачу данных между клиентами и серверами Kafka.
		Addr:  kafka.TCP(brokers...),
		Topic: topic,
		// LeastBytes - это алгоритм балансировки нагрузки,
		// который выбирает партицию с наименьшим количеством байт
		// для отправки следующего сообщения.
		Balancer: &kafka.LeastBytes{},
	}
}

func Message(key, value string) kafka.Message {
	return kafka.Message{
		Key:   []byte(key),
		Value: []byte(value),
	}
}
