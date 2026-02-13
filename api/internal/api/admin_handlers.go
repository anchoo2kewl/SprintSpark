package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// UserWithStats represents a user with admin and activity stats
type UserWithStats struct {
	ID             int64     `json:"id"`
	Email          string    `json:"email"`
	IsAdmin        bool      `json:"is_admin"`
	CreatedAt      time.Time `json:"created_at"`
	LoginCount     int       `json:"login_count"`
	LastLoginAt    *string   `json:"last_login_at"`
	LastLoginIP    *string   `json:"last_login_ip"`
	FailedAttempts int       `json:"failed_attempts"`
	InviteCount    int       `json:"invite_count"`
}

// UserActivity represents a user activity log entry
type UserActivity struct {
	ID           int64   `json:"id"`
	UserID       int64   `json:"user_id"`
	ActivityType string  `json:"activity_type"` // 'login', 'logout', 'failed_login'
	IPAddress    *string `json:"ip_address"`
	UserAgent    *string `json:"user_agent"`
	CreatedAt    string  `json:"created_at"`
}

// HandleGetUsers returns all users (admin only)
func (s *Server) HandleGetUsers(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated", "unauthorized")
		return
	}

	// Check if user is admin
	if !s.isAdmin(r.Context(), userID) {
		respondError(w, http.StatusForbidden, "admin access required", "forbidden")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	query := `
		SELECT
			u.id,
			u.email,
			u.is_admin,
			u.created_at,
			COALESCE(login_stats.login_count, 0) as login_count,
			login_stats.last_login_at,
			login_stats.last_login_ip,
			COALESCE(failed_stats.failed_count, 0) as failed_attempts,
			u.invite_count
		FROM users u
		LEFT JOIN (
			SELECT user_id, COUNT(*) as login_count, MAX(created_at) as last_login_at,
				(SELECT ip_address FROM user_activity WHERE user_id = ua.user_id AND activity_type = 'login' ORDER BY created_at DESC LIMIT 1) as last_login_ip
			FROM user_activity ua
			WHERE activity_type = 'login'
			GROUP BY user_id
		) login_stats ON u.id = login_stats.user_id
		LEFT JOIN (
			SELECT user_id, COUNT(*) as failed_count
			FROM user_activity
			WHERE activity_type = 'failed_login'
			GROUP BY user_id
		) failed_stats ON u.id = failed_stats.user_id
		ORDER BY u.created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		s.logger.Error("Failed to query users", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to get users", "internal_error")
		return
	}
	defer rows.Close()

	users := []UserWithStats{}
	for rows.Next() {
		var u UserWithStats
		err := rows.Scan(
			&u.ID,
			&u.Email,
			&u.IsAdmin,
			&u.CreatedAt,
			&u.LoginCount,
			&u.LastLoginAt,
			&u.LastLoginIP,
			&u.FailedAttempts,
			&u.InviteCount,
		)
		if err != nil {
			s.logger.Error("Failed to scan user row", zap.Error(err))
			continue
		}
		users = append(users, u)
	}

	respondJSON(w, http.StatusOK, users)
}

// HandleGetUserActivity returns activity log for a specific user (admin only)
func (s *Server) HandleGetUserActivity(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated", "unauthorized")
		return
	}

	// Check if user is admin
	if !s.isAdmin(r.Context(), userID) {
		respondError(w, http.StatusForbidden, "admin access required", "forbidden")
		return
	}

	// Get target user ID from path
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

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, user_id, activity_type, ip_address, user_agent, created_at
		FROM user_activity
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT 100
	`

	rows, err := s.db.QueryContext(ctx, query, targetUserID)
	if err != nil {
		s.logger.Error("Failed to query user activity", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to get activity", "internal_error")
		return
	}
	defer rows.Close()

	activities := []UserActivity{}
	for rows.Next() {
		var a UserActivity
		err := rows.Scan(
			&a.ID,
			&a.UserID,
			&a.ActivityType,
			&a.IPAddress,
			&a.UserAgent,
			&a.CreatedAt,
		)
		if err != nil {
			s.logger.Error("Failed to scan activity row", zap.Error(err))
			continue
		}
		activities = append(activities, a)
	}

	respondJSON(w, http.StatusOK, activities)
}

// HandleUpdateUserAdmin updates admin status of a user (admin only)
func (s *Server) HandleUpdateUserAdmin(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "user not authenticated", "unauthorized")
		return
	}

	// Check if user is admin
	if !s.isAdmin(r.Context(), userID) {
		respondError(w, http.StatusForbidden, "admin access required", "forbidden")
		return
	}

	// Get target user ID from path
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
		IsAdmin bool `json:"is_admin"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", "invalid_request")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	query := `UPDATE users SET is_admin = ? WHERE id = ?`
	result, err := s.db.ExecContext(ctx, query, req.IsAdmin, targetUserID)
	if err != nil {
		s.logger.Error("Failed to update user admin status", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to update user", "internal_error")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		s.logger.Error("Failed to get rows affected", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to update user", "internal_error")
		return
	}

	if rowsAffected == 0 {
		respondError(w, http.StatusNotFound, "user not found", "not_found")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"id":       targetUserID,
		"is_admin": req.IsAdmin,
	})
}

// isAdmin checks if a user is an admin
func (s *Server) isAdmin(ctx context.Context, userID int64) bool {
	var isAdmin bool
	query := `SELECT is_admin FROM users WHERE id = ?`
	err := s.db.QueryRowContext(ctx, query, userID).Scan(&isAdmin)
	if err != nil {
		s.logger.Error("Failed to check admin status", zap.Error(err))
		return false
	}
	return isAdmin
}

// logUserActivity logs a user activity event
func (s *Server) logUserActivity(ctx context.Context, userID int64, activityType, ipAddress, userAgent string) error {
	query := `INSERT INTO user_activity (user_id, activity_type, ip_address, user_agent) VALUES (?, ?, ?, ?)`

	// Use NULL for empty strings
	var ip, ua interface{}
	if ipAddress != "" {
		ip = ipAddress
	}
	if userAgent != "" {
		ua = userAgent
	}

	_, err := s.db.ExecContext(ctx, query, userID, activityType, ip, ua)
	if err != nil {
		s.logger.Error("Failed to log user activity", zap.Error(err))
	}
	return err
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies/load balancers)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// XFF can contain multiple IPs, get the first one
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}
