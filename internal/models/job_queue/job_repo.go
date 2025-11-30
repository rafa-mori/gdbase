package jobqueue

import (
	"context"
	"errors"
	"fmt"
	"time"

	svc "github.com/kubex-ecosystem/gdbase/internal/services"
	gl "github.com/kubex-ecosystem/logz"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IJobQueueRepo interface {
	Create(ctx context.Context, job *JobQueue) (*JobQueue, error)
	FindByID(ctx context.Context, id uuid.UUID) (*JobQueue, error)
	FindAll(ctx context.Context) ([]*JobQueue, error)
	Update(ctx context.Context, job *JobQueue) (*JobQueue, error)
	Delete(ctx context.Context, id uuid.UUID) error
	FindByStatus(ctx context.Context, status string) ([]*JobQueue, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*JobQueue, error)
	FindByType(ctx context.Context, jobType string) ([]*JobQueue, error)
	FindByCreatedAt(ctx context.Context, createdAt time.Time) ([]*JobQueue, error)
	FindByUpdatedAt(ctx context.Context, updatedAt time.Time) ([]*JobQueue, error)
	FindByCreatedBy(ctx context.Context, createdBy uuid.UUID) ([]*JobQueue, error)
	FindByUpdatedBy(ctx context.Context, updatedBy uuid.UUID) ([]*JobQueue, error)
	FindByLastExecutedAt(ctx context.Context, lastExecutedAt time.Time) ([]*JobQueue, error)
	FindByLastExecutedBy(ctx context.Context, lastExecutedBy uuid.UUID) ([]*JobQueue, error)
	FindByStatusAndUserID(ctx context.Context, status string, userID uuid.UUID) ([]*JobQueue, error)
	FindByStatusAndType(ctx context.Context, status string, jobType string) ([]*JobQueue, error)
	FindByStatusAndCreatedAt(ctx context.Context, status string, createdAt time.Time) ([]*JobQueue, error)
	FindByStatusAndUpdatedAt(ctx context.Context, status string, updatedAt time.Time) ([]*JobQueue, error)
	FindByStatusAndCreatedBy(ctx context.Context, status string, createdBy uuid.UUID) ([]*JobQueue, error)
	FindByStatusAndUpdatedBy(ctx context.Context, status string, updatedBy uuid.UUID) ([]*JobQueue, error)

	FindByStatusAndLastExecutedAt(ctx context.Context, status string, lastExecutedAt time.Time) ([]*JobQueue, error)
	FindByStatusAndLastExecutedBy(ctx context.Context, status string, lastExecutedBy uuid.UUID) ([]*JobQueue, error)
	FindByUserIDAndType(ctx context.Context, userID uuid.UUID, jobType string) ([]*JobQueue, error)
	FindByUserIDAndCreatedAt(ctx context.Context, userID uuid.UUID, createdAt time.Time) ([]*JobQueue, error)
	FindByUserIDAndCreatedBy(ctx context.Context, userID uuid.UUID, createdBy uuid.UUID) ([]*JobQueue, error)
	FindByUserIDAndUpdatedAt(ctx context.Context, userID uuid.UUID, updatedAt time.Time) ([]*JobQueue, error)
	FindByUserIDAndUpdatedBy(ctx context.Context, userID uuid.UUID, updatedBy uuid.UUID) ([]*JobQueue, error)
	FindByUserIDAndLastExecutedAt(ctx context.Context, userID uuid.UUID, lastExecutedAt time.Time) ([]*JobQueue, error)
	FindByUserIDAndLastExecutedBy(ctx context.Context, userID uuid.UUID, lastExecutedBy uuid.UUID) ([]*JobQueue, error)

	ExecuteJobManually(ctx context.Context, jobID uuid.UUID) error
	RetryFailedJob(ctx context.Context, jobID uuid.UUID) error
	RescheduleJob(ctx context.Context, jobID uuid.UUID, newSchedule time.Time) error
	ValidateJobSchedule(ctx context.Context, schedule string) error
}

type JobQueueRepository struct {
	db *gorm.DB
}

func NewJobQueueRepository(ctx context.Context, dbService *svc.DBServiceImpl) IJobQueueRepo {
	db, err := svc.GetDB(ctx, dbService)
	if err != nil {
		gl.Log("error", fmt.Sprintf("JobQueueRepository: failed to get DB: %v", err))
		return nil
	}
	return &JobQueueRepository{db: db}
}

// Implement repository methods here

func (repo *JobQueueRepository) Create(ctx context.Context, job *JobQueue) (*JobQueue, error) {
	if err := repo.db.Create(job).Error; err != nil {
		return nil, err
	}
	return job, nil
}
func (repo *JobQueueRepository) FindByID(ctx context.Context, id uuid.UUID) (*JobQueue, error) {
	var job JobQueue
	if err := repo.db.First(&job, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &job, nil
}
func (repo *JobQueueRepository) FindAll(ctx context.Context) ([]*JobQueue, error) {
	var jobs []*JobQueue
	if err := repo.db.Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}
func (repo *JobQueueRepository) Update(ctx context.Context, job *JobQueue) (*JobQueue, error) {
	if err := repo.db.Save(job).Error; err != nil {
		return nil, err
	}
	return job, nil
}
func (repo *JobQueueRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := repo.db.Delete(&JobQueue{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}
func (repo *JobQueueRepository) FindByStatus(ctx context.Context, status string) ([]*JobQueue, error) {
	var jobs []*JobQueue
	if err := repo.db.Where("status = ?", status).Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}
func (repo *JobQueueRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*JobQueue, error) {
	var jobs []*JobQueue
	if err := repo.db.Where("user_id = ?", userID).Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}
func (repo *JobQueueRepository) FindByType(ctx context.Context, jobType string) ([]*JobQueue, error) {
	var jobs []*JobQueue
	if err := repo.db.Where("type = ?", jobType).Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}
func (repo *JobQueueRepository) FindByCreatedAt(ctx context.Context, createdAt time.Time) ([]*JobQueue, error) {
	var jobs []*JobQueue
	if err := repo.db.Where("created_at = ?", createdAt).Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}
func (repo *JobQueueRepository) FindByUpdatedAt(ctx context.Context, updatedAt time.Time) ([]*JobQueue, error) {
	var jobs []*JobQueue
	if err := repo.db.Where("updated_at = ?", updatedAt).Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}
func (repo *JobQueueRepository) FindByCreatedBy(ctx context.Context, createdBy uuid.UUID) ([]*JobQueue, error) {
	var jobs []*JobQueue
	if err := repo.db.Where("created_by = ?", createdBy).Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}
func (repo *JobQueueRepository) FindByUpdatedBy(ctx context.Context, updatedBy uuid.UUID) ([]*JobQueue, error) {
	var jobs []*JobQueue
	if err := repo.db.Where("updated_by = ?", updatedBy).Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}
func (repo *JobQueueRepository) FindByLastExecutedAt(ctx context.Context, lastExecutedAt time.Time) ([]*JobQueue, error) {
	var jobs []*JobQueue
	if err := repo.db.Where("last_executed_at = ?", lastExecutedAt).Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}
func (repo *JobQueueRepository) FindByLastExecutedBy(ctx context.Context, lastExecutedBy uuid.UUID) ([]*JobQueue, error) {
	var jobs []*JobQueue
	if err := repo.db.Where("last_executed_by = ?", lastExecutedBy).Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}
func (repo *JobQueueRepository) FindByStatusAndUserID(ctx context.Context, status string, userID uuid.UUID) ([]*JobQueue, error) {
	var jobs []*JobQueue
	if err := repo.db.Where("status = ? AND user_id = ?", status, userID).Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}
func (repo *JobQueueRepository) FindByStatusAndType(ctx context.Context, status string, jobType string) ([]*JobQueue, error) {
	var jobs []*JobQueue
	if err := repo.db.Where("status = ? AND type = ?", status, jobType).Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}
func (repo *JobQueueRepository) FindByStatusAndCreatedAt(ctx context.Context, status string, createdAt time.Time) ([]*JobQueue, error) {
	var jobs []*JobQueue
	if err := repo.db.Where("status = ? AND created_at = ?", status, createdAt).Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}
func (repo *JobQueueRepository) FindByStatusAndCreatedBy(ctx context.Context, status string, createdBy uuid.UUID) ([]*JobQueue, error) {
	var jobs []*JobQueue
	if err := repo.db.Where("status = ? AND created_by = ?", status, createdBy).Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}
func (repo *JobQueueRepository) FindByStatusAndUpdatedAt(ctx context.Context, status string, updatedAt time.Time) ([]*JobQueue, error) {
	var jobs []*JobQueue
	if err := repo.db.Where("status = ? AND updated_at = ?", status, updatedAt).Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}
func (repo *JobQueueRepository) FindByStatusAndUpdatedBy(ctx context.Context, status string, updatedBy uuid.UUID) ([]*JobQueue, error) {
	var jobs []*JobQueue
	if err := repo.db.Where("status = ? AND updated_by = ?", status, updatedBy).Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}
func (repo *JobQueueRepository) FindByStatusAndLastExecutedAt(ctx context.Context, status string, lastExecutedAt time.Time) ([]*JobQueue, error) {
	var jobs []*JobQueue
	if err := repo.db.Where("status = ? AND last_executed_at = ?", status, lastExecutedAt).Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}
func (repo *JobQueueRepository) FindByStatusAndLastExecutedBy(ctx context.Context, status string, lastExecutedBy uuid.UUID) ([]*JobQueue, error) {
	var jobs []*JobQueue
	if err := repo.db.Where("status = ? AND last_executed_by = ?", status, lastExecutedBy).Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}
func (repo *JobQueueRepository) FindByUserIDAndType(ctx context.Context, userID uuid.UUID, jobType string) ([]*JobQueue, error) {
	var jobs []*JobQueue
	if err := repo.db.Where("user_id = ? AND type = ?", userID, jobType).Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}
func (repo *JobQueueRepository) FindByUserIDAndCreatedAt(ctx context.Context, userID uuid.UUID, createdAt time.Time) ([]*JobQueue, error) {
	var jobs []*JobQueue
	if err := repo.db.Where("user_id = ? AND created_at = ?", userID, createdAt).Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}
func (repo *JobQueueRepository) FindByUserIDAndCreatedBy(ctx context.Context, userID uuid.UUID, createdBy uuid.UUID) ([]*JobQueue, error) {
	var jobs []*JobQueue
	if err := repo.db.Where("user_id = ? AND created_by = ?", userID, createdBy).Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}
func (repo *JobQueueRepository) FindByUserIDAndUpdatedAt(ctx context.Context, userID uuid.UUID, updatedAt time.Time) ([]*JobQueue, error) {
	var jobs []*JobQueue
	if err := repo.db.Where("user_id = ? AND updated_at = ?", userID, updatedAt).Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}
func (repo *JobQueueRepository) FindByUserIDAndUpdatedBy(ctx context.Context, userID uuid.UUID, updatedBy uuid.UUID) ([]*JobQueue, error) {
	var jobs []*JobQueue
	if err := repo.db.Where("user_id = ? AND updated_by = ?", userID, updatedBy).Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}
func (repo *JobQueueRepository) FindByUserIDAndLastExecutedAt(ctx context.Context, userID uuid.UUID, lastExecutedAt time.Time) ([]*JobQueue, error) {
	var jobs []*JobQueue
	if err := repo.db.Where("user_id = ? AND last_executed_at = ?", userID, lastExecutedAt).Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}
func (repo *JobQueueRepository) FindByUserIDAndLastExecutedBy(ctx context.Context, userID uuid.UUID, lastExecutedBy uuid.UUID) ([]*JobQueue, error) {
	var jobs []*JobQueue
	if err := repo.db.Where("user_id = ? AND last_executed_by = ?", userID, lastExecutedBy).Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}
func (repo *JobQueueRepository) ExecuteJobManually(ctx context.Context, jobID uuid.UUID) error {
	var job JobQueue
	if err := repo.db.First(&job, "id = ?", jobID).Error; err != nil {
		return err
	}
	job.Status = "executing"
	if err := repo.db.Save(&job).Error; err != nil {
		return err
	}
	// Execute the job logic here
	return nil
}
func (repo *JobQueueRepository) RetryFailedJob(ctx context.Context, jobID uuid.UUID) error {
	var job JobQueue
	if err := repo.db.First(&job, "id = ?", jobID).Error; err != nil {
		return err
	}
	job.Status = "retrying"
	if err := repo.db.Save(&job).Error; err != nil {
		return err
	}
	// Retry the job logic here
	return nil
}
func (repo *JobQueueRepository) RescheduleJob(ctx context.Context, jobID uuid.UUID, newSchedule time.Time) error {
	var job JobQueue
	if err := repo.db.First(&job, "id = ?", jobID).Error; err != nil {
		return err
	}
	job.ScheduledAt = newSchedule
	if err := repo.db.Save(&job).Error; err != nil {
		return err
	}
	return nil
}
func (repo *JobQueueRepository) ValidateJobSchedule(ctx context.Context, schedule string) error {
	if schedule == "" {
		return errors.New("schedule cannot be empty")
	}
	// Add your validation logic here
	return nil
}
