package api

import (
	"sprintspark/internal/config"
	"sprintspark/internal/db"
)

// Server holds the application dependencies
type Server struct {
	db     *db.DB
	config *config.Config
}

// NewServer creates a new API server
func NewServer(database *db.DB, cfg *config.Config) *Server {
	return &Server{
		db:     database,
		config: cfg,
	}
}
