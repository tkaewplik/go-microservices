package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/tkaewplik/go-microservices/auth-service/internal/service"
	"github.com/tkaewplik/go-microservices/pkg/jwt"
	pb "github.com/tkaewplik/go-microservices/proto/auth"
)

// AuthServer implements the gRPC AuthService
type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	authService *service.AuthService
	jwtSecret   string
}

// NewAuthServer creates a new gRPC AuthServer
func NewAuthServer(authService *service.AuthService, jwtSecret string) *AuthServer {
	return &AuthServer{
		authService: authService,
		jwtSecret:   jwtSecret,
	}
}

// Register creates a new user account
func (s *AuthServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.AuthResponse, error) {
	if req.Username == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password are required")
	}

	resp, err := s.authService.Register(ctx, req.Username, req.Password)
	if err != nil {
		if err == service.ErrUserAlreadyExists {
			return nil, status.Error(codes.AlreadyExists, "username already exists")
		}
		return nil, status.Error(codes.Internal, "failed to register user")
	}

	return &pb.AuthResponse{
		Id:       int32(resp.ID),
		Username: resp.Username,
		Token:    resp.Token,
	}, nil
}

// Login authenticates a user and returns a token
func (s *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.AuthResponse, error) {
	if req.Username == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password are required")
	}

	resp, err := s.authService.Login(ctx, req.Username, req.Password)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, "failed to login")
	}

	return &pb.AuthResponse{
		Id:       int32(resp.ID),
		Username: resp.Username,
		Token:    resp.Token,
	}, nil
}

// ValidateToken validates a JWT token and returns user info
func (s *AuthServer) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	if req.Token == "" {
		return &pb.ValidateTokenResponse{Valid: false}, nil
	}

	claims, err := jwt.ValidateToken(req.Token, s.jwtSecret)
	if err != nil {
		return &pb.ValidateTokenResponse{Valid: false}, nil
	}

	return &pb.ValidateTokenResponse{
		Valid:    true,
		UserId:   int32(claims.UserID),
		Username: claims.Username,
	}, nil
}
