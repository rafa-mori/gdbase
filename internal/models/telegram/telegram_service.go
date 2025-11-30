package telegram

import (
	"context"
	"errors"
	"fmt"
	"time"

	gl "github.com/kubex-ecosystem/logz"
)

type ITelegramService interface {
	CreateIntegration(ctx context.Context, telegram *TelegramModel) (*TelegramModel, error)
	UpdateIntegration(ctx context.Context, telegram *TelegramModel) (*TelegramModel, error)
	DeleteIntegration(ctx context.Context, id string) error
	GetIntegrationByID(ctx context.Context, id string) (*TelegramModel, error)
	GetIntegrationByTelegramUserID(ctx context.Context, telegramUserID string) (*TelegramModel, error)
	GetIntegrationByUsername(ctx context.Context, username string) (*TelegramModel, error)
	GetIntegrationsByChatID(ctx context.Context, chatID string) ([]*TelegramModel, error)
	GetIntegrationsByUserID(ctx context.Context, userID string) ([]*TelegramModel, error)
	GetIntegrationsByStatus(ctx context.Context, status TelegramStatus) ([]*TelegramModel, error)
	GetActiveIntegrations(ctx context.Context) ([]*TelegramModel, error)
	GetAllIntegrations(ctx context.Context) ([]*TelegramModel, error)
	ActivateIntegration(ctx context.Context, id string) error
	DeactivateIntegration(ctx context.Context, id string) error
	UpdateLastActivity(ctx context.Context, id string) error
	ValidateIntegration(ctx context.Context, telegram *TelegramModel) error
	GetIntegrationsByTargetTaskID(ctx context.Context, targetTaskID string) ([]*TelegramModel, error)
	SetupBotIntegration(ctx context.Context, telegram *TelegramModel, botToken string) error
	SetupWebhookIntegration(ctx context.Context, telegram *TelegramModel, webhookURL string) error
	TestConnection(ctx context.Context, id string) error
}

type TelegramService struct {
	Repo ITelegramRepo
}

func NewTelegramService(repo ITelegramRepo) ITelegramService {
	return &TelegramService{Repo: repo}
}

func (s *TelegramService) CreateIntegration(ctx context.Context, telegram *TelegramModel) (*TelegramModel, error) {
	if telegram == nil {
		return nil, errors.New("telegram integration cannot be nil")
	}

	if err := s.ValidateIntegration(ctx, telegram); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Check if integration already exists for this telegram user
	if telegram.TelegramUserID != "" {
		existing, _ := s.Repo.FindByTelegramUserID(ctx, telegram.TelegramUserID)
		if existing != nil {
			return nil, fmt.Errorf("integration already exists for telegram user ID: %s", telegram.TelegramUserID)
		}
	}

	return s.Repo.Create(ctx, telegram)
}

func (s *TelegramService) UpdateIntegration(ctx context.Context, telegram *TelegramModel) (*TelegramModel, error) {
	if telegram == nil {
		return nil, errors.New("telegram integration cannot be nil")
	}
	if telegram.ID == "" {
		return nil, errors.New("telegram integration ID cannot be empty")
	}

	if err := s.ValidateIntegration(ctx, telegram); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Check if integration exists
	existing, err := s.Repo.FindByID(ctx, telegram.ID)
	if err != nil {
		return nil, fmt.Errorf("integration not found: %w", err)
	}

	// Preserve creation data
	telegram.CreatedAt = existing.CreatedAt
	telegram.CreatedBy = existing.CreatedBy
	telegram.UpdatedAt = time.Now().Format(time.RFC3339)

	return s.Repo.Update(ctx, telegram)
}

func (s *TelegramService) DeleteIntegration(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("integration ID cannot be empty")
	}
	return s.Repo.Delete(ctx, id)
}

func (s *TelegramService) GetIntegrationByID(ctx context.Context, id string) (*TelegramModel, error) {
	if id == "" {
		return nil, errors.New("integration ID cannot be empty")
	}
	return s.Repo.FindByID(ctx, id)
}

func (s *TelegramService) GetIntegrationByTelegramUserID(ctx context.Context, telegramUserID string) (*TelegramModel, error) {
	if telegramUserID == "" {
		return nil, errors.New("telegram user ID cannot be empty")
	}
	return s.Repo.FindByTelegramUserID(ctx, telegramUserID)
}

func (s *TelegramService) GetIntegrationByUsername(ctx context.Context, username string) (*TelegramModel, error) {
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}
	return s.Repo.FindByUsername(ctx, username)
}

func (s *TelegramService) GetIntegrationsByChatID(ctx context.Context, chatID string) ([]*TelegramModel, error) {
	if chatID == "" {
		return nil, errors.New("chat ID cannot be empty")
	}
	return s.Repo.FindByChatID(ctx, chatID)
}

func (s *TelegramService) GetIntegrationsByUserID(ctx context.Context, userID string) ([]*TelegramModel, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}
	return s.Repo.FindByUserID(ctx, userID)
}

func (s *TelegramService) GetIntegrationsByStatus(ctx context.Context, status TelegramStatus) ([]*TelegramModel, error) {
	if status == "" {
		return nil, errors.New("status cannot be empty")
	}
	return s.Repo.FindByStatus(ctx, status)
}

func (s *TelegramService) GetActiveIntegrations(ctx context.Context) ([]*TelegramModel, error) {
	return s.Repo.FindActiveIntegrations(ctx)
}

func (s *TelegramService) GetAllIntegrations(ctx context.Context) ([]*TelegramModel, error) {
	return s.Repo.FindAll(ctx)
}

