package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/tkaewplik/go-microservices/payment-service/internal/domain"
)

const MaxTransactionTotal = 1000.0

// Common errors
var (
	ErrInvalidAmount  = errors.New("amount must be positive")
	ErrExceedsMaximum = errors.New("total amount exceeds maximum")
	ErrInvalidUserID  = errors.New("invalid user ID")
)

// PaymentService handles payment business logic
type PaymentService struct {
	txRepo domain.TransactionRepository
}

// NewPaymentService creates a new PaymentService
func NewPaymentService(txRepo domain.TransactionRepository) *PaymentService {
	return &PaymentService{txRepo: txRepo}
}

// CreateTransaction creates a new transaction with validation
func (s *PaymentService) CreateTransaction(ctx context.Context, req *domain.CreateTransactionRequest) (*domain.Transaction, error) {
	// Validate amount
	if req.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	if req.UserID <= 0 {
		return nil, ErrInvalidUserID
	}

	// Check if total amount exceeds maximum
	currentTotal, err := s.txRepo.GetTotalAmountByUserID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current total: %w", err)
	}

	if currentTotal+req.Amount > MaxTransactionTotal {
		return nil, fmt.Errorf("%w: current total %.2f, requested %.2f, max %.2f",
			ErrExceedsMaximum, currentTotal, req.Amount, MaxTransactionTotal)
	}

	// Create transaction
	tx := &domain.Transaction{
		UserID:      req.UserID,
		Amount:      req.Amount,
		Description: req.Description,
	}

	createdTx, err := s.txRepo.Create(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	return createdTx, nil
}

// GetTransactions returns all transactions for a user
func (s *PaymentService) GetTransactions(ctx context.Context, userID int) ([]domain.Transaction, error) {
	if userID <= 0 {
		return nil, ErrInvalidUserID
	}

	transactions, err := s.txRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	// Return empty slice instead of nil
	if transactions == nil {
		transactions = []domain.Transaction{}
	}

	return transactions, nil
}

// PayAllTransactions marks all unpaid transactions for a user as paid
func (s *PaymentService) PayAllTransactions(ctx context.Context, userID int) (int64, error) {
	if userID <= 0 {
		return 0, ErrInvalidUserID
	}

	rowsAffected, err := s.txRepo.MarkAllAsPaid(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to pay transactions: %w", err)
	}

	return rowsAffected, nil
}

// GetCurrentTotal returns the current total amount for a user
func (s *PaymentService) GetCurrentTotal(ctx context.Context, userID int) (float64, error) {
	if userID <= 0 {
		return 0, ErrInvalidUserID
	}

	return s.txRepo.GetTotalAmountByUserID(ctx, userID)
}
