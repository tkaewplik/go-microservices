package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

// KafkaConfig holds Kafka connection configuration
type KafkaConfig struct {
	Brokers []string // e.g., ["localhost:9092"]
}

// KafkaProducer represents a Kafka producer
type KafkaProducer struct {
	writer *kafka.Writer
	logger *slog.Logger
}

// KafkaConsumer represents a Kafka consumer
type KafkaConsumer struct {
	reader *kafka.Reader
	logger *slog.Logger
}

// NewKafkaProducer creates a new Kafka producer
func NewKafkaProducer(cfg KafkaConfig, topic string, logger *slog.Logger) *KafkaProducer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
	}

	logger.Info("Kafka producer created", "brokers", cfg.Brokers, "topic", topic)

	return &KafkaProducer{
		writer: writer,
		logger: logger,
	}
}

// Publish publishes a message to Kafka
func (p *KafkaProducer) Publish(ctx context.Context, key string, message interface{}) error {
	value, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = p.writer.WriteMessages(ctx,
		kafka.Message{
			Key:   []byte(key),
			Value: value,
			Time:  time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	p.logger.Debug("message published to Kafka", "key", key)
	return nil
}

// Close closes the Kafka producer
func (p *KafkaProducer) Close() error {
	if err := p.writer.Close(); err != nil {
		return fmt.Errorf("failed to close Kafka producer: %w", err)
	}
	p.logger.Info("Kafka producer closed")
	return nil
}

// NewKafkaConsumer creates a new Kafka consumer
func NewKafkaConsumer(cfg KafkaConfig, topic, groupID string, logger *slog.Logger) *KafkaConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       1,    // 1B
		MaxBytes:       10e6, // 10MB
		CommitInterval: time.Second,
	})

	logger.Info("Kafka consumer created", "brokers", cfg.Brokers, "topic", topic, "group", groupID)

	return &KafkaConsumer{
		reader: reader,
		logger: logger,
	}
}

// KafkaMessageHandler is a function that handles a Kafka message
type KafkaMessageHandler func(key, value []byte) error

// Consume starts consuming messages from Kafka
func (c *KafkaConsumer) Consume(ctx context.Context, handler KafkaMessageHandler) error {
	c.logger.Info("starting Kafka consumer")

	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				// Context cancelled, graceful shutdown
				return nil
			}
			c.logger.Error("failed to read message", "error", err)
			continue
		}

		c.logger.Debug("message received from Kafka",
			"topic", msg.Topic,
			"partition", msg.Partition,
			"offset", msg.Offset,
			"key", string(msg.Key),
		)

		if err := handler(msg.Key, msg.Value); err != nil {
			c.logger.Error("failed to handle message", "error", err)
			// Continue processing other messages
		}
	}
}

// Close closes the Kafka consumer
func (c *KafkaConsumer) Close() error {
	if err := c.reader.Close(); err != nil {
		return fmt.Errorf("failed to close Kafka consumer: %w", err)
	}
	c.logger.Info("Kafka consumer closed")
	return nil
}

// Kafka topics
const (
	TopicTransactions = "transactions"
	TopicUserEvents   = "user-events"
)

// TransactionEvent represents a transaction event for Kafka
type TransactionEvent struct {
	EventType     string    `json:"event_type"`
	TransactionID int       `json:"transaction_id"`
	UserID        int       `json:"user_id"`
	Amount        float64   `json:"amount"`
	Description   string    `json:"description"`
	IsPaid        bool      `json:"is_paid"`
	Timestamp     time.Time `json:"timestamp"`
}
