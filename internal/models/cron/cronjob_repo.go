package cron

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	svc "github.com/kubex-ecosystem/gdbase/internal/services"
	gl "github.com/kubex-ecosystem/logz"

	"gorm.io/gorm"
)

type ICronJobRepo interface {
	Create(ctx context.Context, job *CronJob) (*CronJob, error)
	FindByID(ctx context.Context, id uuid.UUID) (*CronJob, error)
	FindAll(ctx context.Context) ([]*CronJob, error)
	Update(ctx context.Context, job *CronJob) (*CronJob, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type CronJobRepo struct {
	DB *gorm.DB
}

func NewCronJobRepoImpl(ctx context.Context, dbService *svc.DBServiceImpl) *CronJobRepo {
	db, err := svc.GetDB(ctx, dbService)
	if err != nil {
		gl.Log("error", fmt.Sprintf("CronJobRepo: failed to get DB: %v", err))
		return nil
	}
	return &CronJobRepo{DB: db}
}
func NewCronJobRepo(ctx context.Context, dbService *svc.DBServiceImpl) ICronJobRepo {
	return NewCronJobRepoImpl(ctx, dbService)
}

func (r *CronJobRepo) Create(ctx context.Context, job *CronJob) (*CronJob, error) {
	userID, ok := ctx.Value("userID").(uuid.UUID)
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	job.CreatedAt = time.Now()
	job.UserID = userID
	job.CreatedBy = userID
	job.UpdatedBy = userID
	job.LastExecutedBy = userID
	job.UpdatedAt = &job.CreatedAt
	job.LastExecutedAt = nil
	job.LastRunTime = nil
	if job.LastRunStatus == "" {
		job.LastRunStatus = "pending"
	}
	if job.ID == uuid.Nil {
		var err error
		job.ID, err = uuid.NewRandom()
		if err != nil {
			return nil, err
		}
	}
	if err := r.DB.WithContext(ctx).Create(job).Error; err != nil {
		return nil, err
	}
	return job, nil
}

func (r *CronJobRepo) FindByID(ctx context.Context, id uuid.UUID) (*CronJob, error) {
	var job CronJob
	if err := r.DB.WithContext(ctx).First(&job, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *CronJobRepo) FindAll(ctx context.Context) ([]*CronJob, error) {
	var jobs []*CronJob
	if err := r.DB.WithContext(ctx).Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}

func (r *CronJobRepo) Update(ctx context.Context, job *CronJob) (*CronJob, error) {
	if err := r.DB.WithContext(ctx).Save(job).Error; err != nil {
		return nil, err
	}
	return job, nil
}

func (r *CronJobRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.DB.WithContext(ctx).Delete(&CronJob{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}

func (r *CronJobRepo) GetScheduledCronJobs(ctx context.Context) ([]*CronJob, error) {
	var jobs []*CronJob
	if err := r.DB.WithContext(ctx).Where("is_active = ?", true).Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}
