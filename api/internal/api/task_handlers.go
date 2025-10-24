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
	ID             int64     `json:"id"`
	ProjectID      int64     `json:"project_id"`
	Title          string    `json:"title"`
	Description    *string   `json:"description,omitempty"`
	Status         string    `json:"status"`
	DueDate        *string   `json:"due_date,omitempty"`
	SprintID       *int64    `json:"sprint_id,omitempty"`
	Priority       string    `json:"priority"`
	AssigneeID     *int64    `json:"assignee_id,omitempty"`
	EstimatedHours *float64  `json:"estimated_hours,omitempty"`
	ActualHours    *float64  `json:"actual_hours,omitempty"`
	Tags           []Tag     `json:"tags,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CreateTaskRequest struct {
	Title          string   `json:"title"`
	Description    *string  `json:"description,omitempty"`
	Status         *string  `json:"status,omitempty"`
	DueDate        *string  `json:"due_date,omitempty"`
	SprintID       *int64   `json:"sprint_id,omitempty"`
	Priority       *string  `json:"priority,omitempty"`
	AssigneeID     *int64   `json:"assignee_id,omitempty"`
	EstimatedHours *float64 `json:"estimated_hours,omitempty"`
	ActualHours    *float64 `json:"actual_hours,omitempty"`
	TagIDs         []int64  `json:"tag_ids,omitempty"`
}

type UpdateTaskRequest struct {
	Title          *string  `json:"title,omitempty"`
	Description    *string  `json:"description,omitempty"`
	Status         *string  `json:"status,omitempty"`
	DueDate        *string  `json:"due_date,omitempty"`
	SprintID       *int64   `json:"sprint_id,omitempty"`
	Priority       *string  `json:"priority,omitempty"`
	AssigneeID     *int64   `json:"assignee_id,omitempty"`
	EstimatedHours *float64 `json:"estimated_hours,omitempty"`
	ActualHours    *float64 `json:"actual_hours,omitempty"`
	TagIDs         *[]int64 `json:"tag_ids,omitempty"`
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

	// First, fetch all tasks
	query := `
		SELECT id, project_id, title, description, status, due_date, sprint_id, priority, assignee_id, estimated_hours, actual_hours, created_at, updated_at
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
	taskIDs := []int64{}
	taskMap := make(map[int64]*Task)

	for rows.Next() {
		var t Task
		var priority sql.NullString
		if err := rows.Scan(&t.ID, &t.ProjectID, &t.Title, &t.Description, &t.Status, &t.DueDate, &t.SprintID, &priority, &t.AssigneeID, &t.EstimatedHours, &t.ActualHours, &t.CreatedAt, &t.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to scan task", "internal_error")
			return
		}
		t.Priority = "medium" // default
		if priority.Valid {
			t.Priority = priority.String
		}
		t.Tags = []Tag{} // Initialize empty tags array

		tasks = append(tasks, t)
		taskIDs = append(taskIDs, t.ID)
		taskMap[t.ID] = &tasks[len(tasks)-1]
	}

	if err := rows.Err(); err != nil {
		respondError(w, http.StatusInternalServerError, "error iterating tasks", "internal_error")
		return
	}

	// If there are tasks, fetch all tags in a single query
	if len(taskIDs) > 0 {
		// Build IN clause for task IDs
		placeholders := ""
		args := make([]interface{}, len(taskIDs))
		for i, id := range taskIDs {
			if i > 0 {
				placeholders += ","
			}
			placeholders += "?"
			args[i] = id
		}

		tagQuery := `
			SELECT tt.task_id, t.id, t.user_id, t.name, t.color, t.created_at
			FROM tags t
			JOIN task_tags tt ON t.id = tt.tag_id
			WHERE tt.task_id IN (` + placeholders + `)
			ORDER BY tt.task_id, t.name
		`

		tagRows, err := s.db.QueryContext(ctx, tagQuery, args...)
		if err == nil {
			defer tagRows.Close()
			for tagRows.Next() {
				var taskID int64
				var tag Tag
				if err := tagRows.Scan(&taskID, &tag.ID, &tag.UserID, &tag.Name, &tag.Color, &tag.CreatedAt); err == nil {
					if task, exists := taskMap[taskID]; exists {
						task.Tags = append(task.Tags, tag)
					}
				}
			}
		}
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

	// Default priority
	priority := "medium"
	if req.Priority != nil {
		priority = *req.Priority
	}

	// Validate priority
	if priority != "low" && priority != "medium" && priority != "high" && priority != "urgent" {
		respondError(w, http.StatusBadRequest, "invalid priority (must be: low, medium, high, or urgent)", "invalid_input")
		return
	}

	query := `
		INSERT INTO tasks (project_id, title, description, status, due_date, sprint_id, priority, assignee_id, estimated_hours, actual_hours, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	result, err := s.db.ExecContext(ctx, query, projectID, req.Title, req.Description, status, req.DueDate, req.SprintID, priority, req.AssigneeID, req.EstimatedHours, req.ActualHours)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create task", "internal_error")
		return
	}

	taskID, err := result.LastInsertId()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get task ID", "internal_error")
		return
	}

	// Add tags if provided
	if len(req.TagIDs) > 0 {
		for _, tagID := range req.TagIDs {
			_, err := s.db.ExecContext(ctx, "INSERT INTO task_tags (task_id, tag_id) VALUES (?, ?)", taskID, tagID)
			if err != nil {
				// Continue even if tag insertion fails
				continue
			}
		}
	}

	// Fetch the created task
	var t Task
	var priorityVal sql.NullString
	fetchQuery := `
		SELECT id, project_id, title, description, status, due_date, sprint_id, priority, assignee_id, estimated_hours, actual_hours, created_at, updated_at
		FROM tasks
		WHERE id = ?
	`
	err = s.db.QueryRowContext(ctx, fetchQuery, taskID).Scan(
		&t.ID, &t.ProjectID, &t.Title, &t.Description, &t.Status, &t.DueDate, &t.SprintID, &priorityVal, &t.AssigneeID, &t.EstimatedHours, &t.ActualHours, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch created task", "internal_error")
		return
	}

	t.Priority = "medium" // default
	if priorityVal.Valid {
		t.Priority = priorityVal.String
	}

	// Fetch tags
	tagQuery := `
		SELECT t.id, t.user_id, t.name, t.color, t.created_at
		FROM tags t
		JOIN task_tags tt ON t.id = tt.tag_id
		WHERE tt.task_id = ?
	`
	tagRows, err := s.db.QueryContext(ctx, tagQuery, taskID)
	if err == nil {
		defer tagRows.Close()
		t.Tags = []Tag{}
		for tagRows.Next() {
			var tag Tag
			if err := tagRows.Scan(&tag.ID, &tag.UserID, &tag.Name, &tag.Color, &tag.CreatedAt); err == nil {
				t.Tags = append(t.Tags, tag)
			}
		}
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

	if req.DueDate != nil {
		query += ", due_date = ?"
		args = append(args, *req.DueDate)
	}

	if req.Priority != nil {
		if *req.Priority != "low" && *req.Priority != "medium" && *req.Priority != "high" && *req.Priority != "urgent" {
			respondError(w, http.StatusBadRequest, "invalid priority (must be: low, medium, high, or urgent)", "invalid_input")
			return
		}
		query += ", priority = ?"
		args = append(args, *req.Priority)
	}

	if req.SprintID != nil {
		query += ", sprint_id = ?"
		args = append(args, *req.SprintID)
	}

	if req.AssigneeID != nil {
		query += ", assignee_id = ?"
		args = append(args, *req.AssigneeID)
	}

	if req.EstimatedHours != nil {
		query += ", estimated_hours = ?"
		args = append(args, *req.EstimatedHours)
	}

	if req.ActualHours != nil {
		query += ", actual_hours = ?"
		args = append(args, *req.ActualHours)
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
	var priorityVal sql.NullString
	fetchQuery := `
		SELECT id, project_id, title, description, status, due_date, sprint_id, priority, assignee_id, estimated_hours, actual_hours, created_at, updated_at
		FROM tasks
		WHERE id = ?
	`
	err = s.db.QueryRowContext(ctx, fetchQuery, taskID).Scan(
		&t.ID, &t.ProjectID, &t.Title, &t.Description, &t.Status, &t.DueDate, &t.SprintID, &priorityVal, &t.AssigneeID, &t.EstimatedHours, &t.ActualHours, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch updated task", "internal_error")
		return
	}

	t.Priority = "medium" // default
	if priorityVal.Valid {
		t.Priority = priorityVal.String
	}

	// Fetch tags
	tagQuery := `
		SELECT t.id, t.user_id, t.name, t.color, t.created_at
		FROM tags t
		JOIN task_tags tt ON t.id = tt.tag_id
		WHERE tt.task_id = ?
	`
	tagRows, err := s.db.QueryContext(ctx, tagQuery, taskID)
	if err == nil {
		defer tagRows.Close()
		t.Tags = []Tag{}
		for tagRows.Next() {
			var tag Tag
			if err := tagRows.Scan(&tag.ID, &tag.UserID, &tag.Name, &tag.Color, &tag.CreatedAt); err == nil {
				t.Tags = append(t.Tags, tag)
			}
		}
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
