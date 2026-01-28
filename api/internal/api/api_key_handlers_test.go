package api

import (
	"context"
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

func intPtr(i int) *int {
	return &i
}
