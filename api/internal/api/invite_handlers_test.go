package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestHandleListInvites(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		ts := NewTestServer(t)
		defer ts.Close()

		userID := ts.CreateTestUser(t, "inviter@example.com", "password123")

		rec, req := ts.MakeAuthRequest(t, http.MethodGet, "/api/invites", nil, userID, nil)
		ts.HandleListInvites(rec, req)

		AssertStatusCode(t, rec.Code, http.StatusOK)

		var resp map[string]json.RawMessage
		DecodeJSON(t, rec, &resp)

		var invites []json.RawMessage
		if err := json.Unmarshal(resp["invites"], &invites); err != nil {
			t.Fatalf("Failed to unmarshal invites: %v", err)
		}

		if len(invites) != 0 {
			t.Errorf("Expected 0 invites, got %d", len(invites))
		}

		// Check invite_count is returned (default 3)
		var inviteCount float64
		if err := json.Unmarshal(resp["invite_count"], &inviteCount); err != nil {
			t.Fatalf("Failed to unmarshal invite_count: %v", err)
		}
		if inviteCount != 3 {
			t.Errorf("Expected invite_count 3, got %v", inviteCount)
		}
	})

	t.Run("with invites", func(t *testing.T) {
		ts := NewTestServer(t)
		defer ts.Close()

		userID := ts.CreateTestUser(t, "inviter@example.com", "password123")

		// Create two invites
		ts.CreateTestInvite(t, userID)
		ts.CreateTestInvite(t, userID)

		rec, req := ts.MakeAuthRequest(t, http.MethodGet, "/api/invites", nil, userID, nil)
		ts.HandleListInvites(rec, req)

		AssertStatusCode(t, rec.Code, http.StatusOK)

		var resp map[string]json.RawMessage
		DecodeJSON(t, rec, &resp)

		var invites []json.RawMessage
		if err := json.Unmarshal(resp["invites"], &invites); err != nil {
			t.Fatalf("Failed to unmarshal invites: %v", err)
		}

		if len(invites) != 2 {
			t.Errorf("Expected 2 invites, got %d", len(invites))
		}
	})
}

func TestHandleCreateInvite(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		ts := NewTestServer(t)
		defer ts.Close()

		userID := ts.CreateTestUser(t, "inviter@example.com", "password123")

		rec, req := ts.MakeAuthRequest(t, http.MethodPost, "/api/invites", nil, userID, nil)
		ts.HandleCreateInvite(rec, req)

		AssertStatusCode(t, rec.Code, http.StatusCreated)

		var resp map[string]interface{}
		DecodeJSON(t, rec, &resp)

		code, ok := resp["code"].(string)
		if !ok || code == "" {
			t.Errorf("Expected non-empty invite code, got %v", resp["code"])
		}

		expiresAt, ok := resp["expires_at"].(string)
		if !ok || expiresAt == "" {
			t.Errorf("Expected non-empty expires_at, got %v", resp["expires_at"])
		}

		// Verify invite count was decremented
		var inviteCount int
		err := ts.DB.QueryRow("SELECT invite_count FROM users WHERE id = ?", userID).Scan(&inviteCount)
		if err != nil {
			t.Fatalf("Failed to query invite count: %v", err)
		}
		if inviteCount != 2 {
			t.Errorf("Expected invite_count 2 after creating one invite, got %d", inviteCount)
		}
	})

	t.Run("out of invites", func(t *testing.T) {
		ts := NewTestServer(t)
		defer ts.Close()

		userID := ts.CreateTestUser(t, "inviter@example.com", "password123")

		// Set invite_count to 0
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := ts.DB.ExecContext(ctx, `UPDATE users SET invite_count = 0 WHERE id = ?`, userID)
		if err != nil {
			t.Fatalf("Failed to update invite count: %v", err)
		}

		rec, req := ts.MakeAuthRequest(t, http.MethodPost, "/api/invites", nil, userID, nil)
		ts.HandleCreateInvite(rec, req)

		AssertStatusCode(t, rec.Code, http.StatusForbidden)

		var errResp ErrorResponse
		DecodeJSON(t, rec, &errResp)

		if errResp.Error != "no invites remaining" {
			t.Errorf("Expected error 'no invites remaining', got '%s'", errResp.Error)
		}
		if errResp.Code != "no_invites" {
			t.Errorf("Expected code 'no_invites', got '%s'", errResp.Code)
		}
	})
}

