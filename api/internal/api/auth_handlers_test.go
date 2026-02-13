package api

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestHandleSignup(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		password      string
		inviteCode    string
		wantStatus    int
		wantError     string
		wantErrorCode string
		checkToken    bool
		setupFunc     func(*TestServer) string // returns invite code
	}{
		{
			name:       "valid signup",
			email:      "newuser@example.com",
			password:   "password123",
			wantStatus: http.StatusCreated,
			checkToken: true,
			setupFunc: func(ts *TestServer) string {
				inviterID := ts.CreateTestUser(t, "inviter@example.com", "password123")
				return ts.CreateTestInvite(t, inviterID)
			},
		},
		{
			name:          "duplicate email",
			email:         "existing@example.com",
			password:      "password123",
			wantStatus:    http.StatusConflict,
			wantError:     "email already exists",
			wantErrorCode: "email_exists",
			setupFunc: func(ts *TestServer) string {
				inviterID := ts.CreateTestUser(t, "existing@example.com", "password123")
				return ts.CreateTestInvite(t, inviterID)
			},
		},
		{
			name:          "missing email",
			email:         "",
			password:      "password123",
			wantStatus:    http.StatusBadRequest,
			wantError:     "email is required",
			wantErrorCode: "validation_error",
		},
		{
			name:          "invalid email format",
			email:         "notanemail",
			password:      "password123",
			wantStatus:    http.StatusBadRequest,
			wantError:     "invalid email format",
			wantErrorCode: "validation_error",
		},
		{
			name:          "missing password",
			email:         "user@example.com",
			password:      "",
			wantStatus:    http.StatusBadRequest,
			wantError:     "password is required",
			wantErrorCode: "validation_error",
		},
		{
			name:          "password too short",
			email:         "user@example.com",
			password:      "short",
			wantStatus:    http.StatusBadRequest,
			wantError:     "at least 8 characters",
			wantErrorCode: "validation_error",
		},
		{
			name:          "missing invite code",
			email:         "user@example.com",
			password:      "password123",
			wantStatus:    http.StatusBadRequest,
			wantError:     "invite code is required",
			wantErrorCode: "invite_required",
		},
		{
			name:          "invalid invite code",
			email:         "user@example.com",
			password:      "password123",
			inviteCode:    "bad-code",
			wantStatus:    http.StatusBadRequest,
			wantError:     "invalid invite code",
			wantErrorCode: "invalid_invite",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := NewTestServer(t)
			defer ts.Close()

			inviteCode := tt.inviteCode
			if tt.setupFunc != nil {
				inviteCode = tt.setupFunc(ts)
			}

			req := SignupRequest{
				Email:      tt.email,
				Password:   tt.password,
				InviteCode: inviteCode,
			}

			rec, httpReq := MakeRequest(t, http.MethodPost, "/api/auth/signup", req, nil)
			ts.HandleSignup(rec, httpReq)

			AssertStatusCode(t, rec.Code, tt.wantStatus)

			if tt.wantError != "" {
				AssertError(t, rec, tt.wantStatus, tt.wantError, tt.wantErrorCode)
			}

			if tt.checkToken {
				var resp AuthResponse
				DecodeJSON(t, rec, &resp)

				if resp.Token == "" {
					t.Error("Expected token in response, got empty string")
				}

				if resp.User.Email != tt.email {
					t.Errorf("User email mismatch: got %s, want %s", resp.User.Email, tt.email)
				}

				if resp.User.ID == 0 {
					t.Error("Expected user ID to be set")
				}
			}
		})
	}
}

func TestHandleLogin(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		password      string
		wantStatus    int
		wantError     string
		wantErrorCode string
		checkToken    bool
		setupFunc     func(*TestServer)
	}{
		{
			name:       "valid login",
			email:      "user@example.com",
			password:   "password123",
			wantStatus: http.StatusOK,
			checkToken: true,
			setupFunc: func(ts *TestServer) {
				ts.CreateTestUser(t, "user@example.com", "password123")
			},
		},
		{
			name:          "wrong password",
			email:         "user@example.com",
			password:      "wrongpassword",
			wantStatus:    http.StatusUnauthorized,
			wantError:     "invalid email or password",
			wantErrorCode: "invalid_credentials",
			setupFunc: func(ts *TestServer) {
				ts.CreateTestUser(t, "user@example.com", "password123")
			},
		},
		{
			name:          "user not found",
			email:         "nonexistent@example.com",
			password:      "password123",
			wantStatus:    http.StatusUnauthorized,
			wantError:     "invalid email or password",
			wantErrorCode: "invalid_credentials",
		},
		{
			name:          "missing email",
			email:         "",
			password:      "password123",
			wantStatus:    http.StatusBadRequest,
			wantError:     "email and password are required",
			wantErrorCode: "validation_error",
		},
		{
			name:          "missing password",
			email:         "user@example.com",
			password:      "",
			wantStatus:    http.StatusBadRequest,
			wantError:     "email and password are required",
			wantErrorCode: "validation_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := NewTestServer(t)
			defer ts.Close()

			if tt.setupFunc != nil {
				tt.setupFunc(ts)
			}

			req := LoginRequest{
				Email:    tt.email,
				Password: tt.password,
			}

			rec, httpReq := MakeRequest(t, http.MethodPost, "/api/auth/login", req, nil)
			ts.HandleLogin(rec, httpReq)

			AssertStatusCode(t, rec.Code, tt.wantStatus)

			if tt.wantError != "" {
				AssertError(t, rec, tt.wantStatus, tt.wantError, tt.wantErrorCode)
			}

			if tt.checkToken {
				var resp AuthResponse
				DecodeJSON(t, rec, &resp)

				if resp.Token == "" {
					t.Error("Expected token in response, got empty string")
				}

				if resp.User.Email != tt.email {
					t.Errorf("User email mismatch: got %s, want %s", resp.User.Email, tt.email)
				}
			}
		})
	}
}

