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
		reader: {
			Brokers: brokers,
			Topic:   topic,
			GroupID: groupID,
			MinBytes: 10e3, // 10KB
			MaxBytes: 10e6, // 10MB
		},
		consumer: consumer,
	}
}

func (kc *KafkaConsumer[T]) Start(ctx context.Context) error {
	log.Printf("Starting Kafka consumer for topic: %s", kc.reader.Config().Topic)

	for {
		msg, err := kc.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				log.Printf("Kafka consumer stopped: %v", ctx.Err())
				return nil
			}
			log.Printf("Error reading message: %v", err)
			continue
		}

		var t T
		if err := json.Unmarshal(msg.Value, &t); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		if err := kc.consumer.Consume(ctx, &t); err != nil {
			log.Printf("Error consuming message: %v", err)
		}
	}
}

func (kc *KafkaConsumer[T]) Close() error {
	return kc.reader.Close()
}