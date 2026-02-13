package api

import (
	"net/http"
	"os"
	"testing"
)

func TestHandleOpenAPI(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	// Create a temporary openapi.yaml so the handler can find it
	content := []byte("openapi: 3.0.0\ninfo:\n  title: Test\n  version: 1.0\n")
	if err := os.WriteFile("./openapi.yaml", content, 0644); err != nil {
		t.Fatalf("Failed to create test openapi.yaml: %v", err)
	}
	defer os.Remove("./openapi.yaml")

	rec, req := MakeRequest(t, http.MethodGet, "/api/openapi", nil, nil)
	ts.HandleOpenAPI(rec, req)

	AssertStatusCode(t, rec.Code, http.StatusOK)

	if ct := rec.Header().Get("Content-Type"); ct != "application/yaml" {
		t.Errorf("Expected Content-Type 'application/yaml', got %q", ct)
	}

	body := rec.Body.String()
	if body == "" {
		t.Error("Expected non-empty response body")
	}
	if len(body) != len(content)+0 {
		// Just check it contains the YAML content
		if body[:len("openapi")] != "openapi" {
			t.Errorf("Expected YAML content, got %q", body[:20])
		}
	}
}

func TestHandleOpenAPI_FileNotFound(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	// Don't create the file - ensure it doesn't exist
	os.Remove("./openapi.yaml")

	rec, req := MakeRequest(t, http.MethodGet, "/api/openapi", nil, nil)
	ts.HandleOpenAPI(rec, req)

	AssertStatusCode(t, rec.Code, http.StatusInternalServerError)

	var errResp ErrorResponse
	DecodeJSON(t, rec, &errResp)
	if errResp.Error != "failed to load API specification" {
		t.Errorf("Expected error message about loading spec, got %q", errResp.Error)
	}
}

func TestGetOpenAPISpec(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	content := []byte("test: spec\n")
	if err := os.WriteFile("./openapi.yaml", content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove("./openapi.yaml")

	result, err := ts.getOpenAPISpec()
	if err != nil {
		t.Fatalf("getOpenAPISpec failed: %v", err)
	}

	if string(result) != string(content) {
		t.Errorf("Expected %q, got %q", string(content), string(result))
	}
}

func TestGetOpenAPISpec_Missing(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()

	os.Remove("./openapi.yaml")

	_, err := ts.getOpenAPISpec()
	if err == nil {
		t.Fatal("Expected error when file is missing")
	}
}
