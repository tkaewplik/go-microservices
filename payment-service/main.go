package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/tkaewplik/go-microservices/pkg/database"
	"github.com/tkaewplik/go-microservices/pkg/middleware"
)

type Transaction struct {
	ID          int     `json:"id"`
	UserID      int     `json:"user_id"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	IsPaid      bool    `json:"is_paid"`
	CreatedAt   string  `json:"created_at"`
}

type CreateTransactionRequest struct {
	UserID      int     `json:"user_id"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
}

type PaymentService struct {
	db        *sql.DB
	secretKey string
}

func NewPaymentService(db *sql.DB, secretKey string) *PaymentService {
	return &PaymentService{db: db, secretKey: secretKey}
}

func (s *PaymentService) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var req CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// Validate amount
	if req.Amount <= 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "amount must be positive"})
		return
	}

	// Check if total amount exceeds 1000
	var totalAmount float64
	err := s.db.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM transactions WHERE user_id = $1", req.UserID).Scan(&totalAmount)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "database error"})
		return
	}

	if totalAmount+req.Amount > 1000 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error":         "total amount exceeds maximum of 1000",
			"current_total": fmt.Sprintf("%.2f", totalAmount),
			"max_allowed":   "1000.00",
		})
		return
	}

	var transaction Transaction
	err = s.db.QueryRow(`
		INSERT INTO transactions (user_id, amount, description, is_paid) 
		VALUES ($1, $2, $3, false) 
		RETURNING id, user_id, amount, description, is_paid, created_at`,
		req.UserID, req.Amount, req.Description).Scan(
		&transaction.ID, &transaction.UserID, &transaction.Amount,
		&transaction.Description, &transaction.IsPaid, &transaction.CreatedAt)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "error creating transaction"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(transaction)
}

func (s *PaymentService) GetTransactions(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "user_id query parameter required"})
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid user_id"})
		return
	}

	rows, err := s.db.Query(`
		SELECT id, user_id, amount, description, is_paid, created_at 
		FROM transactions 
		WHERE user_id = $1 
		ORDER BY created_at DESC`, userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "database error"})
		return
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		var t Transaction
		if err := rows.Scan(&t.ID, &t.UserID, &t.Amount, &t.Description, &t.IsPaid, &t.CreatedAt); err != nil {
			log.Printf("Error scanning transaction row: %v", err)
			continue
		}
		transactions = append(transactions, t)
	}

	if transactions == nil {
		transactions = []Transaction{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}

func (s *PaymentService) PayAllTransactions(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "user_id query parameter required"})
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid user_id"})
		return
	}

	result, err := s.db.Exec("UPDATE transactions SET is_paid = true WHERE user_id = $1 AND is_paid = false", userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "database error"})
		return
	}

	rowsAffected, _ := result.RowsAffected()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":           "transactions paid successfully",
		"transactions_paid": rowsAffected,
	})
}

func main() {
	dbConfig := database.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnvInt("DB_PORT", 5432),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		DBName:   getEnv("DB_NAME", "paymentdb"),
	}

	db, err := database.Connect(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	secretKey := getEnv("JWT_SECRET", "your-secret-key")
	service := NewPaymentService(db, secretKey)
	authMiddleware := middleware.NewAuthMiddleware(secretKey)

	mux := http.NewServeMux()
	mux.HandleFunc("/transactions", authMiddleware.Authenticate(service.CreateTransaction))
	mux.HandleFunc("/transactions/list", authMiddleware.Authenticate(service.GetTransactions))
	mux.HandleFunc("/transactions/pay", authMiddleware.Authenticate(service.PayAllTransactions))

	port := getEnv("PORT", "8082")
	log.Printf("Payment service starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
