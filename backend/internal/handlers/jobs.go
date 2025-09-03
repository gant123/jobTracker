package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gant123/jobTracker/internal/models"
	"github.com/gant123/jobTracker/internal/services"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type JobHandler struct {
	jobService *services.JobService
	logger     *logrus.Logger
}

func NewJobHandler(jobService *services.JobService, logger *logrus.Logger) *JobHandler {
	return &JobHandler{
		jobService: jobService,
		logger:     logger,
	}
}

func (h *JobHandler) CreateJob(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	var req models.CreateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	job, err := h.jobService.CreateJob(userID, &req)
	if err != nil {
		h.logger.Error("Failed to create job:", err)
		h.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.respondJSON(w, job, http.StatusCreated)
}

func (h *JobHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
	vars := mux.Vars(r)

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.respondError(w, "Invalid job ID", http.StatusBadRequest)
		return
	}

	job, err := h.jobService.GetJob(id, userID)
	if err != nil {
		h.respondError(w, "Job not found", http.StatusNotFound)
		return
	}

	h.respondJSON(w, job, http.StatusOK)
}

func (h *JobHandler) GetJobs(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	filter := &models.JobFilter{
		Status:   r.URL.Query().Get("status"),
		Company:  r.URL.Query().Get("company"),
		Location: r.URL.Query().Get("location"),
		Search:   r.URL.Query().Get("search"),
	}

	jobs, err := h.jobService.GetJobs(userID, filter)
	if err != nil {
		h.logger.Error("Failed to get jobs:", err)
		h.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get stats
	stats, _ := h.jobService.GetStats(userID)

	response := map[string]interface{}{
		"jobs":  jobs,
		"stats": stats,
	}

	h.respondJSON(w, response, http.StatusOK)
}

func (h *JobHandler) UpdateJob(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
	vars := mux.Vars(r)

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.respondError(w, "Invalid job ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	job, err := h.jobService.UpdateJob(id, userID, &req)
	if err != nil {
		h.logger.Error("Failed to update job:", err)
		h.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.respondJSON(w, job, http.StatusOK)
}

func (h *JobHandler) DeleteJob(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)
	vars := mux.Vars(r)

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.respondError(w, "Invalid job ID", http.StatusBadRequest)
		return
	}

	if err := h.jobService.DeleteJob(id, userID); err != nil {
		h.logger.Error("Failed to delete job:", err)
		h.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.respondJSON(w, map[string]string{"message": "Job deleted successfully"}, http.StatusOK)
}

func (h *JobHandler) respondJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *JobHandler) respondError(w http.ResponseWriter, message string, status int) {
	h.respondJSON(w, map[string]string{"error": message}, status)
}
