package repository

import (
	"database/sql"
	"encoding/json"
)

type JobQueueRepository struct {
	db *sql.DB
}

func NewJobQueueRepository(db *sql.DB) *JobQueueRepository {
	return &JobQueueRepository{db: db}
}

func (r *JobQueueRepository) CreateJob(jobType string, userID int, payload interface{}) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(`
        INSERT INTO background_jobs (type, user_id, payload, status, process_after)
        VALUES ($1, $2, $3, 'pending', NOW())
    `, jobType, userID, payloadJSON)

	return err
}

func (r *JobQueueRepository) GetNextJob() (*BackgroundJob, error) {
	var job BackgroundJob
	var payloadJSON []byte

	err := r.db.QueryRow(`
        UPDATE background_jobs
        SET status = 'processing', 
            attempts = attempts + 1,
            updated_at = NOW()
        WHERE id = (
            SELECT id FROM background_jobs
            WHERE status = 'pending' 
            AND process_after <= NOW()
            AND attempts < 3
            ORDER BY created_at
            LIMIT 1
            FOR UPDATE SKIP LOCKED
        )
        RETURNING id, type, user_id, payload, attempts
    `).Scan(&job.ID, &job.Type, &job.UserID, &payloadJSON, &job.Attempts)

	if err == nil && payloadJSON != nil {
		json.Unmarshal(payloadJSON, &job.Payload)
	}

	return &job, err
}

func (r *JobQueueRepository) MarkJobComplete(jobID int) error {
	_, err := r.db.Exec(`
        UPDATE background_jobs 
        SET status = 'completed', updated_at = NOW() 
        WHERE id = $1
    `, jobID)
	return err
}

func (r *JobQueueRepository) MarkJobFailed(jobID int, errMsg string) error {
	_, err := r.db.Exec(`
        UPDATE background_jobs 
        SET status = 'failed', error = $1, updated_at = NOW() 
        WHERE id = $2
    `, errMsg, jobID)
	return err
}

type BackgroundJob struct {
	ID       int
	Type     string
	UserID   int
	Payload  map[string]interface{}
	Attempts int
}
