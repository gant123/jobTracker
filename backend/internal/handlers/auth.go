package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gant123/jobTracker/internal/models"
	"github.com/gant123/jobTracker/internal/services"
	"github.com/sirupsen/logrus"
)

type AuthHandler struct {
	authService *services.AuthService
	logger      *logrus.Logger
}

func NewAuthHandler(authService *services.AuthService, logger *logrus.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := h.authService.Register(&req)
	if err != nil {
		h.logger.Error("Registration failed:", err)
		h.respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.respondJSON(w, response, http.StatusCreated)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := h.authService.Login(&req)
	if err != nil {
		h.logger.Error("Login failed:", err)
		h.respondError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	h.respondJSON(w, response, http.StatusOK)
}

func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		h.respondError(w, "User not found", http.StatusNotFound)
		return
	}

	h.respondJSON(w, user, http.StatusOK)
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// Implement token refresh logic
	h.respondJSON(w, map[string]string{"message": "Token refreshed"}, http.StatusOK)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// In a JWT-based system, logout is typically handled client-side
	h.respondJSON(w, map[string]string{"message": "Logged out successfully"}, http.StatusOK)
}

func (h *AuthHandler) respondJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *AuthHandler) respondError(w http.ResponseWriter, message string, status int) {
	h.respondJSON(w, map[string]string{"error": message}, status)
}
