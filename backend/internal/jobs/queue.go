package jobs

import (
	"database/sql"
	"encoding/json"
	"time"
)

type JobType string

const (
	JobTypeInitialSync  JobType = "initial_sync"
	JobTypeProcessEmail JobType = "process_email"
	JobTypeRenewWatch   JobType = "renew_watch"
)

type Job struct {
	ID        int
	Type      JobType
	UserID    int
	Payload   json.RawMessage
	Status    string // pending, processing, completed, failed
	Attempts  int
	CreatedAt time.Time
	ProcessAt time.Time
}

type JobQueue struct {
	db *sql.DB
}

func (q *JobQueue) Enqueue(jobType JobType, userID int, payload interface{}) error {
	data, _ := json.Marshal(payload)
	_, err := q.db.Exec(`
        INSERT INTO jobs (type, user_id, payload, status, process_at)
        VALUES ($1, $2, $3, 'pending', NOW())
    `, jobType, userID, data)
	return err
}

func (q *JobQueue) ProcessNext() (*Job, error) {
	var job Job
	err := q.db.QueryRow(`
        UPDATE jobs
        SET status = 'processing', 
            attempts = attempts + 1,
            updated_at = NOW()
        WHERE id = (
            SELECT id FROM jobs
            WHERE status = 'pending' 
            AND process_at <= NOW()
            AND attempts < 3
            ORDER BY process_at
            LIMIT 1
            FOR UPDATE SKIP LOCKED
        )
        RETURNING id, type, user_id, payload, attempts
    `).Scan(&job.ID, &job.Type, &job.UserID, &job.Payload, &job.Attempts)

	return &job, err
}
