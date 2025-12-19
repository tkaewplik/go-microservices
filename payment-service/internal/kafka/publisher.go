package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/tkaewplik/go-microservices/payment-service/internal/domain"
)

// Publisher implements domain.EventPublisher using Kafka
type Publisher struct {
	writer *kafka.Writer
	logger *slog.Logger
}

// Config holds Kafka publisher configuration
type Config struct {
	Brokers []string
	Topic   string
}

// NewPublisher creates a new Kafka publisher
func NewPublisher(cfg Config, logger *slog.Logger) *Publisher {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.Topic,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
	}

	logger.Info("Kafka publisher created", "brokers", cfg.Brokers, "topic", cfg.Topic)

	return &Publisher{
		writer: writer,
		logger: logger,
	}
}

// PublishTransactionCreated publishes a transaction created event
func (p *Publisher) PublishTransactionCreated(ctx context.Context, event *domain.TransactionCreatedEvent) error {
	event.EventType = "transaction.created"
	event.Timestamp = time.Now()

	value, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	key := strconv.Itoa(event.UserID)

	err = p.writer.WriteMessages(ctx,
		kafka.Message{
			Key:   []byte(key),
			Value: value,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	p.logger.Info("transaction.created event published",
		"transaction_id", event.TransactionID,
		"user_id", event.UserID,
		"amount", event.Amount,
	)

	return nil
}

// PublishTransactionPaid publishes a transaction paid event
func (p *Publisher) PublishTransactionPaid(ctx context.Context, event *domain.TransactionPaidEvent) error {
	event.EventType = "transaction.paid"
	event.Timestamp = time.Now()

	value, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	key := strconv.Itoa(event.UserID)

	err = p.writer.WriteMessages(ctx,
		kafka.Message{
			Key:   []byte(key),
			Value: value,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	p.logger.Info("transaction.paid event published",
		"user_id", event.UserID,
		"transactions_paid", event.TransactionsPaid,
	)

	return nil
}

// Close closes the Kafka writer
func (p *Publisher) Close() error {
	if err := p.writer.Close(); err != nil {
		return fmt.Errorf("failed to close Kafka writer: %w", err)
	}
	p.logger.Info("Kafka publisher closed")
	return nil
}