func TestHandleValidateInvite(t *testing.T) {
	t.Run("valid code", func(t *testing.T) {
		ts := NewTestServer(t)
		defer ts.Close()

		userID := ts.CreateTestUser(t, "inviter@example.com", "password123")
		code := ts.CreateTestInvite(t, userID)

		// ValidateInvite is a public endpoint - no auth required
		rec, req := MakeRequest(t, http.MethodGet, "/api/invites/validate?code="+code, nil, nil)
		ts.HandleValidateInvite(rec, req)

		AssertStatusCode(t, rec.Code, http.StatusOK)

		var status InviteStatus
		DecodeJSON(t, rec, &status)

		if !status.Valid {
			t.Errorf("Expected invite to be valid, got invalid: %s", status.Message)
		}
	})

	t.Run("invalid code", func(t *testing.T) {
		ts := NewTestServer(t)
		defer ts.Close()

		rec, req := MakeRequest(t, http.MethodGet, "/api/invites/validate?code=nonexistent-code", nil, nil)
		ts.HandleValidateInvite(rec, req)

		AssertStatusCode(t, rec.Code, http.StatusOK)

		var status InviteStatus
		DecodeJSON(t, rec, &status)

		if status.Valid {
			t.Errorf("Expected invite to be invalid")
		}
		if status.Message != "invalid invite code" {
			t.Errorf("Expected message 'invalid invite code', got '%s'", status.Message)
		}
	})

	t.Run("used code", func(t *testing.T) {
		ts := NewTestServer(t)
		defer ts.Close()

		userID := ts.CreateTestUser(t, "inviter@example.com", "password123")
		code := ts.CreateTestInvite(t, userID)

		// Mark the invite as used
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := ts.DB.ExecContext(ctx,
			`UPDATE invites SET used_at = CURRENT_TIMESTAMP, invitee_id = ? WHERE code = ?`,
			userID, code,
		)
		if err != nil {
			t.Fatalf("Failed to mark invite as used: %v", err)
		}

		rec, req := MakeRequest(t, http.MethodGet, "/api/invites/validate?code="+code, nil, nil)
		ts.HandleValidateInvite(rec, req)

		AssertStatusCode(t, rec.Code, http.StatusOK)

		var status InviteStatus
		DecodeJSON(t, rec, &status)

		if status.Valid {
			t.Errorf("Expected used invite to be invalid")
		}
		if status.Message != "this invite has already been used" {
			t.Errorf("Expected message 'this invite has already been used', got '%s'", status.Message)
		}
	})

	t.Run("missing code parameter", func(t *testing.T) {
		ts := NewTestServer(t)
		defer ts.Close()

		rec, req := MakeRequest(t, http.MethodGet, "/api/invites/validate", nil, nil)
		ts.HandleValidateInvite(rec, req)

		AssertStatusCode(t, rec.Code, http.StatusOK)

		var status InviteStatus
		DecodeJSON(t, rec, &status)

		if status.Valid {
			t.Errorf("Expected invite to be invalid when no code provided")
		}
		if status.Message != "invite code is required" {
			t.Errorf("Expected message 'invite code is required', got '%s'", status.Message)
		}
	})
}

// makeAdmin sets a user as admin in the database
func makeAdmin(t *testing.T, ts *TestServer, userID int64) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := ts.DB.ExecContext(ctx, `UPDATE users SET is_admin = 1 WHERE id = ?`, userID)
	if err != nil {
		t.Fatalf("Failed to make user admin: %v", err)
	}
}

