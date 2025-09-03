package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	//
	// "github.com/gant123/jobTracker/internal/jobs"
	"github.com/gant123/jobTracker/internal/repository"
	"github.com/gant123/jobTracker/internal/services"
	"github.com/sirupsen/logrus"
	gmail "google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

const providerGmail = "gmail"
const cookieUID = "g_uid"
const cookieState = "g_state"

type GoogleHandler struct {
	OAuth     *services.GoogleOAuth
	Logger    *logrus.Logger
	TokenRepo repository.TokenRepository
	Scanner   *services.GmailScanner
	JobRepo   *repository.JobRepository
	JobQueue  *repository.JobQueueRepository
	SyncRepo  *repository.GmailSyncRepository
}

func NewGoogleHandler(o *services.GoogleOAuth, logger *logrus.Logger, tr repository.TokenRepository, jr *repository.JobRepository, jq *repository.JobQueueRepository, sr *repository.GmailSyncRepository) *GoogleHandler {
	return &GoogleHandler{
		OAuth:     o,
		Logger:    logger,
		TokenRepo: tr,
		Scanner:   services.NewGmailScanner(),
		JobRepo:   jr,
		JobQueue:  jq,
		SyncRepo:  sr,
	}
}

// GET /api/google/auth-url  (PROTECTED)
func (h *GoogleHandler) BeginAuthURL(w http.ResponseWriter, r *http.Request) {
	uid, err := userIDFromContext(r.Context()) // reads "user_id"
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	state := "st_" + strconv.FormatInt(time.Now().UnixNano(), 10)
	setShortCookie(w, cookieUID, strconv.Itoa(uid))
	setShortCookie(w, cookieState, state)

	url := h.OAuth.AuthCodeURL(state)
	writeJSON(w, http.StatusOK, map[string]string{"url": url})
}

// GET /api/google/callback  (PUBLIC)
func (h *GoogleHandler) Callback(w http.ResponseWriter, r *http.Request) {
	qState := r.URL.Query().Get("state")
	if qState == "" {
		http.Error(w, "missing state", http.StatusBadRequest)
		return
	}
	cState, _ := r.Cookie(cookieState)
	if cState == nil || cState.Value != qState {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	uidCookie, err := r.Cookie(cookieUID)
	if err != nil {
		http.Error(w, "unauthorized (no uid)", http.StatusUnauthorized)
		return
	}
	uid, err := strconv.Atoi(uidCookie.Value)
	if err != nil || uid <= 0 {
		http.Error(w, "unauthorized (bad uid)", http.StatusUnauthorized)
		return
	}

	tok, err := h.OAuth.Exchange(r.Context(), code)
	if err != nil {
		h.Logger.WithError(err).Warn("oauth exchange failed")
		http.Error(w, "oauth exchange failed", http.StatusBadRequest)
		return
	}

	if err := h.TokenRepo.Save(r.Context(), uid, providerGmail, tok); err != nil {
		h.Logger.WithError(err).Error("saving gmail token failed")
		http.Error(w, "saving token failed", http.StatusInternalServerError)
		return
	}
	if err := h.TokenRepo.Save(r.Context(), uid, providerGmail, tok); err != nil {
		h.Logger.WithError(err).Error("saving gmail token failed")
		http.Error(w, "saving token failed", http.StatusInternalServerError)
		return
	}

	// Queue initial sync job
	if err := h.JobQueue.CreateJob("gmail_initial_sync", uid, nil); err != nil {
		h.Logger.WithError(err).Error("failed to queue initial sync")
		// Don't fail the OAuth flow for this
	}
	clearCookie(w, cookieUID)
	clearCookie(w, cookieState)

	front := os.Getenv("FRONTEND_URL")
	if front == "" {
		front = "http://localhost:5173"
	}
	http.Redirect(w, r, front+"/dashboard?gmail=connected", http.StatusFound)
}

// GET /api/google/status  (PROTECTED)
func (h *GoogleHandler) Status(w http.ResponseWriter, r *http.Request) {
	uid, err := userIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	_, err = h.TokenRepo.Get(r.Context(), uid, providerGmail)
	writeJSON(w, http.StatusOK, map[string]any{"connected": err == nil})
}

// POST /api/google/disconnect  (PROTECTED)
func (h *GoogleHandler) Disconnect(w http.ResponseWriter, r *http.Request) {
	uid, err := userIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if err := h.TokenRepo.Delete(r.Context(), uid, providerGmail); err != nil {
		h.Logger.WithError(err).Warn("disconnect gmail failed")
		http.Error(w, "failed to disconnect", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "disconnected"})
}

// GET /api/google/scan  (PROTECTED)
func (h *GoogleHandler) Scan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	uid, err := userIDFromContext(ctx)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// 1. Fetch the IDs of jobs already in our database FIRST.
	existingIDs, err := h.JobRepo.GetAllGmailMessageIDsByUserID(uid)
	if err != nil {
		h.Logger.WithError(err).Error("failed to get existing job ids")
		http.Error(w, "failed to get existing jobs", http.StatusInternalServerError)
		return
	}

	tok, err := h.TokenRepo.Get(ctx, uid, providerGmail)
	if err != nil || tok == nil {
		http.Error(w, "gmail not connected", http.StatusUnauthorized)
		return
	}

	client := h.OAuth.Client(ctx, tok)
	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		h.Logger.WithError(err).Error("gmail client init error")
		http.Error(w, "gmail client error", http.StatusInternalServerError)
		return
	}

	// ... (The query param parsing for since, until, limit, etc. stays the same)
	since := time.Now().AddDate(-1, 0, 0)
	if qs := r.URL.Query().Get("since"); qs != "" {
		if t, e := time.Parse("2006-01-02", qs); e == nil {
			since = t
		}
	}
	var until time.Time
	if qu := r.URL.Query().Get("until"); qu != "" {
		if t, e := time.Parse("2006-01-02", qu); e == nil {
			until = t
		}
	}
	limit := int64(200)
	if ql := r.URL.Query().Get("limit"); ql != "" {
		if n, e := strconv.ParseInt(ql, 10, 64); e == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	cursor := r.URL.Query().Get("cursor")
	only := r.URL.Query().Get("only")

	// 2. Call the scanner, passing the set of existing IDs.
	res, err := h.Scanner.ScanPage(ctx, srv, since, until, limit, cursor, only, existingIDs)
	if err != nil {
		h.Logger.WithError(err).Error("gmail scan failed")
		http.Error(w, "scan failed", http.StatusInternalServerError)
		return
	}

	// The events in the response are now pre-filtered.
	payload := map[string]any{
		"events":        res.Events,
		"count":         len(res.Events),
		"nextPageToken": res.NextPageToken,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		h.Logger.WithError(err).Error("encode response")
	}
}

