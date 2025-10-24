package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	_ "modernc.org/sqlite"
	"sprintspark/internal/db"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *db.DB {
	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create tables
	schema := `
	CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		name TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE projects (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		owner_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		description TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		project_id INTEGER NOT NULL,
		title TEXT NOT NULL,
		description TEXT,
		status TEXT NOT NULL DEFAULT 'todo' CHECK(status IN ('todo', 'in_progress', 'done')),
		due_date DATE,
		sprint_id INTEGER,
		priority TEXT DEFAULT 'medium' CHECK(priority IN ('low', 'medium', 'high', 'urgent')),
		assignee_id INTEGER,
		estimated_hours REAL,
		actual_hours REAL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
	);

	CREATE TABLE tags (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		color TEXT NOT NULL DEFAULT '#3B82F6',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		UNIQUE(user_id, name)
	);

	CREATE TABLE task_tags (
		task_id INTEGER NOT NULL,
		tag_id INTEGER NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (task_id, tag_id),
		FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
		FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
	);
	`

	if _, err := sqlDB.Exec(schema); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	return &db.DB{DB: sqlDB}
}

// createTestUser creates a test user and returns the user ID
func createTestUser(t *testing.T, database *db.DB) int64 {
	result, err := database.Exec(
		"INSERT INTO users (email, password_hash, name) VALUES (?, ?, ?)",
		"test@example.com", "hashed_password", "Test User",
	)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	userID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get user ID: %v", err)
	}

	return userID
}

// createTestProject creates a test project and returns the project ID
func createTestProject(t *testing.T, database *db.DB, userID int64) int64 {
	result, err := database.Exec(
		"INSERT INTO projects (owner_id, name, description) VALUES (?, ?, ?)",
		userID, "Test Project", "Test Description",
	)
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	projectID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get project ID: %v", err)
	}

	return projectID
}

func TestHandleListTasks(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userID := createTestUser(t, db)
	projectID := createTestProject(t, db, userID)

	// Create test tasks
	_, err := db.Exec(
		`INSERT INTO tasks (project_id, title, description, status, priority) VALUES (?, ?, ?, ?, ?)`,
		projectID, "Task 1", "Description 1", "todo", "high",
	)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	_, err = db.Exec(
		`INSERT INTO tasks (project_id, title, description, status, priority) VALUES (?, ?, ?, ?, ?)`,
		projectID, "Task 2", "Description 2", "in_progress", "medium",
	)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	server := &Server{db: db}

	// Create request
	req := httptest.NewRequest("GET", "/api/projects/1/tasks", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(context.WithValue(req.Context(), UserIDKey, userID))

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	server.HandleListTasks(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}

	var tasks []Task
	if err := json.NewDecoder(w.Body).Decode(&tasks); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(tasks))
	}

	// Verify task data - check that both tasks exist with correct priorities
	taskMap := make(map[string]string)
	for _, task := range tasks {
		taskMap[task.Title] = task.Priority
	}

	if taskMap["Task 1"] != "high" {
		t.Errorf("Expected Task 1 priority 'high', got '%s'", taskMap["Task 1"])
	}

	if taskMap["Task 2"] != "medium" {
		t.Errorf("Expected Task 2 priority 'medium', got '%s'", taskMap["Task 2"])
	}
}

