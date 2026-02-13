package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

var (
	// BcryptCost is the cost factor for bcrypt hashing (12 = ~250ms per hash)
	// Tests can lower this via SetBcryptCost for speed.
	BcryptCost = 12
)

// SetBcryptCost overrides the bcrypt cost (use bcrypt.MinCost in tests for speed)
func SetBcryptCost(cost int) {
	BcryptCost = cost
}

// HashPassword generates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hash), nil
}

// VerifyPassword checks if the provided password matches the hash
func VerifyPassword(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return fmt.Errorf("invalid password")
	}
	return nil
}
