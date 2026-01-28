package db

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"
)

// APIKey represents an API key for user authentication
type APIKey struct {
	ID         int64      `json:"id"`
	UserID     int64      `json:"user_id"`
	Name       string     `json:"name"`
	KeyPrefix  string     `json:"key_prefix"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

// APIKeyWithSecret includes the full key (only returned on creation)
type APIKeyWithSecret struct {
	APIKey
	Key string `json:"key"`
}

// GenerateAPIKey creates a new API key with a cryptographically secure random value
func GenerateAPIKey() (key, keyHash, prefix string, err error) {
	// Generate 32 random bytes
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Encode to base64
	key = base64.URLEncoding.EncodeToString(b)

	// Create hash for storage
	hash := sha256.Sum256([]byte(key))
	keyHash = base64.URLEncoding.EncodeToString(hash[:])

	// Prefix for display (first 8 chars)
	prefix = key[:8]

	return key, keyHash, prefix, nil
}

// HashAPIKey creates a hash of an API key for comparison
func HashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return base64.URLEncoding.EncodeToString(hash[:])
}

// CreateAPIKey creates a new API key for a user
func (db *DB) CreateAPIKey(ctx context.Context, userID int64, name string, expiresAt *time.Time) (*APIKeyWithSecret, error) {
	// Generate API key
	key, keyHash, prefix, err := GenerateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	// Insert into database
	query := `
		INSERT INTO api_keys (user_id, name, key_hash, key_prefix, expires_at)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := db.ExecContext(ctx, query, userID, name, keyHash, prefix, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to insert API key: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get API key ID: %w", err)
	}

	return &APIKeyWithSecret{
		APIKey: APIKey{
			ID:        id,
			UserID:    userID,
			Name:      name,
			KeyPrefix: prefix,
			CreatedAt: time.Now(),
			ExpiresAt: expiresAt,
		},
		Key: key,
	}, nil
}

// GetAPIKeysByUserID retrieves all API keys for a user
func (db *DB) GetAPIKeysByUserID(ctx context.Context, userID int64) ([]APIKey, error) {
	query := `
		SELECT id, user_id, name, key_prefix, last_used_at, created_at, expires_at
		FROM api_keys
		WHERE user_id = ?
		ORDER BY created_at DESC
	`

	rows, err := db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query API keys: %w", err)
	}
	defer rows.Close()

	var keys []APIKey
	for rows.Next() {
		var key APIKey
		err := rows.Scan(
			&key.ID,
			&key.UserID,
			&key.Name,
			&key.KeyPrefix,
			&key.LastUsedAt,
			&key.CreatedAt,
			&key.ExpiresAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan API key: %w", err)
		}
		keys = append(keys, key)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating API keys: %w", err)
	}

	return keys, nil
}

// ValidateAPIKey checks if an API key is valid and returns the user ID
func (db *DB) ValidateAPIKey(ctx context.Context, key string) (int64, error) {
	keyHash := HashAPIKey(key)

	query := `
		SELECT user_id, expires_at
		FROM api_keys
		WHERE key_hash = ?
	`

	var userID int64
	var expiresAt *time.Time
	err := db.QueryRowContext(ctx, query, keyHash).Scan(&userID, &expiresAt)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("invalid API key")
	}
	if err != nil {
		return 0, fmt.Errorf("failed to validate API key: %w", err)
	}

	// Check expiration
	if expiresAt != nil && expiresAt.Before(time.Now()) {
		return 0, fmt.Errorf("API key expired")
	}

	// Update last used timestamp
	updateQuery := `UPDATE api_keys SET last_used_at = ? WHERE key_hash = ?`
	_, err = db.ExecContext(ctx, updateQuery, time.Now(), keyHash)
	if err != nil {
		// Log but don't fail on update error
		fmt.Printf("failed to update API key last_used_at: %v\n", err)
	}

	return userID, nil
}

// DeleteAPIKey removes an API key
func (db *DB) DeleteAPIKey(ctx context.Context, keyID, userID int64) error {
	query := `DELETE FROM api_keys WHERE id = ? AND user_id = ?`

	result, err := db.ExecContext(ctx, query, keyID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete API key: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("API key not found or access denied")
	}

	return nil
}

// GetUserByAPIKey retrieves user info using an API key
func (db *DB) GetUserByAPIKey(ctx context.Context, key string) (int64, string, error) {
	userID, err := db.ValidateAPIKey(ctx, key)
	if err != nil {
		return 0, "", err
	}

	// Get user email
	query := `SELECT email FROM users WHERE id = ?`
	var email string
	err = db.QueryRowContext(ctx, query, userID).Scan(&email)
	if err == sql.ErrNoRows {
		return 0, "", fmt.Errorf("user not found")
	}
	if err != nil {
		return 0, "", fmt.Errorf("failed to get user: %w", err)
	}

	return userID, email, nil
}
