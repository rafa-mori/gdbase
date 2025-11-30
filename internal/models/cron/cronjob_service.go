// Package cron provides functionality for managing cron jobs.
package cron

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	jobqueue "github.com/kubex-ecosystem/gdbase/internal/models/job_queue"
	svc "github.com/kubex-ecosystem/gdbase/internal/services"
	gl "github.com/kubex-ecosystem/logz"
	amqp "github.com/rabbitmq/amqp091-go"
)

type ICronJobService interface {
	CreateCronJob(ctx context.Context, job *CronJob) (*CronJob, error)
	GetCronJobByID(ctx context.Context, id uuid.UUID) (*CronJob, error)
	ListCronJobs(ctx context.Context) ([]*CronJob, error)
	UpdateCronJob(ctx context.Context, job *CronJob) (*CronJob, error)
	DeleteCronJob(ctx context.Context, id uuid.UUID) error
	EnableCronJob(ctx context.Context, id uuid.UUID) error
	DisableCronJob(ctx context.Context, id uuid.UUID) error
	ExecuteCronJobManually(ctx context.Context, id uuid.UUID) error
	ListActiveCronJobs(ctx context.Context) ([]*CronJob, error)
	RescheduleCronJob(ctx context.Context, id uuid.UUID, newExpression string) error
	ValidateCronExpression(ctx context.Context, expression string) error
	GetJobQueue(ctx context.Context) ([]jobqueue.JobQueue, error)
	ReprocessFailedJobs(ctx context.Context) error
	GetExecutionLogs(ctx context.Context, cronJobID uuid.UUID) ([]jobqueue.ExecutionLog, error)
	GetScheduledCronJobs(ctx context.Context) ([]Job, error)
}

var (
	dbConfig *svc.DBConfig
)

type CronJobService struct {
	Repo ICronJobRepo
}

func NewCronJobServiceImpl(repo *CronJobRepo) *CronJobService {
	return &CronJobService{Repo: repo}
}
func NewCronJobService(repo ICronJobRepo) ICronJobService {
	if rp, ok := repo.(*CronJobRepo); !ok {
		gl.Log("error", "Invalid repository type provided to NewCronJobService")
		return nil
	} else {
		return NewCronJobServiceImpl(rp)
	}
}

func (s *CronJobService) publishToRabbitMQ(ctx context.Context, queueName string, message string) error {
	iDBConfig := svc.NewDBConfig(dbConfig)
	if iDBConfig == nil {
		gl.Log("error", "Failed to create database config")
		return errors.New("failed to create database config")
	}
	url := getRabbitMQURL(iDBConfig)
	if url == "" {
		gl.Log("error", "Failed to get RabbitMQ URL")
		return errors.New("failed to get RabbitMQ URL")
	}
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Printf("Failed to connect to RabbitMQ: %s", err)
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Printf("Failed to open a channel: %s", err)
		return err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		queueName,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("Failed to declare a queue: %s", err)
		return err
	}

	err = ch.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)
	if err != nil {
		log.Printf("Failed to publish a message: %s", err)
		return err
	}

	log.Printf("Message published to queue %s: %s", queueName, message)
	return nil
}

func (s *CronJobService) CreateCronJob(ctx context.Context, job *CronJob) (*CronJob, error) {
	if job.Name == "" {
		return nil, errors.New("job name is required")
	}

	artifact := NewCronJob(ctx, job, false).CronJobObject()
	if artifact == nil {
		gl.Log("error", "Failed to create cron job object")
		return nil, errors.New("failed to create cron job object")
	}

	createdJob, err := s.Repo.Create(ctx, artifact)
	if err != nil {
		return nil, err
	}

	// Publish to RabbitMQ
	if err := s.publishToRabbitMQ(ctx, "cronjob_events", "CronJob Created: "+createdJob.ID.String()); err != nil {
		log.Printf("Failed to publish create event: %s", err)
	}

	return createdJob, nil
}

func (s *CronJobService) GetCronJobByID(ctx context.Context, id uuid.UUID) (*CronJob, error) {
	return s.Repo.FindByID(ctx, id)
}

func (s *CronJobService) ListCronJobs(ctx context.Context) ([]*CronJob, error) {
	return s.Repo.FindAll(ctx)
}

func (s *CronJobService) UpdateCronJob(ctx context.Context, job *CronJob) (*CronJob, error) {
	if job.ID == uuid.Nil {
		return nil, errors.New("job ID is required")
	}
	updatedJob, err := s.Repo.Update(ctx, job)
	if err != nil {
		return nil, err
	}

	// Publish to RabbitMQ
	if err := s.publishToRabbitMQ(ctx, "cronjob_events", "CronJob Updated: "+updatedJob.ID.String()); err != nil {
		log.Printf("Failed to publish update event: %s", err)
	}

	return updatedJob, nil
}

