// Package rabbitmq provides a RabbitMQ client, which is suitable for sending and receiving messages.
// It is expexted to work with a single queue.
package rabbitmq

import (
	"context"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go" //nolint:depguard,nolintlint
)

// RabbitMQ represents a RabbitMQ client.
type RabbitMQ struct {
	mu sync.RWMutex

	conn *amqp.Connection
	ch   *amqp.Channel

	url          string
	timeout      time.Duration
	retryTimeout time.Duration
	retries      int

	topic       string
	durable     bool
	contentType string

	routingKey string

	consumerDone chan struct{}
	autoAck      bool
	requeue      bool

	l Logger
}

// NewRabbitMQ creates a new RabbitMQ client.
func NewRabbitMQ(logger Logger, config map[string]any) (*RabbitMQ, error) {
	// Args validation.
	missing := make([]string, 0)
	if logger == nil {
		missing = append(missing, "logger")
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("RabbitMQ sender: some of the required parameters are missing: args=%v", missing)
	}

	// Field types validation.
	missing, wrongType := validateFields(config, expectedFields)
	if len(missing) > 0 || len(wrongType) > 0 {
		return nil, fmt.Errorf("invalid RabbitMQ config: missing=%v invalid_type=%v",
			missing, wrongType)
	}

	return &RabbitMQ{
		l: logger,
		url: fmt.Sprintf(
			"amqp://%s:%s@%s:%s/",
			config["user"].(string),
			config["password"].(string),
			config["host"].(string),
			config["port"].(string),
		),
		timeout:      config["timeout"].(time.Duration),
		retryTimeout: config["retry_timeout"].(time.Duration),
		retries:      config["retries"].(int),
		topic:        config["topic"].(string),
		durable:      config["durable"].(bool),
		contentType:  config["content_type"].(string),
		routingKey:   config["routing_key"].(string),
		autoAck:      config["auto_ack"].(bool),
		requeue:      config["requeue"].(bool),
	}, nil
}

// Connect to the message queue.
func (r *RabbitMQ) Connect(_ context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	conn, err := amqp.Dial(r.url)
	if err != nil {
		return fmt.Errorf("message queue connection: %w", err)
	}
	r.conn = conn

	ch, err := r.conn.Channel()
	if err != nil {
		return fmt.Errorf("message queue channel creation: %w", err)
	}
	r.ch = ch

	err = r.ch.ExchangeDeclare(
		r.topic,
		"direct",
		r.durable,
		false, // autoDelete
		false, // internal
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("message queue exchange declaration: %w", err)
	}

	q, err := r.ch.QueueDeclare(
		r.topic,
		r.durable,
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("message queue declaration: %w", err)
	}

	err = r.ch.QueueBind(
		q.Name,
		r.topic,
		r.topic,
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("message queue binding: %w", err)
	}

	return nil
}

// Close the connection to the message queue.
func (r *RabbitMQ) Close(_ context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.conn == nil && r.ch == nil {
		return nil
	}

	var errs []error

	if r.ch != nil {
		err := r.ch.Close()
		if err != nil {
			errs = append(errs, err)
		}
		r.ch = nil
	}

	if r.conn != nil {
		err := r.conn.Close()
		if err != nil {
			errs = append(errs, err)
		}
		r.conn = nil
	}

	var err error
	for i := range errs {
		err = fmt.Errorf("%w: %w", errs[i], err)
	}
	if err != nil {
		return fmt.Errorf("message queue close: %w", err)
	}

	return nil
}
