package service

import (
	"context"
	"errors"
	"testing"

	"github.com/tkaewplik/go-microservices/payment-service/internal/domain"
)

// MockTransactionRepository is a mock implementation of TransactionRepository
type MockTransactionRepository struct {
	transactions []domain.Transaction
	createErr    error
	findErr      error
	updateErr    error
	nextID       int
}

func NewMockTransactionRepository() *MockTransactionRepository {
	return &MockTransactionRepository{
		transactions: []domain.Transaction{},
		nextID:       1,
	}
}

func (m *MockTransactionRepository) Create(ctx context.Context, tx *domain.Transaction) (*domain.Transaction, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	tx.ID = m.nextID
	m.nextID++
	m.transactions = append(m.transactions, *tx)
	return tx, nil
}

func (m *MockTransactionRepository) FindByUserID(ctx context.Context, userID int) ([]domain.Transaction, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	var result []domain.Transaction
	for _, tx := range m.transactions {
		if tx.UserID == userID {
			result = append(result, tx)
		}
	}
	return result, nil
}

func (m *MockTransactionRepository) GetTotalAmountByUserID(ctx context.Context, userID int) (float64, error) {
	if m.findErr != nil {
		return 0, m.findErr
	}
	var total float64
	for _, tx := range m.transactions {
		if tx.UserID == userID {
			total += tx.Amount
		}
	}
	return total, nil
}

func (m *MockTransactionRepository) MarkAllAsPaid(ctx context.Context, userID int) (int64, error) {
	if m.updateErr != nil {
		return 0, m.updateErr
	}
	var count int64
	for i := range m.transactions {
		if m.transactions[i].UserID == userID && !m.transactions[i].IsPaid {
			m.transactions[i].IsPaid = true
			count++
		}
	}
	return count, nil
}

func TestPaymentService_CreateTransaction_Success(t *testing.T) {
	repo := NewMockTransactionRepository()
	svc := NewPaymentService(repo)

	req := &domain.CreateTransactionRequest{
		UserID:      1,
		Amount:      100.50,
		Description: "Test transaction",
	}

	tx, err := svc.CreateTransaction(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if tx.ID != 1 {
		t.Errorf("expected ID 1, got %d", tx.ID)
	}
	if tx.Amount != 100.50 {
		t.Errorf("expected amount 100.50, got %f", tx.Amount)
	}
}

func TestPaymentService_CreateTransaction_InvalidAmount(t *testing.T) {
	repo := NewMockTransactionRepository()
	svc := NewPaymentService(repo)

	req := &domain.CreateTransactionRequest{
		UserID:      1,
		Amount:      0,
		Description: "Invalid transaction",
	}

	_, err := svc.CreateTransaction(context.Background(), req)
	if !errors.Is(err, ErrInvalidAmount) {
		t.Errorf("expected ErrInvalidAmount, got %v", err)
	}
}

func TestPaymentService_CreateTransaction_NegativeAmount(t *testing.T) {
	repo := NewMockTransactionRepository()
	svc := NewPaymentService(repo)

	req := &domain.CreateTransactionRequest{
		UserID:      1,
		Amount:      -50,
		Description: "Negative transaction",
	}

	_, err := svc.CreateTransaction(context.Background(), req)
	if !errors.Is(err, ErrInvalidAmount) {
		t.Errorf("expected ErrInvalidAmount, got %v", err)
	}
}

func TestPaymentService_CreateTransaction_ExceedsMaximum(t *testing.T) {
	repo := NewMockTransactionRepository()
	svc := NewPaymentService(repo)

	// Create first transaction close to limit
	req1 := &domain.CreateTransactionRequest{
		UserID:      1,
		Amount:      900,
		Description: "First transaction",
	}
	_, err := svc.CreateTransaction(context.Background(), req1)
	if err != nil {
		t.Fatalf("first transaction should succeed: %v", err)
	}

	// Try to create second transaction that exceeds limit
	req2 := &domain.CreateTransactionRequest{
		UserID:      1,
		Amount:      200,
		Description: "Second transaction",
	}
	_, err = svc.CreateTransaction(context.Background(), req2)
	if !errors.Is(err, ErrExceedsMaximum) {
		t.Errorf("expected ErrExceedsMaximum, got %v", err)
	}
}

func TestPaymentService_CreateTransaction_ExactlyAtMaximum(t *testing.T) {
	repo := NewMockTransactionRepository()
	svc := NewPaymentService(repo)

	// Create transaction exactly at limit
	req := &domain.CreateTransactionRequest{
		UserID:      1,
		Amount:      1000,
		Description: "Max transaction",
	}
	tx, err := svc.CreateTransaction(context.Background(), req)
	if err != nil {
		t.Fatalf("transaction at exactly max should succeed: %v", err)
	}

	if tx.Amount != 1000 {
		t.Errorf("expected amount 1000, got %f", tx.Amount)
	}
}

func TestPaymentService_GetTransactions_Success(t *testing.T) {
	repo := NewMockTransactionRepository()
	svc := NewPaymentService(repo)

	// Create a transaction
	req := &domain.CreateTransactionRequest{
		UserID:      1,
		Amount:      50,
		Description: "Test",
	}
	_, _ = svc.CreateTransaction(context.Background(), req)

	transactions, err := svc.GetTransactions(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(transactions) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(transactions))
	}
}

func TestPaymentService_GetTransactions_EmptyList(t *testing.T) {
	repo := NewMockTransactionRepository()
	svc := NewPaymentService(repo)

	transactions, err := svc.GetTransactions(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if transactions == nil {
		t.Error("expected empty slice, got nil")
	}
	if len(transactions) != 0 {
		t.Errorf("expected 0 transactions, got %d", len(transactions))
	}
}

func TestPaymentService_PayAllTransactions_Success(t *testing.T) {
	repo := NewMockTransactionRepository()
	svc := NewPaymentService(repo)

	// Create two transactions
	for i := 0; i < 2; i++ {
		req := &domain.CreateTransactionRequest{
			UserID:      1,
			Amount:      50,
			Description: "Test",
		}
		_, _ = svc.CreateTransaction(context.Background(), req)
	}

	count, err := svc.PayAllTransactions(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if count != 2 {
		t.Errorf("expected 2 transactions paid, got %d", count)
	}
}

func TestPaymentService_InvalidUserID(t *testing.T) {
	repo := NewMockTransactionRepository()
	svc := NewPaymentService(repo)

	req := &domain.CreateTransactionRequest{
		UserID:      0,
		Amount:      50,
		Description: "Test",
	}

	_, err := svc.CreateTransaction(context.Background(), req)
	if !errors.Is(err, ErrInvalidUserID) {
		t.Errorf("expected ErrInvalidUserID, got %v", err)
	}
}
