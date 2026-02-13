package api

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"sprintspark/internal/config"
	"sprintspark/internal/db"
	"sprintspark/internal/email"
)

// Server holds the application dependencies
type Server struct {
	db           *db.DB
	config       *config.Config
	logger       *zap.Logger
	emailService *email.BrevoService
	emailMu      sync.RWMutex
}

// NewServer creates a new API server
func NewServer(database *db.DB, cfg *config.Config, logger *zap.Logger) *Server {
	return &Server{
		db:     database,
		config: cfg,
		logger: logger,
	}
}

// getAppURL returns the application URL for use in email links
func (s *Server) getAppURL() string {
	if len(s.config.CORSAllowedOrigins) > 0 {
		return s.config.CORSAllowedOrigins[0]
	}
	return "http://localhost:5173"
}

// invalidateEmailService clears the cached email service so it's reloaded on next use
func (s *Server) invalidateEmailService() {
	s.emailMu.Lock()
	defer s.emailMu.Unlock()
	s.emailService = nil
}

// GetEmailService returns the email service, loading from DB if needed. Returns nil if not configured.
func (s *Server) GetEmailService() *email.BrevoService {
	s.emailMu.RLock()
	if s.emailService != nil {
		svc := s.emailService
		s.emailMu.RUnlock()
		return svc
	}
	s.emailMu.RUnlock()

	// Load from DB
	s.emailMu.Lock()
	defer s.emailMu.Unlock()

	// Double-check after acquiring write lock
	if s.emailService != nil {
		return s.emailService
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var apiKey, senderEmail, senderName, status string
	err := s.db.QueryRowContext(ctx,
		`SELECT api_key, sender_email, sender_name, status FROM email_provider WHERE id = 1`,
	).Scan(&apiKey, &senderEmail, &senderName, &status)
	if err != nil {
		return nil
	}

	if status == "suspended" {
		return nil
	}

	s.emailService = email.NewBrevoService(apiKey, senderEmail, senderName, s.logger)
	return s.emailService
}
