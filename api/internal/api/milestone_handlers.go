package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

// Milestone represents a milestone in the API response.
type Milestone struct {
	ID          int64     `json:"id"`
	ProjectID   int64     `json:"project_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Color       string    `json:"color"`
	TargetDate  string    `json:"target_date,omitempty"`
	Status      string    `json:"status"`
	SortOrder   int       `json:"sort_order"`
	TaskCount   int       `json:"task_count,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// MilestoneProgress represents computed progress for a milestone.
type MilestoneProgress struct {
	MilestoneID    int64                    `json:"milestone_id"`
	MilestoneName  string                   `json:"milestone_name"`
	TotalTasks     int                      `json:"total_tasks"`
	CompletedTasks int                      `json:"completed_tasks"`
	Percentage     float64                  `json:"percentage"`
	ByStatus       map[string]int           `json:"by_status"`
	ByAssignee     []MilestoneAssigneeStats `json:"by_assignee"`
	EstimatedHours float64                  `json:"estimated_hours"`
	ActualHours    float64                  `json:"actual_hours"`
}

// MilestoneAssigneeStats represents task stats per assignee within a milestone.
type MilestoneAssigneeStats struct {
	UserID    int64  `json:"user_id"`
	UserName  string `json:"user_name"`
	TaskCount int    `json:"task_count"`
	Completed int    `json:"completed"`
}

// CreateMilestoneRequest represents a request to create a milestone.
type CreateMilestoneRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
	TargetDate  string `json:"target_date,omitempty"`
	Status      string `json:"status,omitempty"`
	SortOrder   *int   `json:"sort_order,omitempty"`
}

// UpdateMilestoneRequest represents a request to update a milestone.
type UpdateMilestoneRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Color       *string `json:"color,omitempty"`
	TargetDate  *string `json:"target_date,omitempty"`
	Status      *string `json:"status,omitempty"`
	SortOrder   *int    `json:"sort_order,omitempty"`
}

const milestoneSelectCols = `id, project_id, name, description, color, target_date, status, sort_order, created_at, updated_at`

func scanMilestone(row interface{ Scan(...interface{}) error }) (Milestone, error) {
	var m Milestone
	var description *string
	var targetDate *time.Time
	if err := row.Scan(&m.ID, &m.ProjectID, &m.Name, &description, &m.Color, &targetDate, &m.Status, &m.SortOrder, &m.CreatedAt, &m.UpdatedAt); err != nil {
		return m, err
	}
	if description != nil {
		m.Description = *description
	}
	if targetDate != nil {
		m.TargetDate = targetDate.Format("2006-01-02")
	}
	return m, nil
}

// HandleListMilestones returns all milestones for a project with task counts.
// Route: GET /api/projects/{id}/milestones
func (s *Server) HandleListMilestones(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	projectID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project ID", "bad_request")
		return
	}

	hasAccess, err := s.checkProjectAccess(ctx, userID, projectID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to check access", "internal_error")
		return
	}
	if !hasAccess {
		respondError(w, http.StatusForbidden, "Forbidden", "forbidden")
		return
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT `+milestoneSelectCols+`,
		        (SELECT COUNT(*) FROM tasks WHERE tasks.milestone_id = milestones.id) AS task_count
		 FROM milestones
		 WHERE project_id = $1
		 ORDER BY sort_order ASC, created_at ASC`,
		projectID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch milestones", "internal_error")
		return
	}
	defer rows.Close()

	milestones := []Milestone{}
	for rows.Next() {
		var m Milestone
		var description *string
		var targetDate *time.Time
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.Name, &description, &m.Color, &targetDate, &m.Status, &m.SortOrder, &m.CreatedAt, &m.UpdatedAt, &m.TaskCount); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to scan milestone", "internal_error")
			return
		}
		if description != nil {
			m.Description = *description
		}
		if targetDate != nil {
			m.TargetDate = targetDate.Format("2006-01-02")
		}
		milestones = append(milestones, m)
	}

	respondJSON(w, http.StatusOK, milestones)
}

// HandleCreateMilestone creates a new milestone in a project.
// Route: POST /api/projects/{id}/milestones
func (s *Server) HandleCreateMilestone(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	projectID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project ID", "bad_request")
		return
	}

	hasAccess, err := s.checkProjectAccess(ctx, userID, projectID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to check access", "internal_error")
		return
	}
	if !hasAccess {
		respondError(w, http.StatusForbidden, "Forbidden", "forbidden")
		return
	}

	var req CreateMilestoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Milestone name is required", http.StatusBadRequest)
		return
	}

	status := req.Status
	if status == "" {
		status = "active"
	}
	if status != "active" && status != "completed" && status != "cancelled" {
		http.Error(w, "Invalid status. Must be active, completed, or cancelled", http.StatusBadRequest)
		return
	}

	color := req.Color
	if color == "" {
		color = "#5e6ad2"
	}

	var descriptionVal *string
	if req.Description != "" {
		descriptionVal = &req.Description
	}
	var targetDateVal *string
	if req.TargetDate != "" {
		targetDateVal = &req.TargetDate
	}

	sortOrder := 0
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}

	var newID int64
	err = s.db.QueryRowContext(ctx,
		`INSERT INTO milestones (project_id, name, description, color, target_date, status, sort_order)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		projectID, req.Name, descriptionVal, color, targetDateVal, status, sortOrder,
	).Scan(&newID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create milestone", "internal_error")
		return
	}

	m, err := scanMilestone(s.db.QueryRowContext(ctx,
		`SELECT `+milestoneSelectCols+` FROM milestones WHERE id = $1`, newID))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch new milestone", "internal_error")
		return
	}

	respondJSON(w, http.StatusCreated, m)
}

