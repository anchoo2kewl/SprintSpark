package api

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Invite represents an invite record
type Invite struct {
	ID          int64   `json:"id"`
	Code        string  `json:"code"`
	InviterID   int64   `json:"inviter_id"`
	InviterName *string `json:"inviter_name,omitempty"`
	InviteeID   *int64  `json:"invitee_id,omitempty"`
	InviteeName *string `json:"invitee_name,omitempty"`
	UsedAt      *string `json:"used_at,omitempty"`
	ExpiresAt   *string `json:"expires_at,omitempty"`
	CreatedAt   string  `json:"created_at"`
}

// InviteStatus is returned for invite validation
type InviteStatus struct {
	Valid       bool   `json:"valid"`
	InviterName string `json:"inviter_name,omitempty"`
	Message     string `json:"message,omitempty"`
}

// generateInviteCode creates a random URL-safe invite code
func generateInviteCode() (string, error) {
	b := make([]byte, 18)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b), nil
}

// HandleListInvites returns the current user's invites
func (s *Server) HandleListInvites(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated", "unauthorized")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	query := `
		SELECT i.id, i.code, i.inviter_id, i.invitee_id, i.used_at, i.expires_at, i.created_at,
			   u.name as invitee_name, u.email as invitee_email
		FROM invites i
		LEFT JOIN users u ON i.invitee_id = u.id
		WHERE i.inviter_id = ?
		ORDER BY i.created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		s.logger.Error("Failed to query invites", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to list invites", "internal_error")
		return
	}
	defer rows.Close()

	type InviteWithDetails struct {
		ID          int64   `json:"id"`
		Code        string  `json:"code"`
		InviterID   int64   `json:"inviter_id"`
		InviteeID   *int64  `json:"invitee_id,omitempty"`
		InviteeName *string `json:"invitee_name,omitempty"`
		UsedAt      *string `json:"used_at,omitempty"`
		ExpiresAt   *string `json:"expires_at,omitempty"`
		CreatedAt   string  `json:"created_at"`
	}

	invites := []InviteWithDetails{}
	for rows.Next() {
		var inv InviteWithDetails
		var inviteeName, inviteeEmail sql.NullString
		err := rows.Scan(&inv.ID, &inv.Code, &inv.InviterID, &inv.InviteeID, &inv.UsedAt, &inv.ExpiresAt, &inv.CreatedAt, &inviteeName, &inviteeEmail)
		if err != nil {
			s.logger.Error("Failed to scan invite row", zap.Error(err))
			continue
		}
		if inviteeName.Valid && inviteeName.String != "" {
			inv.InviteeName = &inviteeName.String
		} else if inviteeEmail.Valid {
			inv.InviteeName = &inviteeEmail.String
		}
		invites = append(invites, inv)
	}

	// Also get the user's invite count
	var inviteCount int
	var isAdmin bool
	err = s.db.QueryRowContext(ctx, `SELECT invite_count, is_admin FROM users WHERE id = ?`, userID).Scan(&inviteCount, &isAdmin)
	if err != nil {
		s.logger.Error("Failed to get invite count", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to get invite count", "internal_error")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"invites":      invites,
		"invite_count": inviteCount,
		"is_admin":     isAdmin,
	})
}

