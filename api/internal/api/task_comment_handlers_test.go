package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestHandleListTaskCommentsEmpty(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "test@example.com", "password123")
	projectID := ts.CreateTestProject(t, userID, "Test Project")
	taskID := ts.CreateTestTask(t, projectID, "Test Task")

	rec, req := ts.MakeAuthRequest(t, http.MethodGet, fmt.Sprintf("/api/tasks/%d/comments", taskID), nil, userID,
		map[string]string{"taskId": fmt.Sprintf("%d", taskID)})

	ts.HandleListTaskComments(rec, req)

	AssertStatusCode(t, rec.Code, http.StatusOK)

	var comments []TaskComment
	DecodeJSON(t, rec, &comments)

	if len(comments) != 0 {
		t.Errorf("Expected 0 comments, got %d", len(comments))
	}
}

func TestHandleListTaskCommentsWithComments(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "test@example.com", "password123")
	projectID := ts.CreateTestProject(t, userID, "Test Project")
	taskID := ts.CreateTestTask(t, projectID, "Test Task")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Insert comments directly
	_, err := ts.DB.ExecContext(ctx,
		`INSERT INTO task_comments (task_id, user_id, comment) VALUES (?, ?, ?)`,
		taskID, userID, "First comment",
	)
	if err != nil {
		t.Fatalf("Failed to create comment: %v", err)
	}

	_, err = ts.DB.ExecContext(ctx,
		`INSERT INTO task_comments (task_id, user_id, comment) VALUES (?, ?, ?)`,
		taskID, userID, "Second comment",
	)
	if err != nil {
		t.Fatalf("Failed to create comment: %v", err)
	}

	rec, req := ts.MakeAuthRequest(t, http.MethodGet, fmt.Sprintf("/api/tasks/%d/comments", taskID), nil, userID,
		map[string]string{"taskId": fmt.Sprintf("%d", taskID)})

	ts.HandleListTaskComments(rec, req)

	AssertStatusCode(t, rec.Code, http.StatusOK)

	var comments []TaskComment
	DecodeJSON(t, rec, &comments)

	if len(comments) != 2 {
		t.Fatalf("Expected 2 comments, got %d", len(comments))
	}

	// Verify ordering (ASC by created_at)
	if comments[0].Comment != "First comment" {
		t.Errorf("Expected first comment 'First comment', got %q", comments[0].Comment)
	}
	if comments[1].Comment != "Second comment" {
		t.Errorf("Expected second comment 'Second comment', got %q", comments[1].Comment)
	}

	// Verify fields
	if comments[0].TaskID != taskID {
		t.Errorf("Expected task_id %d, got %d", taskID, comments[0].TaskID)
	}
	if comments[0].UserID != userID {
		t.Errorf("Expected user_id %d, got %d", userID, comments[0].UserID)
	}
	if comments[0].ID == 0 {
		t.Error("Expected non-zero comment ID")
	}
}

func TestHandleCreateTaskComment(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "test@example.com", "password123")
	projectID := ts.CreateTestProject(t, userID, "Test Project")
	taskID := ts.CreateTestTask(t, projectID, "Test Task")

	body := CreateCommentRequest{
		Comment: "This is a new comment",
	}

	rec, req := ts.MakeAuthRequest(t, http.MethodPost, fmt.Sprintf("/api/tasks/%d/comments", taskID), body, userID,
		map[string]string{"taskId": fmt.Sprintf("%d", taskID)})

	ts.HandleCreateTaskComment(rec, req)

	AssertStatusCode(t, rec.Code, http.StatusCreated)

	var comment TaskComment
	DecodeJSON(t, rec, &comment)

	if comment.Comment != "This is a new comment" {
		t.Errorf("Expected comment 'This is a new comment', got %q", comment.Comment)
	}
	if comment.TaskID != taskID {
		t.Errorf("Expected task_id %d, got %d", taskID, comment.TaskID)
	}
	if comment.UserID != userID {
		t.Errorf("Expected user_id %d, got %d", userID, comment.UserID)
	}
	if comment.ID == 0 {
		t.Error("Expected non-zero comment ID")
	}
}

