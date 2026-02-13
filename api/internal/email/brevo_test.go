package email

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap/zaptest"
)

func newTestService(t *testing.T, handler http.HandlerFunc) *BrevoService {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	svc := NewBrevoService("test-api-key", "noreply@test.com", "TaskAI Test", zaptest.NewLogger(t))
	svc.apiBaseURL = server.URL
	return svc
}

func TestNewBrevoService(t *testing.T) {
	logger := zaptest.NewLogger(t)
	svc := NewBrevoService("key123", "sender@test.com", "Test Sender", logger)

	if svc.apiKey != "key123" {
		t.Errorf("Expected apiKey 'key123', got '%s'", svc.apiKey)
	}
	if svc.senderEmail != "sender@test.com" {
		t.Errorf("Expected senderEmail 'sender@test.com', got '%s'", svc.senderEmail)
	}
	if svc.senderName != "Test Sender" {
		t.Errorf("Expected senderName 'Test Sender', got '%s'", svc.senderName)
	}
	if svc.apiBaseURL != "https://api.brevo.com/v3" {
		t.Errorf("Expected default apiBaseURL, got '%s'", svc.apiBaseURL)
	}
	if svc.httpClient == nil {
		t.Error("Expected httpClient to be set")
	}
}

func TestSendEmail_Success(t *testing.T) {
	var receivedBody brevoEmailRequest
	var receivedAPIKey string

	svc := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		receivedAPIKey = r.Header.Get("api-key")
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &receivedBody)

		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/smtp/email" {
			t.Errorf("Expected path /smtp/email, got %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"messageId":"abc123"}`))
	})

	err := svc.SendEmail(context.Background(), "recipient@test.com", "Test Subject", "<h1>Hello</h1>")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if receivedAPIKey != "test-api-key" {
		t.Errorf("Expected api-key header 'test-api-key', got '%s'", receivedAPIKey)
	}
	if receivedBody.Sender.Email != "noreply@test.com" {
		t.Errorf("Expected sender email 'noreply@test.com', got '%s'", receivedBody.Sender.Email)
	}
	if receivedBody.Sender.Name != "TaskAI Test" {
		t.Errorf("Expected sender name 'TaskAI Test', got '%s'", receivedBody.Sender.Name)
	}
	if len(receivedBody.To) != 1 || receivedBody.To[0].Email != "recipient@test.com" {
		t.Errorf("Expected recipient 'recipient@test.com', got %+v", receivedBody.To)
	}
	if receivedBody.Subject != "Test Subject" {
		t.Errorf("Expected subject 'Test Subject', got '%s'", receivedBody.Subject)
	}
	if receivedBody.HTMLContent != "<h1>Hello</h1>" {
		t.Errorf("Expected HTML content '<h1>Hello</h1>', got '%s'", receivedBody.HTMLContent)
	}
}

func TestSendEmail_APIError(t *testing.T) {
	svc := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"code":"unauthorized","message":"Invalid API key"}`))
	})

	err := svc.SendEmail(context.Background(), "recipient@test.com", "Test", "<p>Hi</p>")
	if err == nil {
		t.Fatal("Expected error for unauthorized response")
	}
	if !strings.Contains(err.Error(), "HTTP 401") {
		t.Errorf("Expected error to contain 'HTTP 401', got: %v", err)
	}
	if !strings.Contains(err.Error(), "Invalid API key") {
		t.Errorf("Expected error to contain API response body, got: %v", err)
	}
}

func TestSendEmail_ServerError(t *testing.T) {
	svc := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"Internal error"}`))
	})

	err := svc.SendEmail(context.Background(), "recipient@test.com", "Test", "<p>Hi</p>")
	if err == nil {
		t.Fatal("Expected error for 500 response")
	}
	if !strings.Contains(err.Error(), "HTTP 500") {
		t.Errorf("Expected error to contain 'HTTP 500', got: %v", err)
	}
}

func TestSendEmail_NetworkError(t *testing.T) {
	logger := zaptest.NewLogger(t)
	svc := NewBrevoService("key", "sender@test.com", "Test", logger)
	svc.apiBaseURL = "http://localhost:1" // invalid port

	err := svc.SendEmail(context.Background(), "recipient@test.com", "Test", "<p>Hi</p>")
	if err == nil {
		t.Fatal("Expected error for network failure")
	}
	if !strings.Contains(err.Error(), "failed to send email") {
		t.Errorf("Expected 'failed to send email' in error, got: %v", err)
	}
}

func TestSendEmail_ContextCanceled(t *testing.T) {
	svc := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err := svc.SendEmail(ctx, "recipient@test.com", "Test", "<p>Hi</p>")
	if err == nil {
		t.Fatal("Expected error for canceled context")
	}
}

