// Package notifications implements a notification system that integrates with analysis jobs.
package notifications

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	analysisJobs "github.com/kubex-ecosystem/gdbase/internal/models/mcp/analysis_jobs"
	gl "github.com/kubex-ecosystem/logz"
)

// AnalysisJobNotificationIntegration integra o sistema de notificações com analysis_jobs
type AnalysisJobNotificationIntegration struct {
	eventProcessor INotificationEventProcessor
	enabled        bool
}

// NewAnalysisJobNotificationIntegration cria uma nova integração
func NewAnalysisJobNotificationIntegration(eventProcessor INotificationEventProcessor) *AnalysisJobNotificationIntegration {
	return &AnalysisJobNotificationIntegration{
		eventProcessor: eventProcessor,
		enabled:        true,
	}
}

// Enable ativa a integração
func (a *AnalysisJobNotificationIntegration) Enable() {
	a.enabled = true
	gl.Log("info", "Analysis job notification integration enabled")
}

// Disable desativa a integração
func (a *AnalysisJobNotificationIntegration) Disable() {
	a.enabled = false
	gl.Log("info", "Analysis job notification integration disabled")
}

// IsEnabled retorna se a integração está ativa
func (a *AnalysisJobNotificationIntegration) IsEnabled() bool {
	return a.enabled
}

// OnJobStatusChanged é chamado quando o status de um job muda
func (a *AnalysisJobNotificationIntegration) OnJobStatusChanged(ctx context.Context, job analysisJobs.IAnalysisJob, oldStatus, newStatus string) error {
	if !a.enabled {
		return nil
	}

	gl.Log("info", "Job status changed notification trigger",
		"job_id", job.GetID(),
		"old_status", oldStatus,
		"new_status", newStatus,
		"job_type", job.GetJobType())

	// Create base event data from job
	jobData := a.extractJobData(job)

	// Create and process event based on status change
	var event *NotificationEvent
	projectID := job.GetProjectID()

	switch newStatus {
	case "COMPLETED":
		event = NewJobCompletedEvent(job.GetID(), job.GetUserID(), &projectID, jobData)

		// Check for score alerts if job completed successfully
		if err := a.checkScoreAlert(ctx, job, jobData); err != nil {
			gl.Log("error", "Failed to check score alert", "job_id", job.GetID(), "error", err)
		}

	case "FAILED":
		errorMessage := job.GetErrorMessage()
		if errorMessage == "" {
			errorMessage = "Job failed without specific error message"
		}
		event = NewJobFailedEvent(job.GetID(), job.GetUserID(), &projectID, errorMessage, jobData)

	case "RUNNING":
		if oldStatus == "PENDING" {
			event = a.createJobStartedEvent(job, jobData)
		}

	case "PENDING":
		if oldStatus == "FAILED" {
			event = a.createJobRetriedEvent(job, jobData)
		}
	}

	// Process the event if one was created
	if event != nil {
		if err := a.eventProcessor.ProcessEvent(ctx, event); err != nil {
			gl.Log("error", "Failed to process notification event",
				"job_id", job.GetID(),
				"event_type", event.EventType,
				"error", err)
			return fmt.Errorf("failed to process notification event: %w", err)
		}

		gl.Log("info", "Successfully processed notification event",
			"job_id", job.GetID(),
			"event_type", event.EventType,
			"event_id", event.ID)
	}

	return nil
}

// OnJobProgressChanged é chamado quando o progresso de um job muda
func (a *AnalysisJobNotificationIntegration) OnJobProgressChanged(ctx context.Context, job analysisJobs.IAnalysisJob, oldProgress, newProgress float64) error {
	if !a.enabled {
		return nil
	}

	// Only trigger notifications for significant progress milestones
	milestones := []float64{25, 50, 75, 90}

	for _, milestone := range milestones {
		if oldProgress < milestone && newProgress >= milestone {
			gl.Log("info", "Job progress milestone reached",
				"job_id", job.GetID(),
				"progress", newProgress,
				"milestone", milestone)

			// Create progress event (could be extended to create notifications)
			jobData := a.extractJobData(job)
			jobData["progress_milestone"] = milestone
			jobData["previous_progress"] = oldProgress
			jobData["current_progress"] = newProgress

			// For now, just log the milestone. Could be extended to create notifications
			// based on user preferences for progress updates
			break
		}
	}

	return nil
}

