package jwt

import (
	"testing"
	"time"
)

func TestGenerateToken_Success(t *testing.T) {
	token, err := GenerateToken(1, "testuser", "secret-key")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if token == "" {
		t.Error("expected token to be generated")
	}
}

func TestValidateToken_Success(t *testing.T) {
	secretKey := "test-secret-key"

	// Generate token
	token, err := GenerateToken(42, "johndoe", secretKey)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Validate token
	claims, err := ValidateToken(token, secretKey)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if claims.UserID != 42 {
		t.Errorf("expected user ID 42, got %d", claims.UserID)
	}
	if claims.Username != "johndoe" {
		t.Errorf("expected username 'johndoe', got %s", claims.Username)
	}
}

func TestValidateToken_InvalidSecret(t *testing.T) {
	// Generate token with one secret
	token, err := GenerateToken(1, "testuser", "correct-secret")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Validate with different secret
	_, err = ValidateToken(token, "wrong-secret")
	if err == nil {
		t.Error("expected error for invalid secret")
	}
}

func TestValidateToken_InvalidToken(t *testing.T) {
	_, err := ValidateToken("invalid-token-string", "secret-key")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestValidateToken_EmptyToken(t *testing.T) {
	_, err := ValidateToken("", "secret-key")
	if err == nil {
		t.Error("expected error for empty token")
	}
}

func TestTokenExpiration(t *testing.T) {
	secretKey := "test-secret"

	// Generate a valid token
	token, err := GenerateToken(1, "testuser", secretKey)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Validate the token
	claims, err := ValidateToken(token, secretKey)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	// Check that expiration is set (24 hours from now)
	expectedExpiry := time.Now().Add(24 * time.Hour)
	actualExpiry := claims.ExpiresAt.Time

	// Allow 1 minute tolerance
	if actualExpiry.Before(expectedExpiry.Add(-1*time.Minute)) || actualExpiry.After(expectedExpiry.Add(1*time.Minute)) {
		t.Errorf("expected expiry around %v, got %v", expectedExpiry, actualExpiry)
	}
}
