package repository

import (
	"database/sql"
	"time"
)

type GmailSyncRepository struct {
	db *sql.DB
}

func NewGmailSyncRepository(db *sql.DB) *GmailSyncRepository {
	return &GmailSyncRepository{db: db}
}

func (r *GmailSyncRepository) GetOrCreateStatus(userID int) (*GmailSyncStatus, error) {
	var status GmailSyncStatus

	err := r.db.QueryRow(`
        INSERT INTO gmail_sync_status (user_id)
        VALUES ($1)
        ON CONFLICT (user_id) DO UPDATE SET updated_at = NOW()
        RETURNING id, user_id, initial_sync_completed, initial_sync_started_at, 
                  initial_sync_completed_at, last_history_id, watch_expiration, total_imported
    `, userID).Scan(
		&status.ID, &status.UserID, &status.InitialSyncCompleted,
		&status.InitialSyncStartedAt, &status.InitialSyncCompletedAt,
		&status.LastHistoryID, &status.WatchExpiration, &status.TotalImported,
	)

	return &status, err
}

func (r *GmailSyncRepository) UpdateSyncStarted(userID int) error {
	_, err := r.db.Exec(`
        UPDATE gmail_sync_status 
        SET initial_sync_started_at = NOW(), updated_at = NOW()
        WHERE user_id = $1
    `, userID)
	return err
}

func (r *GmailSyncRepository) UpdateSyncCompleted(userID int, totalImported int) error {
	_, err := r.db.Exec(`
        UPDATE gmail_sync_status 
        SET initial_sync_completed = true,
            initial_sync_completed_at = NOW(),
            total_imported = $2,
            updated_at = NOW()
        WHERE user_id = $1
    `, userID, totalImported)
	return err
}

func (r *GmailSyncRepository) UpdateLastHistoryID(userID int, historyID string) error {
	_, err := r.db.Exec(`
        UPDATE gmail_sync_status 
        SET last_history_id = $2, updated_at = NOW()
        WHERE user_id = $1
    `, userID, historyID)
	return err
}

type GmailSyncStatus struct {
	ID                     int
	UserID                 int
	InitialSyncCompleted   bool
	InitialSyncStartedAt   *time.Time
	InitialSyncCompletedAt *time.Time
	LastHistoryID          *string
	WatchExpiration        *time.Time
	TotalImported          int
}
