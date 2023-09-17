package kafka

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	"time"
)

type Service struct {
	Writer *kafka.Writer
	Reader *kafka.Reader
}

func New(brokers []string, topic string) *Service {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Balancer:     &kafka.LeastBytes{},
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:         brokers,
		Topic:           topic,
		GroupID:         "consumer-group-id",
		MinBytes:        10e3,
		MaxBytes:        10e6,
		MaxWait:         1 * time.Second,
		ReadLagInterval: -1,
	})

	return &Service{
		Writer: writer,
		Reader: reader,
	}
}

func (ks *Service) Close() error {
	if err := ks.Writer.Close(); err != nil {
		return fmt.Errorf("failed to close Kafka writer: %w", err)
	}

	if err := ks.Reader.Close(); err != nil {
		return fmt.Errorf("failed to close Kafka reader: %w", err)
	}
	return nil
}

func (ks *Service) ReadMessage(ctx context.Context) (kafka.Message, error) {
	return ks.Reader.ReadMessage(ctx)
}

func (ks *Service) PublishToTopic(topic string, value []byte) error {
	message := kafka.Message{
		Topic: topic,
		Value: value,
	}
	return ks.Writer.WriteMessages(context.Background(), message)
}
