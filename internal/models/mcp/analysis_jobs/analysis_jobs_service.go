package analysisjobs

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type IAnalysisJobService interface {
	CreateJob(ctx context.Context, job *AnalysisJob) (*AnalysisJob, error)
	GetJobByID(ctx context.Context, id uuid.UUID) (*AnalysisJob, error)
	ListJobs(ctx context.Context) ([]*AnalysisJob, error)
	UpdateJob(ctx context.Context, job *AnalysisJob) (*AnalysisJob, error)
	DeleteJob(ctx context.Context, id uuid.UUID) error
	ListJobsByStatus(ctx context.Context, status string) ([]*AnalysisJob, error)
	ListJobsByType(ctx context.Context, jobType string) ([]*AnalysisJob, error)
	ListJobsByUserID(ctx context.Context, userID uuid.UUID) ([]*AnalysisJob, error)
	ListJobsByProjectID(ctx context.Context, projectID uuid.UUID) ([]*AnalysisJob, error)
	StartJob(ctx context.Context, id uuid.UUID) error
	CompleteJob(ctx context.Context, id uuid.UUID, outputData interface{}) error
	FailJob(ctx context.Context, id uuid.UUID, errorMessage string) error
	UpdateJobProgress(ctx context.Context, id uuid.UUID, progress float64) error
	RetryJob(ctx context.Context, id uuid.UUID) error
	GetPendingJobs(ctx context.Context) ([]*AnalysisJob, error)
	GetRunningJobs(ctx context.Context) ([]*AnalysisJob, error)
	GetFailedJobs(ctx context.Context) ([]*AnalysisJob, error)
	ValidateJobData(ctx context.Context, job *AnalysisJob) error
}

type AnalysisJobService struct {
	Repo IAnalysisJobRepo
}

func NewAnalysisJobService(repo IAnalysisJobRepo) IAnalysisJobService {
	return &AnalysisJobService{Repo: repo}
}

func (s *AnalysisJobService) CreateJob(ctx context.Context, job *AnalysisJob) (*AnalysisJob, error) {
	if job == nil {
		return nil, errors.New("job cannot be nil")
	}

	if err := s.ValidateJobData(ctx, job); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	if job.ID == uuid.Nil {
		job.ID = uuid.New()
	}

	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Now()
	}
	job.UpdatedAt = time.Now()

	return s.Repo.Create(ctx, job)
}

func (s *AnalysisJobService) GetJobByID(ctx context.Context, id uuid.UUID) (*AnalysisJob, error) {
	if id == uuid.Nil {
		return nil, errors.New("job ID cannot be empty")
	}
	return s.Repo.FindByID(ctx, id)
}

func (s *AnalysisJobService) ListJobs(ctx context.Context) ([]*AnalysisJob, error) {
	return s.Repo.FindAll(ctx)
}

func (s *AnalysisJobService) UpdateJob(ctx context.Context, job *AnalysisJob) (*AnalysisJob, error) {
	if job == nil {
		return nil, errors.New("job cannot be nil")
	}
	if job.ID == uuid.Nil {
		return nil, errors.New("job ID cannot be empty")
	}

	if err := s.ValidateJobData(ctx, job); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	job.UpdatedAt = time.Now()
	return s.Repo.Update(ctx, job)
}

func (s *AnalysisJobService) DeleteJob(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.New("job ID cannot be empty")
	}
	return s.Repo.Delete(ctx, id)
}

func (s *AnalysisJobService) ListJobsByStatus(ctx context.Context, status string) ([]*AnalysisJob, error) {
	if status == "" {
		return nil, errors.New("status cannot be empty")
	}
	return s.Repo.FindByStatus(ctx, status)
}

func (s *AnalysisJobService) ListJobsByType(ctx context.Context, jobType string) ([]*AnalysisJob, error) {
	if jobType == "" {
		return nil, errors.New("job type cannot be empty")
	}
	return s.Repo.FindByJobType(ctx, jobType)
}

func (s *AnalysisJobService) ListJobsByUserID(ctx context.Context, userID uuid.UUID) ([]*AnalysisJob, error) {
	if userID == uuid.Nil {
		return nil, errors.New("user ID cannot be empty")
	}
	return s.Repo.FindByUserID(ctx, userID)
}

