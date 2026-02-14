package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

// stringPtr returns a pointer to a string value
func stringPtr(s string) *string {
	return &s
}

func TestHandleListTasks(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "test@example.com", "password123")
	projectID := ts.CreateTestProject(t, userID, "Test Project")

	// Create test tasks
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := ts.DB.ExecContext(ctx,
		`INSERT INTO tasks (project_id, task_number, title, description, status, priority) VALUES (?, ?, ?, ?, ?, ?)`,
		projectID, 1, "Task 1", "Description 1", "todo", "high",
	)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	_, err = ts.DB.ExecContext(ctx,
		`INSERT INTO tasks (project_id, task_number, title, description, status, priority) VALUES (?, ?, ?, ?, ?, ?)`,
		projectID, 2, "Task 2", "Description 2", "in_progress", "medium",
	)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	rec, req := ts.MakeAuthRequest(t, http.MethodGet, "/api/projects/1/tasks", nil, userID,
		map[string]string{"projectId": fmt.Sprintf("%d", projectID)})

	ts.HandleListTasks(rec, req)

	AssertStatusCode(t, rec.Code, http.StatusOK)

	var tasks []Task
	DecodeJSON(t, rec, &tasks)

	if len(tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(tasks))
	}

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
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "test@example.com", "password123")
	projectID := ts.CreateTestProject(t, userID, "Test Project")

	taskData := CreateTaskRequest{
		Title:       "New Task",
		Description: stringPtr("Task description with **markdown**"),
		Status:      stringPtr("todo"),
		Priority:    stringPtr("urgent"),
	}

	rec, req := ts.MakeAuthRequest(t, http.MethodPost, "/api/projects/1/tasks", taskData, userID,
		map[string]string{"projectId": fmt.Sprintf("%d", projectID)})

	ts.HandleCreateTask(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status Created (201), got %d: %s", rec.Code, rec.Body.String())
	}

	var createdTask Task
	DecodeJSON(t, rec, &createdTask)

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

func TestHandleCreateTaskValidation(t *testing.T) {
	tests := []struct {
		name       string
		body       CreateTaskRequest
		wantStatus int
		wantError  string
	}{
		{
			name:       "missing title",
			body:       CreateTaskRequest{Title: ""},
			wantStatus: http.StatusBadRequest,
			wantError:  "task title is required",
		},
		{
			name:       "invalid status",
			body:       CreateTaskRequest{Title: "Test", Status: stringPtr("invalid")},
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid status",
		},
		{
			name:       "invalid priority",
			body:       CreateTaskRequest{Title: "Test", Priority: stringPtr("critical")},
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid priority",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := NewTestServer(t)
			defer ts.Close()

			userID := ts.CreateTestUser(t, "test@example.com", "password123")
			projectID := ts.CreateTestProject(t, userID, "Test Project")

			rec, req := ts.MakeAuthRequest(t, http.MethodPost, "/api/projects/1/tasks", tt.body, userID,
				map[string]string{"projectId": fmt.Sprintf("%d", projectID)})

			ts.HandleCreateTask(rec, req)

			AssertStatusCode(t, rec.Code, tt.wantStatus)
			AssertError(t, rec, tt.wantStatus, tt.wantError, "")
		})
	}
}

func TestHandleUpdateTask(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "test@example.com", "password123")
	projectID := ts.CreateTestProject(t, userID, "Test Project")
	taskID := ts.CreateTestTask(t, projectID, "Original Task")

	updateData := UpdateTaskRequest{
		Title:    stringPtr("Updated Task"),
		Status:   stringPtr("in_progress"),
		Priority: stringPtr("high"),
	}

	rec, req := ts.MakeAuthRequest(t, http.MethodPatch, fmt.Sprintf("/api/tasks/%d", taskID), updateData, userID,
		map[string]string{"id": fmt.Sprintf("%d", taskID)})

	ts.HandleUpdateTask(rec, req)

	AssertStatusCode(t, rec.Code, http.StatusOK)

	var updatedTask Task
	DecodeJSON(t, rec, &updatedTask)

	if updatedTask.Title != "Updated Task" {
		t.Errorf("Expected title 'Updated Task', got '%s'", updatedTask.Title)
	}

	if updatedTask.Status != "in_progress" {
		t.Errorf("Expected status 'in_progress', got '%s'", updatedTask.Status)
	}

	if updatedTask.Priority != "high" {
		t.Errorf("Expected priority 'high', got '%s'", updatedTask.Priority)
	}
}

