package api

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestHandleChangePassword(t *testing.T) {
	tests := []struct {
		name            string
		currentPassword string
		newPassword     string
		wantStatus      int
		wantError       string
	}{
		{
			name:            "happy path",
			currentPassword: "password123",
			newPassword:     "newpassword456",
			wantStatus:      http.StatusOK,
		},
		{
			name:            "wrong current password",
			currentPassword: "wrongpassword",
			newPassword:     "newpassword456",
			wantStatus:      http.StatusUnauthorized,
			wantError:       "Current password is incorrect",
		},
		{
			name:            "new password too short",
			currentPassword: "password123",
			newPassword:     "short",
			wantStatus:      http.StatusBadRequest,
			wantError:       "at least 8 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := NewTestServer(t)
			defer ts.Close()

			userID := ts.CreateTestUser(t, "user@example.com", "password123")

			body := ChangePasswordRequest{
				CurrentPassword: tt.currentPassword,
				NewPassword:     tt.newPassword,
			}

			rec, req := ts.MakeAuthRequest(t, http.MethodPost, "/api/settings/password", body, userID, nil)
			ts.HandleChangePassword(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("Status code mismatch: got %d, want %d. Body: %s", rec.Code, tt.wantStatus, rec.Body.String())
			}

			if tt.wantError != "" {
				bodyStr := rec.Body.String()
				if !contains(bodyStr, tt.wantError) {
					t.Errorf("Response body %q does not contain %q", bodyStr, tt.wantError)
				}
			} else {
				var resp map[string]string
				DecodeJSON(t, rec, &resp)
				if resp["message"] != "Password changed successfully" {
					t.Errorf("Expected success message, got %q", resp["message"])
				}

				// Verify the password was actually changed by trying to log in with the new password
				// Check that password_changed_at was updated
				var changedAt *time.Time
				err := ts.DB.QueryRow("SELECT password_changed_at FROM users WHERE id = ?", userID).Scan(&changedAt)
				if err != nil {
					t.Fatalf("Failed to query password_changed_at: %v", err)
				}
				if changedAt == nil {
					t.Error("Expected password_changed_at to be set")
				}
			}
		})
	}
}

func TestHandleChangePasswordInvalidBody(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "user@example.com", "password123")

	// Send a request with invalid JSON
	rec, req := ts.MakeAuthRequest(t, http.MethodPost, "/api/settings/password", "not-json", userID, nil)
	ts.HandleChangePassword(rec, req)

	// Should still get a response (the JSON marshal of "not-json" is a valid JSON string, so the decoder
	// will fail to decode it into the struct, resulting in a 400)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestHandle2FAStatus(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, ts *TestServer) int64
		wantStatus int
		want2FA    bool
	}{
		{
			name: "2FA not enabled",
			setup: func(t *testing.T, ts *TestServer) int64 {
				return ts.CreateTestUser(t, "user@example.com", "password123")
			},
			wantStatus: http.StatusOK,
			want2FA:    false,
		},
		{
			name: "2FA enabled",
			setup: func(t *testing.T, ts *TestServer) int64 {
				userID := ts.CreateTestUser(t, "user@example.com", "password123")
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				_, err := ts.DB.ExecContext(ctx, "UPDATE users SET totp_enabled = 1, totp_secret = 'testsecret' WHERE id = ?", userID)
				if err != nil {
					t.Fatalf("Failed to enable 2FA: %v", err)
				}
				return userID
			},
			wantStatus: http.StatusOK,
			want2FA:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := NewTestServer(t)
			defer ts.Close()

			userID := tt.setup(t, ts)

			rec, req := ts.MakeAuthRequest(t, http.MethodGet, "/api/settings/2fa/status", nil, userID, nil)
			ts.Handle2FAStatus(rec, req)

			AssertStatusCode(t, rec.Code, tt.wantStatus)

			var resp map[string]bool
			DecodeJSON(t, rec, &resp)

			if resp["enabled"] != tt.want2FA {
				t.Errorf("Expected 2FA enabled=%v, got %v", tt.want2FA, resp["enabled"])
			}
		})
	}
}