// HandleGetMilestone returns a single milestone by ID.
// Route: GET /api/milestones/{id}
func (s *Server) HandleGetMilestone(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	milestoneID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid milestone ID", http.StatusBadRequest)
		return
	}

	userID, ok := GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var projectID int64
	err = s.db.QueryRowContext(ctx, `SELECT project_id FROM milestones WHERE id = $1`, milestoneID).Scan(&projectID)
	if err == sql.ErrNoRows {
		http.Error(w, "Milestone not found", http.StatusNotFound)
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

	m, err := scanMilestone(s.db.QueryRowContext(ctx,
		`SELECT `+milestoneSelectCols+` FROM milestones WHERE id = $1`, milestoneID))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch milestone", "internal_error")
		return
	}

	respondJSON(w, http.StatusOK, m)
}

// HandleUpdateMilestone updates a milestone.
// Route: PATCH /api/milestones/{id}
func (s *Server) HandleUpdateMilestone(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	milestoneID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid milestone ID", http.StatusBadRequest)
		return
	}

	userID, ok := GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var projectID int64
	err = s.db.QueryRowContext(ctx, `SELECT project_id FROM milestones WHERE id = $1`, milestoneID).Scan(&projectID)
	if err == sql.ErrNoRows {
		http.Error(w, "Milestone not found", http.StatusNotFound)
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

	var req UpdateMilestoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updateBuilder := s.db.Client.Milestone.UpdateOneID(milestoneID)
	hasUpdates := false

	if req.Name != nil {
		updateBuilder.SetName(*req.Name)
		hasUpdates = true
	}
	if req.Description != nil {
		if *req.Description == "" {
			updateBuilder.ClearDescription()
		} else {
			updateBuilder.SetDescription(*req.Description)
		}
		hasUpdates = true
	}
	if req.Color != nil {
		updateBuilder.SetColor(*req.Color)
		hasUpdates = true
	}
	if req.TargetDate != nil {
		if *req.TargetDate != "" {
			targetDate, err := time.Parse("2006-01-02", *req.TargetDate)
			if err == nil {
				updateBuilder.SetTargetDate(targetDate)
				hasUpdates = true
			}
		} else {
			updateBuilder.ClearTargetDate()
			hasUpdates = true
		}
	}
	if req.Status != nil {
		if *req.Status != "active" && *req.Status != "completed" && *req.Status != "cancelled" {
			http.Error(w, "Invalid status. Must be active, completed, or cancelled", http.StatusBadRequest)
			return
		}
		updateBuilder.SetStatus(*req.Status)
		hasUpdates = true
	}
	if req.SortOrder != nil {
		updateBuilder.SetSortOrder(*req.SortOrder)
		hasUpdates = true
	}

	if !hasUpdates {
		http.Error(w, "No fields to update", http.StatusBadRequest)
		return
	}

	updatedMilestone, err := updateBuilder.Save(ctx)
	if err != nil {
		http.Error(w, "Failed to update milestone", http.StatusInternalServerError)
		return
	}

	m := Milestone{
		ID:        updatedMilestone.ID,
		ProjectID: updatedMilestone.ProjectID,
		Name:      updatedMilestone.Name,
		Color:     updatedMilestone.Color,
		Status:    updatedMilestone.Status,
		SortOrder: updatedMilestone.SortOrder,
		CreatedAt: updatedMilestone.CreatedAt,
		UpdatedAt: updatedMilestone.UpdatedAt,
	}
	if updatedMilestone.Description != nil {
		m.Description = *updatedMilestone.Description
	}
	if updatedMilestone.TargetDate != nil {
		m.TargetDate = updatedMilestone.TargetDate.Format("2006-01-02")
	}

	respondJSON(w, http.StatusOK, m)
}

