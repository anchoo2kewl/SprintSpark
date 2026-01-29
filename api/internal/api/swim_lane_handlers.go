package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type SwimLane struct {
	ID        int64     `json:"id"`
	ProjectID int64     `json:"project_id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateSwimLaneRequest struct {
	Name     string `json:"name"`
	Color    string `json:"color"`
	Position int    `json:"position"`
}

type UpdateSwimLaneRequest struct {
	Name     *string `json:"name,omitempty"`
	Color    *string `json:"color,omitempty"`
	Position *int    `json:"position,omitempty"`
}

// HandleListSwimLanes returns all swim lanes for a project
func (s *Server) HandleListSwimLanes(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	userID := r.Context().Value(UserIDKey).(int64)
	projectID, err := strconv.ParseInt(chi.URLParam(r, "projectId"), 10, 64)
	if err != nil {
		s.logger.Warn("Invalid project ID", zap.Error(err))
		respondError(w, http.StatusBadRequest, "invalid project ID", "invalid_input")
		return
	}

	// Verify user has access to this project
	hasAccess, err := s.checkProjectAccess(ctx, userID, projectID)
	if err != nil {
		s.logger.Error("Failed to verify project access", zap.Error(err), zap.Int64("userID", userID), zap.Int64("projectID", projectID))
		respondError(w, http.StatusInternalServerError, "failed to verify project access", "internal_error")
		return
	}
	if !hasAccess {
		s.logger.Warn("Access denied to project", zap.Int64("userID", userID), zap.Int64("projectID", projectID))
		respondError(w, http.StatusForbidden, "access denied", "forbidden")
		return
	}

	query := `
		SELECT id, project_id, name, color, position, created_at, updated_at
		FROM swim_lanes
		WHERE project_id = ?
		ORDER BY position ASC
	`

	rows, err := s.db.QueryContext(ctx, query, projectID)
	if err != nil {
		s.logger.Error("Failed to fetch swim lanes", zap.Error(err), zap.Int64("projectID", projectID))
		respondError(w, http.StatusInternalServerError, "failed to fetch swim lanes", "internal_error")
		return
	}
	defer rows.Close()

	swimLanes := []SwimLane{}
	for rows.Next() {
		var sl SwimLane
		if err := rows.Scan(&sl.ID, &sl.ProjectID, &sl.Name, &sl.Color, &sl.Position, &sl.CreatedAt, &sl.UpdatedAt); err != nil {
			s.logger.Error("Failed to scan swim lane", zap.Error(err))
			respondError(w, http.StatusInternalServerError, "failed to scan swim lane", "internal_error")
			return
		}
		swimLanes = append(swimLanes, sl)
	}

	if err := rows.Err(); err != nil {
		s.logger.Error("Error iterating swim lanes", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "error iterating swim lanes", "internal_error")
		return
	}

	respondJSON(w, http.StatusOK, swimLanes)
}

// HandleCreateSwimLane creates a new swim lane
func (s *Server) HandleCreateSwimLane(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	userID := r.Context().Value(UserIDKey).(int64)
	projectID, err := strconv.ParseInt(chi.URLParam(r, "projectId"), 10, 64)
	if err != nil {
		s.logger.Warn("Invalid project ID", zap.Error(err))
		respondError(w, http.StatusBadRequest, "invalid project ID", "invalid_input")
		return
	}

	// Verify user has access to this project
	hasAccess, err := s.checkProjectAccess(ctx, userID, projectID)
	if err != nil {
		s.logger.Error("Failed to verify project access", zap.Error(err), zap.Int64("userID", userID), zap.Int64("projectID", projectID))
		respondError(w, http.StatusInternalServerError, "failed to verify project access", "internal_error")
		return
	}
	if !hasAccess {
		s.logger.Warn("Access denied to project", zap.Int64("userID", userID), zap.Int64("projectID", projectID))
		respondError(w, http.StatusForbidden, "access denied", "forbidden")
		return
	}

	var req CreateSwimLaneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Warn("Invalid request body", zap.Error(err))
		respondError(w, http.StatusBadRequest, "invalid request body", "invalid_input")
		return
	}

	// Validation
	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "swim lane name is required", "invalid_input")
		return
	}
	if len(req.Name) > 50 {
		respondError(w, http.StatusBadRequest, "swim lane name is too long (max 50 characters)", "invalid_input")
		return
	}
	if req.Color == "" {
		req.Color = "#6B7280" // default gray
	}

	// Check swim lane count limit (max 6)
	var count int
	countQuery := `SELECT COUNT(*) FROM swim_lanes WHERE project_id = ?`
	if err := s.db.QueryRowContext(ctx, countQuery, projectID).Scan(&count); err != nil {
		s.logger.Error("Failed to count swim lanes", zap.Error(err), zap.Int64("projectID", projectID))
		respondError(w, http.StatusInternalServerError, "failed to count swim lanes", "internal_error")
		return
	}
	if count >= 6 {
		respondError(w, http.StatusBadRequest, "maximum 6 swim lanes allowed per project", "max_limit_reached")
		return
	}

	// Check minimum (need at least 2)
	if req.Position < 0 {
		req.Position = 0
	}

	query := `
		INSERT INTO swim_lanes (project_id, name, color, position, created_at, updated_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	result, err := s.db.ExecContext(ctx, query, projectID, req.Name, req.Color, req.Position)
	if err != nil {
		s.logger.Error("Failed to create swim lane", zap.Error(err), zap.Int64("projectID", projectID), zap.String("name", req.Name))
		respondError(w, http.StatusInternalServerError, "failed to create swim lane", "internal_error")
		return
	}

	swimLaneID, err := result.LastInsertId()
	if err != nil {
		s.logger.Error("Failed to get swim lane ID", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to get swim lane ID", "internal_error")
		return
	}

	// Fetch the created swim lane
	var sl SwimLane
	fetchQuery := `
		SELECT id, project_id, name, color, position, created_at, updated_at
		FROM swim_lanes
		WHERE id = ?
	`
	err = s.db.QueryRowContext(ctx, fetchQuery, swimLaneID).Scan(
		&sl.ID, &sl.ProjectID, &sl.Name, &sl.Color, &sl.Position, &sl.CreatedAt, &sl.UpdatedAt,
	)
	if err != nil {
		s.logger.Error("Failed to fetch created swim lane", zap.Error(err), zap.Int64("swimLaneID", swimLaneID))
		respondError(w, http.StatusInternalServerError, "failed to fetch created swim lane", "internal_error")
		return
	}

	s.logger.Info("Swim lane created", zap.Int64("swimLaneID", swimLaneID), zap.Int64("projectID", projectID), zap.String("name", req.Name))
	respondJSON(w, http.StatusCreated, sl)
}

// HandleUpdateSwimLane updates an existing swim lane
func (s *Server) HandleUpdateSwimLane(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	userID := r.Context().Value(UserIDKey).(int64)
	swimLaneID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		s.logger.Warn("Invalid swim lane ID", zap.Error(err))
		respondError(w, http.StatusBadRequest, "invalid swim lane ID", "invalid_input")
		return
	}

	var req UpdateSwimLaneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Warn("Invalid request body", zap.Error(err))
		respondError(w, http.StatusBadRequest, "invalid request body", "invalid_input")
		return
	}

	// Get swim lane's project ID and verify user has access
	var projectID int64
	projectQuery := `SELECT project_id FROM swim_lanes WHERE id = ?`
	if err := s.db.QueryRowContext(ctx, projectQuery, swimLaneID).Scan(&projectID); err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "swim lane not found", "not_found")
		return
	} else if err != nil {
		s.logger.Error("Failed to get swim lane project", zap.Error(err), zap.Int64("swimLaneID", swimLaneID))
		respondError(w, http.StatusInternalServerError, "failed to get swim lane project", "internal_error")
		return
	}

	// Verify user has access to the project
	hasAccess, err := s.checkProjectAccess(ctx, userID, projectID)
	if err != nil {
		s.logger.Error("Failed to verify project access", zap.Error(err), zap.Int64("userID", userID), zap.Int64("projectID", projectID))
		respondError(w, http.StatusInternalServerError, "failed to verify project access", "internal_error")
		return
	}
	if !hasAccess {
		s.logger.Warn("Access denied to project", zap.Int64("userID", userID), zap.Int64("projectID", projectID))
		respondError(w, http.StatusForbidden, "access denied", "forbidden")
		return
	}

	// Build update query dynamically
	query := "UPDATE swim_lanes SET updated_at = CURRENT_TIMESTAMP"
	args := []interface{}{}

	if req.Name != nil {
		if *req.Name == "" {
			respondError(w, http.StatusBadRequest, "swim lane name cannot be empty", "invalid_input")
			return
		}
		if len(*req.Name) > 50 {
			respondError(w, http.StatusBadRequest, "swim lane name is too long (max 50 characters)", "invalid_input")
			return
		}
		query += ", name = ?"
		args = append(args, *req.Name)
	}

	if req.Color != nil {
		query += ", color = ?"
		args = append(args, *req.Color)
	}

	if req.Position != nil {
		if *req.Position < 0 {
			respondError(w, http.StatusBadRequest, "position cannot be negative", "invalid_input")
			return
		}
		query += ", position = ?"
		args = append(args, *req.Position)
	}

	query += " WHERE id = ?"
	args = append(args, swimLaneID)

	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		s.logger.Error("Failed to update swim lane", zap.Error(err), zap.Int64("swimLaneID", swimLaneID))
		respondError(w, http.StatusInternalServerError, "failed to update swim lane", "internal_error")
		return
	}

	// Fetch the updated swim lane
	var sl SwimLane
	fetchQuery := `
		SELECT id, project_id, name, color, position, created_at, updated_at
		FROM swim_lanes
		WHERE id = ?
	`
	err = s.db.QueryRowContext(ctx, fetchQuery, swimLaneID).Scan(
		&sl.ID, &sl.ProjectID, &sl.Name, &sl.Color, &sl.Position, &sl.CreatedAt, &sl.UpdatedAt,
	)
	if err != nil {
		s.logger.Error("Failed to fetch updated swim lane", zap.Error(err), zap.Int64("swimLaneID", swimLaneID))
		respondError(w, http.StatusInternalServerError, "failed to fetch updated swim lane", "internal_error")
		return
	}

	s.logger.Info("Swim lane updated", zap.Int64("swimLaneID", swimLaneID), zap.Int64("projectID", projectID))
	respondJSON(w, http.StatusOK, sl)
}