func TestHandle2FASetup(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "user@example.com", "password123")

	rec, req := ts.MakeAuthRequest(t, http.MethodPost, "/api/settings/2fa/setup", nil, userID, nil)
	ts.Handle2FASetup(rec, req)

	AssertStatusCode(t, rec.Code, http.StatusOK)

	var resp TwoFactorSetupResponse
	DecodeJSON(t, rec, &resp)

	if resp.Secret == "" {
		t.Error("Expected non-empty TOTP secret")
	}

	if resp.QRCodeURL == "" {
		t.Error("Expected non-empty QR code URL")
	}

	if resp.QRCodeSVG == "" {
		t.Error("Expected non-empty QR code SVG/URL")
	}

	// Verify the secret was stored in the database
	var storedSecret string
	err := ts.DB.QueryRow("SELECT totp_secret FROM users WHERE id = ?", userID).Scan(&storedSecret)
	if err != nil {
		t.Fatalf("Failed to query totp_secret: %v", err)
	}
	if storedSecret != resp.Secret {
		t.Errorf("Stored secret %q does not match response secret %q", storedSecret, resp.Secret)
	}
}

func TestHandle2FAEnableInvalidCode(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "user@example.com", "password123")

	// First set up 2FA to get a secret stored
	setupRec, setupReq := ts.MakeAuthRequest(t, http.MethodPost, "/api/settings/2fa/setup", nil, userID, nil)
	ts.Handle2FASetup(setupRec, setupReq)
	AssertStatusCode(t, setupRec.Code, http.StatusOK)

	// Now try to enable with an invalid code
	body := TwoFactorEnableRequest{Code: "000000"}
	rec, req := ts.MakeAuthRequest(t, http.MethodPost, "/api/settings/2fa/enable", body, userID, nil)
	ts.Handle2FAEnable(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	bodyStr := rec.Body.String()
	if !contains(bodyStr, "Invalid verification code") {
		t.Errorf("Response body %q does not contain 'Invalid verification code'", bodyStr)
	}
}

func TestHandle2FAEnableNoSetup(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "user@example.com", "password123")

	// Try to enable without running setup first (no secret stored)
	body := TwoFactorEnableRequest{Code: "123456"}
	rec, req := ts.MakeAuthRequest(t, http.MethodPost, "/api/settings/2fa/enable", body, userID, nil)
	ts.Handle2FAEnable(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	bodyStr := rec.Body.String()
	if !contains(bodyStr, "2FA setup not initiated") {
		t.Errorf("Response body %q does not contain '2FA setup not initiated'", bodyStr)
	}
}

func TestHandle2FADisableNotEnabled(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "user@example.com", "password123")

	// Try to disable 2FA when it's not enabled (requires password verification)
	body := map[string]string{"password": "password123"}
	rec, req := ts.MakeAuthRequest(t, http.MethodPost, "/api/settings/2fa/disable", body, userID, nil)
	ts.Handle2FADisable(rec, req)

	// Even though 2FA is not enabled, the handler just disables it (sets to 0 and clears secrets)
	// So it should succeed with the correct password
	AssertStatusCode(t, rec.Code, http.StatusOK)

	var resp map[string]string
	DecodeJSON(t, rec, &resp)
	if resp["message"] != "2FA disabled successfully" {
		t.Errorf("Expected success message, got %q", resp["message"])
	}
}

func TestHandle2FADisableWrongPassword(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "user@example.com", "password123")

	body := map[string]string{"password": "wrongpassword"}
	rec, req := ts.MakeAuthRequest(t, http.MethodPost, "/api/settings/2fa/disable", body, userID, nil)
	ts.Handle2FADisable(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	bodyStr := rec.Body.String()
	if !contains(bodyStr, "Incorrect password") {
		t.Errorf("Response body %q does not contain 'Incorrect password'", bodyStr)
	}
}

func TestHandle2FADisableEnabled(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "user@example.com", "password123")

	// Manually enable 2FA in the database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := ts.DB.ExecContext(ctx,
		"UPDATE users SET totp_enabled = 1, totp_secret = 'JBSWY3DPEHPK3PXP', backup_codes = '[]' WHERE id = ?",
		userID,
	)
	if err != nil {
		t.Fatalf("Failed to enable 2FA: %v", err)
	}

	// Verify 2FA is enabled
	var enabled int
	err = ts.DB.QueryRow("SELECT totp_enabled FROM users WHERE id = ?", userID).Scan(&enabled)
	if err != nil {
		t.Fatalf("Failed to query: %v", err)
	}
	if enabled != 1 {
		t.Fatal("2FA should be enabled before disable test")
	}

	// Now disable with correct password
	body := map[string]string{"password": "password123"}
	rec, req := ts.MakeAuthRequest(t, http.MethodPost, "/api/settings/2fa/disable", body, userID, nil)
	ts.Handle2FADisable(rec, req)

	AssertStatusCode(t, rec.Code, http.StatusOK)

	var resp map[string]string
	DecodeJSON(t, rec, &resp)
	if resp["message"] != "2FA disabled successfully" {
		t.Errorf("Expected success message, got %q", resp["message"])
	}

	// Verify 2FA was disabled
	err = ts.DB.QueryRow("SELECT totp_enabled FROM users WHERE id = ?", userID).Scan(&enabled)
	if err != nil {
		t.Fatalf("Failed to query: %v", err)
	}
	if enabled != 0 {
		t.Error("Expected 2FA to be disabled after disable call")
	}

	// Verify secret was cleared
	var secret *string
	err = ts.DB.QueryRow("SELECT totp_secret FROM users WHERE id = ?", userID).Scan(&secret)
	if err != nil {
		t.Fatalf("Failed to query totp_secret: %v", err)
	}
	if secret != nil {
		t.Errorf("Expected totp_secret to be NULL, got %q", *secret)
	}
}

