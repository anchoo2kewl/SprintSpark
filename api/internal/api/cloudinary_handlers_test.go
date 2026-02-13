package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

// createTestProjectForCloudinary creates a project owned by the given user and returns its ID
func createTestProjectForCloudinary(t *testing.T, ts *TestServer, ownerID int64) int64 {
	t.Helper()

	ctx := context.Background()
	result, err := ts.DB.ExecContext(ctx,
		`INSERT INTO projects (owner_id, name, description) VALUES (?, ?, ?)`,
		ownerID, "Test Project", "Test project for cloudinary tests",
	)
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get project ID: %v", err)
	}

	return id
}

// createTestTaskForCloudinary creates a task in the given project and returns its ID
func createTestTaskForCloudinary(t *testing.T, ts *TestServer, projectID int64) int64 {
	t.Helper()

	ctx := context.Background()
	result, err := ts.DB.ExecContext(ctx,
		`INSERT INTO tasks (project_id, title, status, priority) VALUES (?, ?, ?, ?)`,
		projectID, "Test Task", "todo", "medium",
	)
	if err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get task ID: %v", err)
	}

	return id
}

// createTestAttachment inserts a task_attachments row and returns the attachment ID
func createTestAttachment(t *testing.T, ts *TestServer, taskID, userID, projectID int64, fileType, filename, altName string) int64 {
	t.Helper()

	ctx := context.Background()
	result, err := ts.DB.ExecContext(ctx,
		`INSERT INTO task_attachments (task_id, project_id, user_id, filename, alt_name, file_type, content_type, file_size, cloudinary_url, cloudinary_public_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		taskID, projectID, userID, filename, altName, fileType, "application/octet-stream", 1024,
		"https://res.cloudinary.com/test/"+filename, "test/"+filename,
	)
	if err != nil {
		t.Fatalf("Failed to create test attachment: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get attachment ID: %v", err)
	}

	return id
}

// addProjectMember adds a user as a project member
func addProjectMember(t *testing.T, ts *TestServer, projectID, userID, grantedBy int64, role string) {
	t.Helper()

	ctx := context.Background()
	_, err := ts.DB.ExecContext(ctx,
		`INSERT INTO project_members (project_id, user_id, role, granted_by) VALUES (?, ?, ?, ?)`,
		projectID, userID, role, grantedBy,
	)
	if err != nil {
		t.Fatalf("Failed to add project member: %v", err)
	}
}

// makeAuthRequest creates a request with auth context and optional chi URL params
func makeAuthRequest(t *testing.T, method, path string, body interface{}, userID int64, urlParams map[string]string) (*httptest.ResponseRecorder, *http.Request) {
	t.Helper()

	rec, req := MakeRequest(t, method, path, body, nil)

	// Set user context
	ctx := context.WithValue(req.Context(), UserIDKey, userID)

	// Set chi URL params if any
	if len(urlParams) > 0 {
		rctx := chi.NewRouteContext()
		for k, v := range urlParams {
			rctx.URLParams.Add(k, v)
		}
		ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	}

	req = req.WithContext(ctx)
	return rec, req
}

func TestHandleListAssets(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		wantStatus int
		wantCount  int
		setupFunc  func(t *testing.T, ts *TestServer) int64 // returns userID to use for request
		checkOwner bool
	}{
		{
			name:       "empty list when no attachments",
			wantStatus: http.StatusOK,
			wantCount:  0,
			setupFunc: func(t *testing.T, ts *TestServer) int64 {
				return ts.CreateTestUser(t, "user@example.com", "password123")
			},
		},
		{
			name:       "returns own attachments with is_owner true",
			wantStatus: http.StatusOK,
			wantCount:  2,
			checkOwner: true,
			setupFunc: func(t *testing.T, ts *TestServer) int64 {
				userID := ts.CreateTestUser(t, "user@example.com", "password123")
				projectID := createTestProjectForCloudinary(t, ts, userID)
				taskID := createTestTaskForCloudinary(t, ts, projectID)
				createTestAttachment(t, ts, taskID, userID, projectID, "image", "photo1.jpg", "My photo")
				createTestAttachment(t, ts, taskID, userID, projectID, "pdf", "doc.pdf", "My doc")
				return userID
			},
		},
		{
			name:       "returns shared project members attachments",
			wantStatus: http.StatusOK,
			wantCount:  3, // 2 own + 1 shared
			setupFunc: func(t *testing.T, ts *TestServer) int64 {
				user1ID := ts.CreateTestUser(t, "user1@example.com", "password123")
				user2ID := ts.CreateTestUser(t, "user2@example.com", "password123")

				projectID := createTestProjectForCloudinary(t, ts, user1ID)
				taskID := createTestTaskForCloudinary(t, ts, projectID)

				// Add both users as project members
				addProjectMember(t, ts, projectID, user1ID, user1ID, "admin")
				addProjectMember(t, ts, projectID, user2ID, user1ID, "editor")

				// user1's attachments
				createTestAttachment(t, ts, taskID, user1ID, projectID, "image", "user1_photo.jpg", "User1 photo")
				createTestAttachment(t, ts, taskID, user1ID, projectID, "pdf", "user1_doc.pdf", "User1 doc")
				// user2's attachment (should appear for user1 via shared project)
				createTestAttachment(t, ts, taskID, user2ID, projectID, "image", "user2_photo.jpg", "User2 photo")

				return user1ID
			},
		},
		{
			name:       "filters by type=image",
			query:      "?type=image",
			wantStatus: http.StatusOK,
			wantCount:  1,
			setupFunc: func(t *testing.T, ts *TestServer) int64 {
				userID := ts.CreateTestUser(t, "user@example.com", "password123")
				projectID := createTestProjectForCloudinary(t, ts, userID)
				taskID := createTestTaskForCloudinary(t, ts, projectID)
				createTestAttachment(t, ts, taskID, userID, projectID, "image", "photo.jpg", "Photo")
				createTestAttachment(t, ts, taskID, userID, projectID, "pdf", "doc.pdf", "Doc")
				createTestAttachment(t, ts, taskID, userID, projectID, "video", "clip.mp4", "Video")
				return userID
			},
		},
		{
			name:       "filters by type=video",
			query:      "?type=video",
			wantStatus: http.StatusOK,
			wantCount:  1,
			setupFunc: func(t *testing.T, ts *TestServer) int64 {
				userID := ts.CreateTestUser(t, "user@example.com", "password123")
				projectID := createTestProjectForCloudinary(t, ts, userID)
				taskID := createTestTaskForCloudinary(t, ts, projectID)
				createTestAttachment(t, ts, taskID, userID, projectID, "image", "photo.jpg", "Photo")
				createTestAttachment(t, ts, taskID, userID, projectID, "video", "clip.mp4", "Video")
				return userID
			},
		},
		{
			name:       "filters by type=pdf",
			query:      "?type=pdf",
			wantStatus: http.StatusOK,
			wantCount:  1,
			setupFunc: func(t *testing.T, ts *TestServer) int64 {
				userID := ts.CreateTestUser(t, "user@example.com", "password123")
				projectID := createTestProjectForCloudinary(t, ts, userID)
				taskID := createTestTaskForCloudinary(t, ts, projectID)
				createTestAttachment(t, ts, taskID, userID, projectID, "image", "photo.jpg", "Photo")
				createTestAttachment(t, ts, taskID, userID, projectID, "pdf", "report.pdf", "Report")
				return userID
			},
		},
		{
			name:       "search by query matches alt_name",
			query:      "?q=sunset",
			wantStatus: http.StatusOK,
			wantCount:  1,
			setupFunc: func(t *testing.T, ts *TestServer) int64 {
				userID := ts.CreateTestUser(t, "user@example.com", "password123")
				projectID := createTestProjectForCloudinary(t, ts, userID)
				taskID := createTestTaskForCloudinary(t, ts, projectID)
				createTestAttachment(t, ts, taskID, userID, projectID, "image", "img001.jpg", "Beautiful sunset")
				createTestAttachment(t, ts, taskID, userID, projectID, "image", "img002.jpg", "Mountain view")
				return userID
			},
		},
		{
			name:       "search by query matches filename",
			query:      "?q=report",
			wantStatus: http.StatusOK,
			wantCount:  1,
			setupFunc: func(t *testing.T, ts *TestServer) int64 {
				userID := ts.CreateTestUser(t, "user@example.com", "password123")
				projectID := createTestProjectForCloudinary(t, ts, userID)
				taskID := createTestTaskForCloudinary(t, ts, projectID)
				createTestAttachment(t, ts, taskID, userID, projectID, "pdf", "report_2024.pdf", "")
				createTestAttachment(t, ts, taskID, userID, projectID, "pdf", "invoice.pdf", "")
				return userID
			},
		},
		{
			name:       "pagination with limit and offset",
			query:      "?limit=2&offset=1",
			wantStatus: http.StatusOK,
			wantCount:  2,
			setupFunc: func(t *testing.T, ts *TestServer) int64 {
				userID := ts.CreateTestUser(t, "user@example.com", "password123")
				projectID := createTestProjectForCloudinary(t, ts, userID)
				taskID := createTestTaskForCloudinary(t, ts, projectID)
				createTestAttachment(t, ts, taskID, userID, projectID, "image", "a.jpg", "A")
				createTestAttachment(t, ts, taskID, userID, projectID, "image", "b.jpg", "B")
				createTestAttachment(t, ts, taskID, userID, projectID, "image", "c.jpg", "C")
				return userID
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := NewTestServer(t)
			defer ts.Close()

			userID := tt.setupFunc(t, ts)

			path := "/api/assets"
			if tt.query != "" {
				path += tt.query
			}

			rec, req := makeAuthRequest(t, http.MethodGet, path, nil, userID, nil)
			ts.HandleListAssets(rec, req)

			AssertStatusCode(t, rec.Code, tt.wantStatus)

			var assets []AssetResponse
			DecodeJSON(t, rec, &assets)

			if len(assets) != tt.wantCount {
				t.Errorf("Expected %d assets, got %d", tt.wantCount, len(assets))
			}

			if tt.checkOwner && len(assets) > 0 {
				for _, a := range assets {
					if a.UserID == userID && !a.IsOwner {
						t.Errorf("Expected is_owner=true for own asset %d", a.ID)
					}
				}
			}
		})
	}
}

func TestHandleListAssetsRequiresAuth(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "user@example.com", "password123")
	token := ts.GenerateTestToken(t, userID, "user@example.com")

	// With valid auth through middleware
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", token),
	}
	rec, req := MakeRequest(t, http.MethodGet, "/api/assets", nil, headers)
	ts.JWTAuth(http.HandlerFunc(ts.HandleListAssets)).ServeHTTP(rec, req)
	AssertStatusCode(t, rec.Code, http.StatusOK)

	// Without auth through middleware
	rec2, req2 := MakeRequest(t, http.MethodGet, "/api/assets", nil, nil)
	ts.JWTAuth(http.HandlerFunc(ts.HandleListAssets)).ServeHTTP(rec2, req2)
	AssertStatusCode(t, rec2.Code, http.StatusUnauthorized)
}

func TestHandleDeleteAttachment(t *testing.T) {
	tests := []struct {
		name          string
		attachmentID  string
		wantStatus    int
		wantError     string
		wantErrorCode string
		setupFunc     func(t *testing.T, ts *TestServer) int64 // returns userID for request
	}{
		{
			name:       "owner can delete own attachment",
			wantStatus: http.StatusOK,
			setupFunc: func(t *testing.T, ts *TestServer) int64 {
				userID := ts.CreateTestUser(t, "user@example.com", "password123")
				projectID := createTestProjectForCloudinary(t, ts, userID)
				taskID := createTestTaskForCloudinary(t, ts, projectID)
				createTestAttachment(t, ts, taskID, userID, projectID, "image", "photo.jpg", "Photo")
				return userID
			},
		},
		{
			name:          "non-owner gets 403",
			wantStatus:    http.StatusForbidden,
			wantError:     "you can only delete your own attachments",
			wantErrorCode: "forbidden",
			setupFunc: func(t *testing.T, ts *TestServer) int64 {
				ownerID := ts.CreateTestUser(t, "owner@example.com", "password123")
				otherID := ts.CreateTestUser(t, "other@example.com", "password123")
				projectID := createTestProjectForCloudinary(t, ts, ownerID)
				taskID := createTestTaskForCloudinary(t, ts, projectID)
				createTestAttachment(t, ts, taskID, ownerID, projectID, "image", "photo.jpg", "Photo")
				return otherID
			},
		},
		{
			name:          "invalid ID returns 400",
			attachmentID:  "abc",
			wantStatus:    http.StatusBadRequest,
			wantError:     "invalid attachment ID",
			wantErrorCode: "bad_request",
			setupFunc: func(t *testing.T, ts *TestServer) int64 {
				return ts.CreateTestUser(t, "user@example.com", "password123")
			},
		},
		{
			name:          "non-existent ID returns 404",
			attachmentID:  "99999",
			wantStatus:    http.StatusNotFound,
			wantError:     "attachment not found",
			wantErrorCode: "not_found",
			setupFunc: func(t *testing.T, ts *TestServer) int64 {
				return ts.CreateTestUser(t, "user@example.com", "password123")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := NewTestServer(t)
			defer ts.Close()

			userID := tt.setupFunc(t, ts)

			// Determine attachment ID
			attachmentID := tt.attachmentID
			if attachmentID == "" {
				attachmentID = "1" // the first auto-increment attachment
			}

			rec, req := makeAuthRequest(t, http.MethodDelete, "/api/attachments/"+attachmentID, nil, userID,
				map[string]string{"id": attachmentID})
			ts.HandleDeleteAttachment(rec, req)

			AssertStatusCode(t, rec.Code, tt.wantStatus)

			if tt.wantError != "" {
				AssertError(t, rec, tt.wantStatus, tt.wantError, tt.wantErrorCode)
			}

			// Verify deletion actually happened for success case
			if tt.wantStatus == http.StatusOK {
				var resp map[string]string
				DecodeJSON(t, rec, &resp)
				if resp["message"] != "Attachment deleted" {
					t.Errorf("Expected deletion message, got %q", resp["message"])
				}
			}
		})
	}
}

func TestHandleUpdateAttachment(t *testing.T) {
	tests := []struct {
		name          string
		body          interface{}
		wantStatus    int
		wantError     string
		wantErrorCode string
		setupFunc     func(t *testing.T, ts *TestServer) (userID int64, attachmentID string)
	}{
		{
			name:       "owner updates alt_name",
			body:       UpdateAttachmentRequest{AltName: stringPtr("Updated alt text")},
			wantStatus: http.StatusOK,
			setupFunc: func(t *testing.T, ts *TestServer) (int64, string) {
				userID := ts.CreateTestUser(t, "user@example.com", "password123")
				projectID := createTestProjectForCloudinary(t, ts, userID)
				taskID := createTestTaskForCloudinary(t, ts, projectID)
				attachID := createTestAttachment(t, ts, taskID, userID, projectID, "image", "photo.jpg", "Old alt")
				return userID, fmt.Sprintf("%d", attachID)
			},
		},
		{
			name:          "non-owner gets 403",
			body:          UpdateAttachmentRequest{AltName: stringPtr("Hacked")},
			wantStatus:    http.StatusForbidden,
			wantError:     "you can only update your own attachments",
			wantErrorCode: "forbidden",
			setupFunc: func(t *testing.T, ts *TestServer) (int64, string) {
				ownerID := ts.CreateTestUser(t, "owner@example.com", "password123")
				otherID := ts.CreateTestUser(t, "other@example.com", "password123")
				projectID := createTestProjectForCloudinary(t, ts, ownerID)
				taskID := createTestTaskForCloudinary(t, ts, projectID)
				attachID := createTestAttachment(t, ts, taskID, ownerID, projectID, "image", "photo.jpg", "Original")
				return otherID, fmt.Sprintf("%d", attachID)
			},
		},
		{
			name:          "non-existent attachment returns 404",
			body:          UpdateAttachmentRequest{AltName: stringPtr("test")},
			wantStatus:    http.StatusNotFound,
			wantError:     "attachment not found",
			wantErrorCode: "not_found",
			setupFunc: func(t *testing.T, ts *TestServer) (int64, string) {
				userID := ts.CreateTestUser(t, "user@example.com", "password123")
				return userID, "99999"
			},
		},
		{
			name:          "invalid ID returns 400",
			body:          UpdateAttachmentRequest{AltName: stringPtr("test")},
			wantStatus:    http.StatusBadRequest,
			wantError:     "invalid attachment ID",
			wantErrorCode: "bad_request",
			setupFunc: func(t *testing.T, ts *TestServer) (int64, string) {
				userID := ts.CreateTestUser(t, "user@example.com", "password123")
				return userID, "invalid"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := NewTestServer(t)
			defer ts.Close()

			userID, attachmentID := tt.setupFunc(t, ts)

			rec, req := makeAuthRequest(t, http.MethodPatch, "/api/attachments/"+attachmentID, tt.body, userID,
				map[string]string{"id": attachmentID})
			ts.HandleUpdateAttachment(rec, req)

			AssertStatusCode(t, rec.Code, tt.wantStatus)

			if tt.wantError != "" {
				AssertError(t, rec, tt.wantStatus, tt.wantError, tt.wantErrorCode)
			}

			if tt.wantStatus == http.StatusOK {
				var resp TaskAttachment
				DecodeJSON(t, rec, &resp)
				if resp.AltName != "Updated alt text" {
					t.Errorf("Expected alt_name 'Updated alt text', got %q", resp.AltName)
				}
			}
		})
	}
}

func TestHandleGetStorageUsage(t *testing.T) {
	tests := []struct {
		name       string
		projectID  string
		wantStatus int
		wantCount  int
		setupFunc  func(t *testing.T, ts *TestServer) int64 // returns userID
	}{
		{
			name:       "returns storage usage per user",
			projectID:  "1",
			wantStatus: http.StatusOK,
			wantCount:  1,
			setupFunc: func(t *testing.T, ts *TestServer) int64 {
				userID := ts.CreateTestUser(t, "user@example.com", "password123")
				projectID := createTestProjectForCloudinary(t, ts, userID)
				taskID := createTestTaskForCloudinary(t, ts, projectID)
				createTestAttachment(t, ts, taskID, userID, projectID, "image", "a.jpg", "A")
				createTestAttachment(t, ts, taskID, userID, projectID, "image", "b.jpg", "B")
				return userID
			},
		},
		{
			name:       "returns multiple users",
			projectID:  "1",
			wantStatus: http.StatusOK,
			wantCount:  2,
			setupFunc: func(t *testing.T, ts *TestServer) int64 {
				user1ID := ts.CreateTestUser(t, "user1@example.com", "password123")
				user2ID := ts.CreateTestUser(t, "user2@example.com", "password123")
				projectID := createTestProjectForCloudinary(t, ts, user1ID)
				taskID := createTestTaskForCloudinary(t, ts, projectID)
				createTestAttachment(t, ts, taskID, user1ID, projectID, "image", "a.jpg", "A")
				createTestAttachment(t, ts, taskID, user2ID, projectID, "pdf", "b.pdf", "B")
				return user1ID
			},
		},
		{
			name:       "empty project returns empty list",
			projectID:  "1",
			wantStatus: http.StatusOK,
			wantCount:  0,
			setupFunc: func(t *testing.T, ts *TestServer) int64 {
				userID := ts.CreateTestUser(t, "user@example.com", "password123")
				createTestProjectForCloudinary(t, ts, userID)
				return userID
			},
		},
		{
			name:       "invalid project ID returns 400",
			projectID:  "abc",
			wantStatus: http.StatusBadRequest,
			setupFunc: func(t *testing.T, ts *TestServer) int64 {
				return ts.CreateTestUser(t, "user@example.com", "password123")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := NewTestServer(t)
			defer ts.Close()

			userID := tt.setupFunc(t, ts)

			rec, req := makeAuthRequest(t, http.MethodGet, "/api/projects/"+tt.projectID+"/storage", nil, userID,
				map[string]string{"id": tt.projectID})
			ts.HandleGetStorageUsage(rec, req)

			AssertStatusCode(t, rec.Code, tt.wantStatus)

			if tt.wantStatus == http.StatusOK {
				var usage []StorageUsage
				DecodeJSON(t, rec, &usage)
				if len(usage) != tt.wantCount {
					t.Errorf("Expected %d usage entries, got %d", tt.wantCount, len(usage))
				}
			}
		})
	}
}

func TestHandleListImages(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		wantStatus int
		wantCount  int
		setupFunc  func(t *testing.T, ts *TestServer) int64
	}{
		{
			name:       "returns only images",
			wantStatus: http.StatusOK,
			wantCount:  2,
			setupFunc: func(t *testing.T, ts *TestServer) int64 {
				userID := ts.CreateTestUser(t, "user@example.com", "password123")
				projectID := createTestProjectForCloudinary(t, ts, userID)
				taskID := createTestTaskForCloudinary(t, ts, projectID)
				createTestAttachment(t, ts, taskID, userID, projectID, "image", "photo1.jpg", "Photo 1")
				createTestAttachment(t, ts, taskID, userID, projectID, "image", "photo2.jpg", "Photo 2")
				createTestAttachment(t, ts, taskID, userID, projectID, "pdf", "doc.pdf", "Document")
				createTestAttachment(t, ts, taskID, userID, projectID, "video", "clip.mp4", "Video")
				return userID
			},
		},
		{
			name:       "search filters images",
			query:      "?q=sunset",
			wantStatus: http.StatusOK,
			wantCount:  1,
			setupFunc: func(t *testing.T, ts *TestServer) int64 {
				userID := ts.CreateTestUser(t, "user@example.com", "password123")
				projectID := createTestProjectForCloudinary(t, ts, userID)
				taskID := createTestTaskForCloudinary(t, ts, projectID)
				createTestAttachment(t, ts, taskID, userID, projectID, "image", "sunset.jpg", "Beautiful sunset")
				createTestAttachment(t, ts, taskID, userID, projectID, "image", "mountain.jpg", "Mountain view")
				return userID
			},
		},
		{
			name:       "empty result when no images",
			wantStatus: http.StatusOK,
			wantCount:  0,
			setupFunc: func(t *testing.T, ts *TestServer) int64 {
				userID := ts.CreateTestUser(t, "user@example.com", "password123")
				projectID := createTestProjectForCloudinary(t, ts, userID)
				taskID := createTestTaskForCloudinary(t, ts, projectID)
				createTestAttachment(t, ts, taskID, userID, projectID, "pdf", "doc.pdf", "Doc")
				return userID
			},
		},
		{
			name:       "includes shared project members images",
			wantStatus: http.StatusOK,
			wantCount:  2, // 1 own + 1 shared
			setupFunc: func(t *testing.T, ts *TestServer) int64 {
				user1ID := ts.CreateTestUser(t, "user1@example.com", "password123")
				user2ID := ts.CreateTestUser(t, "user2@example.com", "password123")
				projectID := createTestProjectForCloudinary(t, ts, user1ID)
				taskID := createTestTaskForCloudinary(t, ts, projectID)
				addProjectMember(t, ts, projectID, user1ID, user1ID, "admin")
				addProjectMember(t, ts, projectID, user2ID, user1ID, "editor")
				createTestAttachment(t, ts, taskID, user1ID, projectID, "image", "user1.jpg", "User1 image")
				createTestAttachment(t, ts, taskID, user2ID, projectID, "image", "user2.jpg", "User2 image")
				return user1ID
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := NewTestServer(t)
			defer ts.Close()

			userID := tt.setupFunc(t, ts)

			path := "/api/images"
			if tt.query != "" {
				path += tt.query
			}

			rec, req := makeAuthRequest(t, http.MethodGet, path, nil, userID, nil)
			ts.HandleListImages(rec, req)

			AssertStatusCode(t, rec.Code, tt.wantStatus)

			var images []TaskAttachment
			DecodeJSON(t, rec, &images)

			if len(images) != tt.wantCount {
				t.Errorf("Expected %d images, got %d", tt.wantCount, len(images))
			}

			// Verify all returned are images
			for _, img := range images {
				if img.FileType != "image" {
					t.Errorf("Expected file_type 'image', got %q", img.FileType)
				}
			}
		})
	}
}

func TestHandleDeleteAttachmentVerifyRemoval(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "user@example.com", "password123")
	projectID := createTestProjectForCloudinary(t, ts, userID)
	taskID := createTestTaskForCloudinary(t, ts, projectID)
	attachID := createTestAttachment(t, ts, taskID, userID, projectID, "image", "photo.jpg", "Photo")

	attachmentID := fmt.Sprintf("%d", attachID)

	// Delete the attachment
	rec, req := makeAuthRequest(t, http.MethodDelete, "/api/attachments/"+attachmentID, nil, userID,
		map[string]string{"id": attachmentID})
	ts.HandleDeleteAttachment(rec, req)
	AssertStatusCode(t, rec.Code, http.StatusOK)

	// Verify it's gone by trying to delete again
	rec2, req2 := makeAuthRequest(t, http.MethodDelete, "/api/attachments/"+attachmentID, nil, userID,
		map[string]string{"id": attachmentID})
	ts.HandleDeleteAttachment(rec2, req2)
	AssertStatusCode(t, rec2.Code, http.StatusNotFound)
}

func TestHandleListAssetsOwnerFlag(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	user1ID := ts.CreateTestUser(t, "user1@example.com", "password123")
	user2ID := ts.CreateTestUser(t, "user2@example.com", "password123")

	projectID := createTestProjectForCloudinary(t, ts, user1ID)
	taskID := createTestTaskForCloudinary(t, ts, projectID)

	addProjectMember(t, ts, projectID, user1ID, user1ID, "admin")
	addProjectMember(t, ts, projectID, user2ID, user1ID, "editor")

	createTestAttachment(t, ts, taskID, user1ID, projectID, "image", "mine.jpg", "My photo")
	createTestAttachment(t, ts, taskID, user2ID, projectID, "image", "theirs.jpg", "Their photo")

	// Request as user1
	rec, req := makeAuthRequest(t, http.MethodGet, "/api/assets", nil, user1ID, nil)
	ts.HandleListAssets(rec, req)
	AssertStatusCode(t, rec.Code, http.StatusOK)

	var assets []json.RawMessage
	DecodeJSON(t, rec, &assets)

	if len(assets) != 2 {
		t.Fatalf("Expected 2 assets, got %d", len(assets))
	}

	for _, raw := range assets {
		var a struct {
			UserID  int64 `json:"user_id"`
			IsOwner bool  `json:"is_owner"`
		}
		if err := json.Unmarshal(raw, &a); err != nil {
			t.Fatal(err)
		}
		if a.UserID == user1ID && !a.IsOwner {
			t.Error("Expected is_owner=true for user1's asset")
		}
		if a.UserID == user2ID && a.IsOwner {
			t.Error("Expected is_owner=false for user2's asset (viewed by user1)")
		}
	}
}

