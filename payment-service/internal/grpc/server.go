package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/tkaewplik/go-microservices/payment-service/internal/domain"
	"github.com/tkaewplik/go-microservices/payment-service/internal/service"
	pb "github.com/tkaewplik/go-microservices/proto/payment"
)

// PaymentServer implements the gRPC PaymentService
type PaymentServer struct {
	pb.UnimplementedPaymentServiceServer
	paymentService *service.PaymentService
}

// NewPaymentServer creates a new gRPC PaymentServer
func NewPaymentServer(paymentService *service.PaymentService) *PaymentServer {
	return &PaymentServer{
		paymentService: paymentService,
	}
}

// CreateTransaction creates a new transaction
func (s *PaymentServer) CreateTransaction(ctx context.Context, req *pb.CreateTransactionRequest) (*pb.Transaction, error) {
	if req.UserId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	if req.Amount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount must be positive")
	}

	tx, err := s.paymentService.CreateTransaction(ctx, &domain.CreateTransactionRequest{
		UserID:      int(req.UserId),
		Amount:      req.Amount,
		Description: req.Description,
	})
	if err != nil {
		if err == service.ErrInvalidAmount {
			return nil, status.Error(codes.InvalidArgument, "amount must be positive")
		}
		if err == service.ErrInvalidUserID {
			return nil, status.Error(codes.InvalidArgument, "invalid user_id")
		}
		if err == service.ErrExceedsMaximum {
			return nil, status.Error(codes.FailedPrecondition, "total amount exceeds maximum of 1000")
		}
		return nil, status.Error(codes.Internal, "failed to create transaction")
	}

	return &pb.Transaction{
		Id:          int32(tx.ID),
		UserId:      int32(tx.UserID),
		Amount:      tx.Amount,
		Description: tx.Description,
		IsPaid:      tx.IsPaid,
		CreatedAt:   timestamppb.New(tx.CreatedAt),
	}, nil
}

// GetTransactions returns all transactions for a user
func (s *PaymentServer) GetTransactions(ctx context.Context, req *pb.GetTransactionsRequest) (*pb.TransactionList, error) {
	if req.UserId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	transactions, err := s.paymentService.GetTransactions(ctx, int(req.UserId))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get transactions")
	}

	pbTransactions := make([]*pb.Transaction, len(transactions))
	for i, tx := range transactions {
		pbTransactions[i] = &pb.Transaction{
			Id:          int32(tx.ID),
			UserId:      int32(tx.UserID),
			Amount:      tx.Amount,
			Description: tx.Description,
			IsPaid:      tx.IsPaid,
			CreatedAt:   timestamppb.New(tx.CreatedAt),
		}
	}

	return &pb.TransactionList{Transactions: pbTransactions}, nil
}

// PayAllTransactions marks all unpaid transactions as paid
func (s *PaymentServer) PayAllTransactions(ctx context.Context, req *pb.PayRequest) (*pb.PayResponse, error) {
	if req.UserId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	count, err := s.paymentService.PayAllTransactions(ctx, int(req.UserId))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to pay transactions")
	}

	return &pb.PayResponse{
		Message:          "transactions paid successfully",
		TransactionsPaid: count,
	}, nil
}