func TestSendUserInvite(t *testing.T) {
	var receivedBody brevoEmailRequest

	svc := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &receivedBody)
		w.WriteHeader(http.StatusCreated)
	})

	err := svc.SendUserInvite(context.Background(), "newuser@test.com", "Alice", "abc123", "https://app.taskai.cc")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if receivedBody.To[0].Email != "newuser@test.com" {
		t.Errorf("Expected to 'newuser@test.com', got '%s'", receivedBody.To[0].Email)
	}
	if !strings.Contains(receivedBody.Subject, "Alice") {
		t.Errorf("Expected subject to contain inviter name 'Alice', got '%s'", receivedBody.Subject)
	}
	if !strings.Contains(receivedBody.HTMLContent, "https://app.taskai.cc/signup?code=abc123") {
		t.Error("Expected HTML to contain signup URL with invite code")
	}
	if !strings.Contains(receivedBody.HTMLContent, "Alice") {
		t.Error("Expected HTML to contain inviter name")
	}
}

func TestSendProjectInvitation(t *testing.T) {
	var receivedBody brevoEmailRequest

	svc := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &receivedBody)
		w.WriteHeader(http.StatusCreated)
	})

	err := svc.SendProjectInvitation(context.Background(), "member@test.com", "Bob", "My Project", "token123", "https://app.taskai.cc")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if receivedBody.To[0].Email != "member@test.com" {
		t.Errorf("Expected to 'member@test.com', got '%s'", receivedBody.To[0].Email)
	}
	if !strings.Contains(receivedBody.Subject, "My Project") {
		t.Errorf("Expected subject to contain project name, got '%s'", receivedBody.Subject)
	}
	if !strings.Contains(receivedBody.HTMLContent, "https://app.taskai.cc/accept-invite?token=token123") {
		t.Error("Expected HTML to contain accept-invite URL with token")
	}
	if !strings.Contains(receivedBody.HTMLContent, "Bob") {
		t.Error("Expected HTML to contain inviter name")
	}
	if !strings.Contains(receivedBody.HTMLContent, "My Project") {
		t.Error("Expected HTML to contain project name")
	}
	if !strings.Contains(receivedBody.HTMLContent, "Accept Invitation") {
		t.Error("Expected HTML to contain 'Accept Invitation' CTA label")
	}
}

func TestSendProjectInvitationNewUser(t *testing.T) {
	var receivedBody brevoEmailRequest

	svc := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &receivedBody)
		w.WriteHeader(http.StatusCreated)
	})

	err := svc.SendProjectInvitationNewUser(context.Background(), "new@test.com", "Carol", "Sprint Board", "tokenXYZ", "https://app.taskai.cc")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if receivedBody.To[0].Email != "new@test.com" {
		t.Errorf("Expected to 'new@test.com', got '%s'", receivedBody.To[0].Email)
	}
	if !strings.Contains(receivedBody.Subject, "Carol") {
		t.Errorf("Expected subject to contain inviter name, got '%s'", receivedBody.Subject)
	}
	if !strings.Contains(receivedBody.Subject, "Sprint Board") {
		t.Errorf("Expected subject to contain project name, got '%s'", receivedBody.Subject)
	}
	if !strings.Contains(receivedBody.HTMLContent, "https://app.taskai.cc/accept-invite?token=tokenXYZ") {
		t.Error("Expected HTML to contain accept-invite URL with token")
	}
	if !strings.Contains(receivedBody.HTMLContent, "Accept Invitation") {
		t.Error("Expected HTML to contain 'Accept Invitation' CTA label")
	}
}

func TestBuildEmailTemplate(t *testing.T) {
	html := buildEmailTemplate("Test Heading", "Test body text", "https://example.com", "Click Me", "Footer note")

	checks := []struct {
		name     string
		contains string
	}{
		{"heading", "Test Heading"},
		{"body text", "Test body text"},
		{"CTA URL", "https://example.com"},
		{"CTA label", "Click Me"},
		{"footer note", "Footer note"},
		{"TaskAI branding", "TaskAI"},
		{"DOCTYPE", "<!DOCTYPE html>"},
		{"responsive meta", "viewport"},
	}

	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			if !strings.Contains(html, c.contains) {
				t.Errorf("Expected template to contain '%s'", c.contains)
			}
		})
	}
}

func TestSendEmail_StatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    bool
	}{
		{"200 OK", http.StatusOK, false},
		{"201 Created", http.StatusCreated, false},
		{"202 Accepted", http.StatusAccepted, false},
		{"400 Bad Request", http.StatusBadRequest, true},
		{"403 Forbidden", http.StatusForbidden, true},
		{"429 Too Many Requests", http.StatusTooManyRequests, true},
		{"500 Internal Server Error", http.StatusInternalServerError, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(`{"message":"test"}`))
			})

			err := svc.SendEmail(context.Background(), "test@test.com", "Test", "<p>Hi</p>")
			if (err != nil) != tt.wantErr {
				t.Errorf("SendEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