// OnJobTimeout é chamado quando um job excede o tempo limite
func (a *AnalysisJobNotificationIntegration) OnJobTimeout(ctx context.Context, job analysisJobs.IAnalysisJob, duration time.Duration) error {
	if !a.enabled {
		return nil
	}

	gl.Log("warn", "Job timeout detected",
		"job_id", job.GetID(),
		"duration", duration,
		"job_type", job.GetJobType())

	jobData := a.extractJobData(job)
	jobData["timeout_duration"] = duration
	jobData["timeout_occurred"] = true
	projectID := job.GetProjectID()
	// Create time alert event
	event := &NotificationEvent{
		ID:         uuid.New(),
		EventType:  NotificationEventTypeTimeAlert,
		SourceType: "analysis_job",
		SourceID:   job.GetID(),
		UserID:     job.GetUserID(),
		ProjectID:  &projectID,
		Data:       jobData,
		Metadata: map[string]interface{}{
			"alert_reason":                "job_timeout",
			"duration_threshold_exceeded": true,
		},
		Timestamp: time.Now(),
	}

	if err := a.eventProcessor.ProcessEvent(ctx, event); err != nil {
		gl.Log("error", "Failed to process timeout notification event",
			"job_id", job.GetID(),
			"error", err)
		return fmt.Errorf("failed to process timeout notification event: %w", err)
	}

	return nil
}

// checkScoreAlert verifica se o score do job está abaixo do limite
func (a *AnalysisJobNotificationIntegration) checkScoreAlert(ctx context.Context, job analysisJobs.IAnalysisJob, jobData map[string]interface{}) error {
	outputData := job.GetOutputData()
	if outputData == nil {
		return nil
	}

	var score float64
	var scoreFound bool
	for key, value := range outputData {
		if key == "overall_score" {
			if scoreFloat, ok := value.(float64); ok {
				score = scoreFloat
				scoreFound = true
			}
		}
	}

	if !scoreFound {
		return nil // No score to check
	}

	// Add score to job data
	jobData["score"] = score

	// Check if score is below common thresholds
	thresholds := []float64{0.5, 0.6, 0.7, 0.8}

	for _, threshold := range thresholds {
		if score < threshold {
			gl.Log("info", "Score alert threshold met",
				"job_id", job.GetID(),
				"score", score,
				"threshold", threshold)
			var projectID uuid.UUID
			if job.GetProjectID() != uuid.Nil {
				projectID = job.GetProjectID()
			}
			// Create score alert event
			event := NewScoreAlertEvent(job.GetID(), job.GetUserID(), &projectID, score, threshold, jobData)

			if err := a.eventProcessor.ProcessEvent(ctx, event); err != nil {
				return fmt.Errorf("failed to process score alert event: %w", err)
			}

			// Only trigger alert for the highest threshold that was crossed
			break
		}
	}

	return nil
}

// createJobStartedEvent cria evento de job iniciado
func (a *AnalysisJobNotificationIntegration) createJobStartedEvent(job analysisJobs.IAnalysisJob, jobData map[string]interface{}) *NotificationEvent {
	var projectID uuid.UUID
	if job.GetProjectID() != uuid.Nil {
		projectID = job.GetProjectID()
	}
	return &NotificationEvent{
		ID:         uuid.New(),
		EventType:  NotificationEventTypeJobStarted,
		SourceType: "analysis_job",
		SourceID:   job.GetID(),
		UserID:     job.GetUserID(),
		ProjectID:  &projectID,
		Data:       jobData,
		Metadata: map[string]interface{}{
			"started_at": job.GetStartedAt(),
		},
		Timestamp: time.Now(),
	}
}

// createJobRetriedEvent cria evento de job tentado novamente
func (a *AnalysisJobNotificationIntegration) createJobRetriedEvent(job analysisJobs.IAnalysisJob, jobData map[string]interface{}) *NotificationEvent {
	var projectID uuid.UUID
	if job.GetProjectID() != uuid.Nil {
		projectID = job.GetProjectID()
	}
	return &NotificationEvent{
		ID:         uuid.New(),
		EventType:  NotificationEventTypeJobRetried,
		SourceType: "analysis_job",
		SourceID:   job.GetID(),
		UserID:     job.GetUserID(),
		ProjectID:  &projectID,
		Data:       jobData,
		Metadata: map[string]interface{}{
			"retry_count": job.GetRetryCount(),
			"max_retries": job.GetMaxRetries(),
		},
		Timestamp: time.Now(),
	}
}

// extractJobData extrai dados relevantes do job para os eventos
func (a *AnalysisJobNotificationIntegration) extractJobData(job analysisJobs.IAnalysisJob) map[string]interface{} {
	data := map[string]interface{}{
		"job_id":      job.GetID(),
		"job_type":    job.GetJobType(),
		"job_status":  job.GetStatus(),
		"progress":    job.GetProgress(),
		"project_id":  job.GetProjectID(),
		"user_id":     job.GetUserID(),
		"source_url":  job.GetSourceURL(),
		"source_type": job.GetSourceType(),
		"retry_count": job.GetRetryCount(),
		"max_retries": job.GetMaxRetries(),
		"created_at":  job.GetCreatedAt(),
		"updated_at":  job.GetUpdatedAt(),
	}

	// Add started_at if available
	if !job.GetStartedAt().IsZero() {
		data["started_at"] = job.GetStartedAt()
	}

	// Add completed_at if available
	if !job.GetCompletedAt().IsZero() {
		data["completed_at"] = job.GetCompletedAt()

		// Calculate duration if both started_at and completed_at are available
		if !job.GetStartedAt().IsZero() {
			duration := job.GetCompletedAt().Sub(job.GetStartedAt())
			data["duration"] = duration.String()
			data["duration_seconds"] = duration.Seconds()
		}
	}

	// Add error message if present
	if job.GetErrorMessage() != "" {
		data["error_message"] = job.GetErrorMessage()
	}

	// Add input data if present
	if job.GetInputData() != nil {
		data["input_data"] = job.GetInputData()
	}

	// Add output data if present
	if job.GetOutputData() != nil {
		data["output_data"] = job.GetOutputData()

		// Extract common fields from output data
		if outputMap, ok := data["output_data"].(map[string]interface{}); ok {
			if score, exists := outputMap["overall_score"]; exists {
				data["score"] = score
			}
			if tags, exists := outputMap["tags"]; exists {
				data["tags"] = tags
			}
		}
	}

	// Add metadata if present
	if job.GetMetadata() != nil {
		data["metadata"] = job.GetMetadata()
	}

	return data
}

