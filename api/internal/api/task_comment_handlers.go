package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type TaskComment struct {
	ID        int64     `json:"id"`
	TaskID    int64     `json:"task_id"`
	UserID    int64     `json:"user_id"`
	UserName  *string   `json:"user_name,omitempty"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateCommentRequest struct {
	Comment string `json:"comment"`
}

// HandleListTaskComments returns all comments for a task
func (s *Server) HandleListTaskComments(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	userID := r.Context().Value(UserIDKey).(int64)
	taskID, err := strconv.ParseInt(chi.URLParam(r, "taskId"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid task ID", "invalid_input")
		return
	}

	// Get task's project ID and verify user has access
	var projectID int64
	projectQuery := `SELECT project_id FROM tasks WHERE id = ?`
	if err := s.db.QueryRowContext(ctx, projectQuery, taskID).Scan(&projectID); err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "task not found", "not_found")
		return
	} else if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get task project", "internal_error")
		return
	}

	// Verify user has access to the project
	hasAccess, err := s.checkProjectAccess(ctx, userID, projectID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to verify project access", "internal_error")
		return
	}
	if !hasAccess {
		respondError(w, http.StatusForbidden, "access denied", "forbidden")
		return
	}

	// Fetch comments with user names
	query := `
		SELECT c.id, c.task_id, c.user_id, u.name as user_name, c.comment, c.created_at, c.updated_at
		FROM task_comments c
		LEFT JOIN users u ON c.user_id = u.id
		WHERE c.task_id = ?
		ORDER BY c.created_at ASC
	`

	rows, err := s.db.QueryContext(ctx, query, taskID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch comments", "internal_error")
		return
	}
	defer rows.Close()

	comments := []TaskComment{}
	for rows.Next() {
		var c TaskComment
		if err := rows.Scan(&c.ID, &c.TaskID, &c.UserID, &c.UserName, &c.Comment, &c.CreatedAt, &c.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to scan comment", "internal_error")
			return
		}
		comments = append(comments, c)
	}

	if err := rows.Err(); err != nil {
		respondError(w, http.StatusInternalServerError, "error iterating comments", "internal_error")
		return
	}

	respondJSON(w, http.StatusOK, comments)
}

// HandleCreateTaskComment creates a new comment on a task
func (s *Server) HandleCreateTaskComment(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	userID := r.Context().Value(UserIDKey).(int64)
	taskID, err := strconv.ParseInt(chi.URLParam(r, "taskId"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid task ID", "invalid_input")
		return
	}

	// Get task's project ID and verify user has access
	var projectID int64
	projectQuery := `SELECT project_id FROM tasks WHERE id = ?`
	if err := s.db.QueryRowContext(ctx, projectQuery, taskID).Scan(&projectID); err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "task not found", "not_found")
		return
	} else if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get task project", "internal_error")
		return
	}

	// Verify user has access to the project
	hasAccess, err := s.checkProjectAccess(ctx, userID, projectID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to verify project access", "internal_error")
		return
	}
	if !hasAccess {
		respondError(w, http.StatusForbidden, "access denied", "forbidden")
		return
	}

	var req CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", "invalid_input")
		return
	}

	// Validation
	if req.Comment == "" {
		respondError(w, http.StatusBadRequest, "comment is required", "invalid_input")
		return
	}
	if len(req.Comment) > 5000 {
		respondError(w, http.StatusBadRequest, "comment is too long (max 5000 characters)", "invalid_input")
		return
	}

	query := `
		INSERT INTO task_comments (task_id, user_id, comment, created_at, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	result, err := s.db.ExecContext(ctx, query, taskID, userID, req.Comment)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create comment", "internal_error")
		return
	}

	commentID, err := result.LastInsertId()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get comment ID", "internal_error")
		return
	}

	// Fetch the created comment
	var c TaskComment
	fetchQuery := `
		SELECT c.id, c.task_id, c.user_id, u.name as user_name, c.comment, c.created_at, c.updated_at
		FROM task_comments c
		LEFT JOIN users u ON c.user_id = u.id
		WHERE c.id = ?
	`
	err = s.db.QueryRowContext(ctx, fetchQuery, commentID).Scan(
		&c.ID, &c.TaskID, &c.UserID, &c.UserName, &c.Comment, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch created comment", "internal_error")
		return
	}

	respondJSON(w, http.StatusCreated, c)
}
