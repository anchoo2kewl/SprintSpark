package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"sprintspark/internal/api"
	"sprintspark/internal/config"
	"sprintspark/internal/db"
)

func main() {
	// Load configuration
	cfg := config.Load()

	log.Printf("Starting SprintSpark API in %s mode", cfg.Env)

	// Initialize database with auto-migrations
	dbCfg := db.Config{
		DBPath:         cfg.DBPath,
		MigrationsPath: cfg.MigrationsPath,
	}

	database, err := db.New(dbCfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Create server
	server := api.NewServer(database, cfg)

	// Setup router
	r := chi.NewRouter()

	// Middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(api.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// Request size limit (1MB)
	r.Use(middleware.SetHeader("Content-Type", "application/json"))
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB
			next.ServeHTTP(w, r)
		})
	})

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSAllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Public routes
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message":"SprintSpark API","version":"0.1.0"}`)
	})

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := database.HealthCheck(ctx); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, `{"status":"error","message":"database unavailable"}`)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","database":"connected"}`)
	})

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Legacy health endpoint
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"status":"ok"}`)
		})

		// OpenAPI specification
		r.Get("/openapi", server.HandleOpenAPI)

		// Auth routes (public) with rate limiting
		r.Route("/auth", func(r chi.Router) {
			// Apply stricter rate limiting to auth endpoints (20 req/min)
			r.Use(api.RateLimitMiddleware(20))
			r.Post("/signup", server.HandleSignup)
			r.Post("/login", server.HandleLogin)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(server.JWTAuth)
			// Apply general rate limiting (100 req/min)
			r.Use(api.RateLimitMiddleware(cfg.RateLimitRequests))

			r.Get("/me", server.HandleMe)

			// Project routes
			r.Get("/projects", server.HandleListProjects)
			r.Post("/projects", server.HandleCreateProject)
			r.Get("/projects/{id}", server.HandleGetProject)
			r.Patch("/projects/{id}", server.HandleUpdateProject)
			r.Delete("/projects/{id}", server.HandleDeleteProject)

			// Task routes
			r.Get("/projects/{projectId}/tasks", server.HandleListTasks)
			r.Post("/projects/{projectId}/tasks", server.HandleCreateTask)
			r.Patch("/tasks/{id}", server.HandleUpdateTask)
			r.Delete("/tasks/{id}", server.HandleDeleteTask)
		})
	})

	// Create HTTP server
	addr := fmt.Sprintf(":%s", cfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Channel to listen for server errors
	serverErrors := make(chan error, 1)

	// Start server in goroutine
	go func() {
		log.Printf("Server listening on %s", addr)
		serverErrors <- srv.ListenAndServe()
	}()

	// Channel to listen for interrupt signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until we receive shutdown signal or server error
	select {
	case err := <-serverErrors:
		log.Fatalf("Server error: %v", err)
	case sig := <-shutdown:
		log.Printf("Received %v signal, starting graceful shutdown", sig)

		// Give outstanding requests 30 seconds to complete
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Gracefully shutdown server
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Graceful shutdown failed: %v", err)
			if err := srv.Close(); err != nil {
				log.Fatalf("Failed to close server: %v", err)
			}
		}

		log.Println("Server stopped gracefully")
	}
}
