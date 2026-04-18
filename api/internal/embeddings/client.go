package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Client calls an Ollama-compatible embedding API.
// A nil *Client is safe to use — all methods return nil embeddings.
type Client struct {
	baseURL    string
	model      string
	httpClient *http.Client
	logger     *zap.Logger
}

// NewClient creates a new embedding client. Returns nil if baseURL is empty
// (graceful degradation for environments without Ollama).
func NewClient(baseURL, model string, logger *zap.Logger) *Client {
	if baseURL == "" {
		logger.Info("Embeddings disabled (OLLAMA_URL not set)")
		return nil
	}
	if model == "" {
		model = "all-minilm:l6-v2"
	}
	return &Client{
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// Model returns the model name used for embeddings.
func (c *Client) Model() string {
	if c == nil {
		return ""
	}
	return c.model
}

// embedRequest is the Ollama /api/embed request body.
type embedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

// embedResponse is the Ollama /api/embed response body.
type embedResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
}

// Embed generates an embedding for a single text.
// Returns nil if the client is nil (embeddings disabled).
func (c *Client) Embed(ctx context.Context, text string) ([]float32, error) {
	if c == nil {
		return nil, nil
	}
	vectors, err := c.EmbedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(vectors) == 0 {
		return nil, fmt.Errorf("empty embedding response")
	}
	return vectors[0], nil
}

// EmbedBatch generates embeddings for multiple texts in a single call.
// Returns nil if the client is nil (embeddings disabled).
func (c *Client) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if c == nil {
		return nil, nil
	}
	if len(texts) == 0 {
		return nil, nil
	}

	reqBody := embedRequest{
		Model: c.model,
		Input: texts,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal embed request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/embed", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create embed request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("embed request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("embed request returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result embedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode embed response: %w", err)
	}

	if len(result.Embeddings) != len(texts) {
		return nil, fmt.Errorf("expected %d embeddings, got %d", len(texts), len(result.Embeddings))
	}

	return result.Embeddings, nil
}

// Healthy checks if Ollama is reachable. Returns false if client is nil.
func (c *Client) Healthy(ctx context.Context) bool {
	if c == nil {
		return false
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/tags", nil)
	if err != nil {
		return false
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
