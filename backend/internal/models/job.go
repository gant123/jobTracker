package models

import (
	"time"
)

type Job struct {
	ID             int        `json:"id"`
	UserID         int        `json:"user_id"`
	Company        string     `json:"company"`
	Position       string     `json:"position"`
	Location       string     `json:"location,omitempty"`
	JobType        string     `json:"job_type,omitempty"`
	SalaryMin      *int       `json:"salary_min,omitempty"`
	SalaryMax      *int       `json:"salary_max,omitempty"`
	Currency       string     `json:"currency,omitempty"`
	Status         string     `json:"status"`
	URL            string     `json:"url,omitempty"`
	Description    string     `json:"description,omitempty"`
	Notes          string     `json:"notes,omitempty"`
	AppliedDate    *time.Time `json:"applied_date,omitempty"`
	InterviewDate  *time.Time `json:"interview_date,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	GmailMessageID string     `json:"gmail_message_id,omitempty"`
}

type CreateJobRequest struct {
	Company        string     `json:"company" validate:"required"`
	Position       string     `json:"position" validate:"required"`
	Location       string     `json:"location,omitempty"`
	JobType        string     `json:"job_type,omitempty"`
	SalaryMin      *int       `json:"salary_min,omitempty"`
	SalaryMax      *int       `json:"salary_max,omitempty"`
	Currency       string     `json:"currency,omitempty"`
	Status         string     `json:"status,omitempty"`
	URL            string     `json:"url,omitempty"`
	Description    string     `json:"description,omitempty"`
	Notes          string     `json:"notes,omitempty"`
	AppliedDate    *time.Time `json:"applied_date,omitempty"`
	InterviewDate  *time.Time `json:"interview_date,omitempty"`
	GmailMessageID string     `json:"gmail_message_id,omitempty"`
}

type UpdateJobRequest struct {
	Company       string     `json:"company,omitempty"`
	Position      string     `json:"position,omitempty"`
	Location      string     `json:"location,omitempty"`
	JobType       string     `json:"job_type,omitempty"`
	SalaryMin     *int       `json:"salary_min,omitempty"`
	SalaryMax     *int       `json:"salary_max,omitempty"`
	Currency      string     `json:"currency,omitempty"`
	Status        string     `json:"status,omitempty"`
	URL           string     `json:"url,omitempty"`
	Description   string     `json:"description,omitempty"`
	Notes         string     `json:"notes,omitempty"`
	AppliedDate   *time.Time `json:"applied_date,omitempty"`
	InterviewDate *time.Time `json:"interview_date,omitempty"`
}

type JobFilter struct {
	Status   string `json:"status,omitempty"`
	Company  string `json:"company,omitempty"`
	Location string `json:"location,omitempty"`
	Search   string `json:"search,omitempty"`
}
