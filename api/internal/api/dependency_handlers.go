package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

// TaskDependencyResponse represents a dependency in the API response.
type TaskDependencyResponse struct {
	ID             int64     `json:"id"`
	TaskID         int64     `json:"task_id"`
	TaskTitle      string    `json:"task_title,omitempty"`
	TaskNumber     int       `json:"task_number,omitempty"`
	DependsOnID    int64     `json:"depends_on_id"`
	DependsOnTitle string    `json:"depends_on_title,omitempty"`
	DependsOnNum   int       `json:"depends_on_number,omitempty"`
	DependencyType string    `json:"dependency_type"`
	CreatedAt      time.Time `json:"created_at"`
}

// TaskDependencies groups dependencies by direction.
type TaskDependencies struct {
	BlockedBy []TaskDependencyResponse `json:"blocked_by"`
	Blocks    []TaskDependencyResponse `json:"blocks"`
}

// CreateDependencyRequest represents a request to create a dependency.
type CreateDependencyRequest struct {
	DependsOnID    int64  `json:"depends_on_id"`
	DependencyType string `json:"dependency_type,omitempty"`
}

// HandleListDependencies returns all dependencies for a task (both directions).
// Route: GET /api/tasks/{taskId}/dependencies
func (s *Server) HandleListDependencies(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	taskID, err := strconv.ParseInt(chi.URLParam(r, "taskId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	userID, ok := GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get project_id for access check
	var projectID int64
	err = s.db.QueryRowContext(ctx, `SELECT project_id FROM tasks WHERE id = $1`, taskID).Scan(&projectID)
	if err == sql.ErrNoRows {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	hasAccess, err := s.checkProjectAccess(ctx, userID, projectID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !hasAccess {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// "blocked_by" — this task depends on these tasks
	blockedByRows, err := s.db.QueryContext(ctx,
		`SELECT td.id, td.task_id, td.depends_on_id, td.dependency_type, td.created_at,
		        t.title, COALESCE(t.task_number, 0)
		 FROM task_dependencies td
		 JOIN tasks t ON t.id = td.depends_on_id
		 WHERE td.task_id = $1
		 ORDER BY td.created_at`, taskID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch dependencies", "internal_error")
		return
	}
	defer blockedByRows.Close()

	blockedBy := []TaskDependencyResponse{}
	for blockedByRows.Next() {
		var dep TaskDependencyResponse
		if err := blockedByRows.Scan(&dep.ID, &dep.TaskID, &dep.DependsOnID, &dep.DependencyType, &dep.CreatedAt, &dep.DependsOnTitle, &dep.DependsOnNum); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to scan dependency", "internal_error")
			return
		}
		blockedBy = append(blockedBy, dep)
	}

	// "blocks" — these tasks depend on this task
	blocksRows, err := s.db.QueryContext(ctx,
		`SELECT td.id, td.task_id, td.depends_on_id, td.dependency_type, td.created_at,
		        t.title, COALESCE(t.task_number, 0)
		 FROM task_dependencies td
		 JOIN tasks t ON t.id = td.task_id
		 WHERE td.depends_on_id = $1
		 ORDER BY td.created_at`, taskID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch dependents", "internal_error")
		return
	}
	defer blocksRows.Close()

	blocks := []TaskDependencyResponse{}
	for blocksRows.Next() {
		var dep TaskDependencyResponse
		if err := blocksRows.Scan(&dep.ID, &dep.TaskID, &dep.DependsOnID, &dep.DependencyType, &dep.CreatedAt, &dep.TaskTitle, &dep.TaskNumber); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to scan dependent", "internal_error")
			return
		}
		blocks = append(blocks, dep)
	}

	respondJSON(w, http.StatusOK, TaskDependencies{
		BlockedBy: blockedBy,
		Blocks:    blocks,
	})
}

// HandleCreateDependency creates a new task dependency with cycle detection.
// Route: POST /api/tasks/{taskId}/dependencies
func (s *Server) HandleCreateDependency(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	taskID, err := strconv.ParseInt(chi.URLParam(r, "taskId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	userID, ok := GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var projectID int64
	err = s.db.QueryRowContext(ctx, `SELECT project_id FROM tasks WHERE id = $1`, taskID).Scan(&projectID)
	if err == sql.ErrNoRows {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	hasAccess, err := s.checkProjectAccess(ctx, userID, projectID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !hasAccess {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req CreateDependencyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.DependsOnID == 0 {
		http.Error(w, "depends_on_id is required", http.StatusBadRequest)
		return
	}

	if req.DependsOnID == taskID {
		http.Error(w, "A task cannot depend on itself", http.StatusBadRequest)
		return
	}

	depType := req.DependencyType
	if depType == "" {
		depType = "blocks"
	}
	if depType != "blocks" && depType != "related" {
		http.Error(w, "Invalid dependency_type. Must be blocks or related", http.StatusBadRequest)
		return
	}

	// Verify the depends_on task exists and is in the same project
	var depProjectID int64
	err = s.db.QueryRowContext(ctx, `SELECT project_id FROM tasks WHERE id = $1`, req.DependsOnID).Scan(&depProjectID)
	if err == sql.ErrNoRows {
		http.Error(w, "Depends-on task not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if depProjectID != projectID {
		http.Error(w, "Dependencies must be within the same project", http.StatusBadRequest)
		return
	}

	// Cycle detection: walk from depends_on_id and ensure we never reach taskID
	var hasCycle bool
	err = s.db.QueryRowContext(ctx,
		`WITH RECURSIVE chain AS (
			SELECT depends_on_id FROM task_dependencies WHERE task_id = $1
			UNION ALL
			SELECT td.depends_on_id FROM task_dependencies td JOIN chain c ON td.task_id = c.depends_on_id
		)
		SELECT EXISTS(SELECT 1 FROM chain WHERE depends_on_id = $2)`,
		req.DependsOnID, taskID).Scan(&hasCycle)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if hasCycle {
		respondError(w, http.StatusConflict, "Adding this dependency would create a cycle", "cycle_detected")
		return
	}

	var newID int64
	err = s.db.QueryRowContext(ctx,
		`INSERT INTO task_dependencies (task_id, depends_on_id, dependency_type)
		 VALUES ($1, $2, $3) RETURNING id`,
		taskID, req.DependsOnID, depType).Scan(&newID)
	if err != nil {
		// Check for unique constraint violation
		respondError(w, http.StatusConflict, "This dependency already exists", "duplicate")
		return
	}

	dep := TaskDependencyResponse{
		ID:             newID,
		TaskID:         taskID,
		DependsOnID:    req.DependsOnID,
		DependencyType: depType,
		CreatedAt:      time.Now(),
	}

	respondJSON(w, http.StatusCreated, dep)
}

// HandleDeleteDependency removes a task dependency.
// Route: DELETE /api/task-dependencies/{id}
func (s *Server) HandleDeleteDependency(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	depID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid dependency ID", http.StatusBadRequest)
		return
	}

	userID, ok := GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get task_id to check project access
	var taskID int64
	err = s.db.QueryRowContext(ctx, `SELECT task_id FROM task_dependencies WHERE id = $1`, depID).Scan(&taskID)
	if err == sql.ErrNoRows {
		http.Error(w, "Dependency not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var projectID int64
	err = s.db.QueryRowContext(ctx, `SELECT project_id FROM tasks WHERE id = $1`, taskID).Scan(&projectID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	hasAccess, err := s.checkProjectAccess(ctx, userID, projectID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !hasAccess {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	_, err = s.db.ExecContext(ctx, `DELETE FROM task_dependencies WHERE id = $1`, depID)
	if err != nil {
		http.Error(w, "Failed to delete dependency", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Dependency deleted successfully"})
}
