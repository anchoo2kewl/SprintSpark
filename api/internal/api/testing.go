package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"sprintspark/internal/auth"
	"sprintspark/internal/config"
	"sprintspark/internal/db"
)

// TestServer holds test server dependencies
type TestServer struct {
	*Server
	DB *db.DB
}

// NewTestServer creates a new test server with in-memory SQLite database
func NewTestServer(t *testing.T) *TestServer {
	t.Helper()

	// Create in-memory database
	cfg := db.Config{
		DBPath:         ":memory:",
		MigrationsPath: "./../../internal/db/migrations",
	}

	database, err := db.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create test config
	testCfg := &config.Config{
		JWTSecret:      "test-secret-key",
		JWTExpiryHours: 24,
	}

	server := NewServer(database, testCfg)

	return &TestServer{
		Server: server,
		DB:     database,
	}
}

// Close cleans up test server resources
func (ts *TestServer) Close() {
	if ts.DB != nil {
		ts.DB.Close()
	}
}

// CreateTestUser creates a user for testing and returns the user ID
func (ts *TestServer) CreateTestUser(t *testing.T, email, password string) int64 {
	t.Helper()

	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `INSERT INTO users (email, password_hash) VALUES (?, ?) RETURNING id`
	var userID int64
	err = ts.DB.QueryRowContext(ctx, query, email, hashedPassword).Scan(&userID)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return userID
}

// GenerateTestToken generates a JWT token for testing
func (ts *TestServer) GenerateTestToken(t *testing.T, userID int64, email string) string {
	t.Helper()

	token, err := auth.GenerateToken(userID, email, ts.config.JWTSecret, ts.config.JWTExpiry())
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}

	return token
}

// MakeRequest is a helper to make HTTP requests in tests
// Returns both the ResponseRecorder and the Request for testing
func MakeRequest(t *testing.T, method, path string, body interface{}, headers map[string]string) (*httptest.ResponseRecorder, *http.Request) {
	t.Helper()

	var reqBody []byte
	var err error
	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return httptest.NewRecorder(), req
}

// DecodeJSON decodes a JSON response into the provided interface
func DecodeJSON(t *testing.T, rec *httptest.ResponseRecorder, v interface{}) {
	t.Helper()

	if err := json.NewDecoder(rec.Body).Decode(v); err != nil {
		t.Fatalf("Failed to decode JSON response: %v", err)
	}
}

// AssertStatusCode checks if the response status code matches expected
func AssertStatusCode(t *testing.T, got, want int) {
	t.Helper()

	if got != want {
		t.Errorf("Status code mismatch: got %d, want %d", got, want)
	}
}

// AssertJSONField checks if a JSON field has the expected value
func AssertJSONField(t *testing.T, data map[string]interface{}, field string, want interface{}) {
	t.Helper()

	got, ok := data[field]
	if !ok {
		t.Errorf("Field %q not found in response", field)
		return
	}

	if got != want {
		t.Errorf("Field %q mismatch: got %v, want %v", field, got, want)
	}
}

// AssertError checks if the error response matches expected error and code
func AssertError(t *testing.T, rec *httptest.ResponseRecorder, wantCode int, wantErrorContains, wantCodeContains string) {
	t.Helper()

	AssertStatusCode(t, rec.Code, wantCode)

	var errResp ErrorResponse
	DecodeJSON(t, rec, &errResp)

	if wantErrorContains != "" && !contains(errResp.Error, wantErrorContains) {
		t.Errorf("Error message %q does not contain %q", errResp.Error, wantErrorContains)
	}

	if wantCodeContains != "" && !contains(errResp.Code, wantCodeContains) {
		t.Errorf("Error code %q does not contain %q", errResp.Code, wantCodeContains)
	}
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || indexOf(s, substr) >= 0)
}

// indexOf returns the index of the first occurrence of substr in s, or -1
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
