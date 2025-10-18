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

type Task struct {
	ID          int64     `json:"id"`
	ProjectID   int64     `json:"project_id"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateTaskRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty"`
}

type UpdateTaskRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty"`
}

// HandleListTasks returns all tasks for a project
func (s *Server) HandleListTasks(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	userID := r.Context().Value(UserIDKey).(int64)
	projectID, err := strconv.ParseInt(chi.URLParam(r, "projectId"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project ID", "invalid_input")
		return
	}

	// Verify user owns the project
	var ownerID int64
	checkQuery := `SELECT owner_id FROM projects WHERE id = ?`
	if err := s.db.QueryRowContext(ctx, checkQuery, projectID).Scan(&ownerID); err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "project not found", "not_found")
		return
	} else if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to verify project ownership", "internal_error")
		return
	}

	if ownerID != userID {
		respondError(w, http.StatusForbidden, "access denied", "forbidden")
		return
	}

	query := `
		SELECT id, project_id, title, description, status, created_at, updated_at
		FROM tasks
		WHERE project_id = ?
		ORDER BY created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, projectID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch tasks", "internal_error")
		return
	}
	defer rows.Close()

	tasks := []Task{}
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.ProjectID, &t.Title, &t.Description, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to scan task", "internal_error")
			return
		}
		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		respondError(w, http.StatusInternalServerError, "error iterating tasks", "internal_error")
		return
	}

	respondJSON(w, http.StatusOK, tasks)
}

// HandleCreateTask creates a new task
func (s *Server) HandleCreateTask(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	userID := r.Context().Value(UserIDKey).(int64)
	projectID, err := strconv.ParseInt(chi.URLParam(r, "projectId"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project ID", "invalid_input")
		return
	}

	// Verify user owns the project
	var ownerID int64
	checkQuery := `SELECT owner_id FROM projects WHERE id = ?`
	if err := s.db.QueryRowContext(ctx, checkQuery, projectID).Scan(&ownerID); err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "project not found", "not_found")
		return
	} else if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to verify project ownership", "internal_error")
		return
	}

	if ownerID != userID {
		respondError(w, http.StatusForbidden, "access denied", "forbidden")
		return
	}

	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", "invalid_input")
		return
	}

	// Validation
	if req.Title == "" {
		respondError(w, http.StatusBadRequest, "task title is required", "invalid_input")
		return
	}
	if len(req.Title) > 255 {
		respondError(w, http.StatusBadRequest, "task title is too long (max 255 characters)", "invalid_input")
		return
	}

	// Default status to 'todo' if not provided
	status := "todo"
	if req.Status != nil {
		status = *req.Status
	}

	// Validate status
	if status != "todo" && status != "in_progress" && status != "done" {
		respondError(w, http.StatusBadRequest, "invalid status (must be: todo, in_progress, or done)", "invalid_input")
		return
	}

	query := `
		INSERT INTO tasks (project_id, title, description, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	result, err := s.db.ExecContext(ctx, query, projectID, req.Title, req.Description, status)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create task", "internal_error")
		return
	}

	taskID, err := result.LastInsertId()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get task ID", "internal_error")
		return
	}

	// Fetch the created task
	var t Task
	fetchQuery := `
		SELECT id, project_id, title, description, status, created_at, updated_at
		FROM tasks
		WHERE id = ?
	`
	err = s.db.QueryRowContext(ctx, fetchQuery, taskID).Scan(
		&t.ID, &t.ProjectID, &t.Title, &t.Description, &t.Status, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch created task", "internal_error")
		return
	}

	respondJSON(w, http.StatusCreated, t)
}

// HandleUpdateTask updates an existing task
func (s *Server) HandleUpdateTask(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	userID := r.Context().Value(UserIDKey).(int64)
	taskID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid task ID", "invalid_input")
		return
	}

	var req UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", "invalid_input")
		return
	}

	// Verify user owns the project that contains this task
	checkQuery := `
		SELECT p.owner_id FROM projects p
		JOIN tasks t ON t.project_id = p.id
		WHERE t.id = ?
	`
	var ownerID int64
	if err := s.db.QueryRowContext(ctx, checkQuery, taskID).Scan(&ownerID); err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "task not found", "not_found")
		return
	} else if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to verify task ownership", "internal_error")
		return
	}

	if ownerID != userID {
		respondError(w, http.StatusForbidden, "access denied", "forbidden")
		return
	}

	// Build update query dynamically
	query := "UPDATE tasks SET updated_at = CURRENT_TIMESTAMP"
	args := []interface{}{}

	if req.Title != nil {
		if *req.Title == "" {
			respondError(w, http.StatusBadRequest, "task title cannot be empty", "invalid_input")
			return
		}
		if len(*req.Title) > 255 {
			respondError(w, http.StatusBadRequest, "task title is too long (max 255 characters)", "invalid_input")
			return
		}
		query += ", title = ?"
		args = append(args, *req.Title)
	}

	if req.Description != nil {
		query += ", description = ?"
		args = append(args, *req.Description)
	}

	if req.Status != nil {
		if *req.Status != "todo" && *req.Status != "in_progress" && *req.Status != "done" {
			respondError(w, http.StatusBadRequest, "invalid status (must be: todo, in_progress, or done)", "invalid_input")
			return
		}
		query += ", status = ?"
		args = append(args, *req.Status)
	}

	query += " WHERE id = ?"
	args = append(args, taskID)

	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update task", "internal_error")
		return
	}

	// Fetch the updated task
	var t Task
	fetchQuery := `
		SELECT id, project_id, title, description, status, created_at, updated_at
		FROM tasks
		WHERE id = ?
	`
	err = s.db.QueryRowContext(ctx, fetchQuery, taskID).Scan(
		&t.ID, &t.ProjectID, &t.Title, &t.Description, &t.Status, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch updated task", "internal_error")
		return
	}

	respondJSON(w, http.StatusOK, t)
}

// HandleDeleteTask deletes a task
func (s *Server) HandleDeleteTask(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	userID := r.Context().Value(UserIDKey).(int64)
	taskID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid task ID", "invalid_input")
		return
	}

	// Verify user owns the project that contains this task
	checkQuery := `
		SELECT p.owner_id FROM projects p
		JOIN tasks t ON t.project_id = p.id
		WHERE t.id = ?
	`
	var ownerID int64
	if err := s.db.QueryRowContext(ctx, checkQuery, taskID).Scan(&ownerID); err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "task not found", "not_found")
		return
	} else if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to verify task ownership", "internal_error")
		return
	}

	if ownerID != userID {
		respondError(w, http.StatusForbidden, "access denied", "forbidden")
		return
	}

	query := `DELETE FROM tasks WHERE id = ?`
	result, err := s.db.ExecContext(ctx, query, taskID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to delete task", "internal_error")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to verify deletion", "internal_error")
		return
	}

	if rowsAffected == 0 {
		respondError(w, http.StatusNotFound, "task not found", "not_found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