func TestHandleCreateTaskCommentValidation(t *testing.T) {
	tests := []struct {
		name       string
		body       CreateCommentRequest
		wantStatus int
		wantError  string
	}{
		{
			name:       "empty comment",
			body:       CreateCommentRequest{Comment: ""},
			wantStatus: http.StatusBadRequest,
			wantError:  "comment is required",
		},
		{
			name:       "comment too long",
			body:       CreateCommentRequest{Comment: strings.Repeat("x", 5001)},
			wantStatus: http.StatusBadRequest,
			wantError:  "comment is too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := NewTestServer(t)
			defer ts.Close()

			userID := ts.CreateTestUser(t, "test@example.com", "password123")
			projectID := ts.CreateTestProject(t, userID, "Test Project")
			taskID := ts.CreateTestTask(t, projectID, "Test Task")

			rec, req := ts.MakeAuthRequest(t, http.MethodPost, fmt.Sprintf("/api/tasks/%d/comments", taskID), tt.body, userID,
				map[string]string{"taskId": fmt.Sprintf("%d", taskID)})

			ts.HandleCreateTaskComment(rec, req)

			AssertError(t, rec, tt.wantStatus, tt.wantError, "invalid_input")
		})
	}
}

func TestHandleListTaskCommentsUnauthorized(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	user1ID := ts.CreateTestUser(t, "user1@example.com", "password123")
	user2ID := ts.CreateTestUser(t, "user2@example.com", "password123")
	projectID := ts.CreateTestProject(t, user1ID, "User1 Project")
	taskID := ts.CreateTestTask(t, projectID, "Test Task")

	// user2 is NOT a member of user1's project
	rec, req := ts.MakeAuthRequest(t, http.MethodGet, fmt.Sprintf("/api/tasks/%d/comments", taskID), nil, user2ID,
		map[string]string{"taskId": fmt.Sprintf("%d", taskID)})

	ts.HandleListTaskComments(rec, req)

	AssertError(t, rec, http.StatusForbidden, "access denied", "forbidden")
}

func TestHandleCreateTaskCommentUnauthorized(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	user1ID := ts.CreateTestUser(t, "user1@example.com", "password123")
	user2ID := ts.CreateTestUser(t, "user2@example.com", "password123")
	projectID := ts.CreateTestProject(t, user1ID, "User1 Project")
	taskID := ts.CreateTestTask(t, projectID, "Test Task")

	body := CreateCommentRequest{
		Comment: "Unauthorized comment",
	}

	// user2 is NOT a member of user1's project
	rec, req := ts.MakeAuthRequest(t, http.MethodPost, fmt.Sprintf("/api/tasks/%d/comments", taskID), body, user2ID,
		map[string]string{"taskId": fmt.Sprintf("%d", taskID)})

	ts.HandleCreateTaskComment(rec, req)

	AssertError(t, rec, http.StatusForbidden, "access denied", "forbidden")
}

func TestHandleListTaskCommentsInvalidTaskID(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "test@example.com", "password123")

	rec, req := ts.MakeAuthRequest(t, http.MethodGet, "/api/tasks/abc/comments", nil, userID,
		map[string]string{"taskId": "abc"})

	ts.HandleListTaskComments(rec, req)

	AssertError(t, rec, http.StatusBadRequest, "invalid task ID", "invalid_input")
}

func TestHandleCreateTaskCommentInvalidTaskID(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "test@example.com", "password123")

	body := CreateCommentRequest{
		Comment: "Some comment",
	}

	rec, req := ts.MakeAuthRequest(t, http.MethodPost, "/api/tasks/abc/comments", body, userID,
		map[string]string{"taskId": "abc"})

	ts.HandleCreateTaskComment(rec, req)

	AssertError(t, rec, http.StatusBadRequest, "invalid task ID", "invalid_input")
}

func TestHandleListTaskCommentsTaskNotFound(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "test@example.com", "password123")

	rec, req := ts.MakeAuthRequest(t, http.MethodGet, "/api/tasks/99999/comments", nil, userID,
		map[string]string{"taskId": "99999"})

	ts.HandleListTaskComments(rec, req)

	AssertError(t, rec, http.StatusNotFound, "task not found", "not_found")
}

func TestHandleCreateTaskCommentTaskNotFound(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "test@example.com", "password123")

	body := CreateCommentRequest{
		Comment: "Comment on ghost task",
	}

	rec, req := ts.MakeAuthRequest(t, http.MethodPost, "/api/tasks/99999/comments", body, userID,
		map[string]string{"taskId": "99999"})

	ts.HandleCreateTaskComment(rec, req)

	AssertError(t, rec, http.StatusNotFound, "task not found", "not_found")
}
