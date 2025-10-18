package auth

import (
	"testing"
	"time"
)

func TestGenerateToken(t *testing.T) {
	tests := []struct {
		name      string
		userID    int64
		email     string
		secret    string
		expiry    time.Duration
		wantError bool
	}{
		{
			name:      "valid token",
			userID:    1,
			email:     "user@example.com",
			secret:    "test-secret",
			expiry:    24 * time.Hour,
			wantError: false,
		},
		{
			name:      "short expiry",
			userID:    1,
			email:     "user@example.com",
			secret:    "test-secret",
			expiry:    1 * time.Minute,
			wantError: false,
		},
		{
			name:      "long expiry",
			userID:    1,
			email:     "user@example.com",
			secret:    "test-secret",
			expiry:    720 * time.Hour, // 30 days
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateToken(tt.userID, tt.email, tt.secret, tt.expiry)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if token == "" {
				t.Error("Expected non-empty token")
			}

			// Token should have 3 parts separated by dots
			parts := splitString(token, '.')
			if len(parts) != 3 {
				t.Errorf("Expected 3 token parts, got %d", len(parts))
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	secret := "test-secret"
	userID := int64(1)
	email := "user@example.com"

	validToken, err := GenerateToken(userID, email, secret, 24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	expiredToken, err := GenerateToken(userID, email, secret, -1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate expired token: %v", err)
	}

	tests := []struct {
		name      string
		token     string
		secret    string
		wantError bool
		checkID   bool
		checkEmail bool
	}{
		{
			name:       "valid token",
			token:      validToken,
			secret:     secret,
			wantError:  false,
			checkID:    true,
			checkEmail: true,
		},
		{
			name:      "expired token",
			token:     expiredToken,
			secret:    secret,
			wantError: true,
		},
		{
			name:      "wrong secret",
			token:     validToken,
			secret:    "wrong-secret",
			wantError: true,
		},
		{
			name:      "invalid token format",
			token:     "invalid.token",
			secret:    secret,
			wantError: true,
		},
		{
			name:      "empty token",
			token:     "",
			secret:    secret,
			wantError: true,
		},
		{
			name:      "malformed token",
			token:     "header.payload",
			secret:    secret,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ValidateToken(tt.token, tt.secret)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if tt.checkID && claims.UserID != userID {
				t.Errorf("UserID mismatch: got %d, want %d", claims.UserID, userID)
			}

			if tt.checkEmail && claims.Email != email {
				t.Errorf("Email mismatch: got %s, want %s", claims.Email, email)
			}
		})
	}
}

func TestTokenExpiry(t *testing.T) {
	secret := "test-secret"
	userID := int64(1)
	email := "user@example.com"

	// Create token that expires in 1 second
	token, err := GenerateToken(userID, email, secret, 1*time.Second)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Should be valid immediately
	claims, err := ValidateToken(token, secret)
	if err != nil {
		t.Errorf("Token should be valid immediately: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("UserID mismatch: got %d, want %d", claims.UserID, userID)
	}

	// Wait for token to expire
	time.Sleep(2 * time.Second)

	// Should be expired now
	_, err = ValidateToken(token, secret)
	if err == nil {
		t.Error("Expected error for expired token, got nil")
	}
}

func TestTokenClaims(t *testing.T) {
	secret := "test-secret"
	userID := int64(42)
	email := "test@example.com"
	expiry := 1 * time.Hour

	token, err := GenerateToken(userID, email, secret, expiry)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	claims, err := ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	// Check custom claims
	if claims.UserID != userID {
		t.Errorf("UserID mismatch: got %d, want %d", claims.UserID, userID)
	}

	if claims.Email != email {
		t.Errorf("Email mismatch: got %s, want %s", claims.Email, email)
	}

	// Check standard claims
	now := time.Now()

	if claims.IssuedAt == nil {
		t.Error("IssuedAt should be set")
	} else if claims.IssuedAt.Time.After(now) {
		t.Error("IssuedAt should not be in the future")
	}

	if claims.ExpiresAt == nil {
		t.Error("ExpiresAt should be set")
	} else if claims.ExpiresAt.Time.Before(now) {
		t.Error("ExpiresAt should be in the future")
	}

	if claims.NotBefore == nil {
		t.Error("NotBefore should be set")
	} else if claims.NotBefore.Time.After(now) {
		t.Error("NotBefore should not be in the future")
	}
}

// splitString splits a string by a delimiter
func splitString(s string, delim rune) []string {
	var parts []string
	var current string

	for _, ch := range s {
		if ch == delim {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(ch)
		}
	}

	if current != "" || len(s) > 0 {
		parts = append(parts, current)
	}

	return parts
}