func TestHandleMe(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(*TestServer) string
		wantStatus    int
		wantError     string
		wantErrorCode string
	}{
		{
			name: "valid request",
			setupFunc: func(ts *TestServer) string {
				userID := ts.CreateTestUser(t, "user@example.com", "password123")
				return ts.GenerateTestToken(t, userID, "user@example.com")
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "missing token",
			setupFunc: func(ts *TestServer) string {
				return ""
			},
			wantStatus:    http.StatusUnauthorized,
			wantError:     "missing authorization header",
			wantErrorCode: "unauthorized",
		},
		{
			name: "invalid token",
			setupFunc: func(ts *TestServer) string {
				return "invalid-token"
			},
			wantStatus:    http.StatusUnauthorized,
			wantError:     "invalid or expired token",
			wantErrorCode: "unauthorized",
		},
		{
			name: "user not found",
			setupFunc: func(ts *TestServer) string {
				// Create token for non-existent user
				return ts.GenerateTestToken(t, 99999, "nonexistent@example.com")
			},
			wantStatus:    http.StatusNotFound,
			wantError:     "user not found",
			wantErrorCode: "not_found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := NewTestServer(t)
			defer ts.Close()

			token := tt.setupFunc(ts)

			headers := map[string]string{}
			if token != "" {
				headers["Authorization"] = fmt.Sprintf("Bearer %s", token)
			}

			rec, httpReq := MakeRequest(t, http.MethodGet, "/api/me", nil, headers)

			// Add JWT middleware context if token is valid
			if token != "" && tt.wantStatus == http.StatusOK {
				ts.JWTAuth(http.HandlerFunc(ts.HandleMe)).ServeHTTP(rec, httpReq)
			} else if token != "" {
				// Test middleware directly for auth errors
				ts.JWTAuth(http.HandlerFunc(ts.HandleMe)).ServeHTTP(rec, httpReq)
			} else {
				// No token, middleware should reject
				ts.JWTAuth(http.HandlerFunc(ts.HandleMe)).ServeHTTP(rec, httpReq)
			}

			AssertStatusCode(t, rec.Code, tt.wantStatus)

			if tt.wantError != "" {
				AssertError(t, rec, tt.wantStatus, tt.wantError, tt.wantErrorCode)
			} else {
				var user User
				DecodeJSON(t, rec, &user)

				if user.Email != "user@example.com" {
					t.Errorf("User email mismatch: got %s, want user@example.com", user.Email)
				}

				if user.ID == 0 {
					t.Error("Expected user ID to be set")
				}
			}
		})
	}
}

