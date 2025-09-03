package database

import (
	"database/sql"
	"fmt"
)

func Migrate(db *sql.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
            id SERIAL PRIMARY KEY,
            email VARCHAR(255) UNIQUE NOT NULL,
            password VARCHAR(255) NOT NULL,
            name VARCHAR(255) NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )`,

		`CREATE TABLE IF NOT EXISTS jobs (
            id SERIAL PRIMARY KEY,
            user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	    gmail_message_id VARCHAR(255),
            company VARCHAR(255) NOT NULL,
            position VARCHAR(255) NOT NULL,
            location VARCHAR(255),
            job_type VARCHAR(50),
            salary_min INTEGER,
            salary_max INTEGER,
            currency VARCHAR(10) DEFAULT 'USD',
            status VARCHAR(50) DEFAULT 'applied',
            url TEXT,
            description TEXT,
            notes TEXT,
            applied_date DATE,
            interview_date TIMESTAMP,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	    UNIQUE(gmail_message_id)
        )`,
		// Gmail/Email OAuth token storage (encrypted at rest)
		`CREATE TABLE IF NOT EXISTS email_tokens (
            id SERIAL PRIMARY KEY,
	    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	    provider VARCHAR(50) NOT NULL,                    -- e.g., 'gmail'
	    access_token_enc BYTEA NOT NULL,
	    refresh_token_enc BYTEA,
	    expiry TIMESTAMP,                                 -- token expiry if present
	    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	    UNIQUE (user_id, provider)
)`,
		`CREATE TABLE IF NOT EXISTS gmail_sync_status (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    initial_sync_completed BOOLEAN DEFAULT FALSE,
    initial_sync_started_at TIMESTAMP,
    initial_sync_completed_at TIMESTAMP,
    last_history_id VARCHAR(255),
    watch_expiration TIMESTAMP,
    total_imported INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id)
)`,
		`CREATE TABLE IF NOT EXISTS background_jobs (
    id SERIAL PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    payload JSONB,
    status VARCHAR(20) DEFAULT 'pending',
    attempts INTEGER DEFAULT 0,
    error TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    process_after TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)`,
		`CREATE INDEX IF NOT EXISTS idx_jobs_status_process ON background_jobs(status, process_after)`,
		`CREATE INDEX IF NOT EXISTS idx_email_tokens_user_provider ON email_tokens(user_id, provider)`,
		`CREATE INDEX IF NOT EXISTS idx_jobs_gmail_message_id ON jobs(gmail_message_id)`,
		`CREATE INDEX IF NOT EXISTS idx_jobs_user_id ON jobs(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status)`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}