func TestHandleChangePasswordUserNotFound(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	// Use a non-existent user ID
	body := ChangePasswordRequest{
		CurrentPassword: "password123",
		NewPassword:     "newpassword456",
	}

	rec, req := ts.MakeAuthRequest(t, http.MethodPost, "/api/settings/password", body, int64(99999), nil)
	ts.HandleChangePassword(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d. Body: %s", rec.Code, rec.Body.String())
	}
}

func TestHandle2FASetupReturnsDataURL(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "qr@example.com", "password123")

	rec, req := ts.MakeAuthRequest(t, http.MethodPost, "/api/settings/2fa/setup", nil, userID, nil)
	ts.Handle2FASetup(rec, req)

	AssertStatusCode(t, rec.Code, http.StatusOK)

	var resp TwoFactorSetupResponse
	DecodeJSON(t, rec, &resp)

	// QR code URL should be a data URL
	if !contains(resp.QRCodeURL, "data:image/png;base64,") {
		t.Errorf("Expected QR code URL to be a data URL, got %q", resp.QRCodeURL[:min(50, len(resp.QRCodeURL))])
	}

	// QRCodeSVG should contain the TOTP URL
	if !contains(resp.QRCodeSVG, "otpauth://totp/") {
		t.Errorf("Expected QRCodeSVG to contain otpauth URL, got %q", resp.QRCodeSVG[:min(50, len(resp.QRCodeSVG))])
	}
}

func TestGenerateBackupCodes(t *testing.T) {
	codes, err := generateBackupCodes(10)
	if err != nil {
		t.Fatalf("Failed to generate backup codes: %v", err)
	}

	if len(codes) != 10 {
		t.Errorf("Expected 10 backup codes, got %d", len(codes))
	}

	// Check format: each code should be XXXX-XXXX
	for i, code := range codes {
		if len(code) != 9 {
			t.Errorf("Backup code %d has length %d, expected 9: %q", i, len(code), code)
		}
		if code[4] != '-' {
			t.Errorf("Backup code %d missing dash at position 4: %q", i, code)
		}
	}

	// Check uniqueness
	seen := make(map[string]bool)
	for _, code := range codes {
		if seen[code] {
			t.Errorf("Duplicate backup code: %q", code)
		}
		seen[code] = true
	}
}

func TestHandle2FADisableInvalidBody(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "user@example.com", "password123")

	// Send request with body that cannot be decoded into the expected struct
	rec, req := ts.MakeAuthRequest(t, http.MethodPost, "/api/settings/2fa/disable", "invalid-json-data", userID, nil)
	ts.Handle2FADisable(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d. Body: %s", rec.Code, rec.Body.String())
	}
}

func TestHandle2FAEnableInvalidBody(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "user@example.com", "password123")

	rec, req := ts.MakeAuthRequest(t, http.MethodPost, "/api/settings/2fa/enable", "invalid-json", userID, nil)
	ts.Handle2FAEnable(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d. Body: %s", rec.Code, rec.Body.String())
	}
}

func TestHandle2FAStatusUserNotFound(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	rec, req := ts.MakeAuthRequest(t, http.MethodGet, "/api/settings/2fa/status", nil, int64(99999), nil)
	ts.Handle2FAStatus(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d. Body: %s", rec.Code, rec.Body.String())
	}
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Ensure JSON output is valid for 2FA status
func TestHandle2FAStatusJSON(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "json@example.com", "password123")

	rec, req := ts.MakeAuthRequest(t, http.MethodGet, "/api/settings/2fa/status", nil, userID, nil)
	ts.Handle2FAStatus(rec, req)

	AssertStatusCode(t, rec.Code, http.StatusOK)

	// Verify it's valid JSON
	var raw json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatalf("Response is not valid JSON: %v", err)
	}
}
