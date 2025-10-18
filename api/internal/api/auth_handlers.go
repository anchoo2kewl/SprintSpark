package api

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"sprintspark/internal/auth"
)

// SignupRequest represents the signup request payload
type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// User represents a user
type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// HandleSignup creates a new user account
func (s *Server) HandleSignup(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", "invalid_request")
		return
	}

	// Validate input
	if err := validateSignupRequest(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error(), "validation_error")
		return
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		respondError(w, http.StatusInternalServerError, "failed to create user", "internal_error")
		return
	}

	// Create user
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO users (email, password_hash)
		VALUES (?, ?)
		RETURNING id, email, created_at
	`

	var user User
	err = s.db.QueryRowContext(ctx, query, req.Email, hashedPassword).
		Scan(&user.ID, &user.Email, &user.CreatedAt)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			respondError(w, http.StatusConflict, "email already exists", "email_exists")
			return
		}
		log.Printf("Failed to create user: %v", err)
		respondError(w, http.StatusInternalServerError, "failed to create user", "internal_error")
		return
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID, user.Email, s.config.JWTSecret, s.config.JWTExpiry())
	if err != nil {
		log.Printf("Failed to generate token: %v", err)
		respondError(w, http.StatusInternalServerError, "failed to generate token", "internal_error")
		return
	}

	respondJSON(w, http.StatusCreated, AuthResponse{
		Token: token,
		User:  user,
	})
}

// HandleLogin authenticates a user and returns a JWT token
func (s *Server) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", "invalid_request")
		return
	}

	// Validate input
	if req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "email and password are required", "validation_error")
		return
	}

	// Get user from database
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	query := `SELECT id, email, password_hash, created_at FROM users WHERE email = ?`

	var user User
	var passwordHash string
	err := s.db.QueryRowContext(ctx, query, req.Email).
		Scan(&user.ID, &user.Email, &passwordHash, &user.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			respondError(w, http.StatusUnauthorized, "invalid email or password", "invalid_credentials")
			return
		}
		log.Printf("Failed to query user: %v", err)
		respondError(w, http.StatusInternalServerError, "failed to authenticate", "internal_error")
		return
	}

	// Verify password
	if err := auth.VerifyPassword(passwordHash, req.Password); err != nil {
		respondError(w, http.StatusUnauthorized, "invalid email or password", "invalid_credentials")
		return
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID, user.Email, s.config.JWTSecret, s.config.JWTExpiry())
	if err != nil {
		log.Printf("Failed to generate token: %v", err)
		respondError(w, http.StatusInternalServerError, "failed to generate token", "internal_error")
		return
	}

	respondJSON(w, http.StatusOK, AuthResponse{
		Token: token,
		User:  user,
	})
}

// HandleMe returns the current authenticated user
func (s *Server) HandleMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated", "unauthorized")
		return
	}

	// Get user from database
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	query := `SELECT id, email, created_at FROM users WHERE id = ?`

	var user User
	err := s.db.QueryRowContext(ctx, query, userID).
		Scan(&user.ID, &user.Email, &user.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			respondError(w, http.StatusNotFound, "user not found", "not_found")
			return
		}
		log.Printf("Failed to query user: %v", err)
		respondError(w, http.StatusInternalServerError, "failed to get user", "internal_error")
		return
	}

	respondJSON(w, http.StatusOK, user)
}

// validateSignupRequest validates the signup request
func validateSignupRequest(req SignupRequest) error {
	// Validate email
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if !strings.Contains(req.Email, "@") || !strings.Contains(req.Email, ".") {
		return fmt.Errorf("invalid email format")
	}

	// Validate password
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}
	if len(req.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	return nil
}