func (h *GoogleHandler) SyncStatus(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// ✅ Use your existing "GetOrCreateStatus" function
	status, err := h.SyncRepo.GetOrCreateStatus(userID)
	if err != nil {
		fmt.Printf("Error getting sync status for user %d: %v\n", userID, err)
		http.Error(w, "could not retrieve sync status", http.StatusInternalServerError)
		return
	}

	// Now, we need a struct that matches the frontend's expectations.
	// Let's create a response struct for clarity.
	type syncStatusResponse struct {
		IsSyncing  bool `json:"is_syncing"`
		FoundCount int  `json:"found_count"`
	}

	// Your logic to determine if syncing is in progress.
	// For example: is it started but not completed?
	isSyncingNow := status.InitialSyncStartedAt != nil && !status.InitialSyncCompleted

	// The `found_count` will come from another table later,
	// where you store the emails found but not yet imported.
	// For now, we can use the `TotalImported` as a placeholder or just 0.
	foundCount := 0 // You will update this logic later

	resp := syncStatusResponse{
		IsSyncing:  isSyncingNow,
		FoundCount: foundCount,
	}

	writeJSON(w, http.StatusOK, resp)
}

// ===== helpers =====

func userIDFromContext(ctx context.Context) (int, error) {
	// your auth middleware sets context key "user_id" (see middleware/auth.go)
	val := ctx.Value("user_id") // <— key from your middleware
	if val == nil {
		return 0, errors.New("no user in context")
	}
	switch v := val.(type) {
	case int:
		return v, nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		n, err := strconv.Atoi(v)
		if err != nil {
			return 0, err
		}
		return n, nil
	default:
		return 0, errors.New("unsupported user id type")
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func setShortCookie(w http.ResponseWriter, name, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode, // works on top-level redirects
		Secure:   false,                // set true in HTTPS
		MaxAge:   300,                  // 5 minutes
	})
}

func clearCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
}
