package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	gl "github.com/kubex-ecosystem/logz"
)

// NotificationEventType representa o tipo de evento de notificação
type NotificationEventType string

const (
	NotificationEventTypeJobStatusChanged NotificationEventType = "JOB_STATUS_CHANGED"
	NotificationEventTypeJobCompleted     NotificationEventType = "JOB_COMPLETED"
	NotificationEventTypeJobFailed        NotificationEventType = "JOB_FAILED"
	NotificationEventTypeJobStarted       NotificationEventType = "JOB_STARTED"
	NotificationEventTypeJobRetried       NotificationEventType = "JOB_RETRIED"
	NotificationEventTypeScoreAlert       NotificationEventType = "SCORE_ALERT"
	NotificationEventTypeTimeAlert        NotificationEventType = "TIME_ALERT"
	NotificationEventTypeSystemAlert      NotificationEventType = "SYSTEM_ALERT"
)

// NotificationEvent representa um evento que pode disparar notificações
type NotificationEvent struct {
	ID           uuid.UUID              `json:"id"`
	EventType    NotificationEventType  `json:"event_type"`
	SourceType   string                 `json:"source_type"` // "analysis_job", "system", "custom"
	SourceID     uuid.UUID              `json:"source_id"`
	UserID       uuid.UUID              `json:"user_id"`
	ProjectID    *uuid.UUID             `json:"project_id,omitempty"`
	Data         map[string]interface{} `json:"data"`
	Metadata     map[string]interface{} `json:"metadata"`
	Timestamp    time.Time              `json:"timestamp"`
	ProcessedAt  *time.Time             `json:"processed_at,omitempty"`
	ErrorMessage string                 `json:"error_message,omitempty"`
}

// INotificationEventProcessor interface para processar eventos de notificação
type INotificationEventProcessor interface {
	ProcessEvent(ctx context.Context, event *NotificationEvent) error
	GetEligibleRules(ctx context.Context, event *NotificationEvent) ([]INotificationRule, error)
	CreateNotifications(ctx context.Context, event *NotificationEvent, rules []INotificationRule) ([]INotificationHistory, error)
	PublishToMessageQueue(ctx context.Context, event *NotificationEvent) error
	RegisterEventHandler(eventType NotificationEventType, handler EventHandler)
	UnregisterEventHandler(eventType NotificationEventType)
}

// EventHandler função para lidar com eventos específicos
type EventHandler func(ctx context.Context, event *NotificationEvent) error

// NotificationEventProcessor implementação do processador de eventos
type NotificationEventProcessor struct {
	ruleRepo      INotificationRuleRepo
	templateRepo  INotificationTemplateRepo
	historyRepo   INotificationHistoryRepo
	messageQueue  MessageQueuePublisher
	eventHandlers map[NotificationEventType]EventHandler
}

// MessageQueuePublisher interface para publicar mensagens na fila
type MessageQueuePublisher interface {
	Publish(exchange, routingKey string, body []byte) error
}

type INotificationRuleRepo interface {
	FindActiveRulesForEvent(ctx context.Context, eventType NotificationEventType, jobType string, userID, projectID uuid.UUID) ([]INotificationRule, error)
	FindByID(ctx context.Context, id uuid.UUID) (INotificationRule, error)
}

type INotificationTemplateRepo interface {
	FindByID(ctx context.Context, id uuid.UUID) (INotificationTemplate, error)
	FindDefaultByType(ctx context.Context, templateType NotificationTemplateType, language string) (INotificationTemplate, error)
}

type INotificationHistoryRepo interface {
	Create(ctx context.Context, history INotificationHistory) (INotificationHistory, error)
	CountRecentNotifications(ctx context.Context, ruleID uuid.UUID, since time.Time) (int, error)
}

// NewNotificationEventProcessor cria um novo processador de eventos
func NewNotificationEventProcessor(
	ruleRepo INotificationRuleRepo,
	templateRepo INotificationTemplateRepo,
	historyRepo INotificationHistoryRepo,
	messageQueue MessageQueuePublisher,
) INotificationEventProcessor {
	return &NotificationEventProcessor{
		ruleRepo:      ruleRepo,
		templateRepo:  templateRepo,
		historyRepo:   historyRepo,
		messageQueue:  messageQueue,
		eventHandlers: make(map[NotificationEventType]EventHandler),
	}
}