// AnalysisJobNotificationHooks fornece hooks para integrar com o sistema de analysis_jobs
type AnalysisJobNotificationHooks struct {
	integration *AnalysisJobNotificationIntegration
}

// NewAnalysisJobNotificationHooks cria novos hooks de notificação
func NewAnalysisJobNotificationHooks(integration *AnalysisJobNotificationIntegration) *AnalysisJobNotificationHooks {
	return &AnalysisJobNotificationHooks{
		integration: integration,
	}
}

// InstallHooks instala os hooks no sistema de analysis_jobs
// Esta função deve ser chamada durante a inicialização do sistema
func (h *AnalysisJobNotificationHooks) InstallHooks(ctx context.Context) error {
	if h.integration == nil {
		return fmt.Errorf("notification integration not configured")
	}

	gl.Log("info", "Installing analysis job notification hooks")

	// Registrar handlers customizados no event processor
	h.integration.eventProcessor.RegisterEventHandler(
		NotificationEventTypeJobCompleted,
		h.handleJobCompletedEvent,
	)

	h.integration.eventProcessor.RegisterEventHandler(
		NotificationEventTypeJobFailed,
		h.handleJobFailedEvent,
	)

	h.integration.eventProcessor.RegisterEventHandler(
		NotificationEventTypeScoreAlert,
		h.handleScoreAlertEvent,
	)

	gl.Log("info", "Analysis job notification hooks installed successfully")
	return nil
}

// handleJobCompletedEvent trata eventos de job concluído
func (h *AnalysisJobNotificationHooks) handleJobCompletedEvent(ctx context.Context, event *NotificationEvent) error {
	gl.Log("info", "Handling job completed event", "job_id", event.SourceID, "event_id", event.ID)

	// Custom logic for job completed events
	// For example, you could:
	// - Update project statistics
	// - Trigger dependent jobs
	// - Send summary emails
	// - Update dashboards

	return nil
}

// handleJobFailedEvent trata eventos de job falhado
func (h *AnalysisJobNotificationHooks) handleJobFailedEvent(ctx context.Context, event *NotificationEvent) error {
	gl.Log("warn", "Handling job failed event", "job_id", event.SourceID, "event_id", event.ID)

	// Custom logic for job failed events
	// For example, you could:
	// - Create incident reports
	// - Alert on-call engineers
	// - Trigger automatic retry logic
	// - Update failure metrics

	return nil
}

// handleScoreAlertEvent trata eventos de alerta de score
func (h *AnalysisJobNotificationHooks) handleScoreAlertEvent(ctx context.Context, event *NotificationEvent) error {
	score, _ := event.Data["score"].(float64)
	threshold, _ := event.Data["threshold"].(float64)

	gl.Log("warn", "Handling score alert event",
		"job_id", event.SourceID,
		"score", score,
		"threshold", threshold,
		"event_id", event.ID)

	// Custom logic for score alert events
	// For example, you could:
	// - Flag repositories for review
	// - Create improvement tasks
	// - Update quality metrics
	// - Trigger additional analysis

	return nil
}

// IntegrateWithAnalysisJobService integra o sistema de notificações com o serviço de análise de jobs
func IntegrateWithAnalysisJobService(jobService analysisJobs.IAnalysisJobService, integration *AnalysisJobNotificationIntegration) {
	// Esta função seria implementada para integrar com o service de analysis_jobs
	// Idealmente, o AnalysisJobService teria callbacks/hooks que poderíamos registrar

	gl.Log("info", "Integrating notification system with analysis job service")

	// Example integration (this would depend on the actual AnalysisJobService implementation):
	// jobService.OnStatusChanged(integration.OnJobStatusChanged)
	// jobService.OnProgressChanged(integration.OnJobProgressChanged)
	// jobService.OnTimeout(integration.OnJobTimeout)

	// For now, we log that integration is set up and ready
	gl.Log("info", "Analysis job notification integration ready - waiting for events")
}