func TestHandleUpdateTaskValidation(t *testing.T) {
	tests := []struct {
		name       string
		body       UpdateTaskRequest
		wantStatus int
		wantError  string
	}{
		{
			name:       "empty title",
			body:       UpdateTaskRequest{Title: stringPtr("")},
			wantStatus: http.StatusBadRequest,
			wantError:  "title cannot be empty",
		},
		{
			name:       "invalid status",
			body:       UpdateTaskRequest{Status: stringPtr("invalid")},
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid status",
		},
		{
			name:       "invalid priority",
			body:       UpdateTaskRequest{Priority: stringPtr("critical")},
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid priority",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := NewTestServer(t)
			defer ts.Close()

			userID := ts.CreateTestUser(t, "test@example.com", "password123")
			projectID := ts.CreateTestProject(t, userID, "Test Project")
			taskID := ts.CreateTestTask(t, projectID, "Test Task")

			rec, req := ts.MakeAuthRequest(t, http.MethodPatch, fmt.Sprintf("/api/tasks/%d", taskID), tt.body, userID,
				map[string]string{"id": fmt.Sprintf("%d", taskID)})

			ts.HandleUpdateTask(rec, req)

			AssertStatusCode(t, rec.Code, tt.wantStatus)
			AssertError(t, rec, tt.wantStatus, tt.wantError, "")
		})
	}
}

func TestHandleUpdateTaskNotFound(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "test@example.com", "password123")

	updateData := UpdateTaskRequest{
		Title: stringPtr("Updated"),
	}

	rec, req := ts.MakeAuthRequest(t, http.MethodPatch, "/api/tasks/99999", updateData, userID,
		map[string]string{"id": "99999"})

	ts.HandleUpdateTask(rec, req)

	AssertStatusCode(t, rec.Code, http.StatusNotFound)
}

func TestHandleDeleteTask(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "test@example.com", "password123")
	projectID := ts.CreateTestProject(t, userID, "Test Project")
	taskID := ts.CreateTestTask(t, projectID, "Task to Delete")

	rec, req := ts.MakeAuthRequest(t, http.MethodDelete, fmt.Sprintf("/api/tasks/%d", taskID), nil, userID,
		map[string]string{"id": fmt.Sprintf("%d", taskID)})

	ts.HandleDeleteTask(rec, req)

	AssertStatusCode(t, rec.Code, http.StatusNoContent)

	// Verify task was deleted
	var count int
	err := ts.DB.QueryRow("SELECT COUNT(*) FROM tasks WHERE id = ?", taskID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query task count: %v", err)
	}

	if count != 0 {
		t.Errorf("Task was not deleted from database")
	}
}

func TestHandleDeleteTaskNotFound(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "test@example.com", "password123")

	rec, req := ts.MakeAuthRequest(t, http.MethodDelete, "/api/tasks/99999", nil, userID,
		map[string]string{"id": "99999"})

	ts.HandleDeleteTask(rec, req)

	AssertStatusCode(t, rec.Code, http.StatusNotFound)
}

func TestHandleListTasksWithTags(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "test@example.com", "password123")
	projectID := ts.CreateTestProject(t, userID, "Test Project")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create tags
	tagResult1, _ := ts.DB.ExecContext(ctx, "INSERT INTO tags (user_id, name, color) VALUES (?, ?, ?)", userID, "bug", "#FF0000")
	tagID1, _ := tagResult1.LastInsertId()

	tagResult2, _ := ts.DB.ExecContext(ctx, "INSERT INTO tags (user_id, name, color) VALUES (?, ?, ?)", userID, "feature", "#00FF00")
	tagID2, _ := tagResult2.LastInsertId()

	// Create task
	taskID := ts.CreateTestTask(t, projectID, "Tagged Task")

	// Associate tags
	_, _ = ts.DB.ExecContext(ctx, "INSERT INTO task_tags (task_id, tag_id) VALUES (?, ?)", taskID, tagID1)
	_, _ = ts.DB.ExecContext(ctx, "INSERT INTO task_tags (task_id, tag_id) VALUES (?, ?)", taskID, tagID2)

	rec, req := ts.MakeAuthRequest(t, http.MethodGet, "/api/projects/1/tasks", nil, userID,
		map[string]string{"projectId": fmt.Sprintf("%d", projectID)})

	ts.HandleListTasks(rec, req)

	AssertStatusCode(t, rec.Code, http.StatusOK)

	var tasks []Task
	DecodeJSON(t, rec, &tasks)

	if len(tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(tasks))
	}

	if len(tasks[0].Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(tasks[0].Tags))
	}

	tagNames := make(map[string]bool)
	for _, tag := range tasks[0].Tags {
		tagNames[tag.Name] = true
	}

	if !tagNames["bug"] || !tagNames["feature"] {
		t.Errorf("Expected tags 'bug' and 'feature', got %v", tagNames)
	}
}