func (s *TelegramService) ActivateIntegration(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("integration ID cannot be empty")
	}

	if err := s.Repo.UpdateStatus(ctx, id, TelegramStatusActive); err != nil {
		return fmt.Errorf("failed to activate integration: %w", err)
	}

	gl.Log("info", "Telegram integration activated", id)
	return nil
}

func (s *TelegramService) DeactivateIntegration(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("integration ID cannot be empty")
	}

	if err := s.Repo.UpdateStatus(ctx, id, TelegramStatusInactive); err != nil {
		return fmt.Errorf("failed to deactivate integration: %w", err)
	}

	gl.Log("info", "Telegram integration deactivated", id)
	return nil
}

func (s *TelegramService) UpdateLastActivity(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("integration ID cannot be empty")
	}
	return s.Repo.UpdateLastActivity(ctx, id)
}

func (s *TelegramService) ValidateIntegration(ctx context.Context, telegram *TelegramModel) error {
	if telegram.TelegramUserID == "" {
		return errors.New("telegram user ID is required")
	}

	if telegram.FirstName == "" {
		return errors.New("first name is required")
	}

	// Validate user type
	validUserTypes := map[TelegramUserType]bool{
		TelegramUserTypeBot:     true,
		TelegramUserTypeUser:    true,
		TelegramUserTypeChannel: true,
		TelegramUserTypeGroup:   true,
		TelegramUserTypeSystem:  true,
	}
	if !validUserTypes[telegram.UserType] {
		return fmt.Errorf("invalid user type: %s", telegram.UserType)
	}

	// Validate status
	validStatuses := map[TelegramStatus]bool{
		TelegramStatusActive:       true,
		TelegramStatusInactive:     true,
		TelegramStatusDisconnected: true,
		TelegramStatusError:        true,
		TelegramStatusBlocked:      true,
	}
	if !validStatuses[telegram.Status] {
		return fmt.Errorf("invalid status: %s", telegram.Status)
	}

	// Validate integration type
	validIntegrationTypes := map[TelegramIntegrationType]bool{
		TelegramIntegrationTypeBot:     true,
		TelegramIntegrationTypeWebhook: true,
		TelegramIntegrationTypeAPI:     true,
	}
	if !validIntegrationTypes[telegram.IntegrationType] {
		return fmt.Errorf("invalid integration type: %s", telegram.IntegrationType)
	}

	// Validate bot token for bot integrations
	if telegram.IntegrationType == TelegramIntegrationTypeBot && telegram.BotToken == "" {
		return errors.New("bot token is required for bot integrations")
	}

	// Validate webhook URL for webhook integrations
	if telegram.IntegrationType == TelegramIntegrationTypeWebhook && telegram.WebhookURL == "" {
		return errors.New("webhook URL is required for webhook integrations")
	}

	return nil
}

func (s *TelegramService) GetIntegrationsByTargetTaskID(ctx context.Context, targetTaskID string) ([]*TelegramModel, error) {
	if targetTaskID == "" {
		return nil, errors.New("target task ID cannot be empty")
	}
	return s.Repo.FindByTargetTaskID(ctx, targetTaskID)
}

func (s *TelegramService) SetupBotIntegration(ctx context.Context, telegram *TelegramModel, botToken string) error {
	if telegram == nil {
		return errors.New("telegram integration cannot be nil")
	}
	if botToken == "" {
		return errors.New("bot token cannot be empty")
	}

	telegram.IntegrationType = TelegramIntegrationTypeBot
	telegram.BotToken = botToken
	telegram.Status = TelegramStatusActive

	if telegram.ID == "" {
		// Create new integration
		_, err := s.CreateIntegration(ctx, telegram)
		return err
	} else {
		// Update existing integration
		_, err := s.UpdateIntegration(ctx, telegram)
		return err
	}
}

func (s *TelegramService) SetupWebhookIntegration(ctx context.Context, telegram *TelegramModel, webhookURL string) error {
	if telegram == nil {
		return errors.New("telegram integration cannot be nil")
	}
	if webhookURL == "" {
		return errors.New("webhook URL cannot be empty")
	}

	telegram.IntegrationType = TelegramIntegrationTypeWebhook
	telegram.WebhookURL = webhookURL
	telegram.Status = TelegramStatusActive

	if telegram.ID == "" {
		// Create new integration
		_, err := s.CreateIntegration(ctx, telegram)
		return err
	} else {
		// Update existing integration
		_, err := s.UpdateIntegration(ctx, telegram)
		return err
	}
}

func (s *TelegramService) TestConnection(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("integration ID cannot be empty")
	}

	integration, err := s.Repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("integration not found: %w", err)
	}

	// Test connection based on integration type
	switch integration.IntegrationType {
	case TelegramIntegrationTypeBot:
		if integration.BotToken == "" {
			return errors.New("bot token is required for testing bot integration")
		}
		// Here you would implement actual Telegram Bot API test
		gl.Log("info", "Testing Telegram bot connection", integration.ID)

	case TelegramIntegrationTypeWebhook:
		if integration.WebhookURL == "" {
			return errors.New("webhook URL is required for testing webhook integration")
		}
		// Here you would implement actual webhook test
		gl.Log("info", "Testing Telegram webhook connection", integration.ID)

	case TelegramIntegrationTypeAPI:
		if integration.APIKey == "" {
			return errors.New("API key is required for testing API integration")
		}
		// Here you would implement actual API test
		gl.Log("info", "Testing Telegram API connection", integration.ID)

	default:
		return fmt.Errorf("unsupported integration type for testing: %s", integration.IntegrationType)
	}

	// Update last activity on successful test
	if err := s.UpdateLastActivity(ctx, id); err != nil {
		gl.Log("warning", "Failed to update last activity after connection test", err)
	}

	return nil
}