func TestHandleUpdateProfile(t *testing.T) {
	tests := []struct {
		name          string
		body          interface{}
		setupFunc     func(*TestServer) int64 // returns userID
		wantStatus    int
		wantError     string
		wantErrorCode string
		wantName      string
		noAuth        bool
	}{
		{
			name: "update name successfully",
			body: UpdateProfileRequest{Name: "John Doe"},
			setupFunc: func(ts *TestServer) int64 {
				return ts.CreateTestUser(t, "user@example.com", "password123")
			},
			wantStatus: http.StatusOK,
			wantName:   "John Doe",
		},
		{
			name: "set empty name",
			body: UpdateProfileRequest{Name: ""},
			setupFunc: func(ts *TestServer) int64 {
				return ts.CreateTestUser(t, "user@example.com", "password123")
			},
			wantStatus: http.StatusOK,
			wantName:   "",
		},
		{
			name: "name at 100 char limit",
			body: UpdateProfileRequest{Name: strings.Repeat("a", 100)},
			setupFunc: func(ts *TestServer) int64 {
				return ts.CreateTestUser(t, "user@example.com", "password123")
			},
			wantStatus: http.StatusOK,
			wantName:   strings.Repeat("a", 100),
		},
		{
			name: "name exceeds 100 chars",
			body: UpdateProfileRequest{Name: strings.Repeat("a", 101)},
			setupFunc: func(ts *TestServer) int64 {
				return ts.CreateTestUser(t, "user@example.com", "password123")
			},
			wantStatus:    http.StatusBadRequest,
			wantError:     "name must be 100 characters or less",
			wantErrorCode: "validation_error",
		},
		{
			name:          "unauthenticated request",
			body:          UpdateProfileRequest{Name: "Test"},
			noAuth:        true,
			wantStatus:    http.StatusUnauthorized,
			wantError:     "user not authenticated",
			wantErrorCode: "unauthorized",
		},
		{
			name: "invalid request body",
			body: "not-json",
			setupFunc: func(ts *TestServer) int64 {
				return ts.CreateTestUser(t, "user@example.com", "password123")
			},
			wantStatus:    http.StatusBadRequest,
			wantError:     "invalid request body",
			wantErrorCode: "invalid_request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := NewTestServer(t)
			defer ts.Close()

			var userID int64
			if tt.setupFunc != nil {
				userID = tt.setupFunc(ts)
			}

			if tt.noAuth {
				rec, req := MakeRequest(t, http.MethodPatch, "/api/me/profile", tt.body, nil)
				ts.HandleUpdateProfile(rec, req)

				AssertStatusCode(t, rec.Code, tt.wantStatus)
				if tt.wantError != "" {
					AssertError(t, rec, tt.wantStatus, tt.wantError, tt.wantErrorCode)
				}
				return
			}

			rec, req := ts.MakeAuthRequest(t, http.MethodPatch, "/api/me/profile", tt.body, userID, nil)
			ts.HandleUpdateProfile(rec, req)

			AssertStatusCode(t, rec.Code, tt.wantStatus)

			if tt.wantError != "" {
				AssertError(t, rec, tt.wantStatus, tt.wantError, tt.wantErrorCode)
			} else {
				var user User
				DecodeJSON(t, rec, &user)

				if user.Name != tt.wantName {
					t.Errorf("User name = %q, want %q", user.Name, tt.wantName)
				}
				if user.Email != "user@example.com" {
					t.Errorf("User email = %q, want %q", user.Email, "user@example.com")
				}
				if user.ID == 0 {
					t.Error("Expected user ID to be set")
				}
			}
		})
	}
}

// TestCompleteAuthFlow tests the complete authentication flow
func TestCompleteAuthFlow(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	email := "flowtest@example.com"
	password := "password123"

	// Create an inviter user and invite code for signup
	inviterID := ts.CreateTestUser(t, "inviter@example.com", "password123")
	inviteCode := ts.CreateTestInvite(t, inviterID)

	// Step 1: Sign up
	signupReq := SignupRequest{
		Email:      email,
		Password:   password,
		InviteCode: inviteCode,
	}

	signupRec, signupHttpReq := MakeRequest(t, http.MethodPost, "/api/auth/signup", signupReq, nil)
	ts.HandleSignup(signupRec, signupHttpReq)

	AssertStatusCode(t, signupRec.Code, http.StatusCreated)

	var signupResp AuthResponse
	DecodeJSON(t, signupRec, &signupResp)

	if signupResp.Token == "" {
		t.Fatal("Expected token from signup")
	}

	signupToken := signupResp.Token

	// Step 2: Log in with same credentials
	loginReq := LoginRequest{
		Email:    email,
		Password: password,
	}

	loginRec, loginHttpReq := MakeRequest(t, http.MethodPost, "/api/auth/login", loginReq, nil)
	ts.HandleLogin(loginRec, loginHttpReq)

	AssertStatusCode(t, loginRec.Code, http.StatusOK)

	var loginResp AuthResponse
	DecodeJSON(t, loginRec, &loginResp)

	if loginResp.Token == "" {
		t.Fatal("Expected token from login")
	}

	// Step 3: Access protected endpoint with token from signup
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", signupToken),
	}

	meRec, meHttpReq := MakeRequest(t, http.MethodGet, "/api/me", nil, headers)
	ts.JWTAuth(http.HandlerFunc(ts.HandleMe)).ServeHTTP(meRec, meHttpReq)

	AssertStatusCode(t, meRec.Code, http.StatusOK)

	var user User
	DecodeJSON(t, meRec, &user)

	if user.Email != email {
		t.Errorf("User email mismatch: got %s, want %s", user.Email, email)
	}

	// Step 4: Access protected endpoint with token from login
	headers["Authorization"] = fmt.Sprintf("Bearer %s", loginResp.Token)

	meRec2, meHttpReq2 := MakeRequest(t, http.MethodGet, "/api/me", nil, headers)
	ts.JWTAuth(http.HandlerFunc(ts.HandleMe)).ServeHTTP(meRec2, meHttpReq2)

	AssertStatusCode(t, meRec2.Code, http.StatusOK)

	var user2 User
	DecodeJSON(t, meRec2, &user2)

	if user2.Email != email {
		t.Errorf("User email mismatch: got %s, want %s", user2.Email, email)
	}

	if user2.ID != user.ID {
		t.Errorf("User ID mismatch: got %d, want %d", user2.ID, user.ID)
	}
}
