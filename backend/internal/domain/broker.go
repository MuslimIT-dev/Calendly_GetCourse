package domain

import "context"

type Publisher[T any] interface {
	Publish(ctx context.Context, topic string, message *T) error
}

type MessageConsumer[T any] interface {
	Consume(ctx context.Context, msg *T) error
}