// HandleDeleteMilestone deletes a milestone (tasks get milestone_id set to NULL).
// Route: DELETE /api/milestones/{id}
func (s *Server) HandleDeleteMilestone(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	milestoneID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid milestone ID", http.StatusBadRequest)
		return
	}

	userID, ok := GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var projectID int64
	err = s.db.QueryRowContext(ctx, `SELECT project_id FROM milestones WHERE id = $1`, milestoneID).Scan(&projectID)
	if err == sql.ErrNoRows {
		http.Error(w, "Milestone not found", http.StatusNotFound)
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

	err = s.db.Client.Milestone.DeleteOneID(milestoneID).Exec(ctx)
	if err != nil {
		http.Error(w, "Failed to delete milestone", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Milestone deleted successfully"})
}

// HandleGetMilestoneProgress returns computed progress stats for a milestone.
// Route: GET /api/milestones/{id}/progress
func (s *Server) HandleGetMilestoneProgress(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	milestoneID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid milestone ID", http.StatusBadRequest)
		return
	}

	userID, ok := GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var projectID int64
	var milestoneName string
	err = s.db.QueryRowContext(ctx, `SELECT project_id, name FROM milestones WHERE id = $1`, milestoneID).Scan(&projectID, &milestoneName)
	if err == sql.ErrNoRows {
		http.Error(w, "Milestone not found", http.StatusNotFound)
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

	// Count tasks by status (using swim_lane status_category)
	byStatus := make(map[string]int)
	totalTasks := 0
	completedTasks := 0

	statusRows, err := s.db.QueryContext(ctx,
		`SELECT COALESCE(sl.status_category, t.status) AS effective_status, COUNT(*)
		 FROM tasks t
		 LEFT JOIN swim_lanes sl ON t.swim_lane_id = sl.id
		 WHERE t.milestone_id = $1
		 GROUP BY effective_status`, milestoneID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to compute progress", "internal_error")
		return
	}
	defer statusRows.Close()

	for statusRows.Next() {
		var status string
		var count int
		if err := statusRows.Scan(&status, &count); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to scan status", "internal_error")
			return
		}
		byStatus[status] = count
		totalTasks += count
		if status == "done" {
			completedTasks = count
		}
	}

	// Hours aggregation
	var estimatedHours, actualHours float64
	err = s.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(estimated_hours), 0), COALESCE(SUM(actual_hours), 0)
		 FROM tasks WHERE milestone_id = $1`, milestoneID).Scan(&estimatedHours, &actualHours)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to compute hours", "internal_error")
		return
	}

	// By assignee (includes both task_assignees and legacy assignee_id)
	assigneeRows, err := s.db.QueryContext(ctx,
		`WITH all_assignees AS (
			SELECT ta.user_id, t.id AS task_id, COALESCE(sl.status_category, t.status) AS effective_status
			FROM tasks t
			JOIN task_assignees ta ON ta.task_id = t.id
			LEFT JOIN swim_lanes sl ON t.swim_lane_id = sl.id
			WHERE t.milestone_id = $1
			UNION
			SELECT t.assignee_id AS user_id, t.id AS task_id, COALESCE(sl.status_category, t.status) AS effective_status
			FROM tasks t
			LEFT JOIN swim_lanes sl ON t.swim_lane_id = sl.id
			WHERE t.milestone_id = $1 AND t.assignee_id IS NOT NULL
			  AND NOT EXISTS (SELECT 1 FROM task_assignees ta2 WHERE ta2.task_id = t.id)
		)
		SELECT u.id, COALESCE(u.name, u.email) AS user_name,
		       COUNT(DISTINCT aa.task_id) AS task_count,
		       COUNT(DISTINCT CASE WHEN aa.effective_status = 'done' THEN aa.task_id END) AS completed
		FROM all_assignees aa
		JOIN users u ON u.id = aa.user_id
		GROUP BY u.id, user_name
		ORDER BY task_count DESC`, milestoneID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch assignee stats", "internal_error")
		return
	}
	defer assigneeRows.Close()

	assigneeStats := []MilestoneAssigneeStats{}
	for assigneeRows.Next() {
		var stat MilestoneAssigneeStats
		if err := assigneeRows.Scan(&stat.UserID, &stat.UserName, &stat.TaskCount, &stat.Completed); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to scan assignee stats", "internal_error")
			return
		}
		assigneeStats = append(assigneeStats, stat)
	}

	percentage := 0.0
	if totalTasks > 0 {
		percentage = float64(completedTasks) / float64(totalTasks) * 100
	}

	progress := MilestoneProgress{
		MilestoneID:    milestoneID,
		MilestoneName:  milestoneName,
		TotalTasks:     totalTasks,
		CompletedTasks: completedTasks,
		Percentage:     percentage,
		ByStatus:       byStatus,
		ByAssignee:     assigneeStats,
		EstimatedHours: estimatedHours,
		ActualHours:    actualHours,
	}

	respondJSON(w, http.StatusOK, progress)
}
