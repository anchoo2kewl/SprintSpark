package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap/zaptest"

	"sprintspark/internal/auth"
	"sprintspark/internal/config"
	"sprintspark/internal/db"
)

func TestAdminHandlers(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Setup test database
	cfg := &config.Config{
		DBPath:         ":memory:",
		MigrationsPath: "../db/migrations",
		JWTSecret:      "test-secret-key",
		JWTExpiryHours: 24,
	}

	dbCfg := db.Config{
		DBPath:         cfg.DBPath,
		MigrationsPath: cfg.MigrationsPath,
	}

	database, err := db.New(dbCfg, logger)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer database.Close()

	server := NewServer(database, cfg, logger)

	// Create test users
	adminPassword, err := auth.HashPassword("admin123")
	if err != nil {
		t.Fatalf("Failed to hash admin password: %v", err)
	}

	userPassword, err := auth.HashPassword("user123")
	if err != nil {
		t.Fatalf("Failed to hash user password: %v", err)
	}

	ctx := context.Background()

	// Create admin user
	query := `INSERT INTO users (email, password_hash, is_admin) VALUES (?, ?, 1) RETURNING id`
	var adminID int64
	err = database.QueryRowContext(ctx, query, "admin@test.com", adminPassword).Scan(&adminID)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	// Create regular user
	query = `INSERT INTO users (email, password_hash, is_admin) VALUES (?, ?, 0) RETURNING id`
	var userID int64
	err = database.QueryRowContext(ctx, query, "user@test.com", userPassword).Scan(&userID)
	if err != nil {
		t.Fatalf("Failed to create regular user: %v", err)
	}

	// Log some activity for the regular user
	activityQuery := `INSERT INTO user_activity (user_id, activity_type, ip_address, user_agent) VALUES (?, ?, ?, ?)`
	_, err = database.ExecContext(ctx, activityQuery, userID, "login", "127.0.0.1", "Test Browser")
	if err != nil {
		t.Fatalf("Failed to create activity log: %v", err)
	}
	_, err = database.ExecContext(ctx, activityQuery, userID, "failed_login", "192.168.1.1", "Test Browser")
	if err != nil {
		t.Fatalf("Failed to create failed activity log: %v", err)
	}

	// Generate tokens
	adminToken, err := auth.GenerateToken(adminID, "admin@test.com", cfg.JWTSecret, cfg.JWTExpiry())
	if err != nil {
		t.Fatalf("Failed to generate admin token: %v", err)
	}

	userToken, err := auth.GenerateToken(userID, "user@test.com", cfg.JWTSecret, cfg.JWTExpiry())
	if err != nil {
		t.Fatalf("Failed to generate user token: %v", err)
	}

	t.Run("Non-admin cannot access admin endpoints", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()

		// Use JWT middleware
		handler := server.JWTAuth(http.HandlerFunc(server.HandleGetUsers))
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status 403, got %d", w.Code)
		}
	})

	t.Run("Admin can list all users", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()

		handler := server.JWTAuth(http.HandlerFunc(server.HandleGetUsers))
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var users []UserWithStats
		if err := json.NewDecoder(w.Body).Decode(&users); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if len(users) != 2 {
			t.Errorf("Expected 2 users, got %d", len(users))
		}

		// Check that stats are included
		for _, u := range users {
			if u.Email == "user@test.com" {
				if u.LoginCount != 1 {
					t.Errorf("Expected login count 1, got %d", u.LoginCount)
				}
				if u.FailedAttempts != 1 {
					t.Errorf("Expected failed attempts 1, got %d", u.FailedAttempts)
				}
				if u.LastLoginIP == nil || *u.LastLoginIP != "127.0.0.1" {
					t.Errorf("Expected last login IP 127.0.0.1, got %v", u.LastLoginIP)
				}
			}
		}
	})

	t.Run("Admin can view user activity", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/admin/users/2/activity", nil)
		req.SetPathValue("id", "2")
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()

		handler := server.JWTAuth(http.HandlerFunc(server.HandleGetUserActivity))
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		var activities []UserActivity
		if err := json.NewDecoder(w.Body).Decode(&activities); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if len(activities) != 2 {
			t.Errorf("Expected 2 activities, got %d", len(activities))
		}
	})

	t.Run("Admin can grant admin privileges", func(t *testing.T) {
		body := map[string]interface{}{
			"is_admin": true,
		}
		bodyJSON, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPatch, "/api/admin/users/2/admin", bytes.NewReader(bodyJSON))
		req.SetPathValue("id", "2")
		req.Header.Set("Authorization", "Bearer "+adminToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler := server.JWTAuth(http.HandlerFunc(server.HandleUpdateUserAdmin))
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		// Verify user is now admin
		var isAdmin bool
		err := database.QueryRowContext(ctx, "SELECT is_admin FROM users WHERE id = ?", userID).Scan(&isAdmin)
		if err != nil {
			t.Fatalf("Failed to query user: %v", err)
		}
		if !isAdmin {
			t.Error("User should be admin after update")
		}
	})

	t.Run("Admin can revoke admin privileges", func(t *testing.T) {
		body := map[string]interface{}{
			"is_admin": false,
		}
		bodyJSON, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPatch, "/api/admin/users/2/admin", bytes.NewReader(bodyJSON))
		req.SetPathValue("id", "2")
		req.Header.Set("Authorization", "Bearer "+adminToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler := server.JWTAuth(http.HandlerFunc(server.HandleUpdateUserAdmin))
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		// Verify user is no longer admin
		var isAdmin bool
		err := database.QueryRowContext(ctx, "SELECT is_admin FROM users WHERE id = ?", userID).Scan(&isAdmin)
		if err != nil {
			t.Fatalf("Failed to query user: %v", err)
		}
		if isAdmin {
			t.Error("User should not be admin after revoke")
		}
	})
}

