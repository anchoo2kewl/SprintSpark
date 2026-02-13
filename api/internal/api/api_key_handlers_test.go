package api

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"sprintspark/internal/db"
)

func TestHandleCreateAPIKey(t *testing.T) {
	tests := []struct {
		name          string
		keyName       string
		expiresIn     *int
		wantStatus    int
		wantError     string
		wantErrorCode string
	}{
		{
			name:       "valid API key",
			keyName:    "Test API Key",
			wantStatus: http.StatusCreated,
		},
		{
			name:          "missing name",
			keyName:       "",
			wantStatus:    http.StatusBadRequest,
			wantError:     "name is required",
			wantErrorCode: "validation_error",
		},
		{
			name:       "with expiration",
			keyName:    "Expiring Key",
			expiresIn:  intPtr(90),
			wantStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := NewTestServer(t)
			defer ts.Close()

			userID := ts.CreateTestUser(t, "apikey@example.com", "password123")

			req := CreateAPIKeyRequest{
				Name:      tt.keyName,
				ExpiresIn: tt.expiresIn,
			}

			rec, httpReq := MakeRequest(t, http.MethodPost, "/api/api-keys", req, nil)

			// Add user to context
			ctx := context.WithValue(httpReq.Context(), UserIDKey, userID)
			httpReq = httpReq.WithContext(ctx)

			ts.HandleCreateAPIKey(rec, httpReq)

			AssertStatusCode(t, rec.Code, tt.wantStatus)

			if tt.wantError != "" {
				AssertError(t, rec, tt.wantStatus, tt.wantError, tt.wantErrorCode)
			} else {
				var resp CreateAPIKeyResponse
				DecodeJSON(t, rec, &resp)

				if resp.Name != tt.keyName {
					t.Errorf("Expected name %q, got %q", tt.keyName, resp.Name)
				}

				if resp.Key == "" {
					t.Error("Expected non-empty API key")
				}

				if len(resp.KeyPrefix) != 8 {
					t.Errorf("Expected key prefix length 8, got %d", len(resp.KeyPrefix))
				}
			}
		})
	}
}

func TestHandleListAPIKeys(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "apikey@example.com", "password123")

	// Create some test API keys
	_, err := ts.DB.CreateAPIKey(context.Background(), userID, "Key 1", nil)
	if err != nil {
		t.Fatalf("Failed to create test API key: %v", err)
	}

	rec, httpReq := MakeRequest(t, http.MethodGet, "/api/api-keys", nil, nil)

	// Add user to context
	ctx := context.WithValue(httpReq.Context(), UserIDKey, userID)
	httpReq = httpReq.WithContext(ctx)

	ts.HandleListAPIKeys(rec, httpReq)

	AssertStatusCode(t, rec.Code, http.StatusOK)

	var keys []APIKeyResponse
	DecodeJSON(t, rec, &keys)

	if len(keys) != 1 {
		t.Errorf("Expected 1 API key, got %d", len(keys))
	}
}

func TestGenerateAPIKey(t *testing.T) {
	key, keyHash, prefix, err := db.GenerateAPIKey()
	if err != nil {
		t.Fatalf("Failed to generate API key: %v", err)
	}

	if len(key) == 0 {
		t.Error("Expected non-empty key")
	}

	if len(keyHash) == 0 {
		t.Error("Expected non-empty key hash")
	}

	if len(prefix) != 8 {
		t.Errorf("Expected prefix length 8, got %d", len(prefix))
	}

	// Verify hash matches
	expectedHash := db.HashAPIKey(key)
	if keyHash != expectedHash {
		t.Error("Key hash does not match expected hash")
	}
}

func TestHandleDeleteAPIKey(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(*TestServer) (userID int64, keyID string)
		wantStatus    int
		wantError     string
		wantErrorCode string
		noAuth        bool
	}{
		{
			name: "delete own API key successfully",
			setupFunc: func(ts *TestServer) (int64, string) {
				userID := ts.CreateTestUser(t, "user@example.com", "password123")
				apiKey, err := ts.DB.CreateAPIKey(context.Background(), userID, "My Key", nil)
				if err != nil {
					t.Fatalf("Failed to create API key: %v", err)
				}
				return userID, fmt.Sprintf("%d", apiKey.ID)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "delete non-existent key returns not found",
			setupFunc: func(ts *TestServer) (int64, string) {
				userID := ts.CreateTestUser(t, "user@example.com", "password123")
				return userID, "99999"
			},
			wantStatus:    http.StatusNotFound,
			wantError:     "API key not found",
			wantErrorCode: "not_found",
		},
		{
			name: "cannot delete another user's key",
			setupFunc: func(ts *TestServer) (int64, string) {
				ownerID := ts.CreateTestUser(t, "owner@example.com", "password123")
				apiKey, err := ts.DB.CreateAPIKey(context.Background(), ownerID, "Owner Key", nil)
				if err != nil {
					t.Fatalf("Failed to create API key: %v", err)
				}
				// Return a different user as the requester
				otherUserID := ts.CreateTestUser(t, "other@example.com", "password123")
				return otherUserID, fmt.Sprintf("%d", apiKey.ID)
			},
			wantStatus:    http.StatusNotFound,
			wantError:     "API key not found",
			wantErrorCode: "not_found",
		},
		{
			name: "invalid key ID format",
			setupFunc: func(ts *TestServer) (int64, string) {
				userID := ts.CreateTestUser(t, "user@example.com", "password123")
				return userID, "not-a-number"
			},
			wantStatus:    http.StatusBadRequest,
			wantError:     "invalid API key ID",
			wantErrorCode: "invalid_request",
		},
		{
			name:          "unauthenticated request",
			noAuth:        true,
			wantStatus:    http.StatusUnauthorized,
			wantError:     "unauthorized",
			wantErrorCode: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := NewTestServer(t)
			defer ts.Close()

			if tt.noAuth {
				rec, req := MakeRequest(t, http.MethodDelete, "/api/api-keys/1", nil, nil)
				ts.HandleDeleteAPIKey(rec, req)

				AssertStatusCode(t, rec.Code, tt.wantStatus)
				if tt.wantError != "" {
					AssertError(t, rec, tt.wantStatus, tt.wantError, tt.wantErrorCode)
				}
				return
			}

			userID, keyIDStr := tt.setupFunc(ts)

			urlParams := map[string]string{"id": keyIDStr}
			rec, req := ts.MakeAuthRequest(t, http.MethodDelete, "/api/api-keys/"+keyIDStr, nil, userID, urlParams)
			ts.HandleDeleteAPIKey(rec, req)

			AssertStatusCode(t, rec.Code, tt.wantStatus)

			if tt.wantError != "" {
				AssertError(t, rec, tt.wantStatus, tt.wantError, tt.wantErrorCode)
			}

			// For successful delete, verify the key is actually gone
			if tt.wantStatus == http.StatusNoContent {
				keys, err := ts.DB.GetAPIKeysByUserID(context.Background(), userID)
				if err != nil {
					t.Fatalf("Failed to list API keys: %v", err)
				}
				if len(keys) != 0 {
					t.Errorf("Expected 0 API keys after delete, got %d", len(keys))
				}
			}
		})
	}
}

