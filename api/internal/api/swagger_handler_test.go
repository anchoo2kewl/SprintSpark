package api

import (
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestHandleSwaggerUI(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	rec, req := MakeRequest(t, http.MethodGet, "/api/docs", nil, nil)
	ts.HandleSwaggerUI(rec, req)

	AssertStatusCode(t, rec.Code, http.StatusOK)

	if ct := rec.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("Expected Content-Type 'text/html; charset=utf-8', got %q", ct)
	}

	body := rec.Body.String()

	// Check essential elements
	checks := []string{
		"<!DOCTYPE html>",
		"<title>TaskAI API Documentation</title>",
		"swagger-ui",
		"SwaggerUIBundle",
		"/api/openapi",
	}

	for _, check := range checks {
		if !strings.Contains(body, check) {
			t.Errorf("Expected body to contain %q", check)
		}
	}
}

func TestHandleOpenAPIYAML(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	content := []byte("openapi: 3.0.0\ninfo:\n  title: Test\n")
	if err := os.WriteFile("./openapi.yaml", content, 0644); err != nil {
		t.Fatalf("Failed to create test openapi.yaml: %v", err)
	}
	defer os.Remove("./openapi.yaml")

	tests := []struct {
		name        string
		query       string
		headers     map[string]string
		wantType    string
		wantDisp    bool
	}{
		{
			name:     "serves as YAML by default",
			wantType: "application/yaml",
		},
		{
			name:     "serves as download with query param",
			query:    "?download=true",
			wantType: "application/octet-stream",
			wantDisp: true,
		},
		{
			name:     "serves as download with Accept header",
			headers:  map[string]string{"Accept": "application/octet-stream"},
			wantType: "application/octet-stream",
			wantDisp: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := "/api/openapi.yaml"
			if tt.query != "" {
				path += tt.query
			}

			rec, req := MakeRequest(t, http.MethodGet, path, nil, tt.headers)

			userID := ts.CreateTestUser(t, tt.name+"@example.com", "password123")
			token := ts.GenerateTestToken(t, userID, tt.name+"@example.com")
			req.Header.Set("Authorization", "Bearer "+token)

			ts.HandleOpenAPIYAML(rec, req)

			AssertStatusCode(t, rec.Code, http.StatusOK)

			if ct := rec.Header().Get("Content-Type"); ct != tt.wantType {
				t.Errorf("Expected Content-Type %q, got %q", tt.wantType, ct)
			}

			if tt.wantDisp {
				disp := rec.Header().Get("Content-Disposition")
				if !strings.Contains(disp, "attachment") {
					t.Errorf("Expected Content-Disposition with 'attachment', got %q", disp)
				}
				if !strings.Contains(disp, "taskai-openapi.yaml") {
					t.Errorf("Expected filename 'taskai-openapi.yaml' in Content-Disposition, got %q", disp)
				}
			}
		})
	}
}

func TestHandleOpenAPIYAML_FileNotFound(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	os.Remove("./openapi.yaml")

	rec, req := MakeRequest(t, http.MethodGet, "/api/openapi.yaml", nil, nil)
	ts.HandleOpenAPIYAML(rec, req)

	AssertStatusCode(t, rec.Code, http.StatusInternalServerError)
}

func TestHandleOpenAPIJSON(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	rec, req := MakeRequest(t, http.MethodGet, "/api/openapi.json", nil, nil)
	ts.HandleOpenAPIJSON(rec, req)

	AssertStatusCode(t, rec.Code, http.StatusNotImplemented)

	var errResp ErrorResponse
	DecodeJSON(t, rec, &errResp)
	if errResp.Code != "not_implemented" {
		t.Errorf("Expected error code 'not_implemented', got %q", errResp.Code)
	}
	if !strings.Contains(errResp.Error, "JSON format not yet supported") {
		t.Errorf("Expected message about JSON not supported, got %q", errResp.Error)
	}
}
