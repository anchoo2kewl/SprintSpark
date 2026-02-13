package api

import (
	"context"
	"database/sql"
	"fmt"

	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"sprintspark/internal/auth"
)

// SignupRequest represents the signup request payload
type SignupRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	InviteCode string `json:"invite_code"`
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
	Name      string    `json:"name,omitempty"`
	IsAdmin   bool      `json:"is_admin"`
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

	// Validate invite code
	if req.InviteCode == "" {
		respondError(w, http.StatusBadRequest, "invite code is required — you need a referral to create an account", "invite_required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Verify invite code is valid
	var inviteID int64
	var usedAt sql.NullString
	var expiresAt sql.NullString
	err := s.db.QueryRowContext(ctx,
		`SELECT id, used_at, expires_at FROM invites WHERE code = ?`, req.InviteCode,
	).Scan(&inviteID, &usedAt, &expiresAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusBadRequest, "invalid invite code", "invalid_invite")
		return
	}
	if err != nil {
		s.logger.Error("Failed to validate invite code", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to create user", "internal_error")
		return
	}
	if usedAt.Valid {
		respondError(w, http.StatusBadRequest, "this invite has already been used", "invite_used")
		return
	}
	if expiresAt.Valid {
		t, parseErr := time.Parse(time.RFC3339, expiresAt.String)
		if parseErr != nil {
			t, parseErr = time.Parse("2006-01-02 15:04:05-07:00", expiresAt.String)
		}
		if parseErr == nil && time.Now().After(t) {
			respondError(w, http.StatusBadRequest, "this invite has expired", "invite_expired")
			return
		}
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to create user", "internal_error")
		return
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.logger.Error("Failed to begin transaction", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to create user", "internal_error")
		return
	}
	defer tx.Rollback()

	// Create user
	userQuery := `
		INSERT INTO users (email, password_hash)
		VALUES (?, ?)
		RETURNING id, email, name, is_admin, created_at
	`

	var user User
	var name sql.NullString
	err = tx.QueryRowContext(ctx, userQuery, req.Email, hashedPassword).
		Scan(&user.ID, &user.Email, &name, &user.IsAdmin, &user.CreatedAt)

	if name.Valid {
		user.Name = name.String
	}

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			respondError(w, http.StatusConflict, "email already exists", "email_exists")
			return
		}
		s.logger.Error("Failed to create user", zap.Error(err), zap.String("email", req.Email))
		respondError(w, http.StatusInternalServerError, "failed to create user", "internal_error")
		return
	}

	// Create team for the user
	teamName := user.Email + "'s Team"
	if user.Name != "" {
		teamName = user.Name + "'s Team"
	}

	teamQuery := `
		INSERT INTO teams (name, owner_id, created_at, updated_at)
		VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	teamResult, err := tx.ExecContext(ctx, teamQuery, teamName, user.ID)
	if err != nil {
		s.logger.Error("Failed to create team", zap.Error(err), zap.Int64("user_id", user.ID))
		respondError(w, http.StatusInternalServerError, "failed to create team", "internal_error")
		return
	}

	teamID, err := teamResult.LastInsertId()
	if err != nil {
		s.logger.Error("Failed to get team ID", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to create team", "internal_error")
		return
	}

	// Add user to team as owner
	memberQuery := `
		INSERT INTO team_members (team_id, user_id, role, status, joined_at)
		VALUES (?, ?, 'owner', 'active', CURRENT_TIMESTAMP)
	`
	_, err = tx.ExecContext(ctx, memberQuery, teamID, user.ID)
	if err != nil {
		s.logger.Error("Failed to add user to team", zap.Error(err), zap.Int64("user_id", user.ID))
		respondError(w, http.StatusInternalServerError, "failed to add user to team", "internal_error")
		return
	}

	// Mark invite as used
	_, err = tx.ExecContext(ctx,
		`UPDATE invites SET invitee_id = ?, used_at = CURRENT_TIMESTAMP WHERE id = ?`,
		user.ID, inviteID,
	)
	if err != nil {
		s.logger.Error("Failed to mark invite as used", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to create user", "internal_error")
		return
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		s.logger.Error("Failed to commit transaction", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to create user", "internal_error")
		return
	}

	s.logger.Info("User and team created",
		zap.Int64("user_id", user.ID),
		zap.Int64("team_id", teamID),
		zap.String("email", user.Email),
	)

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID, user.Email, s.config.JWTSecret, s.config.JWTExpiry())
	if err != nil {
		s.logger.Error("Failed to generate token", zap.Error(err))
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

	query := `SELECT id, email, name, password_hash, is_admin, created_at FROM users WHERE email = ?`

	var user User
	var passwordHash string
	var name sql.NullString
	err := s.db.QueryRowContext(ctx, query, req.Email).
		Scan(&user.ID, &user.Email, &name, &passwordHash, &user.IsAdmin, &user.CreatedAt)

	if name.Valid {
		user.Name = name.String
	}

	if err != nil {
		if err == sql.ErrNoRows {
			// Log failed login attempt
			respondError(w, http.StatusUnauthorized, "invalid email or password", "invalid_credentials")
			return
		}
		s.logger.Error("Failed to query user", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to authenticate", "internal_error")
		return
	}

	// Verify password
	if err := auth.VerifyPassword(passwordHash, req.Password); err != nil {
		// Log failed login attempt
		go s.logUserActivity(context.Background(), user.ID, "failed_login", getClientIP(r), r.UserAgent())
		respondError(w, http.StatusUnauthorized, "invalid email or password", "invalid_credentials")
		return
	}

	// Log successful login
	go s.logUserActivity(context.Background(), user.ID, "login", getClientIP(r), r.UserAgent())

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID, user.Email, s.config.JWTSecret, s.config.JWTExpiry())
	if err != nil {
		s.logger.Error("Failed to generate token", zap.Error(err))
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

	query := `SELECT id, email, name, is_admin, created_at FROM users WHERE id = ?`

	var user User
	var name sql.NullString
	err := s.db.QueryRowContext(ctx, query, userID).
		Scan(&user.ID, &user.Email, &name, &user.IsAdmin, &user.CreatedAt)

	if name.Valid {
		user.Name = name.String
	}

	if err != nil {
		if err == sql.ErrNoRows {
			respondError(w, http.StatusNotFound, "user not found", "not_found")
			return
		}
		s.logger.Error("Failed to query user", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to get user", "internal_error")
		return
	}

	respondJSON(w, http.StatusOK, user)
}

// UpdateProfileRequest represents the update profile request
type UpdateProfileRequest struct {
	Name string `json:"name"`
}

// HandleUpdateProfile updates the current user's profile
func (s *Server) HandleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated", "unauthorized")
		return
	}

	var req UpdateProfileRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", "invalid_request")
		return
	}

	// Validate name
	if len(req.Name) > 100 {
		respondError(w, http.StatusBadRequest, "name must be 100 characters or less", "validation_error")
		return
	}

	// Update user name
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	query := `UPDATE users SET name = ? WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, req.Name, userID)
	if err != nil {
		s.logger.Error("Failed to update user profile", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to update profile", "internal_error")
		return
	}

	// Get updated user
	userQuery := `SELECT id, email, name, is_admin, created_at FROM users WHERE id = ?`
	var user User
	var name sql.NullString
	err = s.db.QueryRowContext(ctx, userQuery, userID).
		Scan(&user.ID, &user.Email, &name, &user.IsAdmin, &user.CreatedAt)

	if name.Valid {
		user.Name = name.String
	}

	if err != nil {
		s.logger.Error("Failed to query user", zap.Error(err))
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

	// Validate password strength
	if err := validatePasswordStrength(req.Password); err != nil {
		return err
	}

	return nil
}

// validatePasswordStrength ensures password meets security requirements
func validatePasswordStrength(password string) error {
	if password == "" {
		return fmt.Errorf("password is required")
	}

	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	// Check for at least one digit
	hasDigit := false
	for _, ch := range password {
		if ch >= '0' && ch <= '9' {
			hasDigit = true
			break
		}
	}
	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}

	// Check for at least one letter (uppercase or lowercase)
	hasLetter := false
	for _, ch := range password {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') {
			hasLetter = true
			break
		}
	}
	if !hasLetter {
		return fmt.Errorf("password must contain at least one letter")
	}

	return nil
}
