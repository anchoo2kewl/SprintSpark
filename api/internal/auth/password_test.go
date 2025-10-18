package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name      string
		password  string
		wantError bool
	}{
		{
			name:      "valid password",
			password:  "password123",
			wantError: false,
		},
		{
			name:      "long password",
			password:  "this-is-a-very-long-password-with-special-chars-123!@#",
			wantError: false,
		},
		{
			name:      "empty password",
			password:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if hash == "" {
				t.Error("Expected non-empty hash")
			}

			// Verify hash starts with bcrypt prefix and cost 12
			if len(hash) < 4 || hash[:4] != "$2a$" {
				t.Errorf("Hash doesn't start with bcrypt prefix: %s", hash[:4])
			}

			// Check cost is 12
			if len(hash) >= 7 && hash[4:6] != "12" {
				t.Errorf("Expected cost 12, got: %s", hash[4:6])
			}
		})
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "correctPassword123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	tests := []struct {
		name      string
		hash      string
		password  string
		wantError bool
	}{
		{
			name:      "correct password",
			hash:      hash,
			password:  password,
			wantError: false,
		},
		{
			name:      "wrong password",
			hash:      hash,
			password:  "wrongPassword",
			wantError: true,
		},
		{
			name:      "empty password",
			hash:      hash,
			password:  "",
			wantError: true,
		},
		{
			name:      "invalid hash",
			hash:      "invalid-hash",
			password:  password,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := VerifyPassword(tt.hash, tt.password)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestHashPasswordUniqueness(t *testing.T) {
	password := "samePassword123"

	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Hashes should be different due to different salts
	if hash1 == hash2 {
		t.Error("Expected different hashes for same password (different salts)")
	}

	// But both should verify correctly
	if err := VerifyPassword(hash1, password); err != nil {
		t.Errorf("Hash1 verification failed: %v", err)
	}

	if err := VerifyPassword(hash2, password); err != nil {
		t.Errorf("Hash2 verification failed: %v", err)
	}
}