func TestHandleCreateTask(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userID := createTestUser(t, db)
	_ = createTestProject(t, db, userID)

	server := &Server{db: db}

	// Create request body
	taskData := CreateTaskRequest{
		Title:       "New Task",
		Description: stringPtr("Task description with **markdown**"),
		Status:      stringPtr("todo"),
		Priority:    stringPtr("urgent"),
	}

	body, _ := json.Marshal(taskData)
	req := httptest.NewRequest("POST", "/api/projects/1/tasks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(context.WithValue(req.Context(), UserIDKey, userID))

	w := httptest.NewRecorder()

	server.HandleCreateTask(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status Created (201), got %d: %s", w.Code, w.Body.String())
	}

	var createdTask Task
	if err := json.NewDecoder(w.Body).Decode(&createdTask); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if createdTask.Title != "New Task" {
		t.Errorf("Expected title 'New Task', got '%s'", createdTask.Title)
	}

	if createdTask.Priority != "urgent" {
		t.Errorf("Expected priority 'urgent', got '%s'", createdTask.Priority)
	}

	if createdTask.Description == nil || *createdTask.Description != "Task description with **markdown**" {
		t.Errorf("Expected description with markdown")
	}
}

func TestHandleUpdateTask(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userID := createTestUser(t, db)
	projectID := createTestProject(t, db, userID)

	// Create a task
	result, err := db.Exec(
		`INSERT INTO tasks (project_id, title, description, status, priority) VALUES (?, ?, ?, ?, ?)`,
		projectID, "Original Task", "Original description", "todo", "medium",
	)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	taskID, _ := result.LastInsertId()

	server := &Server{db: db}

	// Update task
	updateData := map[string]interface{}{
		"title":    "Updated Task",
		"status":   "in_progress",
		"priority": "high",
	}

	body, _ := json.Marshal(updateData)
	req := httptest.NewRequest("PATCH", "/api/tasks/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(context.WithValue(req.Context(), UserIDKey, userID))

	w := httptest.NewRecorder()

	server.HandleUpdateTask(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d: %s", w.Code, w.Body.String())
	}

	var updatedTask Task
	if err := json.NewDecoder(w.Body).Decode(&updatedTask); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if updatedTask.Title != "Updated Task" {
		t.Errorf("Expected title 'Updated Task', got '%s'", updatedTask.Title)
	}

	if updatedTask.Status != "in_progress" {
		t.Errorf("Expected status 'in_progress', got '%s'", updatedTask.Status)
	}

	if updatedTask.Priority != "high" {
		t.Errorf("Expected priority 'high', got '%s'", updatedTask.Priority)
	}

	// Verify database was updated
	var dbTitle, dbStatus, dbPriority string
	err = db.QueryRow("SELECT title, status, priority FROM tasks WHERE id = ?", taskID).
		Scan(&dbTitle, &dbStatus, &dbPriority)
	if err != nil {
		t.Fatalf("Failed to query updated task: %v", err)
	}

	if dbTitle != "Updated Task" {
		t.Errorf("Database not updated: expected 'Updated Task', got '%s'", dbTitle)
	}
}

func TestHandleDeleteTask(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userID := createTestUser(t, db)
	projectID := createTestProject(t, db, userID)

	// Create a task
	result, err := db.Exec(
		`INSERT INTO tasks (project_id, title, status) VALUES (?, ?, ?)`,
		projectID, "Task to Delete", "todo",
	)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	taskID, _ := result.LastInsertId()

	server := &Server{db: db}

	req := httptest.NewRequest("DELETE", "/api/tasks/1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(context.WithValue(req.Context(), UserIDKey, userID))

	w := httptest.NewRecorder()

	server.HandleDeleteTask(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status NoContent (204), got %d: %s", w.Code, w.Body.String())
	}

	// Verify task was deleted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM tasks WHERE id = ?", taskID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query task count: %v", err)
	}

	if count != 0 {
		t.Errorf("Task was not deleted from database")
	}
}

func TestHandleListTasksWithTags(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userID := createTestUser(t, db)
	projectID := createTestProject(t, db, userID)

	// Create tags
	tagResult1, _ := db.Exec("INSERT INTO tags (user_id, name, color) VALUES (?, ?, ?)", userID, "bug", "#FF0000")
	tagID1, _ := tagResult1.LastInsertId()

	tagResult2, _ := db.Exec("INSERT INTO tags (user_id, name, color) VALUES (?, ?, ?)", userID, "feature", "#00FF00")
	tagID2, _ := tagResult2.LastInsertId()

	// Create task
	taskResult, _ := db.Exec(
		`INSERT INTO tasks (project_id, title, status) VALUES (?, ?, ?)`,
		projectID, "Tagged Task", "todo",
	)
	taskID, _ := taskResult.LastInsertId()

	// Associate tags
	db.Exec("INSERT INTO task_tags (task_id, tag_id) VALUES (?, ?)", taskID, tagID1)
	db.Exec("INSERT INTO task_tags (task_id, tag_id) VALUES (?, ?)", taskID, tagID2)

	server := &Server{db: db}

	req := httptest.NewRequest("GET", "/api/projects/1/tasks", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(context.WithValue(req.Context(), UserIDKey, userID))

	w := httptest.NewRecorder()

	server.HandleListTasks(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}

	var tasks []Task
	if err := json.NewDecoder(w.Body).Decode(&tasks); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(tasks))
	}

	if len(tasks[0].Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(tasks[0].Tags))
	}

	// Verify tag names
	tagNames := make(map[string]bool)
	for _, tag := range tasks[0].Tags {
		tagNames[tag.Name] = true
	}

	if !tagNames["bug"] || !tagNames["feature"] {
		t.Errorf("Expected tags 'bug' and 'feature', got %v", tagNames)
	}
}

func TestHandleListTasksUnauthorized(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	user1 := createTestUser(t, db)
	user2, _ := db.Exec("INSERT INTO users (email, password_hash, name) VALUES (?, ?, ?)",
		"user2@example.com", "hash", "User 2")
	user2ID, _ := user2.LastInsertId()

	_ = createTestProject(t, db, user1)

	server := &Server{db: db}

	// Try to access as different user
	req := httptest.NewRequest("GET", "/api/projects/1/tasks", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("projectId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(context.WithValue(req.Context(), UserIDKey, user2ID))

	w := httptest.NewRecorder()

	server.HandleListTasks(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status Forbidden (403), got %d", w.Code)
	}
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