func intPtr(i int) *int {
	return &i
}

func TestHandleCreateAPIKey_Unauthenticated(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	rec, req := MakeRequest(t, http.MethodPost, "/api/api-keys", CreateAPIKeyRequest{Name: "Test"}, nil)
	ts.HandleCreateAPIKey(rec, req)

	AssertError(t, rec, http.StatusUnauthorized, "unauthorized", "unauthorized")
}

func TestHandleCreateAPIKey_InvalidBody(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "user@example.com", "password123")

	rec, req := MakeRequest(t, http.MethodPost, "/api/api-keys", "not-json", nil)
	ctx := context.WithValue(req.Context(), UserIDKey, userID)
	req = req.WithContext(ctx)
	ts.HandleCreateAPIKey(rec, req)

	AssertError(t, rec, http.StatusBadRequest, "invalid request body", "invalid_request")
}

func TestHandleCreateAPIKey_NameTooLong(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "user@example.com", "password123")

	longName := ""
	for i := 0; i < 101; i++ {
		longName += "a"
	}
	req := CreateAPIKeyRequest{Name: longName}

	rec, httpReq := MakeRequest(t, http.MethodPost, "/api/api-keys", req, nil)
	ctx := context.WithValue(httpReq.Context(), UserIDKey, userID)
	httpReq = httpReq.WithContext(ctx)
	ts.HandleCreateAPIKey(rec, httpReq)

	AssertError(t, rec, http.StatusBadRequest, "name must be 100 characters or less", "validation_error")
}

func TestHandleCreateAPIKey_ExpiresInNegative(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "user@example.com", "password123")

	neg := -1
	req := CreateAPIKeyRequest{Name: "Test", ExpiresIn: &neg}

	rec, httpReq := MakeRequest(t, http.MethodPost, "/api/api-keys", req, nil)
	ctx := context.WithValue(httpReq.Context(), UserIDKey, userID)
	httpReq = httpReq.WithContext(ctx)
	ts.HandleCreateAPIKey(rec, httpReq)

	AssertError(t, rec, http.StatusBadRequest, "expires_in must be positive", "validation_error")
}

func TestHandleCreateAPIKey_ExpiresInTooLarge(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "user@example.com", "password123")

	big := 366
	req := CreateAPIKeyRequest{Name: "Test", ExpiresIn: &big}

	rec, httpReq := MakeRequest(t, http.MethodPost, "/api/api-keys", req, nil)
	ctx := context.WithValue(httpReq.Context(), UserIDKey, userID)
	httpReq = httpReq.WithContext(ctx)
	ts.HandleCreateAPIKey(rec, httpReq)

	AssertError(t, rec, http.StatusBadRequest, "expires_in cannot exceed 365 days", "validation_error")
}

func TestHandleListAPIKeys_Unauthenticated(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	rec, req := MakeRequest(t, http.MethodGet, "/api/api-keys", nil, nil)
	ts.HandleListAPIKeys(rec, req)

	AssertError(t, rec, http.StatusUnauthorized, "unauthorized", "unauthorized")
}

func TestHandleListAPIKeys_Empty(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	userID := ts.CreateTestUser(t, "user@example.com", "password123")

	rec, req := MakeRequest(t, http.MethodGet, "/api/api-keys", nil, nil)
	ctx := context.WithValue(req.Context(), UserIDKey, userID)
	req = req.WithContext(ctx)
	ts.HandleListAPIKeys(rec, req)

	AssertStatusCode(t, rec.Code, http.StatusOK)

	var keys []APIKeyResponse
	DecodeJSON(t, rec, &keys)
	if len(keys) != 0 {
		t.Errorf("Expected 0 keys, got %d", len(keys))
	}
}
