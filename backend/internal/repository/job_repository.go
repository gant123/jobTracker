package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gant123/jobTracker/internal/models"
)

var ErrDuplicate = errors.New("record already exists")

type JobRepository struct {
	db *sql.DB
}

func NewJobRepository(db *sql.DB) *JobRepository {
	return &JobRepository{db: db}
}

func (r *JobRepository) Create(job *models.Job) error {
	query := `
        INSERT INTO jobs (
            user_id, company, position, location, job_type,
            salary_min, salary_max, currency, status, url,
            description, notes, applied_date, interview_date, gmail_message_id
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
        ON CONFLICT (gmail_message_id) DO NOTHING
        RETURNING id, created_at, updated_at
    `
	// Add job.GmailMessageID as the 15th paramete
	err := r.db.QueryRow(
		query,
		job.UserID,
		job.Company,
		job.Position,
		job.Location,
		job.JobType,
		job.SalaryMin,
		job.SalaryMax,
		job.Currency,
		job.Status,
		job.URL,
		job.Description,
		job.Notes,
		job.AppliedDate,
		job.InterviewDate,
		job.GmailMessageID,
	).Scan(&job.ID, &job.CreatedAt, &job.UpdatedAt)
	if err != nil {
		// If the error is "no rows", it means our ON CONFLICT was triggered.
		// We return our custom ErrDuplicate so the service layer can handle it.
		if err == sql.ErrNoRows {
			return ErrDuplicate
		}
		// For any other error, we return it as a real problem.
		return fmt.Errorf("failed to create job: %w", err)
	}
	return nil
}

func (r *JobRepository) GetByID(id int, userID int) (*models.Job, error) {
	job := &models.Job{}

	query := `
        SELECT 
            id, user_id, company, position, location, job_type,
            salary_min, salary_max, currency, status, url,
            description, notes, applied_date, interview_date,
            created_at, updated_at
        FROM jobs
        WHERE id = $1 AND user_id = $2
    `

	err := r.db.QueryRow(query, id, userID).Scan(
		&job.ID,
		&job.UserID,
		&job.Company,
		&job.Position,
		&job.Location,
		&job.JobType,
		&job.SalaryMin,
		&job.SalaryMax,
		&job.Currency,
		&job.Status,
		&job.URL,
		&job.Description,
		&job.Notes,
		&job.AppliedDate,
		&job.InterviewDate,
		&job.CreatedAt,
		&job.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("job not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	return job, nil
}

func (r *JobRepository) GetAllByUserID(userID int, filter *models.JobFilter) ([]*models.Job, error) {
	query := `
        SELECT 
            id, user_id, company, position, location, job_type,
            salary_min, salary_max, currency, status, url,
            description, notes, applied_date, interview_date,
            created_at, updated_at
        FROM jobs
        WHERE user_id = $1
    `

	args := []interface{}{userID}
	argCounter := 2

	if filter != nil {
		if filter.Status != "" {
			query += fmt.Sprintf(" AND status = $%d", argCounter)
			args = append(args, filter.Status)
			argCounter++
		}

		if filter.Company != "" {
			query += fmt.Sprintf(" AND LOWER(company) LIKE LOWER($%d)", argCounter)
			args = append(args, "%"+filter.Company+"%")
			argCounter++
		}

		if filter.Location != "" {
			query += fmt.Sprintf(" AND LOWER(location) LIKE LOWER($%d)", argCounter)
			args = append(args, "%"+filter.Location+"%")
			argCounter++
		}

		if filter.Search != "" {
			query += fmt.Sprintf(" AND (LOWER(company) LIKE LOWER($%d) OR LOWER(position) LIKE LOWER($%d))", argCounter, argCounter)
			args = append(args, "%"+filter.Search+"%")
			argCounter++
		}
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*models.Job
	for rows.Next() {
		job := &models.Job{}
		err := rows.Scan(
			&job.ID,
			&job.UserID,
			&job.Company,
			&job.Position,
			&job.Location,
			&job.JobType,
			&job.SalaryMin,
			&job.SalaryMax,
			&job.Currency,
			&job.Status,
			&job.URL,
			&job.Description,
			&job.Notes,
			&job.AppliedDate,
			&job.InterviewDate,
			&job.CreatedAt,
			&job.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (r *JobRepository) Update(job *models.Job) error {
	query := `
        UPDATE jobs
        SET company = $1, position = $2, location = $3, job_type = $4,
            salary_min = $5, salary_max = $6, currency = $7, status = $8,
            url = $9, description = $10, notes = $11, applied_date = $12,
            interview_date = $13, updated_at = CURRENT_TIMESTAMP
        WHERE id = $14 AND user_id = $15
        RETURNING updated_at
    `

	err := r.db.QueryRow(
		query,
		job.Company,
		job.Position,
		job.Location,
		job.JobType,
		job.SalaryMin,
		job.SalaryMax,
		job.Currency,
		job.Status,
		job.URL,
		job.Description,
		job.Notes,
		job.AppliedDate,
		job.InterviewDate,
		job.ID,
		job.UserID,
	).Scan(&job.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	return nil
}

func (r *JobRepository) Delete(id int, userID int) error {
	query := `DELETE FROM jobs WHERE id = $1 AND user_id = $2`

	result, err := r.db.Exec(query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete job: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("job not found")
	}

	return nil
}

func (r *JobRepository) GetStats(userID int) (map[string]int, error) {
	query := `
        SELECT status, COUNT(*) as count
        FROM jobs
        WHERE user_id = $1
        GROUP BY status
    `

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}
	defer rows.Close()

	stats := make(map[string]int)
	stats["total"] = 0

	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			continue
		}
		stats[status] = count
		stats["total"] += count
	}

	return stats, nil
}

// NEW: This function efficiently fetches only the Gmail message IDs for a user.
// GetAllGmailMessageIDsByUserID efficiently fetches only the Gmail message IDs for a user.
func (r *JobRepository) GetAllGmailMessageIDsByUserID(userID int) (map[string]struct{}, error) {
	query := `SELECT gmail_message_id FROM jobs WHERE user_id = $1 AND gmail_message_id IS NOT NULL AND gmail_message_id != ''`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query gmail message ids: %w", err)
	}
	defer rows.Close()

	// Use a map[string]struct{} as a "Set" for fast lookups.
	idSet := make(map[string]struct{})
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan gmail message id: %w", err)
		}
		idSet[id] = struct{}{}
	}

	return idSet, nil
}
