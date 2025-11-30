package cron

import (
	"context"
	"time"

	"github.com/google/uuid"
	t "github.com/kubex-ecosystem/gdbase/internal/types"
	l "github.com/kubex-ecosystem/logz"

	jobqueue "github.com/kubex-ecosystem/gdbase/internal/models/job_queue"

	gl "github.com/kubex-ecosystem/logz"
)

type ICronJobModel interface {
	TableName() string
	PrimaryKey() string
	GetID() uuid.UUID
	SetID(id uuid.UUID)
	GetCreatedAt() time.Time
	SetCreatedAt(createdAt time.Time)
	GetUpdatedAt() *time.Time
	SetUpdatedAt(updatedAt *time.Time)
	GetCreatedBy() uuid.UUID
	SetCreatedBy(createdBy uuid.UUID)
	GetUpdatedBy() uuid.UUID
	SetUpdatedBy(updatedBy uuid.UUID)
	GetLastExecutedAt() *time.Time
	SetLastExecutedAt(lastExecutedAt *time.Time)
	GetLastExecutedBy() uuid.UUID
	SetLastExecutedBy(lastExecutedBy uuid.UUID)
	GetUserID() uuid.UUID
	SetUserID(userID uuid.UUID)
	CronJobObject() *CronJob
}

type CronJob struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;primary_key,default:uuid_generate_v4()" binding:"-"`
	UserID         uuid.UUID `json:"user_id" gorm:"type:uuid;references:users(id)" binding:"omitempty"`
	CreatedBy      uuid.UUID `json:"created_by" gorm:"type:uuid;references:users(id)" binding:"omitempty"`
	UpdatedBy      uuid.UUID `json:"updated_by" gorm:"type:uuid;references:users(id),omitempty" binding:"omitempty"`
	LastExecutedBy uuid.UUID `json:"last_executed_by" gorm:"type:uuid;references:users(id),omitempty" binding:"omitempty"`

	Name           string `json:"name" gorm:"type:varchar(255);not null" binding:"required"`
	Description    string `json:"description" gorm:"type:text" binding:"omitempty"`
	CronType       string `json:"cron_type" gorm:"type:enum('cron', 'interval');default:'cron'" binding:"omitempty"`
	CronExpression string `json:"cron_expression" gorm:"type:text;default:'2 * * * *'" binding:"omitempty"`
	Command        string `json:"command" gorm:"type:text"`                                // ajustar para que não seja obrigatório, mas revisar a lógica de execução antes
	Method         string `json:"method" gorm:"type:enum('GET', 'POST', 'PUT', 'DELETE')"` // ajustar para que não seja obrigatório, mas revisar a lógica de execução antes
	APIEndpoint    string `json:"api_endpoint" gorm:"type:varchar(255)"`
	LastRunStatus  string `json:"last_run_status" gorm:"type:enum('success', 'failure', 'pending');default:'pending'" binding:"omitempty"`
	LastRunMessage string `json:"last_run_message" gorm:"type:text" binding:"omitempty"`

	Retries          int `json:"retries" gorm:"default:0" binding:"omitempty"`
	ExecTimeout      int `json:"exec_timeout" gorm:"default:30" binding:"omitempty"`
	MaxRetries       int `json:"max_retries" gorm:"default:3" binding:"omitempty"`
	RetryInterval    int `json:"retry_interval" gorm:"default:10" binding:"omitempty"`
	MaxExecutionTime int `json:"max_execution_time" gorm:"default:300" binding:"omitempty"`

	IsRecurring bool `json:"is_recurring" gorm:"default:false" binding:"omitempty"`
	IsActive    bool `json:"is_active" gorm:"default:true" binding:"omitempty"`

	StartsAt  time.Time `json:"starts_at" gorm:"default:now()" binding:"omitempty"`
	CreatedAt time.Time `json:"created_at" gorm:"default:now()" binding:"omitempty"`

	EndsAt         *time.Time `json:"ends_at" binding:"omitempty"`
	LastRunTime    *time.Time `json:"last_run_time" binding:"omitempty"`
	UpdatedAt      *time.Time `json:"updated_at" gorm:"default:now()" binding:"omitempty"`
	LastExecutedAt *time.Time `json:"last_executed_at" binding:"omitempty"`

	Payload t.JSONBImpl `json:"payload" binding:"omitempty"`
	Headers t.JSONBImpl `json:"headers" binding:"omitempty"`

	Metadata t.JSONBImpl `json:"metadata" binding:"omitempty"`
}

func NewCronJob(ctx context.Context, cron *CronJob, restrict bool) ICronJobModel {
	if cron == nil {
		cron = &CronJob{}
	}
	cv := NewCronModelValidation(ctx, l.GetLogger(""), false)
	cronV, err := cv.ValidateCronJobProperties(ctx, cron, restrict)
	if err != nil {
		gl.Log("error", "Failed to validate cron job properties")
		return nil
	}
	if cronV == nil {
		gl.Log("error", "Cron job is nil after validation")
		return nil
	}
	return cronV
}