// ProcessEvent processa um evento de notificação
func (p *NotificationEventProcessor) ProcessEvent(ctx context.Context, event *NotificationEvent) error {
	gl.Log("info", "Processing notification event", "event_id", event.ID, "type", event.EventType)

	// Execute custom event handler if registered
	if handler, exists := p.eventHandlers[event.EventType]; exists {
		if err := handler(ctx, event); err != nil {
			gl.Log("error", "Custom event handler failed", "event_id", event.ID, "error", err)
			return fmt.Errorf("custom event handler failed: %w", err)
		}
	}

	// Get eligible rules for this event
	rules, err := p.GetEligibleRules(ctx, event)
	if err != nil {
		return fmt.Errorf("failed to get eligible rules: %w", err)
	}

	if len(rules) == 0 {
		gl.Log("info", "No eligible rules found for event", "event_id", event.ID, "type", event.EventType)
		return nil
	}

	gl.Log("info", "Found eligible rules", "event_id", event.ID, "rules_count", len(rules))

	// Create notifications for each eligible rule
	notifications, err := p.CreateNotifications(ctx, event, rules)
	if err != nil {
		return fmt.Errorf("failed to create notifications: %w", err)
	}

	gl.Log("info", "Created notifications", "event_id", event.ID, "notifications_count", len(notifications))

	// Publish event to message queue for external processing
	if err := p.PublishToMessageQueue(ctx, event); err != nil {
		gl.Log("error", "Failed to publish event to message queue", "event_id", event.ID, "error", err)
		// Don't return error here - we still want to process notifications locally
	}

	// Mark event as processed
	now := time.Now()
	event.ProcessedAt = &now

	return nil
}

// GetEligibleRules obtém regras elegíveis para um evento
func (p *NotificationEventProcessor) GetEligibleRules(ctx context.Context, event *NotificationEvent) ([]INotificationRule, error) {
	// Determine condition based on event type
	var condition NotificationRuleCondition
	switch event.EventType {
	case NotificationEventTypeJobCompleted:
		condition = NotificationRuleConditionJobCompleted
	case NotificationEventTypeJobFailed:
		condition = NotificationRuleConditionJobFailed
	case NotificationEventTypeJobStarted:
		condition = NotificationRuleConditionJobStarted
	case NotificationEventTypeJobRetried:
		condition = NotificationRuleConditionJobRetried
	case NotificationEventTypeScoreAlert:
		condition = NotificationRuleConditionScoreAlert
	case NotificationEventTypeTimeAlert:
		condition = NotificationRuleConditionTimeAlert
	default:
		return []INotificationRule{}, nil
	}

	// Extract job type from event data
	jobType, _ := event.Data["job_type"].(string)
	if jobType == "" {
		jobType = "UNKNOWN"
	}

	// Get project ID
	var projectID uuid.UUID
	if event.ProjectID != nil {
		projectID = *event.ProjectID
	}

	// Find rules that match the criteria
	rules, err := p.ruleRepo.FindActiveRulesForEvent(ctx, NotificationEventType(condition), jobType, event.UserID, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to find rules: %w", err)
	}

	// Filter rules based on specific conditions
	var eligibleRules []INotificationRule
	for _, rule := range rules {
		if p.isRuleEligible(rule, event) {
			eligibleRules = append(eligibleRules, rule)
		}
	}

	return eligibleRules, nil
}

// isRuleEligible verifica se uma regra é elegível para um evento específico
func (p *NotificationEventProcessor) isRuleEligible(rule INotificationRule, event *NotificationEvent) bool {
	// Check if rule can trigger (cooldown, rate limiting, etc.)
	if !rule.CanTrigger() {
		return false
	}

	// Check job type eligibility
	if jobType, exists := event.Data["job_type"].(string); exists {
		if !rule.IsEligibleForJobType(jobType) {
			return false
		}
	}

	// Check user eligibility
	if !rule.IsEligibleForUser(event.UserID) {
		return false
	}

	// Check project eligibility
	if event.ProjectID != nil && !rule.IsEligibleForProject(*event.ProjectID) {
		return false
	}

	// Check score-based conditions
	if event.EventType == NotificationEventTypeScoreAlert {
		if score, exists := event.Data["score"].(float64); exists {
			if !rule.ShouldTriggerForScore(score) {
				return false
			}
		}
	}

	// Check time-based conditions
	if event.EventType == NotificationEventTypeTimeAlert {
		if duration, exists := event.Data["duration"].(time.Duration); exists {
			if !rule.ShouldTriggerForDuration(duration) {
				return false
			}
		}
	}

	return true
}