func TestHandleListTasksUnauthorized(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	user1ID := ts.CreateTestUser(t, "user1@example.com", "password123")
	user2ID := ts.CreateTestUser(t, "user2@example.com", "password123")
	projectID := ts.CreateTestProject(t, user1ID, "User1 Project")

	// user2 is NOT a member of user1's project
	rec, req := ts.MakeAuthRequest(t, http.MethodGet, "/api/projects/1/tasks", nil, user2ID,
		map[string]string{"projectId": fmt.Sprintf("%d", projectID)})

	ts.HandleListTasks(rec, req)

	AssertStatusCode(t, rec.Code, http.StatusForbidden)
}

func TestHandleCreateTaskUnauthorized(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	user1ID := ts.CreateTestUser(t, "user1@example.com", "password123")
	user2ID := ts.CreateTestUser(t, "user2@example.com", "password123")
	projectID := ts.CreateTestProject(t, user1ID, "User1 Project")

	taskData := CreateTaskRequest{
		Title: "Unauthorized Task",
	}

	rec, req := ts.MakeAuthRequest(t, http.MethodPost, "/api/projects/1/tasks", taskData, user2ID,
		map[string]string{"projectId": fmt.Sprintf("%d", projectID)})

	ts.HandleCreateTask(rec, req)

	AssertStatusCode(t, rec.Code, http.StatusForbidden)
}

func TestHandleUpdateTaskUnauthorized(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	user1ID := ts.CreateTestUser(t, "user1@example.com", "password123")
	user2ID := ts.CreateTestUser(t, "user2@example.com", "password123")
	projectID := ts.CreateTestProject(t, user1ID, "User1 Project")
	taskID := ts.CreateTestTask(t, projectID, "Task")

	updateData := UpdateTaskRequest{
		Title: stringPtr("Hacked Title"),
	}

	rec, req := ts.MakeAuthRequest(t, http.MethodPatch, fmt.Sprintf("/api/tasks/%d", taskID), updateData, user2ID,
		map[string]string{"id": fmt.Sprintf("%d", taskID)})

	ts.HandleUpdateTask(rec, req)

	AssertStatusCode(t, rec.Code, http.StatusForbidden)
}

func TestHandleDeleteTaskUnauthorized(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	user1ID := ts.CreateTestUser(t, "user1@example.com", "password123")
	user2ID := ts.CreateTestUser(t, "user2@example.com", "password123")
	projectID := ts.CreateTestProject(t, user1ID, "User1 Project")
	taskID := ts.CreateTestTask(t, projectID, "Task")

	rec, req := ts.MakeAuthRequest(t, http.MethodDelete, fmt.Sprintf("/api/tasks/%d", taskID), nil, user2ID,
		map[string]string{"id": fmt.Sprintf("%d", taskID)})

	ts.HandleDeleteTask(rec, req)

	AssertStatusCode(t, rec.Code, http.StatusForbidden)
}

func TestHandleListTasksInvalidProjectID(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "test@example.com", "password123")

	rec, req := ts.MakeAuthRequest(t, http.MethodGet, "/api/projects/abc/tasks", nil, userID,
		map[string]string{"projectId": "abc"})

	ts.HandleListTasks(rec, req)

	AssertStatusCode(t, rec.Code, http.StatusBadRequest)
}

func TestHandleCreateTaskWithTags(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "test@example.com", "password123")
	projectID := ts.CreateTestProject(t, userID, "Test Project")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create tags
	tagResult1, _ := ts.DB.ExecContext(ctx, "INSERT INTO tags (user_id, name, color) VALUES (?, ?, ?)", userID, "bug", "#FF0000")
	tagID1, _ := tagResult1.LastInsertId()

	tagResult2, _ := ts.DB.ExecContext(ctx, "INSERT INTO tags (user_id, name, color) VALUES (?, ?, ?)", userID, "feature", "#00FF00")
	tagID2, _ := tagResult2.LastInsertId()

	taskData := CreateTaskRequest{
		Title:  "Tagged Task",
		TagIDs: []int64{tagID1, tagID2},
	}

	rec, req := ts.MakeAuthRequest(t, http.MethodPost, "/api/projects/1/tasks", taskData, userID,
		map[string]string{"projectId": fmt.Sprintf("%d", projectID)})

	ts.HandleCreateTask(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var createdTask Task
	DecodeJSON(t, rec, &createdTask)

	if len(createdTask.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(createdTask.Tags))
	}
}

func TestHandleListTasksEmptyProject(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "test@example.com", "password123")
	projectID := ts.CreateTestProject(t, userID, "Empty Project")

	rec, req := ts.MakeAuthRequest(t, http.MethodGet, "/api/projects/1/tasks", nil, userID,
		map[string]string{"projectId": fmt.Sprintf("%d", projectID)})

	ts.HandleListTasks(rec, req)

	AssertStatusCode(t, rec.Code, http.StatusOK)

	var tasks []json.RawMessage
	DecodeJSON(t, rec, &tasks)

	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks for empty project, got %d", len(tasks))
	}
}