func (s *AnalysisJobService) ListJobsByProjectID(ctx context.Context, projectID uuid.UUID) ([]*AnalysisJob, error) {
	if projectID == uuid.Nil {
		return nil, errors.New("project ID cannot be empty")
	}
	return s.Repo.FindByProjectID(ctx, projectID)
}

func (s *AnalysisJobService) StartJob(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.New("job ID cannot be empty")
	}

	job, err := s.Repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find job: %w", err)
	}

	if job.Status != "PENDING" {
		return fmt.Errorf("job cannot be started from status: %s", job.Status)
	}

	return s.Repo.MarkAsStarted(ctx, id)
}

func (s *AnalysisJobService) CompleteJob(ctx context.Context, id uuid.UUID, outputData interface{}) error {
	if id == uuid.Nil {
		return errors.New("job ID cannot be empty")
	}

	job, err := s.Repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find job: %w", err)
	}

	if job.Status != "RUNNING" {
		return fmt.Errorf("job cannot be completed from status: %s", job.Status)
	}

	return s.Repo.MarkAsCompleted(ctx, id, outputData)
}

func (s *AnalysisJobService) FailJob(ctx context.Context, id uuid.UUID, errorMessage string) error {
	if id == uuid.Nil {
		return errors.New("job ID cannot be empty")
	}
	if errorMessage == "" {
		return errors.New("error message cannot be empty")
	}

	return s.Repo.MarkAsFailed(ctx, id, errorMessage)
}

func (s *AnalysisJobService) UpdateJobProgress(ctx context.Context, id uuid.UUID, progress float64) error {
	if id == uuid.Nil {
		return errors.New("job ID cannot be empty")
	}
	if progress < 0 || progress > 100 {
		return errors.New("progress must be between 0 and 100")
	}

	return s.Repo.UpdateProgress(ctx, id, progress)
}

func (s *AnalysisJobService) RetryJob(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.New("job ID cannot be empty")
	}

	job, err := s.Repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find job: %w", err)
	}

	if job.Status != "FAILED" {
		return fmt.Errorf("only failed jobs can be retried, current status: %s", job.Status)
	}

	if job.RetryCount >= job.MaxRetries {
		return fmt.Errorf("job has exceeded maximum retry attempts (%d)", job.MaxRetries)
	}

	if err := s.Repo.IncrementRetryCount(ctx, id); err != nil {
		return fmt.Errorf("failed to increment retry count: %w", err)
	}

	return s.Repo.UpdateStatus(ctx, id, "PENDING")
}

func (s *AnalysisJobService) GetPendingJobs(ctx context.Context) ([]*AnalysisJob, error) {
	return s.Repo.FindPendingJobs(ctx)
}

func (s *AnalysisJobService) GetRunningJobs(ctx context.Context) ([]*AnalysisJob, error) {
	return s.Repo.FindRunningJobs(ctx)
}

func (s *AnalysisJobService) GetFailedJobs(ctx context.Context) ([]*AnalysisJob, error) {
	return s.Repo.FindFailedJobs(ctx)
}

func (s *AnalysisJobService) ValidateJobData(ctx context.Context, job *AnalysisJob) error {
	if job.JobType == "" {
		return errors.New("job type cannot be empty")
	}

	if job.UserID == uuid.Nil {
		return errors.New("user ID cannot be empty")
	}

	validStatuses := map[string]bool{
		"PENDING":   true,
		"RUNNING":   true,
		"COMPLETED": true,
		"FAILED":    true,
		"CANCELLED": true,
	}

	if job.Status != "" && !validStatuses[job.Status] {
		return fmt.Errorf("invalid status: %s", job.Status)
	}

	validJobTypes := map[string]bool{
		"SCORECARD_ANALYSIS":   true,
		"CODE_ANALYSIS":        true,
		"SECURITY_ANALYSIS":    true,
		"PERFORMANCE_ANALYSIS": true,
		"DEPENDENCY_ANALYSIS":  true,
	}

	if !validJobTypes[job.JobType] {
		return fmt.Errorf("invalid job type: %s", job.JobType)
	}

	if job.Progress < 0 || job.Progress > 100 {
		return errors.New("progress must be between 0 and 100")
	}

	if job.MaxRetries < 0 {
		return errors.New("max retries cannot be negative")
	}

	if job.RetryCount < 0 {
		return errors.New("retry count cannot be negative")
	}

	return nil
}
