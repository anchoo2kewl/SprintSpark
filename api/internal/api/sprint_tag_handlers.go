package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

// Sprint represents a sprint
type Sprint struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Name      string    `json:"name"`
	Goal      string    `json:"goal,omitempty"`
	StartDate string    `json:"start_date,omitempty"`
	EndDate   string    `json:"end_date,omitempty"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Tag represents a tag
type Tag struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateSprintRequest represents a request to create a sprint
type CreateSprintRequest struct {
	Name      string `json:"name"`
	Goal      string `json:"goal,omitempty"`
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
	Status    string `json:"status,omitempty"`
}

// UpdateSprintRequest represents a request to update a sprint
type UpdateSprintRequest struct {
	Name      *string `json:"name,omitempty"`
	Goal      *string `json:"goal,omitempty"`
	StartDate *string `json:"start_date,omitempty"`
	EndDate   *string `json:"end_date,omitempty"`
	Status    *string `json:"status,omitempty"`
}

// CreateTagRequest represents a request to create a tag
type CreateTagRequest struct {
	Name  string `json:"name"`
	Color string `json:"color,omitempty"`
}

// UpdateTagRequest represents a request to update a tag
type UpdateTagRequest struct {
	Name  *string `json:"name,omitempty"`
	Color *string `json:"color,omitempty"`
}

// HandleListSprints returns all sprints for the current user
func (s *Server) HandleListSprints(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := `
		SELECT id, user_id, name, COALESCE(goal, ''), COALESCE(start_date, ''), COALESCE(end_date, ''), status, created_at, updated_at
		FROM sprints
		WHERE user_id = ?
		ORDER BY
			CASE status
				WHEN 'active' THEN 1
				WHEN 'planned' THEN 2
				WHEN 'completed' THEN 3
			END,
			start_date DESC,
			created_at DESC
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		http.Error(w, "Failed to fetch sprints", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	sprints := []Sprint{}
	for rows.Next() {
		var sp Sprint
		if err := rows.Scan(&sp.ID, &sp.UserID, &sp.Name, &sp.Goal, &sp.StartDate, &sp.EndDate, &sp.Status, &sp.CreatedAt, &sp.UpdatedAt); err != nil {
			http.Error(w, "Failed to scan sprint", http.StatusInternalServerError)
			return
		}
		sprints = append(sprints, sp)
	}

	respondJSON(w, http.StatusOK, sprints)
}

// HandleCreateSprint creates a new sprint
func (s *Server) HandleCreateSprint(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateSprintRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Sprint name is required", http.StatusBadRequest)
		return
	}

	status := req.Status
	if status == "" {
		status = "planned"
	}

	// Validate status
	if status != "planned" && status != "active" && status != "completed" {
		http.Error(w, "Invalid status. Must be planned, active, or completed", http.StatusBadRequest)
		return
	}

	result, err := s.db.Exec(`
		INSERT INTO sprints (user_id, name, goal, start_date, end_date, status)
		VALUES (?, ?, ?, ?, ?, ?)
	`, userID, req.Name, req.Goal, req.StartDate, req.EndDate, status)

	if err != nil {
		http.Error(w, "Failed to create sprint", http.StatusInternalServerError)
		return
	}

	sprintID, _ := result.LastInsertId()

	// Fetch the created sprint
	var sprint Sprint
	err = s.db.QueryRow(`
		SELECT id, user_id, name, COALESCE(goal, ''), COALESCE(start_date, ''), COALESCE(end_date, ''), status, created_at, updated_at
		FROM sprints WHERE id = ?
	`, sprintID).Scan(&sprint.ID, &sprint.UserID, &sprint.Name, &sprint.Goal, &sprint.StartDate, &sprint.EndDate, &sprint.Status, &sprint.CreatedAt, &sprint.UpdatedAt)

	if err != nil {
		http.Error(w, "Failed to fetch created sprint", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusCreated, sprint)
}