// HandleCreateInvite creates a new invite code
func (s *Server) HandleCreateInvite(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated", "unauthorized")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Check invite count (admins have unlimited)
	var inviteCount int
	var isAdmin bool
	err := s.db.QueryRowContext(ctx, `SELECT invite_count, is_admin FROM users WHERE id = ?`, userID).Scan(&inviteCount, &isAdmin)
	if err != nil {
		s.logger.Error("Failed to get invite count", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to create invite", "internal_error")
		return
	}

	if !isAdmin && inviteCount <= 0 {
		respondError(w, http.StatusForbidden, "no invites remaining", "no_invites")
		return
	}

	// Generate invite code
	code, err := generateInviteCode()
	if err != nil {
		s.logger.Error("Failed to generate invite code", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to create invite", "internal_error")
		return
	}

	// Set expiry to 7 days
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.logger.Error("Failed to begin transaction", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to create invite", "internal_error")
		return
	}
	defer tx.Rollback()

	// Insert invite
	_, err = tx.ExecContext(ctx,
		`INSERT INTO invites (code, inviter_id, expires_at) VALUES (?, ?, ?)`,
		code, userID, expiresAt,
	)
	if err != nil {
		s.logger.Error("Failed to insert invite", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to create invite", "internal_error")
		return
	}

	// Decrement invite count (only for non-admins)
	if !isAdmin {
		_, err = tx.ExecContext(ctx,
			`UPDATE users SET invite_count = invite_count - 1 WHERE id = ? AND invite_count > 0`,
			userID,
		)
		if err != nil {
			s.logger.Error("Failed to decrement invite count", zap.Error(err))
			respondError(w, http.StatusInternalServerError, "failed to create invite", "internal_error")
			return
		}
	}

	if err := tx.Commit(); err != nil {
		s.logger.Error("Failed to commit transaction", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to create invite", "internal_error")
		return
	}

	s.logger.Info("Invite created", zap.Int64("user_id", userID), zap.String("code", code[:8]+"..."))

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"code":       code,
		"expires_at": expiresAt.Format(time.RFC3339),
	})
}

// HandleValidateInvite checks if an invite code is valid (public endpoint)
func (s *Server) HandleValidateInvite(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		respondJSON(w, http.StatusOK, InviteStatus{Valid: false, Message: "invite code is required"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var inviterName sql.NullString
	var usedAt sql.NullString
	var expiresAt sql.NullString
	err := s.db.QueryRowContext(ctx,
		`SELECT u.name, u.email, i.used_at, i.expires_at
		 FROM invites i JOIN users u ON i.inviter_id = u.id
		 WHERE i.code = ?`, code,
	).Scan(&inviterName, new(string), &usedAt, &expiresAt)

	if err == sql.ErrNoRows {
		respondJSON(w, http.StatusOK, InviteStatus{Valid: false, Message: "invalid invite code"})
		return
	}
	if err != nil {
		s.logger.Error("Failed to validate invite", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to validate invite", "internal_error")
		return
	}

	if usedAt.Valid {
		respondJSON(w, http.StatusOK, InviteStatus{Valid: false, Message: "this invite has already been used"})
		return
	}

	if expiresAt.Valid {
		t, err := time.Parse(time.RFC3339, expiresAt.String)
		if err == nil && time.Now().After(t) {
			respondJSON(w, http.StatusOK, InviteStatus{Valid: false, Message: "this invite has expired"})
			return
		}
		// Also try parsing as the SQLite default format
		t2, err2 := time.Parse("2006-01-02 15:04:05-07:00", expiresAt.String)
		if err2 == nil && time.Now().After(t2) {
			respondJSON(w, http.StatusOK, InviteStatus{Valid: false, Message: "this invite has expired"})
			return
		}
	}

	name := ""
	if inviterName.Valid && inviterName.String != "" {
		name = inviterName.String
	}

	respondJSON(w, http.StatusOK, InviteStatus{Valid: true, InviterName: name})
}

// HandleAdminBoostInvites allows admins to set a user's invite count
func (s *Server) HandleAdminBoostInvites(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated", "unauthorized")
		return
	}

	if !s.isAdmin(r.Context(), userID) {
		respondError(w, http.StatusForbidden, "admin access required", "forbidden")
		return
	}

	targetUserIDStr := r.PathValue("id")
	if targetUserIDStr == "" {
		respondError(w, http.StatusBadRequest, "user id required", "validation_error")
		return
	}

	var targetUserID int64
	if _, err := fmt.Sscanf(targetUserIDStr, "%d", &targetUserID); err != nil {
		respondError(w, http.StatusBadRequest, "invalid user id", "validation_error")
		return
	}

	var req struct {
		InviteCount int `json:"invite_count"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", "invalid_request")
		return
	}

	if req.InviteCount < 0 {
		respondError(w, http.StatusBadRequest, "invite count must be non-negative", "validation_error")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	result, err := s.db.ExecContext(ctx, `UPDATE users SET invite_count = ? WHERE id = ?`, req.InviteCount, targetUserID)
	if err != nil {
		s.logger.Error("Failed to update invite count", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to update invite count", "internal_error")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		respondError(w, http.StatusNotFound, "user not found", "not_found")
		return
	}

	s.logger.Info("Admin boosted invites",
		zap.Int64("admin_id", userID),
		zap.Int64("target_user_id", targetUserID),
		zap.Int("invite_count", req.InviteCount),
	)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"id":           targetUserID,
		"invite_count": req.InviteCount,
	})
}
