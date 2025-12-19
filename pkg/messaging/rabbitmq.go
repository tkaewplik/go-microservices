package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQConfig holds RabbitMQ connection configuration
type RabbitMQConfig struct {
	URL string // amqp://user:password@host:port/
}

// RabbitMQ represents a RabbitMQ connection with publisher and consumer capabilities
type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	logger  *slog.Logger
}

// NewRabbitMQ creates a new RabbitMQ connection
func NewRabbitMQ(cfg RabbitMQConfig, logger *slog.Logger) (*RabbitMQ, error) {
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	logger.Info("connected to RabbitMQ", "url", cfg.URL)

	return &RabbitMQ{
		conn:    conn,
		channel: ch,
		logger:  logger,
	}, nil
}

// Close closes the RabbitMQ connection
func (r *RabbitMQ) Close() error {
	if err := r.channel.Close(); err != nil {
		return fmt.Errorf("failed to close channel: %w", err)
	}
	if err := r.conn.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}
	r.logger.Info("RabbitMQ connection closed")
	return nil
}

// DeclareQueue declares a queue with the given name
func (r *RabbitMQ) DeclareQueue(name string) (amqp.Queue, error) {
	q, err := r.channel.QueueDeclare(
		name,  // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return amqp.Queue{}, fmt.Errorf("failed to declare queue %s: %w", name, err)
	}

	r.logger.Info("queue declared", "name", name)
	return q, nil
}

// Publish publishes a message to a queue
func (r *RabbitMQ) Publish(ctx context.Context, queueName string, message interface{}) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = r.channel.PublishWithContext(ctx,
		"",        // exchange
		queueName, // routing key (queue name)
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	r.logger.Debug("message published", "queue", queueName)
	return nil
}

// MessageHandler is a function that handles a consumed message
type MessageHandler func(body []byte) error

// Consume starts consuming messages from a queue
func (r *RabbitMQ) Consume(queueName string, handler MessageHandler) error {
	msgs, err := r.channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	go func() {
		for msg := range msgs {
			r.logger.Debug("message received", "queue", queueName)

			if err := handler(msg.Body); err != nil {
				r.logger.Error("failed to handle message", "error", err, "queue", queueName)
				msg.Nack(false, true) // requeue on failure
				continue
			}

			msg.Ack(false)
		}
	}()

	r.logger.Info("consumer started", "queue", queueName)
	return nil
}

// Common event types
const (
	EventUserRegistered     = "user.registered"
	EventTransactionCreated = "transaction.created"
	EventTransactionPaid    = "transaction.paid"
)

// UserRegisteredEvent represents a user registration event
type UserRegisteredEvent struct {
	EventType string    `json:"event_type"`
	UserID    int       `json:"user_id"`
	Username  string    `json:"username"`
	Timestamp time.Time `json:"timestamp"`
}

// TransactionCreatedEvent represents a transaction created event
type TransactionCreatedEvent struct {
	EventType     string    `json:"event_type"`
	TransactionID int       `json:"transaction_id"`
	UserID        int       `json:"user_id"`
	Amount        float64   `json:"amount"`
	Description   string    `json:"description"`
	Timestamp     time.Time `json:"timestamp"`
}
