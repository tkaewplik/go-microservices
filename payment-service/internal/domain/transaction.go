package domain

import (
	"context"
	"time"
)

// Transaction represents a payment transaction
type Transaction struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
	IsPaid      bool      `json:"is_paid"`
	CreatedAt   time.Time `json:"created_at"`
}

// TransactionRepository defines the interface for transaction data access
type TransactionRepository interface {
	// Create creates a new transaction
	Create(ctx context.Context, tx *Transaction) (*Transaction, error)
	// FindByUserID finds all transactions for a user
	FindByUserID(ctx context.Context, userID int) ([]Transaction, error)
	// GetTotalAmountByUserID returns the total amount of all transactions for a user
	GetTotalAmountByUserID(ctx context.Context, userID int) (float64, error)
	// MarkAllAsPaid marks all unpaid transactions for a user as paid
	MarkAllAsPaid(ctx context.Context, userID int) (int64, error)
}

// CreateTransactionRequest represents the request to create a transaction
type CreateTransactionRequest struct {
	UserID      int     `json:"user_id"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
}