// CreateNotifications cria notificações para as regras elegíveis
func (p *NotificationEventProcessor) CreateNotifications(ctx context.Context, event *NotificationEvent, rules []INotificationRule) ([]INotificationHistory, error) {
	var notifications []INotificationHistory

	for _, rule := range rules {
		// Check rate limiting
		if err := p.checkRateLimit(ctx, rule); err != nil {
			gl.Log("warn", "Rule rate limited", "rule_id", rule.GetID(), "error", err)
			continue
		}

		// Get template for this rule
		template, err := p.getTemplateForRule(ctx, rule, event)
		if err != nil {
			gl.Log("error", "Failed to get template for rule", "rule_id", rule.GetID(), "error", err)
			continue
		}

		// Create notifications for each platform in the rule
		platforms := rule.GetPlatforms()
		for _, platform := range platforms {
			// Get target configuration for this platform
			targets, err := p.getTargetsForPlatform(rule, string(platform))
			if err != nil {
				gl.Log("error", "Failed to get targets for platform", "platform", platform, "rule_id", rule.GetID(), "error", err)
				continue
			}

			// Create notification for each target
			for _, target := range targets {
				notification, err := p.createSingleNotification(ctx, event, rule, template, platform, target)
				if err != nil {
					gl.Log("error", "Failed to create notification", "rule_id", rule.GetID(), "platform", platform, "error", err)
					continue
				}

				notifications = append(notifications, notification)
			}
		}
	}

	return notifications, nil
}

// checkRateLimit verifica se a regra excedeu o limite de notificações
func (p *NotificationEventProcessor) checkRateLimit(ctx context.Context, rule INotificationRule) error {
	if rule.GetMaxNotificationsPerHour() <= 0 {
		return nil // No rate limiting
	}

	since := time.Now().Add(-1 * time.Hour)
	count, err := p.historyRepo.CountRecentNotifications(ctx, rule.GetID(), since)
	if err != nil {
		return fmt.Errorf("failed to count recent notifications: %w", err)
	}

	if count >= rule.GetMaxNotificationsPerHour() {
		return fmt.Errorf("rate limit exceeded: %d notifications in the last hour", count)
	}

	return nil
}

// getTemplateForRule obtém o template apropriado para uma regra
func (p *NotificationEventProcessor) getTemplateForRule(ctx context.Context, rule INotificationRule, event *NotificationEvent) (INotificationTemplate, error) {
	// Try to get specific template if rule has one
	if rule.GetTemplateID() != nil {
		template, err := p.templateRepo.FindByID(ctx, *rule.GetTemplateID())
		if err == nil {
			return template, nil
		}
		gl.Log("warn", "Failed to find specific template, falling back to default", "template_id", *rule.GetTemplateID(), "error", err)
	}

	// Get default template based on event type
	var templateType NotificationTemplateType
	switch event.EventType {
	case NotificationEventTypeJobCompleted:
		templateType = NotificationTemplateTypeJobCompleted
	case NotificationEventTypeJobFailed:
		templateType = NotificationTemplateTypeJobFailed
	case NotificationEventTypeJobStarted:
		templateType = NotificationTemplateTypeJobStarted
	case NotificationEventTypeJobRetried:
		templateType = NotificationTemplateTypeJobRetried
	case NotificationEventTypeScoreAlert:
		templateType = NotificationTemplateTypeScoreAlert
	case NotificationEventTypeTimeAlert:
		templateType = NotificationTemplateTypeTimeAlert
	default:
		templateType = NotificationTemplateTypeCustom
	}

	template, err := p.templateRepo.FindDefaultByType(ctx, templateType, "pt-BR")
	if err != nil {
		return nil, fmt.Errorf("failed to find default template: %w", err)
	}

	return template, nil
}

// getTargetsForPlatform obtém os alvos de notificação para uma plataforma
func (p *NotificationEventProcessor) getTargetsForPlatform(rule INotificationRule, platform string) ([]NotificationTarget, error) {
	targetConfig := rule.GetTargetConfig()
	if targetConfig.IsNil() || targetConfig.IsEmpty() {
		return nil, fmt.Errorf("no target configuration found for rule")
	}

	var configMap = make(map[string]interface{})
	for k, v := range targetConfig {
		configMap[k] = v
	}

	platformConfig, ok := configMap[platform]
	if !ok {
		return nil, fmt.Errorf("no configuration found for platform: %s", platform)
	}

	var targets []NotificationTarget
	switch config := platformConfig.(type) {
	case []interface{}:
		for _, target := range config {
			if targetMap, ok := target.(map[string]interface{}); ok {
				targets = append(targets, NotificationTarget{
					ID:   fmt.Sprintf("%v", targetMap["id"]),
					Name: fmt.Sprintf("%v", targetMap["name"]),
					Type: platform,
				})
			}
		}
	case map[string]interface{}:
		targets = append(targets, NotificationTarget{
			ID:   fmt.Sprintf("%v", config["id"]),
			Name: fmt.Sprintf("%v", config["name"]),
			Type: platform,
		})
	}

	return targets, nil
}

