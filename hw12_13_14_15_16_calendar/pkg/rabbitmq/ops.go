package rabbitmq

import (
	"context"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go" //nolint:depguard,nolintlint
)

// Produce sends a message to the message queue using retry logic and operation timeout.
func (r *RabbitMQ) Produce(ctx context.Context, payload []byte) error {
	err := r.withRetries(ctx, "produce", func() error {
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
	if err != nil {
		return err
	}
	r.l.Debug(ctx, "message successfully sent")
	return nil
}

// Consume consumes messages from the message queue using retry logic and operation timeout.
// The methods abstracts the logic of consuming messages from the message queue.
func (r *RabbitMQ) Consume(ctx context.Context) (<-chan []byte, error) {
	var err error
	var ch <-chan amqp.Delivery
	var autoAck, requeue bool

	err = r.withRetries(ctx, "consume", func() error {
		return r.withTimeout(ctx, func(localCtx context.Context) error {
			r.mu.Lock()
			// Stopping the active consumer if it exists.
			if r.consumerDone != nil {
				close(r.consumerDone)
				r.consumerDone = nil
			}

			autoAck = r.autoAck
			requeue = r.requeue
			ch, err = r.ch.ConsumeWithContext(
				localCtx,
				r.topic,
				"",
				autoAck,
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

	resQueue := make(chan []byte)
	go r.populateConsumer(ctx, ch, resQueue, autoAck, requeue)

	return resQueue, nil
}

// populateConsumer reads messages from the AMPQ channel and sends raw data to the result channel.
func (r *RabbitMQ) populateConsumer(
	ctx context.Context,
	ch <-chan amqp.Delivery,
	resQueue chan<- []byte,
	autoAck,
	requeue bool,
) {
	defer close(resQueue)
	for {
		select {
		case <-ctx.Done():
			r.l.Warn(ctx, "context done before the message extracted")
			return
		case <-r.consumerDone:
			r.l.Info(ctx, "consumer stopped")
			return
		case msg := <-ch:
			if !autoAck {
				if err := msg.Ack(false); err != nil {
					r.l.Warn(
						ctx,
						"message ack failed",
						slog.String("message_id", msg.MessageId),
						slog.Any("error", err),
					)
					err = msg.Reject(requeue)
					if err != nil {
						r.l.Error(
							ctx,
							"message reject failed",
							slog.String("message_id", msg.MessageId),
							slog.Any("error", err),
						)
						continue
					}
				}
			}
			select {
			case <-ctx.Done():
				r.l.Warn(ctx, "context done before the message extracted")
			case resQueue <- msg.Body:
				r.l.Info(ctx, "message successfully received", slog.String("message_id", msg.MessageId))
			}
		}
	}
}
