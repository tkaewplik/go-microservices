package domain

import (
	"context"
	"time"
)

// EventPublisher defines the interface for publishing events
type EventPublisher interface {
	// PublishTransactionCreated publishes a transaction created event
	PublishTransactionCreated(ctx context.Context, event *TransactionCreatedEvent) error
	// PublishTransactionPaid publishes a transaction paid event
	PublishTransactionPaid(ctx context.Context, event *TransactionPaidEvent) error
	// Close closes the publisher
	Close() error
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

// TransactionPaidEvent represents a transaction paid event
type TransactionPaidEvent struct {
	EventType        string    `json:"event_type"`
	UserID           int       `json:"user_id"`
	TransactionsPaid int64     `json:"transactions_paid"`
	Timestamp        time.Time `json:"timestamp"`
}
