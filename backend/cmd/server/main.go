package main

import (
	"github.com/gant123/jobTracker/internal/config"
	"github.com/gant123/jobTracker/internal/crypto"
	"github.com/gant123/jobTracker/internal/database"
	"github.com/gant123/jobTracker/internal/handlers"
	"github.com/gant123/jobTracker/internal/middleware"
	"github.com/gant123/jobTracker/internal/repository"
	"github.com/gant123/jobTracker/internal/services"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.Initialize(cfg)
	if err != nil {
		logger.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()
	box, err := crypto.NewSecretBox(cfg.EncryptionKey)
	if err != nil {
		logger.Fatal("Invalid ENCRYPTION_KEY (must be 32 bytes or 64-char hex): ", err)
	}
	tokenRepo := repository.NewPostgresTokenRepository(db, box)
	// Run migrations
	if err := database.Migrate(db); err != nil {
		logger.Fatal("Failed to run migrations:", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	jobRepo := repository.NewJobRepository(db)

	// Initialize services
	authService := services.NewAuthService(userRepo, cfg)
	jobService := services.NewJobService(jobRepo)
	// google oauth handler
	googleOAuth := services.NewGoogleOAuth()
	googleHandler := handlers.NewGoogleHandler(googleOAuth, logger, tokenRepo, jobRepo)
	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, logger)
	jobHandler := handlers.NewJobHandler(jobService, logger)
	healthHandler := handlers.NewHealthHandler(db)

	// Setup routes
	router := setupRoutes(authHandler, jobHandler, healthHandler, googleHandler, cfg, logger)

	// Start server
	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	logger.Infof("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		logger.Fatal("Server failed to start:", err)
	}
}

func setupRoutes(
	authHandler *handlers.AuthHandler,
	jobHandler *handlers.JobHandler,
	healthHandler *handlers.HealthHandler,
	googleHandler *handlers.GoogleHandler,
	cfg *config.Config,
	logger *logrus.Logger,
) *mux.Router {
	r := mux.NewRouter()

	// Apply CORS middleware
	r.Use(middleware.CORS(cfg))
	r.Use(middleware.Logger(logger))

	// Health check
	r.HandleFunc("/health", healthHandler.Health).Methods("GET")
	// Catch-all OPTIONS so preflights always match a route
	r.PathPrefix("/").Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	// API routes
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/google/callback", googleHandler.Callback).Methods(http.MethodGet)
	api.HandleFunc("/google/callback", googleHandler.Callback).Methods(http.MethodGet)
	// Public routes
	api.HandleFunc("/auth/register", authHandler.Register).Methods("POST")
	api.HandleFunc("/auth/login", authHandler.Login).Methods("POST")
	api.HandleFunc("/auth/refresh", authHandler.RefreshToken).Methods("POST")

	// Protected routes
	protected := api.PathPrefix("").Subrouter()
	protected.Use(middleware.Auth(cfg))
	// protected
	protected.HandleFunc("/google/auth-url", googleHandler.BeginAuthURL).Methods(http.MethodGet)
	protected.HandleFunc("/google/status", googleHandler.Status).Methods(http.MethodGet)
	protected.HandleFunc("/google/disconnect", googleHandler.Disconnect).Methods(http.MethodPost)
	protected.HandleFunc("/google/scan", googleHandler.Scan).Methods(http.MethodGet)
	// protected (requires logged-in user)
	protected.HandleFunc("/google/scan", googleHandler.Scan).Methods("GET")
	// Jobs routes
	protected.HandleFunc("/jobs", jobHandler.GetJobs).Methods("GET")
	protected.HandleFunc("/jobs", jobHandler.CreateJob).Methods("POST")
	protected.HandleFunc("/jobs/{id}", jobHandler.GetJob).Methods("GET")
	protected.HandleFunc("/jobs/{id}", jobHandler.UpdateJob).Methods("PUT")
	protected.HandleFunc("/jobs/{id}", jobHandler.DeleteJob).Methods("DELETE")

	// User routes
	protected.HandleFunc("/auth/me", authHandler.GetProfile).Methods("GET")
	protected.HandleFunc("/auth/logout", authHandler.Logout).Methods("POST")

	return r
}

