package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/tkaewplik/go-microservices/payment-service/internal/domain"
	"github.com/tkaewplik/go-microservices/payment-service/internal/service"
)

// PaymentHandler handles HTTP requests for payments
type PaymentHandler struct {
	paymentService *service.PaymentService
	logger         *slog.Logger
}

// NewPaymentHandler creates a new PaymentHandler
func NewPaymentHandler(paymentService *service.PaymentService, logger *slog.Logger) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
		logger:         logger,
	}
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error        string `json:"error"`
	CurrentTotal string `json:"current_total,omitempty"`
	MaxAllowed   string `json:"max_allowed,omitempty"`
}

// PayResponse represents a pay response
type PayResponse struct {
	Message          string `json:"message"`
	TransactionsPaid int64  `json:"transactions_paid"`
}

// CreateTransaction handles transaction creation
func (h *PaymentHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req domain.CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode create transaction request", "error", err)
		h.respondError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	tx, err := h.paymentService.CreateTransaction(ctx, &req)
	if err != nil {
		h.logger.Error("failed to create transaction", "error", err, "user_id", req.UserID)

		if errors.Is(err, service.ErrInvalidAmount) {
			h.respondError(w, http.StatusBadRequest, "amount must be positive", nil)
			return
		}

		if errors.Is(err, service.ErrInvalidUserID) {
			h.respondError(w, http.StatusBadRequest, "invalid user_id", nil)
			return
		}

		if errors.Is(err, service.ErrExceedsMaximum) {
			// Get current total for detailed error
			currentTotal, _ := h.paymentService.GetCurrentTotal(ctx, req.UserID)
			h.respondJSON(w, http.StatusBadRequest, ErrorResponse{
				Error:        "total amount exceeds maximum of 1000",
				CurrentTotal: formatFloat(currentTotal),
				MaxAllowed:   "1000.00",
			})
			return
		}

		h.respondError(w, http.StatusInternalServerError, "failed to create transaction", nil)
		return
	}

	h.logger.Info("transaction created", "id", tx.ID, "user_id", tx.UserID, "amount", tx.Amount)
	h.respondJSON(w, http.StatusCreated, tx)
}

// GetTransactions handles getting transactions for a user
func (h *PaymentHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		h.respondError(w, http.StatusBadRequest, "user_id query parameter required", nil)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid user_id", nil)
		return
	}

	transactions, err := h.paymentService.GetTransactions(ctx, userID)
	if err != nil {
		h.logger.Error("failed to get transactions", "error", err, "user_id", userID)

		if errors.Is(err, service.ErrInvalidUserID) {
			h.respondError(w, http.StatusBadRequest, "invalid user_id", nil)
			return
		}

		h.respondError(w, http.StatusInternalServerError, "failed to get transactions", nil)
		return
	}

	h.respondJSON(w, http.StatusOK, transactions)
}

// PayAllTransactions handles paying all transactions for a user
func (h *PaymentHandler) PayAllTransactions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		h.respondError(w, http.StatusBadRequest, "user_id query parameter required", nil)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid user_id", nil)
		return
	}

	rowsAffected, err := h.paymentService.PayAllTransactions(ctx, userID)
	if err != nil {
		h.logger.Error("failed to pay transactions", "error", err, "user_id", userID)

		if errors.Is(err, service.ErrInvalidUserID) {
			h.respondError(w, http.StatusBadRequest, "invalid user_id", nil)
			return
		}

		h.respondError(w, http.StatusInternalServerError, "failed to pay transactions", nil)
		return
	}

	h.logger.Info("transactions paid", "user_id", userID, "count", rowsAffected)
	h.respondJSON(w, http.StatusOK, PayResponse{
		Message:          "transactions paid successfully",
		TransactionsPaid: rowsAffected,
	})
}

// respondJSON writes a JSON response
func (h *PaymentHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

// respondError writes an error response
func (h *PaymentHandler) respondError(w http.ResponseWriter, status int, message string, extra map[string]string) {
	resp := ErrorResponse{Error: message}
	if extra != nil {
		resp.CurrentTotal = extra["current_total"]
		resp.MaxAllowed = extra["max_allowed"]
	}
	h.respondJSON(w, status, resp)
}

func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}