// NotificationTarget representa um alvo de notificação
type NotificationTarget struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// createSingleNotification cria uma única notificação
func (p *NotificationEventProcessor) createSingleNotification(
	ctx context.Context,
	event *NotificationEvent,
	rule INotificationRule,
	template INotificationTemplate,
	platform NotificationRulePlatform,
	target NotificationTarget,
) (INotificationHistory, error) {
	// Render template with event data
	subject, err := template.RenderSubject(event.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to render subject: %w", err)
	}

	message, err := template.RenderBody(event.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to render message: %w", err)
	}

	// Create notification history record
	history := NewNotificationHistoryModel().(*NotificationHistory)
	history.RuleID = rule.GetID()
	templateID := template.GetID()
	history.TemplateID = &templateID
	if event.SourceType == "analysis_job" {
		history.AnalysisJobID = &event.SourceID
	}
	history.Platform = NotificationHistoryPlatform(platform)
	history.Subject = subject
	history.Message = message
	history.TargetID = target.ID
	history.TargetName = target.Name
	history.Priority = rule.GetPriority()
	history.Status = NotificationHistoryStatusPending

	// Set platform-specific configuration
	platformConfig := template.GetPlatformConfig(string(platform))
	if len(platformConfig) > 0 {
		history.PlatformConfig = platformConfig
	}

	// Save to database
	createdHistory, err := p.historyRepo.Create(ctx, history)
	if err != nil {
		return nil, fmt.Errorf("failed to save notification history: %w", err)
	}

	return createdHistory, nil
}

// PublishToMessageQueue publica o evento na fila de mensagens
func (p *NotificationEventProcessor) PublishToMessageQueue(ctx context.Context, event *NotificationEvent) error {
	if p.messageQueue == nil {
		return nil // No message queue configured
	}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Publish to GoBE notification queue
	routingKey := fmt.Sprintf("mcp.notification.%s", event.EventType)
	err = p.messageQueue.Publish("gobe.events", routingKey, eventBytes)
	if err != nil {
		return fmt.Errorf("failed to publish to message queue: %w", err)
	}

	gl.Log("info", "Event published to message queue", "event_id", event.ID, "routing_key", routingKey)
	return nil
}

// RegisterEventHandler registra um handler personalizado para um tipo de evento
func (p *NotificationEventProcessor) RegisterEventHandler(eventType NotificationEventType, handler EventHandler) {
	p.eventHandlers[eventType] = handler
	gl.Log("info", "Event handler registered", "event_type", eventType)
}

// UnregisterEventHandler remove um handler de evento
func (p *NotificationEventProcessor) UnregisterEventHandler(eventType NotificationEventType) {
	delete(p.eventHandlers, eventType)
	gl.Log("info", "Event handler unregistered", "event_type", eventType)
}

// Helper functions para criar eventos específicos

func NewJobStatusChangedEvent(jobID, userID uuid.UUID, projectID *uuid.UUID, oldStatus, newStatus string, jobData map[string]interface{}) *NotificationEvent {
	event := &NotificationEvent{
		ID:         uuid.New(),
		EventType:  NotificationEventTypeJobStatusChanged,
		SourceType: "analysis_job",
		SourceID:   jobID,
		UserID:     userID,
		ProjectID:  projectID,
		Data:       make(map[string]interface{}),
		Metadata:   make(map[string]interface{}),
		Timestamp:  time.Now(),
	}

	// Copy job data
	for k, v := range jobData {
		event.Data[k] = v
	}

	// Add status change specific data
	event.Data["old_status"] = oldStatus
	event.Data["new_status"] = newStatus
	event.Data["status_changed_at"] = time.Now()

	return event
}

func NewJobCompletedEvent(jobID, userID uuid.UUID, projectID *uuid.UUID, jobData map[string]interface{}) *NotificationEvent {
	event := NewJobStatusChangedEvent(jobID, userID, projectID, "RUNNING", "COMPLETED", jobData)
	event.EventType = NotificationEventTypeJobCompleted
	return event
}

func NewJobFailedEvent(jobID, userID uuid.UUID, projectID *uuid.UUID, errorMessage string, jobData map[string]interface{}) *NotificationEvent {
	event := NewJobStatusChangedEvent(jobID, userID, projectID, "RUNNING", "FAILED", jobData)
	event.EventType = NotificationEventTypeJobFailed
	event.Data["error_message"] = errorMessage
	return event
}

func NewScoreAlertEvent(jobID, userID uuid.UUID, projectID *uuid.UUID, score float64, threshold float64, jobData map[string]interface{}) *NotificationEvent {
	event := &NotificationEvent{
		ID:         uuid.New(),
		EventType:  NotificationEventTypeScoreAlert,
		SourceType: "analysis_job",
		SourceID:   jobID,
		UserID:     userID,
		ProjectID:  projectID,
		Data:       make(map[string]interface{}),
		Metadata:   make(map[string]interface{}),
		Timestamp:  time.Now(),
	}

	// Copy job data
	for k, v := range jobData {
		event.Data[k] = v
	}

	// Add score alert specific data
	event.Data["score"] = score
	event.Data["threshold"] = threshold
	event.Data["alert_type"] = "score_below_threshold"

	return event
}
