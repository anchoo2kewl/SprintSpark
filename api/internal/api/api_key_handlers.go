package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

// CreateAPIKeyRequest represents the request to create an API key
type CreateAPIKeyRequest struct {
	Name      string `json:"name"`
	ExpiresIn *int   `json:"expires_in,omitempty"` // Days until expiration, null for no expiration
}

// CreateAPIKeyResponse represents the response when creating an API key
type CreateAPIKeyResponse struct {
	ID        int64      `json:"id"`
	Name      string     `json:"name"`
	Key       string     `json:"key"`
	KeyPrefix string     `json:"key_prefix"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// APIKeyResponse represents an API key in responses (without the full key)
type APIKeyResponse struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	KeyPrefix  string     `json:"key_prefix"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

// HandleCreateAPIKey creates a new API key for the authenticated user
func (s *Server) HandleCreateAPIKey(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	// Parse request body
	var req CreateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", "invalid_request")
		return
	}

	// Validate name
	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "name is required", "validation_error")
		return
	}

	if len(req.Name) > 100 {
		respondError(w, http.StatusBadRequest, "name must be 100 characters or less", "validation_error")
		return
	}

	// Calculate expiration date
	var expiresAt *time.Time
	if req.ExpiresIn != nil {
		if *req.ExpiresIn <= 0 {
			respondError(w, http.StatusBadRequest, "expires_in must be positive", "validation_error")
			return
		}
		if *req.ExpiresIn > 365 {
			respondError(w, http.StatusBadRequest, "expires_in cannot exceed 365 days", "validation_error")
			return
		}
		exp := time.Now().AddDate(0, 0, *req.ExpiresIn)
		expiresAt = &exp
	}

	// Create API key
	apiKey, err := s.db.CreateAPIKey(r.Context(), userID, req.Name, expiresAt)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create API key", "internal_error")
		return
	}

	// Return response
	response := CreateAPIKeyResponse{
		ID:        apiKey.ID,
		Name:      apiKey.Name,
		Key:       apiKey.Key,
		KeyPrefix: apiKey.KeyPrefix,
		CreatedAt: apiKey.CreatedAt,
		ExpiresAt: apiKey.ExpiresAt,
	}

	respondJSON(w, http.StatusCreated, response)
}

// HandleListAPIKeys lists all API keys for the authenticated user
func (s *Server) HandleListAPIKeys(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	// Get API keys
	keys, err := s.db.GetAPIKeysByUserID(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to retrieve API keys", "internal_error")
		return
	}

	// Convert to response format
	response := make([]APIKeyResponse, len(keys))
	for i, key := range keys {
		response[i] = APIKeyResponse{
			ID:         key.ID,
			Name:       key.Name,
			KeyPrefix:  key.KeyPrefix,
			LastUsedAt: key.LastUsedAt,
			CreatedAt:  key.CreatedAt,
			ExpiresAt:  key.ExpiresAt,
		}
	}

	respondJSON(w, http.StatusOK, response)
}

// HandleDeleteAPIKey deletes an API key
func (s *Server) HandleDeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	// Get key ID from URL
	keyIDStr := chi.URLParam(r, "id")
	keyID, err := strconv.ParseInt(keyIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid API key ID", "invalid_request")
		return
	}

	// Delete API key
	if err := s.db.DeleteAPIKey(r.Context(), keyID, userID); err != nil {
		respondError(w, http.StatusNotFound, "API key not found", "not_found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
