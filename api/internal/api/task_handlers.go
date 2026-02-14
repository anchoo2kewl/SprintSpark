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
	TaskNumber     int64     `json:"task_number"`
	Title          string    `json:"title"`
	Description    *string   `json:"description,omitempty"`
	Status         string    `json:"status"`
	SwimLaneID     *int64    `json:"swim_lane_id,omitempty"`
	SwimLaneName   *string   `json:"swim_lane_name,omitempty"`
	DueDate        *string   `json:"due_date,omitempty"`
	SprintID       *int64    `json:"sprint_id,omitempty"`
	SprintName     *string   `json:"sprint_name,omitempty"`
	Priority       string    `json:"priority"`
	AssigneeID     *int64    `json:"assignee_id,omitempty"`
	AssigneeName   *string   `json:"assignee_name,omitempty"`
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
	SwimLaneID     *int64   `json:"swim_lane_id,omitempty"`
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
	SwimLaneID     *int64   `json:"swim_lane_id,omitempty"`
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

	// Verify user has access to this project
	hasAccess, err := s.checkProjectAccess(ctx, userID, projectID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to verify project access", "internal_error")
		return
	}
	if !hasAccess {
		respondError(w, http.StatusForbidden, "access denied", "forbidden")
		return
	}

	// First, fetch all tasks with assignee, sprint, and swim lane names
	query := `
		SELECT t.id, t.project_id, t.task_number, t.title, t.description, t.status, t.swim_lane_id, sl.name as swim_lane_name, t.due_date,
		       t.sprint_id, s.name as sprint_name,
		       t.priority, t.assignee_id, u.name as assignee_name,
		       t.estimated_hours, t.actual_hours, t.created_at, t.updated_at
		FROM tasks t
		LEFT JOIN users u ON t.assignee_id = u.id
		LEFT JOIN sprints s ON t.sprint_id = s.id
		LEFT JOIN swim_lanes sl ON t.swim_lane_id = sl.id
		WHERE t.project_id = ?
		ORDER BY t.created_at DESC
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
		var taskNumber sql.NullInt64
		if err := rows.Scan(&t.ID, &t.ProjectID, &taskNumber, &t.Title, &t.Description, &t.Status, &t.SwimLaneID, &t.SwimLaneName, &t.DueDate,
			&t.SprintID, &t.SprintName, &priority, &t.AssigneeID, &t.AssigneeName,
			&t.EstimatedHours, &t.ActualHours, &t.CreatedAt, &t.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to scan task", "internal_error")
			return
		}
		t.Priority = "medium" // default
		if priority.Valid {
			t.Priority = priority.String
		}
		if taskNumber.Valid {
			t.TaskNumber = taskNumber.Int64
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

	// Verify user has access to this project
	hasAccess, err := s.checkProjectAccess(ctx, userID, projectID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to verify project access", "internal_error")
		return
	}
	if !hasAccess {
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

	// Default status to 'todo' if not provided (for backward compatibility)
	status := "todo"
	if req.Status != nil {
		status = *req.Status
	}

	// Validate status
	if status != "todo" && status != "in_progress" && status != "done" {
		respondError(w, http.StatusBadRequest, "invalid status (must be: todo, in_progress, or done)", "invalid_input")
		return
	}

	// Sync swim_lane_id and status
	var swimLaneID *int64
	if req.SwimLaneID != nil {
		swimLaneID = req.SwimLaneID
		// Derive status from the swim lane's status_category
		var category string
		laneQuery := `SELECT status_category FROM swim_lanes WHERE id = ? AND project_id = ?`
		if err := s.db.QueryRowContext(ctx, laneQuery, *req.SwimLaneID, projectID).Scan(&category); err == nil && category != "" {
			status = category
		}
	} else {
		// Find first swim lane matching the status category
		var laneID int64
		laneQuery := `SELECT id FROM swim_lanes WHERE project_id = ? AND status_category = ? ORDER BY position ASC LIMIT 1`
		if err := s.db.QueryRowContext(ctx, laneQuery, projectID, status).Scan(&laneID); err == nil {
			swimLaneID = &laneID
		}
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

	// Assign task_number atomically using a transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to start transaction", "internal_error")
		return
	}
	defer tx.Rollback()

	// Get next task_number for this project
	var nextNumber int64
	err = tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(task_number), 0) + 1 FROM tasks WHERE project_id = ?`, projectID).Scan(&nextNumber)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get next task number", "internal_error")
		return
	}

	query := `
		INSERT INTO tasks (project_id, task_number, title, description, status, swim_lane_id, due_date, sprint_id, priority, assignee_id, estimated_hours, actual_hours, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	result, err := tx.ExecContext(ctx, query, projectID, nextNumber, req.Title, req.Description, status, swimLaneID, req.DueDate, req.SprintID, priority, req.AssigneeID, req.EstimatedHours, req.ActualHours)
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
			_, err := tx.ExecContext(ctx, "INSERT INTO task_tags (task_id, tag_id) VALUES (?, ?)", taskID, tagID)
			if err != nil {
				// Continue even if tag insertion fails
				continue
			}
		}
	}

	if err := tx.Commit(); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to commit task creation", "internal_error")
		return
	}

	// Fetch the created task with assignee, sprint, and swim lane names
	var t Task
	var priorityVal sql.NullString
	fetchQuery := `
		SELECT t.id, t.project_id, t.task_number, t.title, t.description, t.status, t.swim_lane_id, sl.name as swim_lane_name, t.due_date,
		       t.sprint_id, s.name as sprint_name,
		       t.priority, t.assignee_id, u.name as assignee_name,
		       t.estimated_hours, t.actual_hours, t.created_at, t.updated_at
		FROM tasks t
		LEFT JOIN users u ON t.assignee_id = u.id
		LEFT JOIN sprints s ON t.sprint_id = s.id
		LEFT JOIN swim_lanes sl ON t.swim_lane_id = sl.id
		WHERE t.id = ?
	`
	err = s.db.QueryRowContext(ctx, fetchQuery, taskID).Scan(
		&t.ID, &t.ProjectID, &t.TaskNumber, &t.Title, &t.Description, &t.Status, &t.SwimLaneID, &t.SwimLaneName, &t.DueDate,
		&t.SprintID, &t.SprintName, &priorityVal, &t.AssigneeID, &t.AssigneeName,
		&t.EstimatedHours, &t.ActualHours, &t.CreatedAt, &t.UpdatedAt,
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

	// Swim lane/status sync: if swim_lane_id changes, derive status; if status changes, find matching lane
	if req.SwimLaneID != nil && req.Status == nil {
		// Swim lane changed — derive status from lane's status_category
		var category string
		laneQuery := `SELECT status_category FROM swim_lanes WHERE id = ? AND project_id = ?`
		if err := s.db.QueryRowContext(ctx, laneQuery, *req.SwimLaneID, projectID).Scan(&category); err == nil && category != "" {
			query += ", status = ?"
			args = append(args, category)
		}
		query += ", swim_lane_id = ?"
		args = append(args, *req.SwimLaneID)
	} else if req.Status != nil && req.SwimLaneID == nil {
		// Status changed — find first swim lane with matching status_category
		if *req.Status != "todo" && *req.Status != "in_progress" && *req.Status != "done" {
			respondError(w, http.StatusBadRequest, "invalid status (must be: todo, in_progress, or done)", "invalid_input")
			return
		}
		query += ", status = ?"
		args = append(args, *req.Status)
		var laneID int64
		laneQuery := `SELECT id FROM swim_lanes WHERE project_id = ? AND status_category = ? ORDER BY position ASC LIMIT 1`
		if err := s.db.QueryRowContext(ctx, laneQuery, projectID, *req.Status).Scan(&laneID); err == nil {
			query += ", swim_lane_id = ?"
			args = append(args, laneID)
		}
	} else if req.Status != nil && req.SwimLaneID != nil {
		// Both provided — trust swim_lane_id, derive status from it
		if *req.Status != "todo" && *req.Status != "in_progress" && *req.Status != "done" {
			respondError(w, http.StatusBadRequest, "invalid status (must be: todo, in_progress, or done)", "invalid_input")
			return
		}
		var category string
		laneQuery := `SELECT status_category FROM swim_lanes WHERE id = ? AND project_id = ?`
		if err := s.db.QueryRowContext(ctx, laneQuery, *req.SwimLaneID, projectID).Scan(&category); err == nil && category != "" {
			query += ", status = ?"
			args = append(args, category)
		} else {
			query += ", status = ?"
			args = append(args, *req.Status)
		}
		query += ", swim_lane_id = ?"
		args = append(args, *req.SwimLaneID)
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

	// Fetch the updated task with assignee, sprint, and swim lane names
	var t Task
	var priorityVal sql.NullString
	fetchQuery := `
		SELECT t.id, t.project_id, t.task_number, t.title, t.description, t.status, t.swim_lane_id, sl.name as swim_lane_name, t.due_date,
		       t.sprint_id, s.name as sprint_name,
		       t.priority, t.assignee_id, u.name as assignee_name,
		       t.estimated_hours, t.actual_hours, t.created_at, t.updated_at
		FROM tasks t
		LEFT JOIN users u ON t.assignee_id = u.id
		LEFT JOIN sprints s ON t.sprint_id = s.id
		LEFT JOIN swim_lanes sl ON t.swim_lane_id = sl.id
		WHERE t.id = ?
	`
	err = s.db.QueryRowContext(ctx, fetchQuery, taskID).Scan(
		&t.ID, &t.ProjectID, &t.TaskNumber, &t.Title, &t.Description, &t.Status, &t.SwimLaneID, &t.SwimLaneName, &t.DueDate,
		&t.SprintID, &t.SprintName, &priorityVal, &t.AssigneeID, &t.AssigneeName,
		&t.EstimatedHours, &t.ActualHours, &t.CreatedAt, &t.UpdatedAt,
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

// HandleGetTaskByNumber returns a single task by project-scoped task number
func (s *Server) HandleGetTaskByNumber(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	userID := r.Context().Value(UserIDKey).(int64)
	projectID, err := strconv.ParseInt(chi.URLParam(r, "projectId"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project ID", "invalid_input")
		return
	}
	taskNumber, err := strconv.ParseInt(chi.URLParam(r, "taskNumber"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid task number", "invalid_input")
		return
	}

	// Verify user has access to this project
	hasAccess, err := s.checkProjectAccess(ctx, userID, projectID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to verify project access", "internal_error")
		return
	}
	if !hasAccess {
		respondError(w, http.StatusForbidden, "access denied", "forbidden")
		return
	}

	var t Task
	var priorityVal sql.NullString
	fetchQuery := `
		SELECT t.id, t.project_id, t.task_number, t.title, t.description, t.status, t.swim_lane_id, sl.name as swim_lane_name, t.due_date,
		       t.sprint_id, s.name as sprint_name,
		       t.priority, t.assignee_id, u.name as assignee_name,
		       t.estimated_hours, t.actual_hours, t.created_at, t.updated_at
		FROM tasks t
		LEFT JOIN users u ON t.assignee_id = u.id
		LEFT JOIN sprints s ON t.sprint_id = s.id
		LEFT JOIN swim_lanes sl ON t.swim_lane_id = sl.id
		WHERE t.project_id = ? AND t.task_number = ?
	`
	err = s.db.QueryRowContext(ctx, fetchQuery, projectID, taskNumber).Scan(
		&t.ID, &t.ProjectID, &t.TaskNumber, &t.Title, &t.Description, &t.Status, &t.SwimLaneID, &t.SwimLaneName, &t.DueDate,
		&t.SprintID, &t.SprintName, &priorityVal, &t.AssigneeID, &t.AssigneeName,
		&t.EstimatedHours, &t.ActualHours, &t.CreatedAt, &t.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "task not found", "not_found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch task", "internal_error")
		return
	}

	t.Priority = "medium"
	if priorityVal.Valid {
		t.Priority = priorityVal.String
	}

	// Fetch tags
	tagQuery := `
		SELECT tg.id, tg.user_id, tg.name, tg.color, tg.created_at
		FROM tags tg
		JOIN task_tags tt ON tg.id = tt.tag_id
		WHERE tt.task_id = ?
	`
	tagRows, err := s.db.QueryContext(ctx, tagQuery, t.ID)
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

// checkProjectAccess verifies that a user has access to a project via project_members table
func (s *Server) checkProjectAccess(ctx context.Context, userID, projectID int64) (bool, error) {
	var hasAccess bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM project_members
			WHERE project_id = ? AND user_id = ?
		)
	`
	err := s.db.QueryRowContext(ctx, query, projectID, userID).Scan(&hasAccess)
	return hasAccess, err
}
