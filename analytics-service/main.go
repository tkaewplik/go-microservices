package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
)

// TransactionEvent represents a transaction event from Kafka
type TransactionEvent struct {
	EventType        string    `json:"event_type"`
	TransactionID    int       `json:"transaction_id,omitempty"`
	UserID           int       `json:"user_id"`
	Amount           float64   `json:"amount,omitempty"`
	Description      string    `json:"description,omitempty"`
	TransactionsPaid int64     `json:"transactions_paid,omitempty"`
	Timestamp        time.Time `json:"timestamp"`
}

// Analytics holds aggregated analytics data
type Analytics struct {
	mu                    sync.RWMutex
	TotalTransactions     int64           `json:"total_transactions"`
	TotalAmount           float64         `json:"total_amount"`
	TotalPaidTransactions int64           `json:"total_paid_transactions"`
	EventsProcessed       int64           `json:"events_processed"`
	LastEventTime         string          `json:"last_event_time,omitempty"`
	TransactionsByUser    map[int]int64   `json:"transactions_by_user"`
	AmountByUser          map[int]float64 `json:"amount_by_user"`
}

func NewAnalytics() *Analytics {
	return &Analytics{
		TransactionsByUser: make(map[int]int64),
		AmountByUser:       make(map[int]float64),
	}
}

func (a *Analytics) ProcessEvent(event *TransactionEvent) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.EventsProcessed++
	a.LastEventTime = event.Timestamp.Format(time.RFC3339)

	switch event.EventType {
	case "transaction.created":
		a.TotalTransactions++
		a.TotalAmount += event.Amount
		a.TransactionsByUser[event.UserID]++
		a.AmountByUser[event.UserID] += event.Amount
	case "transaction.paid":
		a.TotalPaidTransactions += event.TransactionsPaid
	}
}

func (a *Analytics) GetStats() map[string]interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return map[string]interface{}{
		"total_transactions":      a.TotalTransactions,
		"total_amount":            a.TotalAmount,
		"total_paid_transactions": a.TotalPaidTransactions,
		"events_processed":        a.EventsProcessed,
		"last_event_time":         a.LastEventTime,
		"unique_users":            len(a.TransactionsByUser),
	}
}

func main() {
	// Setup structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Kafka configuration
	brokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ",")
	topic := getEnv("KAFKA_TOPIC", "transactions")
	groupID := getEnv("KAFKA_GROUP_ID", "analytics-consumer")
	port := getEnv("PORT", "8083")

	// Create analytics aggregator
	analytics := NewAnalytics()

	// Create Kafka reader
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: time.Second,
		StartOffset:    kafka.FirstOffset,
	})

	logger.Info("analytics service starting",
		"port", port,
		"kafka_brokers", brokers,
		"kafka_topic", topic,
		"kafka_group", groupID,
	)

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start Kafka consumer in background
	go func() {
		for {
			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return // Context cancelled
				}
				logger.Error("failed to read message", "error", err)
				continue
			}

			var event TransactionEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				logger.Error("failed to unmarshal event", "error", err)
				continue
			}

			analytics.ProcessEvent(&event)

			logger.Info("event processed",
				"event_type", event.EventType,
				"user_id", event.UserID,
				"partition", msg.Partition,
				"offset", msg.Offset,
			)
		}
	}()

	// HTTP server for analytics API
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"status":"ok"}`)); err != nil {
			logger.Error("failed to write response", "error", err)
		}
	})

	// Analytics stats endpoint
	mux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		stats := analytics.GetStats()
		if err := json.NewEncoder(w).Encode(stats); err != nil {
			logger.Error("failed to encode stats", "error", err)
		}
	})

	// Start HTTP server
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server failed", "error", err)
		}
	}()

	logger.Info("analytics service ready", "port", port)

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("shutting down...")
	cancel()

	// Shutdown HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server shutdown error", "error", err)
	}

	// Close Kafka reader
	if err := reader.Close(); err != nil {
		logger.Error("Kafka reader close error", "error", err)
	}

	logger.Info("analytics service stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