func (c *CronJob) TableName() string                           { return "cron_jobs" }
func (c *CronJob) PrimaryKey() string                          { return "id" }
func (c *CronJob) GetID() uuid.UUID                            { return c.ID }
func (c *CronJob) SetID(id uuid.UUID)                          { c.ID = id }
func (c *CronJob) GetCreatedAt() time.Time                     { return c.CreatedAt }
func (c *CronJob) SetCreatedAt(createdAt time.Time)            { c.CreatedAt = createdAt }
func (c *CronJob) GetUpdatedAt() *time.Time                    { return c.UpdatedAt }
func (c *CronJob) SetUpdatedAt(updatedAt *time.Time)           { c.UpdatedAt = updatedAt }
func (c *CronJob) GetCreatedBy() uuid.UUID                     { return c.CreatedBy }
func (c *CronJob) SetCreatedBy(createdBy uuid.UUID)            { c.CreatedBy = createdBy }
func (c *CronJob) GetUpdatedBy() uuid.UUID                     { return c.UpdatedBy }
func (c *CronJob) SetUpdatedBy(updatedBy uuid.UUID)            { c.UpdatedBy = updatedBy }
func (c *CronJob) GetLastExecutedAt() *time.Time               { return c.LastExecutedAt }
func (c *CronJob) SetLastExecutedAt(lastExecutedAt *time.Time) { c.LastExecutedAt = lastExecutedAt }
func (c *CronJob) GetLastExecutedBy() uuid.UUID                { return c.LastExecutedBy }
func (c *CronJob) SetLastExecutedBy(lastExecutedBy uuid.UUID)  { c.LastExecutedBy = lastExecutedBy }
func (c *CronJob) GetUserID() uuid.UUID                        { return c.UserID }
func (c *CronJob) SetUserID(userID uuid.UUID)                  { c.UserID = userID }
func (c *CronJob) GetScheduledCronJobs() []CronJob {
	return []CronJob{*c}
}

// Add a method to enqueue a job into the JobQueue

func (c *CronJob) EnqueueJob(ctx context.Context, jobQueueService jobqueue.IJobQueueService) error {
	job := &jobqueue.JobQueue{
		CronJobID:      c.ID,
		Status:         "pending",
		ScheduledAt:    time.Now(),
		ExecutionTime:  time.Now().Add(time.Duration(c.ExecTimeout) * time.Second),
		JobType:        "cron",
		JobExpression:  c.CronExpression,
		JobCommand:     c.Command,
		JobMethod:      c.Method,
		JobAPIEndpoint: c.APIEndpoint,
		JobPayload:     c.Payload,
		JobHeaders:     c.Headers,
		JobRetries:     c.Retries,
		JobTimeout:     c.ExecTimeout,
		UserID:         c.UserID,
		CreatedBy:      c.CreatedBy,
		UpdatedBy:      c.UpdatedBy,
	}
	_, err := jobQueueService.CreateJob(ctx, job)
	return err
}

// LogExecutionDetails logs execution details into the ExecutionLog using the provided service.
func (c *CronJob) LogExecutionDetails(ctx context.Context, service jobqueue.IExecutionLogService, details jobqueue.ExecutionLog) error {
	return service.CreateLog(ctx, details)
}

// PrepareForSave ensures all required fields are set before saving a CronJob.
func (c *CronJob) PrepareForSave(ctx context.Context, defaultUserID uuid.UUID) {
	if c.UserID == uuid.Nil {
		c.UserID = defaultUserID
	}
	if c.CreatedBy == uuid.Nil {
		c.CreatedBy = defaultUserID
	}
	if c.UpdatedBy == uuid.Nil {
		c.UpdatedBy = defaultUserID
	}
	if c.LastExecutedBy == uuid.Nil {
		c.LastExecutedBy = defaultUserID
	}
}

func (c *CronJob) CronJobObject() *CronJob { return c }

type Job interface {
	Mu() *t.Mutexes
	Ref() *t.Reference
	GetUserID() uuid.UUID
	Run() error
	Retry() error
	Cancel() error
}

func (c *CronJob) Mu() *t.Mutexes {
	return &t.Mutexes{}
}

func (c *CronJob) Ref() *t.Reference {
	return t.NewReference("cron_jobs").GetReference()
}

func (c *CronJob) Run() error {
	// Implementar a lógica de execução do cron job
	gl.Log("info", "Executing cron job: "+c.Name)
	return nil
}

func (c *CronJob) Retry() error {
	// Implementar a lógica de retry do cron job
	gl.Log("info", "Retrying cron job: "+c.Name)
	return nil
}

func (c *CronJob) Cancel() error {
	// Implementar a lógica de cancelamento do cron job
	gl.Log("info", "Cancelling cron job: "+c.Name)
	return nil
}
