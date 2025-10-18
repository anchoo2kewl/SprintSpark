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

type Project struct {
	ID          int64     `json:"id"`
	OwnerID     int64     `json:"owner_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateProjectRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

type UpdateProjectRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// HandleListProjects returns all projects for the authenticated user
func (s *Server) HandleListProjects(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	userID := r.Context().Value(UserIDKey).(int64)

	query := `
		SELECT id, owner_id, name, description, created_at, updated_at
		FROM projects
		WHERE owner_id = ?
		ORDER BY updated_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch projects", "internal_error")
		return
	}
	defer rows.Close()

	projects := []Project{}
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.OwnerID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to scan project", "internal_error")
			return
		}
		projects = append(projects, p)
	}

	if err := rows.Err(); err != nil {
		respondError(w, http.StatusInternalServerError, "error iterating projects", "internal_error")
		return
	}

	respondJSON(w, http.StatusOK, projects)
}

// HandleGetProject returns a single project by ID
func (s *Server) HandleGetProject(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	userID := r.Context().Value(UserIDKey).(int64)
	projectID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project ID", "invalid_input")
		return
	}

	query := `
		SELECT id, owner_id, name, description, created_at, updated_at
		FROM projects
		WHERE id = ? AND owner_id = ?
	`

	var p Project
	err = s.db.QueryRowContext(ctx, query, projectID, userID).Scan(
		&p.ID, &p.OwnerID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "project not found", "not_found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch project", "internal_error")
		return
	}

	respondJSON(w, http.StatusOK, p)
}

// HandleCreateProject creates a new project
func (s *Server) HandleCreateProject(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	userID := r.Context().Value(UserIDKey).(int64)

	var req CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", "invalid_input")
		return
	}

	// Validation
	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "project name is required", "invalid_input")
		return
	}
	if len(req.Name) > 255 {
		respondError(w, http.StatusBadRequest, "project name is too long (max 255 characters)", "invalid_input")
		return
	}

	query := `
		INSERT INTO projects (owner_id, name, description, created_at, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	result, err := s.db.ExecContext(ctx, query, userID, req.Name, req.Description)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create project", "internal_error")
		return
	}

	projectID, err := result.LastInsertId()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get project ID", "internal_error")
		return
	}

	// Fetch the created project
	var p Project
	fetchQuery := `
		SELECT id, owner_id, name, description, created_at, updated_at
		FROM projects
		WHERE id = ?
	`
	err = s.db.QueryRowContext(ctx, fetchQuery, projectID).Scan(
		&p.ID, &p.OwnerID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch created project", "internal_error")
		return
	}

	respondJSON(w, http.StatusCreated, p)
}

// HandleUpdateProject updates an existing project
func (s *Server) HandleUpdateProject(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	userID := r.Context().Value(UserIDKey).(int64)
	projectID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project ID", "invalid_input")
		return
	}

	var req UpdateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", "invalid_input")
		return
	}

	// Check project exists and belongs to user
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM projects WHERE id = ? AND owner_id = ?)`
	if err := s.db.QueryRowContext(ctx, checkQuery, projectID, userID).Scan(&exists); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to check project ownership", "internal_error")
		return
	}
	if !exists {
		respondError(w, http.StatusNotFound, "project not found", "not_found")
		return
	}

	// Build update query dynamically
	query := "UPDATE projects SET updated_at = CURRENT_TIMESTAMP"
	args := []interface{}{}

	if req.Name != nil {
		if *req.Name == "" {
			respondError(w, http.StatusBadRequest, "project name cannot be empty", "invalid_input")
			return
		}
		if len(*req.Name) > 255 {
			respondError(w, http.StatusBadRequest, "project name is too long (max 255 characters)", "invalid_input")
			return
		}
		query += ", name = ?"
		args = append(args, *req.Name)
	}

	if req.Description != nil {
		query += ", description = ?"
		args = append(args, *req.Description)
	}

	query += " WHERE id = ? AND owner_id = ?"
	args = append(args, projectID, userID)

	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update project", "internal_error")
		return
	}

	// Fetch the updated project
	var p Project
	fetchQuery := `
		SELECT id, owner_id, name, description, created_at, updated_at
		FROM projects
		WHERE id = ?
	`
	err = s.db.QueryRowContext(ctx, fetchQuery, projectID).Scan(
		&p.ID, &p.OwnerID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch updated project", "internal_error")
		return
	}

	respondJSON(w, http.StatusOK, p)
}

// HandleDeleteProject deletes a project
func (s *Server) HandleDeleteProject(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	userID := r.Context().Value(UserIDKey).(int64)
	projectID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project ID", "invalid_input")
		return
	}

	query := `DELETE FROM projects WHERE id = ? AND owner_id = ?`
	result, err := s.db.ExecContext(ctx, query, projectID, userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to delete project", "internal_error")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to verify deletion", "internal_error")
		return
	}

	if rowsAffected == 0 {
		respondError(w, http.StatusNotFound, "project not found", "not_found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