// HandleDeleteSwimLane deletes a swim lane
func (s *Server) HandleDeleteSwimLane(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	userID := r.Context().Value(UserIDKey).(int64)
	swimLaneID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		s.logger.Warn("Invalid swim lane ID", zap.Error(err))
		respondError(w, http.StatusBadRequest, "invalid swim lane ID", "invalid_input")
		return
	}

	// Get swim lane's project ID and verify user has access
	var projectID int64
	projectQuery := `SELECT project_id FROM swim_lanes WHERE id = ?`
	if err := s.db.QueryRowContext(ctx, projectQuery, swimLaneID).Scan(&projectID); err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "swim lane not found", "not_found")
		return
	} else if err != nil {
		s.logger.Error("Failed to get swim lane project", zap.Error(err), zap.Int64("swimLaneID", swimLaneID))
		respondError(w, http.StatusInternalServerError, "failed to get swim lane project", "internal_error")
		return
	}

	// Verify user has access to the project
	hasAccess, err := s.checkProjectAccess(ctx, userID, projectID)
	if err != nil {
		s.logger.Error("Failed to verify project access", zap.Error(err), zap.Int64("userID", userID), zap.Int64("projectID", projectID))
		respondError(w, http.StatusInternalServerError, "failed to verify project access", "internal_error")
		return
	}
	if !hasAccess {
		s.logger.Warn("Access denied to project", zap.Int64("userID", userID), zap.Int64("projectID", projectID))
		respondError(w, http.StatusForbidden, "access denied", "forbidden")
		return
	}

	// Check minimum swim lanes (need at least 2)
	var count int
	countQuery := `SELECT COUNT(*) FROM swim_lanes WHERE project_id = ?`
	if err := s.db.QueryRowContext(ctx, countQuery, projectID).Scan(&count); err != nil {
		s.logger.Error("Failed to count swim lanes", zap.Error(err), zap.Int64("projectID", projectID))
		respondError(w, http.StatusInternalServerError, "failed to count swim lanes", "internal_error")
		return
	}
	if count <= 2 {
		respondError(w, http.StatusBadRequest, "minimum 2 swim lanes required per project", "min_limit_reached")
		return
	}

	query := `DELETE FROM swim_lanes WHERE id = ?`
	result, err := s.db.ExecContext(ctx, query, swimLaneID)
	if err != nil {
		s.logger.Error("Failed to delete swim lane", zap.Error(err), zap.Int64("swimLaneID", swimLaneID))
		respondError(w, http.StatusInternalServerError, "failed to delete swim lane", "internal_error")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		s.logger.Error("Failed to verify deletion", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "failed to verify deletion", "internal_error")
		return
	}

	if rowsAffected == 0 {
		respondError(w, http.StatusNotFound, "swim lane not found", "not_found")
		return
	}

	s.logger.Info("Swim lane deleted", zap.Int64("swimLaneID", swimLaneID), zap.Int64("projectID", projectID))
	w.WriteHeader(http.StatusNoContent)
}