// HandleUpdateSprint updates a sprint
func (s *Server) HandleUpdateSprint(w http.ResponseWriter, r *http.Request) {
	sprintID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid sprint ID", http.StatusBadRequest)
		return
	}

	userID, ok := GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if sprint belongs to user
	var ownerID int
	err = s.db.QueryRow("SELECT user_id FROM sprints WHERE id = ?", sprintID).Scan(&ownerID)
	if err == sql.ErrNoRows {
		http.Error(w, "Sprint not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if int64(ownerID) != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req UpdateSprintRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Build update query dynamically
	updates := []string{}
	args := []interface{}{}

	if req.Name != nil {
		updates = append(updates, "name = ?")
		args = append(args, *req.Name)
	}
	if req.Goal != nil {
		updates = append(updates, "goal = ?")
		args = append(args, *req.Goal)
	}
	if req.StartDate != nil {
		updates = append(updates, "start_date = ?")
		args = append(args, *req.StartDate)
	}
	if req.EndDate != nil {
		updates = append(updates, "end_date = ?")
		args = append(args, *req.EndDate)
	}
	if req.Status != nil {
		// Validate status
		if *req.Status != "planned" && *req.Status != "active" && *req.Status != "completed" {
			http.Error(w, "Invalid status", http.StatusBadRequest)
			return
		}
		updates = append(updates, "status = ?")
		args = append(args, *req.Status)
	}

	if len(updates) == 0 {
		http.Error(w, "No fields to update", http.StatusBadRequest)
		return
	}

	args = append(args, sprintID)
	query := "UPDATE sprints SET " + updates[0]
	for i := 1; i < len(updates); i++ {
		query += ", " + updates[i]
	}
	query += " WHERE id = ?"

	_, err = s.db.Exec(query, args...)
	if err != nil {
		http.Error(w, "Failed to update sprint", http.StatusInternalServerError)
		return
	}

	// Fetch updated sprint
	var sprint Sprint
	err = s.db.QueryRow(`
		SELECT id, user_id, name, COALESCE(goal, ''), COALESCE(start_date, ''), COALESCE(end_date, ''), status, created_at, updated_at
		FROM sprints WHERE id = ?
	`, sprintID).Scan(&sprint.ID, &sprint.UserID, &sprint.Name, &sprint.Goal, &sprint.StartDate, &sprint.EndDate, &sprint.Status, &sprint.CreatedAt, &sprint.UpdatedAt)

	if err != nil {
		http.Error(w, "Failed to fetch updated sprint", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, sprint)
}

// HandleDeleteSprint deletes a sprint
func (s *Server) HandleDeleteSprint(w http.ResponseWriter, r *http.Request) {
	sprintID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid sprint ID", http.StatusBadRequest)
		return
	}

	userID, ok := GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if sprint belongs to user
	var ownerID int
	err = s.db.QueryRow("SELECT user_id FROM sprints WHERE id = ?", sprintID).Scan(&ownerID)
	if err == sql.ErrNoRows {
		http.Error(w, "Sprint not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if int64(ownerID) != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	_, err = s.db.Exec("DELETE FROM sprints WHERE id = ?", sprintID)
	if err != nil {
		http.Error(w, "Failed to delete sprint", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Sprint deleted successfully"})
}

// HandleListTags returns all tags for the current user
func (s *Server) HandleListTags(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := `
		SELECT id, user_id, name, color, created_at
		FROM tags
		WHERE user_id = ?
		ORDER BY name ASC
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		http.Error(w, "Failed to fetch tags", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	tags := []Tag{}
	for rows.Next() {
		var tag Tag
		if err := rows.Scan(&tag.ID, &tag.UserID, &tag.Name, &tag.Color, &tag.CreatedAt); err != nil {
			http.Error(w, "Failed to scan tag", http.StatusInternalServerError)
			return
		}
		tags = append(tags, tag)
	}

	respondJSON(w, http.StatusOK, tags)
}

// HandleCreateTag creates a new tag
func (s *Server) HandleCreateTag(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Tag name is required", http.StatusBadRequest)
		return
	}

	color := req.Color
	if color == "" {
		color = "#3B82F6"
	}

	result, err := s.db.Exec(`
		INSERT INTO tags (user_id, name, color)
		VALUES (?, ?, ?)
	`, userID, req.Name, color)

	if err != nil {
		http.Error(w, "Failed to create tag. Tag name must be unique.", http.StatusConflict)
		return
	}

	tagID, _ := result.LastInsertId()

	// Fetch the created tag
	var tag Tag
	err = s.db.QueryRow(`
		SELECT id, user_id, name, color, created_at
		FROM tags WHERE id = ?
	`, tagID).Scan(&tag.ID, &tag.UserID, &tag.Name, &tag.Color, &tag.CreatedAt)

	if err != nil {
		http.Error(w, "Failed to fetch created tag", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusCreated, tag)
}

// HandleUpdateTag updates a tag
func (s *Server) HandleUpdateTag(w http.ResponseWriter, r *http.Request) {
	tagID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid tag ID", http.StatusBadRequest)
		return
	}

	userID, ok := GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if tag belongs to user
	var ownerID int
	err = s.db.QueryRow("SELECT user_id FROM tags WHERE id = ?", tagID).Scan(&ownerID)
	if err == sql.ErrNoRows {
		http.Error(w, "Tag not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if int64(ownerID) != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req UpdateTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Build update query dynamically
	updates := []string{}
	args := []interface{}{}

	if req.Name != nil {
		updates = append(updates, "name = ?")
		args = append(args, *req.Name)
	}
	if req.Color != nil {
		updates = append(updates, "color = ?")
		args = append(args, *req.Color)
	}

	if len(updates) == 0 {
		http.Error(w, "No fields to update", http.StatusBadRequest)
		return
	}

	args = append(args, tagID)
	query := "UPDATE tags SET " + updates[0]
	for i := 1; i < len(updates); i++ {
		query += ", " + updates[i]
	}
	query += " WHERE id = ?"

	_, err = s.db.Exec(query, args...)
	if err != nil {
		http.Error(w, "Failed to update tag", http.StatusInternalServerError)
		return
	}

	// Fetch updated tag
	var tag Tag
	err = s.db.QueryRow(`
		SELECT id, user_id, name, color, created_at
		FROM tags WHERE id = ?
	`, tagID).Scan(&tag.ID, &tag.UserID, &tag.Name, &tag.Color, &tag.CreatedAt)

	if err != nil {
		http.Error(w, "Failed to fetch updated tag", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, tag)
}

// HandleDeleteTag deletes a tag
func (s *Server) HandleDeleteTag(w http.ResponseWriter, r *http.Request) {
	tagID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid tag ID", http.StatusBadRequest)
		return
	}

	userID, ok := GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if tag belongs to user
	var ownerID int
	err = s.db.QueryRow("SELECT user_id FROM tags WHERE id = ?", tagID).Scan(&ownerID)
	if err == sql.ErrNoRows {
		http.Error(w, "Tag not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if int64(ownerID) != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	_, err = s.db.Exec("DELETE FROM tags WHERE id = ?", tagID)
	if err != nil {
		http.Error(w, "Failed to delete tag", http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Tag deleted successfully"})
}
