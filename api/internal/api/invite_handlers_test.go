package api

import (
	"context"
	"encoding/json"
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