func TestHandleAdminBoostInvites(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(*TestServer) (adminID int64, targetUserID int64)
		targetIDPath  string       // the {id} in the path, set from setupFunc if empty
		body          interface{}
		isAdmin       bool
		wantStatus    int
		wantError     string
		wantErrorCode string
		wantCount     int
		noAuth        bool
	}{
		{
			name: "admin boosts invites successfully",
			setupFunc: func(ts *TestServer) (int64, int64) {
				admin := ts.CreateTestUser(t, "admin@example.com", "password123")
				makeAdmin(t, ts, admin)
				target := ts.CreateTestUser(t, "target@example.com", "password123")
				return admin, target
			},
			body:       map[string]interface{}{"invite_count": 10},
			isAdmin:    true,
			wantStatus: http.StatusOK,
			wantCount:  10,
		},
		{
			name: "admin sets invite count to zero",
			setupFunc: func(ts *TestServer) (int64, int64) {
				admin := ts.CreateTestUser(t, "admin@example.com", "password123")
				makeAdmin(t, ts, admin)
				target := ts.CreateTestUser(t, "target@example.com", "password123")
				return admin, target
			},
			body:       map[string]interface{}{"invite_count": 0},
			isAdmin:    true,
			wantStatus: http.StatusOK,
			wantCount:  0,
		},
		{
			name: "non-admin forbidden",
			setupFunc: func(ts *TestServer) (int64, int64) {
				user := ts.CreateTestUser(t, "user@example.com", "password123")
				target := ts.CreateTestUser(t, "target@example.com", "password123")
				return user, target
			},
			body:          map[string]interface{}{"invite_count": 10},
			isAdmin:       false,
			wantStatus:    http.StatusForbidden,
			wantError:     "admin access required",
			wantErrorCode: "forbidden",
		},
		{
			name:          "unauthenticated request",
			noAuth:        true,
			body:          map[string]interface{}{"invite_count": 10},
			targetIDPath:  "1",
			wantStatus:    http.StatusUnauthorized,
			wantError:     "user not authenticated",
			wantErrorCode: "unauthorized",
		},
		{
			name: "negative invite count",
			setupFunc: func(ts *TestServer) (int64, int64) {
				admin := ts.CreateTestUser(t, "admin@example.com", "password123")
				makeAdmin(t, ts, admin)
				target := ts.CreateTestUser(t, "target@example.com", "password123")
				return admin, target
			},
			body:          map[string]interface{}{"invite_count": -1},
			isAdmin:       true,
			wantStatus:    http.StatusBadRequest,
			wantError:     "invite count must be non-negative",
			wantErrorCode: "validation_error",
		},
		{
			name: "invalid user id in path",
			setupFunc: func(ts *TestServer) (int64, int64) {
				admin := ts.CreateTestUser(t, "admin@example.com", "password123")
				makeAdmin(t, ts, admin)
				return admin, 0
			},
			targetIDPath:  "not-a-number",
			body:          map[string]interface{}{"invite_count": 5},
			isAdmin:       true,
			wantStatus:    http.StatusBadRequest,
			wantError:     "invalid user id",
			wantErrorCode: "validation_error",
		},
		{
			name: "target user not found",
			setupFunc: func(ts *TestServer) (int64, int64) {
				admin := ts.CreateTestUser(t, "admin@example.com", "password123")
				makeAdmin(t, ts, admin)
				return admin, 0
			},
			targetIDPath:  "99999",
			body:          map[string]interface{}{"invite_count": 5},
			isAdmin:       true,
			wantStatus:    http.StatusNotFound,
			wantError:     "user not found",
			wantErrorCode: "not_found",
		},
		{
			name: "invalid request body",
			setupFunc: func(ts *TestServer) (int64, int64) {
				admin := ts.CreateTestUser(t, "admin@example.com", "password123")
				makeAdmin(t, ts, admin)
				target := ts.CreateTestUser(t, "target@example.com", "password123")
				return admin, target
			},
			body:          "not-json",
			isAdmin:       true,
			wantStatus:    http.StatusBadRequest,
			wantError:     "invalid request body",
			wantErrorCode: "invalid_request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := NewTestServer(t)
			defer ts.Close()

			if tt.noAuth {
				rec, req := MakeRequest(t, http.MethodPatch, "/api/admin/users/1/invites", tt.body, nil)
				req.SetPathValue("id", tt.targetIDPath)
				ts.HandleAdminBoostInvites(rec, req)

				AssertStatusCode(t, rec.Code, tt.wantStatus)
				if tt.wantError != "" {
					AssertError(t, rec, tt.wantStatus, tt.wantError, tt.wantErrorCode)
				}
				return
			}

			adminID, targetUserID := tt.setupFunc(ts)

			pathID := tt.targetIDPath
			if pathID == "" {
				pathID = fmt.Sprintf("%d", targetUserID)
			}

			rec, req := ts.MakeAuthRequest(t, http.MethodPatch, "/api/admin/users/"+pathID+"/invites", tt.body, adminID, nil)
			req.SetPathValue("id", pathID)
			ts.HandleAdminBoostInvites(rec, req)

			AssertStatusCode(t, rec.Code, tt.wantStatus)

			if tt.wantError != "" {
				AssertError(t, rec, tt.wantStatus, tt.wantError, tt.wantErrorCode)
			} else {
				var resp map[string]interface{}
				DecodeJSON(t, rec, &resp)

				gotID := int64(resp["id"].(float64))
				if gotID != targetUserID {
					t.Errorf("Response id = %d, want %d", gotID, targetUserID)
				}

				gotCount := int(resp["invite_count"].(float64))
				if gotCount != tt.wantCount {
					t.Errorf("Response invite_count = %d, want %d", gotCount, tt.wantCount)
				}

				// Verify it was persisted in the database
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				var dbCount int
				err := ts.DB.QueryRowContext(ctx, `SELECT invite_count FROM users WHERE id = ?`, targetUserID).Scan(&dbCount)
				if err != nil {
					t.Fatalf("Failed to query invite count: %v", err)
				}
				if dbCount != tt.wantCount {
					t.Errorf("DB invite_count = %d, want %d", dbCount, tt.wantCount)
				}
			}
		})
	}
}
