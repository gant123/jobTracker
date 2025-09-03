package services

import (
	"errors"
	"fmt"
	"github.com/gant123/jobTracker/internal/models"
	"github.com/gant123/jobTracker/internal/repository"
)

type JobService struct {
	jobRepo *repository.JobRepository
}

func NewJobService(jobRepo *repository.JobRepository) *JobService {
	return &JobService{
		jobRepo: jobRepo,
	}
}

func (s *JobService) CreateJob(userID int, req *models.CreateJobRequest) (*models.Job, error) {
	job := &models.Job{
		UserID:         userID,
		Company:        req.Company,
		Position:       req.Position,
		Location:       req.Location,
		JobType:        req.JobType,
		SalaryMin:      req.SalaryMin,
		SalaryMax:      req.SalaryMax,
		Currency:       req.Currency,
		Status:         req.Status,
		URL:            req.URL,
		Description:    req.Description,
		Notes:          req.Notes,
		AppliedDate:    req.AppliedDate,
		InterviewDate:  req.InterviewDate,
		GmailMessageID: req.GmailMessageID,
	}

	if job.Status == "" {
		job.Status = "applied"
	}

	if job.Currency == "" {
		job.Currency = "USD"
	}
	if err := s.jobRepo.Create(job); err != nil {
		// Check if the error is our specific duplicate error.
		if errors.Is(err, repository.ErrDuplicate) {
			// If it's a duplicate, we return success (nil error).
			// The job object isn't created, but it's not a failure.
			return job, nil
		}
		// For any other error, report it as a failure.
		return nil, fmt.Errorf("failed to create job: %w", err)
	}
	return job, nil
}

func (s *JobService) GetJob(id int, userID int) (*models.Job, error) {
	return s.jobRepo.GetByID(id, userID)
}

func (s *JobService) GetJobs(userID int, filter *models.JobFilter) ([]*models.Job, error) {
	return s.jobRepo.GetAllByUserID(userID, filter)
}

func (s *JobService) UpdateJob(id int, userID int, req *models.UpdateJobRequest) (*models.Job, error) {
	// Get existing job
	job, err := s.jobRepo.GetByID(id, userID)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Company != "" {
		job.Company = req.Company
	}
	if req.Position != "" {
		job.Position = req.Position
	}
	if req.Location != "" {
		job.Location = req.Location
	}
	if req.JobType != "" {
		job.JobType = req.JobType
	}
	if req.SalaryMin != nil {
		job.SalaryMin = req.SalaryMin
	}
	if req.SalaryMax != nil {
		job.SalaryMax = req.SalaryMax
	}
	if req.Currency != "" {
		job.Currency = req.Currency
	}
	if req.Status != "" {
		job.Status = req.Status
	}
	if req.URL != "" {
		job.URL = req.URL
	}
	if req.Description != "" {
		job.Description = req.Description
	}
	if req.Notes != "" {
		job.Notes = req.Notes
	}
	if req.AppliedDate != nil {
		job.AppliedDate = req.AppliedDate
	}
	if req.InterviewDate != nil {
		job.InterviewDate = req.InterviewDate
	}

	if err := s.jobRepo.Update(job); err != nil {
		return nil, fmt.Errorf("failed to update job: %w", err)
	}

	return job, nil
}

func (s *JobService) DeleteJob(id int, userID int) error {
	return s.jobRepo.Delete(id, userID)
}

func (s *JobService) GetStats(userID int) (map[string]int, error) {
	return s.jobRepo.GetStats(userID)
}
