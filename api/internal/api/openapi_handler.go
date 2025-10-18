package api

import (
	"log"
	"net/http"
	"os"
)

// HandleOpenAPI serves the OpenAPI specification
func (s *Server) HandleOpenAPI(w http.ResponseWriter, r *http.Request) {
	// Read OpenAPI spec file
	specPath := "./openapi.yaml"
	content, err := os.ReadFile(specPath)
	if err != nil {
		log.Printf("Failed to read OpenAPI spec: %v", err)
		respondError(w, http.StatusInternalServerError, "failed to load API specification", "internal_error")
		return
	}

	// Serve as YAML
	w.Header().Set("Content-Type", "application/yaml")
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}
