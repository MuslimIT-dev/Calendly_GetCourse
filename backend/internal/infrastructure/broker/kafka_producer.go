package broker

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
)

type KafkaPublisher[T any] struct {
	writer *kafka.Writer
}

func NewKafkaPublisher[T any](writer *kafka.Writer) *KafkaPublisher[T] {
	return &KafkaPublisher[T]{writer: writer}
}

func (kp *KafkaPublisher[T]) Publish(ctx context.Context, topic string, message *T) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return kp.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Value: data,
	})
}

func (kp *KafkaPublisher[T]) Close() error {
	return kp.writer.Close()
}