package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
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

		// Auth routes (public)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/signup", server.HandleSignup)
			r.Post("/login", server.HandleLogin)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(server.JWTAuth)

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

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Server listening on %s", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}
