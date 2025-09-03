package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gant123/jobTracker/internal/config"
	"github.com/gant123/jobTracker/internal/crypto"
	"github.com/gant123/jobTracker/internal/database"
	"github.com/gant123/jobTracker/internal/handlers"
	"github.com/gant123/jobTracker/internal/middleware"
	"github.com/gant123/jobTracker/internal/models"
	"github.com/gant123/jobTracker/internal/repository"
	"github.com/gant123/jobTracker/internal/services"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	gmail "google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
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
	jobQueueRepo := repository.NewJobQueueRepository(db)
	gmailSyncRepo := repository.NewGmailSyncRepository(db)
	// Initialize services
	authService := services.NewAuthService(userRepo, cfg)
	jobService := services.NewJobService(jobRepo)
	// google oauth handler
	googleOAuth := services.NewGoogleOAuth()
	googleHandler := handlers.NewGoogleHandler(googleOAuth, logger, tokenRepo, jobRepo, jobQueueRepo, gmailSyncRepo)
	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, logger)
	jobHandler := handlers.NewJobHandler(jobService, logger)
	healthHandler := handlers.NewHealthHandler(db)
	worker := NewWorker(db, logger, tokenRepo, jobRepo, jobQueueRepo, gmailSyncRepo)
	go worker.Start()
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
	protected.HandleFunc("/google/sync-status", googleHandler.SyncStatus).Methods("GET")
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

type Worker struct {
	db           *sql.DB
	logger       *logrus.Logger
	tokenRepo    repository.TokenRepository
	jobRepo      *repository.JobRepository
	jobQueueRepo *repository.JobQueueRepository
	syncRepo     *repository.GmailSyncRepository
}

func NewWorker(
	db *sql.DB,
	logger *logrus.Logger,
	tokenRepo repository.TokenRepository,
	jobRepo *repository.JobRepository,
	jobQueueRepo *repository.JobQueueRepository,
	syncRepo *repository.GmailSyncRepository,
) *Worker {
	return &Worker{
		db:           db,
		logger:       logger,
		tokenRepo:    tokenRepo,
		jobRepo:      jobRepo,
		jobQueueRepo: jobQueueRepo,
		syncRepo:     syncRepo,
	}
}

func (w *Worker) Start() {
	w.logger.Info("Starting background worker")

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.processNextJob()
		}
	}
}

func (w *Worker) processNextJob() {
	job, err := w.jobQueueRepo.GetNextJob()
	if err == sql.ErrNoRows {
		return // No jobs
	}
	if err != nil {
		w.logger.WithError(err).Error("Failed to get next job")
		return
	}

	w.logger.WithFields(logrus.Fields{
		"job_id":   job.ID,
		"type":     job.Type,
		"user_id":  job.UserID,
		"attempts": job.Attempts,
	}).Info("Processing job")

	switch job.Type {
	case "gmail_initial_sync":
		err = w.processInitialSync(job.UserID)
	default:
		w.logger.Warn("Unknown job type", "type", job.Type)
		return
	}

	if err != nil {
		w.logger.WithError(err).Error("Job failed")
		w.jobQueueRepo.MarkJobFailed(job.ID, err.Error())
	} else {
		w.jobQueueRepo.MarkJobComplete(job.ID)
	}
}

func (w *Worker) processInitialSync(userID int) error {
	w.logger.Info("Starting initial Gmail sync", "user_id", userID)

	// Get Gmail token
	ctx := context.Background()
	tok, err := w.tokenRepo.Get(ctx, userID, "gmail")
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}

	// Create Gmail service
	oauth := services.NewGoogleOAuth()
	client := oauth.Client(ctx, tok)
	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("failed to create gmail service: %w", err)
	}

	// Mark sync started
	w.syncRepo.UpdateSyncStarted(userID)

	// Get existing job IDs to avoid duplicates
	existingIDs, err := w.jobRepo.GetAllGmailMessageIDsByUserID(userID)
	if err != nil {
		return fmt.Errorf("failed to get existing IDs: %w", err)
	}

	// Scan last year of emails
	scanner := services.NewGmailScanner()
	since := time.Now().AddDate(-1, 0, 0)
	until := time.Now()

	totalImported := 0
	pageToken := ""

	for {
		// Use your existing scanner
		result, err := scanner.ScanPage(ctx, srv, since, until, 100, pageToken, "all", existingIDs)
		if err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}

		// Import each job
		for _, event := range result.Events {
			jobData := &models.CreateJobRequest{
				Company:        event.Company,
				Position:       event.Title,
				Status:         event.Status,
				AppliedDate:    &event.AppliedDate,
				Notes:          fmt.Sprintf("[Gmail Import] %s", event.Subject),
				GmailMessageID: event.MessageID,
			}

			if jobData.Company == "" {
				jobData.Company = "Unknown Company"
			}
			if jobData.Position == "" {
				jobData.Position = "Unknown Position"
			}

			err := w.jobRepo.Create(&models.Job{
				UserID:         userID,
				Company:        jobData.Company,
				Position:       jobData.Position,
				Status:         jobData.Status,
				AppliedDate:    jobData.AppliedDate,
				Notes:          jobData.Notes,
				GmailMessageID: jobData.GmailMessageID,
			})

			if err == nil {
				totalImported++
			}
		}

		if result.NextPageToken == "" {
			break
		}
		pageToken = result.NextPageToken

		// Be nice to Gmail API
		time.Sleep(100 * time.Millisecond)
	}

	// Mark sync completed
	w.syncRepo.UpdateSyncCompleted(userID, totalImported)

	w.logger.Info("Initial sync completed", "user_id", userID, "total_imported", totalImported)

	return nil
}

