package api

import (
	"go.uber.org/zap"

	"sprintspark/internal/config"
	"sprintspark/internal/db"
)

// Server holds the application dependencies
type Server struct {
	db     *db.DB
	config *config.Config
	logger *zap.Logger
}

// NewServer creates a new API server
func NewServer(database *db.DB, cfg *config.Config, logger *zap.Logger) *Server {
	return &Server{
		db:     database,
		config: cfg,
		logger: logger,
	}
}
