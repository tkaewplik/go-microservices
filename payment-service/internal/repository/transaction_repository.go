package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/tkaewplik/go-microservices/payment-service/internal/domain"
)

// PostgresTransactionRepository implements TransactionRepository using PostgreSQL
type PostgresTransactionRepository struct {
	db *sql.DB
}

// NewPostgresTransactionRepository creates a new PostgresTransactionRepository
func NewPostgresTransactionRepository(db *sql.DB) *PostgresTransactionRepository {
	return &PostgresTransactionRepository{db: db}
}

// Create creates a new transaction in the database
func (r *PostgresTransactionRepository) Create(ctx context.Context, tx *domain.Transaction) (*domain.Transaction, error) {
	query := `
		INSERT INTO transactions (user_id, amount, description, is_paid) 
		VALUES ($1, $2, $3, false) 
		RETURNING id, user_id, amount, description, is_paid, created_at`

	err := r.db.QueryRowContext(ctx, query, tx.UserID, tx.Amount, tx.Description).Scan(
		&tx.ID, &tx.UserID, &tx.Amount, &tx.Description, &tx.IsPaid, &tx.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	return tx, nil
}

// FindByUserID finds all transactions for a user
func (r *PostgresTransactionRepository) FindByUserID(ctx context.Context, userID int) ([]domain.Transaction, error) {
	query := `
		SELECT id, user_id, amount, description, is_paid, created_at 
		FROM transactions 
		WHERE user_id = $1 
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer rows.Close()

	var transactions []domain.Transaction
	for rows.Next() {
		var t domain.Transaction
		if err := rows.Scan(&t.ID, &t.UserID, &t.Amount, &t.Description, &t.IsPaid, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transactions: %w", err)
	}

	return transactions, nil
}

// GetTotalAmountByUserID returns the total amount of all transactions for a user
func (r *PostgresTransactionRepository) GetTotalAmountByUserID(ctx context.Context, userID int) (float64, error) {
	query := "SELECT COALESCE(SUM(amount), 0) FROM transactions WHERE user_id = $1"

	var total float64
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to get total amount: %w", err)
	}

	return total, nil
}

// MarkAllAsPaid marks all unpaid transactions for a user as paid
func (r *PostgresTransactionRepository) MarkAllAsPaid(ctx context.Context, userID int) (int64, error) {
	query := "UPDATE transactions SET is_paid = true WHERE user_id = $1 AND is_paid = false"

	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to mark transactions as paid: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}
