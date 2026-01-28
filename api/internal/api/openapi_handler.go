package api

import (
	"fmt"
	"net/http"
	"os"
)

// getOpenAPISpec reads and returns the OpenAPI specification file
func (s *Server) getOpenAPISpec() ([]byte, error) {
	specPath := "./openapi.yaml"
	content, err := os.ReadFile(specPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read OpenAPI spec: %w", err)
	}
	return content, nil
}

// HandleOpenAPI serves the OpenAPI specification (public endpoint)
func (s *Server) HandleOpenAPI(w http.ResponseWriter, r *http.Request) {
	content, err := s.getOpenAPISpec()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load API specification", "internal_error")
		return
	}

	// Serve as YAML
	w.Header().Set("Content-Type", "application/yaml")
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}
