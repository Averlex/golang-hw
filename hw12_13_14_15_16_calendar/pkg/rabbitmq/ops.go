package rabbitmq

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go" //nolint:depguard,nolintlint
)

// Produce sends a message to the message queue using retry logic and operation timeout.
func (r *RabbitMQ) Produce(ctx context.Context, payload []byte) error {
	return r.withRetries(ctx, "produce", func() error {
		return r.withTimeout(ctx, func(localCtx context.Context) error {
			r.mu.Lock()
			err := r.ch.PublishWithContext(
				localCtx,
				r.topic,
				r.routingKey,
				false, // mandatory
				false, // immediate
				amqp.Publishing{
					ContentType: r.contentType,
					Body:        payload,
				},
			)
			r.mu.Unlock()
			return err
		})
	})
}

// Consume consumes messages from the message queue using retry logic and operation timeout.
func (r *RabbitMQ) Consume(ctx context.Context) (<-chan amqp.Delivery, error) {
	var err error
	var ch <-chan amqp.Delivery

	err = r.withRetries(ctx, "consume", func() error {
		return r.withTimeout(ctx, func(localCtx context.Context) error {
			r.mu.Lock()
			ch, err = r.ch.ConsumeWithContext(
				localCtx,
				r.topic,
				"",
				r.autoAck,
				false, // exclusive
				false, // noLocal
				false, // noWait
				nil,   // args
			)
			r.mu.Unlock()
			return err
		})
	})
	if err != nil {
		return nil, err
	}
	return ch, nil
}