func TestActivityLogging(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Setup test database
	cfg := &config.Config{
		DBPath:         ":memory:",
		MigrationsPath: "../db/migrations",
		JWTSecret:      "test-secret-key",
		JWTExpiryHours: 24,
	}

	dbCfg := db.Config{
		DBPath:         cfg.DBPath,
		MigrationsPath: cfg.MigrationsPath,
	}

	database, err := db.New(dbCfg, logger)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer database.Close()

	server := NewServer(database, cfg, logger)
	ctx := context.Background()

	// Create test user
	password, _ := auth.HashPassword("test123")
	query := `INSERT INTO users (email, password_hash) VALUES (?, ?) RETURNING id`
	var userID int64
	err = database.QueryRowContext(ctx, query, "test@example.com", password).Scan(&userID)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	t.Run("Successful login logs activity", func(t *testing.T) {
		body := LoginRequest{
			Email:    "test@example.com",
			Password: "test123",
		}
		bodyJSON, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(bodyJSON))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Forwarded-For", "1.2.3.4")
		req.Header.Set("User-Agent", "TestAgent/1.0")
		w := httptest.NewRecorder()

		server.HandleLogin(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
		}

		// Wait for goroutine to complete
		time.Sleep(100 * time.Millisecond)

		// Verify activity was logged
		var count int
		err := database.QueryRowContext(ctx, "SELECT COUNT(*) FROM user_activity WHERE user_id = ? AND activity_type = 'login'", userID).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to query activity: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected 1 login activity, got %d", count)
		}

		// Verify IP and user agent were logged
		var ipAddr, userAgent string
		err = database.QueryRowContext(ctx, "SELECT ip_address, user_agent FROM user_activity WHERE user_id = ? AND activity_type = 'login' ORDER BY created_at DESC LIMIT 1", userID).Scan(&ipAddr, &userAgent)
		if err != nil {
			t.Fatalf("Failed to query activity details: %v", err)
		}
		if ipAddr != "1.2.3.4" {
			t.Errorf("Expected IP 1.2.3.4, got %s", ipAddr)
		}
		if userAgent != "TestAgent/1.0" {
			t.Errorf("Expected User-Agent TestAgent/1.0, got %s", userAgent)
		}
	})

	t.Run("Failed login logs activity", func(t *testing.T) {
		body := LoginRequest{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}
		bodyJSON, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(bodyJSON))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Forwarded-For", "5.6.7.8")
		w := httptest.NewRecorder()

		server.HandleLogin(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}

		// Wait for goroutine to complete
		time.Sleep(100 * time.Millisecond)

		// Verify failed activity was logged
		var count int
		err := database.QueryRowContext(ctx, "SELECT COUNT(*) FROM user_activity WHERE user_id = ? AND activity_type = 'failed_login'", userID).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to query activity: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected 1 failed_login activity, got %d", count)
		}
	})
}