func (s *CronJobService) DeleteCronJob(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.New("job ID is required")
	}
	if err := s.Repo.Delete(ctx, id); err != nil {
		return err
	}

	// Publish to RabbitMQ
	if err := s.publishToRabbitMQ(ctx, "cronjob_events", "CronJob Deleted: "+id.String()); err != nil {
		log.Printf("Failed to publish delete event: %s", err)
	}

	return nil
}

func (s *CronJobService) EnableCronJob(ctx context.Context, id uuid.UUID) error {
	job, err := s.Repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	job.IsActive = true
	_, err = s.Repo.Update(ctx, job)
	return err
}

func (s *CronJobService) DisableCronJob(ctx context.Context, id uuid.UUID) error {
	job, err := s.Repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	job.IsActive = false
	_, err = s.Repo.Update(ctx, job)
	return err
}

func (s *CronJobService) ExecuteCronJobManually(ctx context.Context, id uuid.UUID) error {
	job, err := s.Repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	// Simulate execution logic here (e.g., log execution or trigger a worker)
	job.LastRunStatus = "success"
	now := time.Now().UTC()
	job.LastRunTime = &now
	_, err = s.Repo.Update(ctx, job)
	return err
}

func (s *CronJobService) ListActiveCronJobs(ctx context.Context) ([]*CronJob, error) {
	cronList, err := s.Repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	var activeCronJobs []*CronJob
	for _, job := range cronList {
		if job.IsActive {
			activeCronJobs = append(activeCronJobs, job)
		}
	}
	return activeCronJobs, nil
}

func (s *CronJobService) RescheduleCronJob(ctx context.Context, id uuid.UUID, newExpression string) error {
	job, err := s.Repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	job.CronExpression = newExpression
	_, err = s.Repo.Update(ctx, job)
	return err
}

func (s *CronJobService) ValidateCronExpression(ctx context.Context, expression string) error {
	expression = strings.TrimSpace(expression)
	if expression == "" {
		return errors.New("cron expression cannot be empty")
	}
	// Implement cron expression validation logic here.
	// For example, you can use a library like "github.com/robfig/cron/v3" to validate the expression.
	// _, err := cron.ParseStandard(expression)
	return fmt.Errorf("cron expression '%s' is not valid", expression)
}

func (s *CronJobService) GetScheduledCronJobs(ctx context.Context) ([]Job, error) {
	// Implement logic to fetch scheduled cron jobs from the repository.
	return nil, errors.New("not implemented")
}

// GetJobQueue retrieves the current state of the job queue.
func (s *CronJobService) GetJobQueue(ctx context.Context) ([]jobqueue.JobQueue, error) {
	// Implement logic to fetch the job queue from the repository.
	return nil, errors.New("not implemented")
}

// ReprocessFailedJobs reprocesses all failed jobs in the queue.
func (s *CronJobService) ReprocessFailedJobs(ctx context.Context) error {
	// Implement logic to reprocess failed jobs.
	return errors.New("not implemented")
}

// GetExecutionLogs retrieves execution logs for a specific cron job.
func (s *CronJobService) GetExecutionLogs(ctx context.Context, cronJobID uuid.UUID) ([]jobqueue.ExecutionLog, error) {
	// Implement logic to fetch execution logs from the repository.
	return nil, errors.New("not implemented")
}

// SaveCronJob ensures all required fields are set and saves the CronJob.
func (s *CronJobService) SaveCronJob(ctx context.Context, job *CronJob, defaultUserID uuid.UUID) error {
	// Preencher campos automaticamente
	job.PrepareForSave(ctx, defaultUserID)

	// Salvar no repositório usando o método Update
	_, err := s.Repo.Update(ctx, job)
	return err
}

func getRabbitMQURL(dbConfig *svc.DBConfig) string {
	if dbConfig != nil {
		if dbConfig.Messagery != nil {
			if dbConfig.Messagery.RabbitMQ != nil {
				return fmt.Sprintf("amqp://%s:%s@%s:%d/",
					dbConfig.Messagery.RabbitMQ.Username,
					dbConfig.Messagery.RabbitMQ.Password,
					dbConfig.Messagery.RabbitMQ.Host,
					dbConfig.Messagery.RabbitMQ.Port,
				)
			}
		}
	}
	return ""
}
