package broker

import (
	"context"
	"encoding/json"
	"log"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
	"github.com/segmentio/kafka-go"
)

type KafkaConsumer[T any] struct {
	reader   *kafka.Reader
	consumer domain.MessageConsumer[T]
}

func NewKafkaConsumer[T any](brokers []string, topic, groupID string, consumer domain.MessageConsumer[T]) *KafkaConsumer[T] {
	return &KafkaConsumer[T]{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			Topic:    topic,
			GroupID:  groupID,
			MinBytes: 10e3, // 10KB
			MaxBytes: 10e6, // 10MB
		}),
		consumer: consumer,
	}
}

func (kc *KafkaConsumer[T]) Start(ctx context.Context) error {
	defer kc.reader.Close()
	log.Printf("Starting Kafka consumer for topic: %s", kc.reader.Config().Topic)

	for {
		msg, err := kc.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				log.Printf("Kafka consumer stopped: %v", ctx.Err())
				return nil
			}
			return err
		}

		var t T
		if err := json.Unmarshal(msg.Value, &t); err != nil {
			return err
		}

		if err := kc.consumer.Consume(ctx, &t); err != nil {
			return err
		}
		if err := kc.reader.CommitMessages(ctx, msg); err != nil {
			return err
		}
	}
}

func (kc *KafkaConsumer[T]) Close() error {
	return kc.reader.Close()
}
