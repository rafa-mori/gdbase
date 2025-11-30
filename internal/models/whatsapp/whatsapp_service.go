package whatsapp

import (
	"context"
	"errors"
	"fmt"
	"time"

	gl "github.com/kubex-ecosystem/logz"
)

type IWhatsAppService interface {
	CreateIntegration(ctx context.Context, whatsapp *WhatsAppModel) (*WhatsAppModel, error)
	UpdateIntegration(ctx context.Context, whatsapp *WhatsAppModel) (*WhatsAppModel, error)
	DeleteIntegration(ctx context.Context, id string) error
	GetIntegrationByID(ctx context.Context, id string) (*WhatsAppModel, error)
	GetIntegrationByPhoneNumber(ctx context.Context, phoneNumber string) (*WhatsAppModel, error)
	GetIntegrationByPhoneNumberID(ctx context.Context, phoneNumberID string) (*WhatsAppModel, error)
	GetIntegrationsByBusinessID(ctx context.Context, businessID string) ([]*WhatsAppModel, error)
	GetIntegrationsByUserID(ctx context.Context, userID string) ([]*WhatsAppModel, error)
	GetIntegrationsByStatus(ctx context.Context, status WhatsAppStatus) ([]*WhatsAppModel, error)
	GetActiveIntegrations(ctx context.Context) ([]*WhatsAppModel, error)
	GetBusinessIntegrations(ctx context.Context) ([]*WhatsAppModel, error)
	GetAllIntegrations(ctx context.Context) ([]*WhatsAppModel, error)
	ActivateIntegration(ctx context.Context, id string) error
	DeactivateIntegration(ctx context.Context, id string) error
	UpdateLastActivity(ctx context.Context, id string) error
	ValidateIntegration(ctx context.Context, whatsapp *WhatsAppModel) error
	GetIntegrationsByTargetTaskID(ctx context.Context, targetTaskID string) ([]*WhatsAppModel, error)
	SetupBusinessAPIIntegration(ctx context.Context, whatsapp *WhatsAppModel, accessToken string, phoneNumberID string) error
	SetupCloudAPIIntegration(ctx context.Context, whatsapp *WhatsAppModel, accessToken string, phoneNumberID string) error
	SetupWebhookIntegration(ctx context.Context, whatsapp *WhatsAppModel, webhookURL string, verifyToken string) error
	SetupGraphAPIIntegration(ctx context.Context, whatsapp *WhatsAppModel, appID string, appSecret string) error
	RefreshAccessToken(ctx context.Context, id string, newToken string, expiresAt time.Time) error
	TestConnection(ctx context.Context, id string) error
	GetSupportedMessageTypes(ctx context.Context, id string) ([]string, error)
	UpdateSupportedMessageTypes(ctx context.Context, id string, messageTypes []string) error
}

type WhatsAppService struct {
	Repo IWhatsAppRepo
}

func NewWhatsAppService(repo IWhatsAppRepo) IWhatsAppService {
	return &WhatsAppService{Repo: repo}
}

func (s *WhatsAppService) CreateIntegration(ctx context.Context, whatsapp *WhatsAppModel) (*WhatsAppModel, error) {
	if whatsapp == nil {
		return nil, errors.New("WhatsApp integration cannot be nil")
	}

	if err := s.ValidateIntegration(ctx, whatsapp); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Check if integration already exists for this phone number
	if whatsapp.PhoneNumber != "" {
		existing, _ := s.Repo.FindByPhoneNumber(ctx, whatsapp.PhoneNumber)
		if existing != nil {
			return nil, fmt.Errorf("integration already exists for phone number: %s", whatsapp.PhoneNumber)
		}
	}

	return s.Repo.Create(ctx, whatsapp)
}

func (s *WhatsAppService) UpdateIntegration(ctx context.Context, whatsapp *WhatsAppModel) (*WhatsAppModel, error) {
	if whatsapp == nil {
		return nil, errors.New("WhatsApp integration cannot be nil")
	}
	if whatsapp.ID == "" {
		return nil, errors.New("WhatsApp integration ID cannot be empty")
	}

	if err := s.ValidateIntegration(ctx, whatsapp); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Check if integration exists
	existing, err := s.Repo.FindByID(ctx, whatsapp.ID)
	if err != nil {
		return nil, fmt.Errorf("integration not found: %w", err)
	}

	// Preserve creation data
	whatsapp.CreatedAt = existing.CreatedAt
	whatsapp.CreatedBy = existing.CreatedBy
	whatsapp.UpdatedAt = time.Now().Format(time.RFC3339)

	return s.Repo.Update(ctx, whatsapp)
}

func (s *WhatsAppService) DeleteIntegration(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("integration ID cannot be empty")
	}
	return s.Repo.Delete(ctx, id)
}

func (s *WhatsAppService) GetIntegrationByID(ctx context.Context, id string) (*WhatsAppModel, error) {
	if id == "" {
		return nil, errors.New("integration ID cannot be empty")
	}
	return s.Repo.FindByID(ctx, id)
}

func (s *WhatsAppService) GetIntegrationByPhoneNumber(ctx context.Context, phoneNumber string) (*WhatsAppModel, error) {
	if phoneNumber == "" {
		return nil, errors.New("phone number cannot be empty")
	}
	return s.Repo.FindByPhoneNumber(ctx, phoneNumber)
}

func (s *WhatsAppService) GetIntegrationByPhoneNumberID(ctx context.Context, phoneNumberID string) (*WhatsAppModel, error) {
	if phoneNumberID == "" {
		return nil, errors.New("phone number ID cannot be empty")
	}
	return s.Repo.FindByPhoneNumberID(ctx, phoneNumberID)
}

func (s *WhatsAppService) GetIntegrationsByBusinessID(ctx context.Context, businessID string) ([]*WhatsAppModel, error) {
	if businessID == "" {
		return nil, errors.New("business ID cannot be empty")
	}
	return s.Repo.FindByBusinessID(ctx, businessID)
}

func (s *WhatsAppService) GetIntegrationsByUserID(ctx context.Context, userID string) ([]*WhatsAppModel, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}
	return s.Repo.FindByUserID(ctx, userID)
}

func (s *WhatsAppService) GetIntegrationsByStatus(ctx context.Context, status WhatsAppStatus) ([]*WhatsAppModel, error) {
	if status == "" {
		return nil, errors.New("status cannot be empty")
	}
	return s.Repo.FindByStatus(ctx, status)
}

func (s *WhatsAppService) GetActiveIntegrations(ctx context.Context) ([]*WhatsAppModel, error) {
	return s.Repo.FindActiveIntegrations(ctx)
}

func (s *WhatsAppService) GetBusinessIntegrations(ctx context.Context) ([]*WhatsAppModel, error) {
	return s.Repo.FindBusinessIntegrations(ctx)
}

func (s *WhatsAppService) GetAllIntegrations(ctx context.Context) ([]*WhatsAppModel, error) {
	return s.Repo.FindAll(ctx)
}

func (s *WhatsAppService) ActivateIntegration(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("integration ID cannot be empty")
	}

	if err := s.Repo.UpdateStatus(ctx, id, WhatsAppStatusActive); err != nil {
		return fmt.Errorf("failed to activate integration: %w", err)
	}

	gl.Log("info", "WhatsApp integration activated", id)
	return nil
}

func (s *WhatsAppService) DeactivateIntegration(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("integration ID cannot be empty")
	}

	if err := s.Repo.UpdateStatus(ctx, id, WhatsAppStatusInactive); err != nil {
		return fmt.Errorf("failed to deactivate integration: %w", err)
	}

	gl.Log("info", "WhatsApp integration deactivated", id)
	return nil
}

func (s *WhatsAppService) UpdateLastActivity(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("integration ID cannot be empty")
	}
	return s.Repo.UpdateLastActivity(ctx, id)
}

func (s *WhatsAppService) ValidateIntegration(ctx context.Context, whatsapp *WhatsAppModel) error {
	if whatsapp.PhoneNumber == "" {
		return errors.New("phone number is required")
	}

	// Validate user type
	validUserTypes := map[WhatsAppUserType]bool{
		WhatsAppUserTypeBusiness: true,
		WhatsAppUserTypeUser:     true,
		WhatsAppUserTypeBot:      true,
		WhatsAppUserTypeSystem:   true,
	}
	if !validUserTypes[whatsapp.UserType] {
		return fmt.Errorf("invalid user type: %s", whatsapp.UserType)
	}

	// Validate status
	validStatuses := map[WhatsAppStatus]bool{
		WhatsAppStatusActive:       true,
		WhatsAppStatusInactive:     true,
		WhatsAppStatusDisconnected: true,
		WhatsAppStatusError:        true,
		WhatsAppStatusBlocked:      true,
		WhatsAppStatusPending:      true,
	}
	if !validStatuses[whatsapp.Status] {
		return fmt.Errorf("invalid status: %s", whatsapp.Status)
	}

	// Validate integration type
	validIntegrationTypes := map[WhatsAppIntegrationType]bool{
		WhatsAppIntegrationTypeBusinessAPI: true,
		WhatsAppIntegrationTypeCloudAPI:    true,
		WhatsAppIntegrationTypeWebhook:     true,
		WhatsAppIntegrationTypeGraph:       true,
	}
	if !validIntegrationTypes[whatsapp.IntegrationType] {
		return fmt.Errorf("invalid integration type: %s", whatsapp.IntegrationType)
	}

	return whatsapp.Validate()
}

func (s *WhatsAppService) GetIntegrationsByTargetTaskID(ctx context.Context, targetTaskID string) ([]*WhatsAppModel, error) {
	if targetTaskID == "" {
		return nil, errors.New("target task ID cannot be empty")
	}
	return s.Repo.FindByTargetTaskID(ctx, targetTaskID)
}

func (s *WhatsAppService) SetupBusinessAPIIntegration(ctx context.Context, whatsapp *WhatsAppModel, accessToken string, phoneNumberID string) error {
	if whatsapp == nil {
		return errors.New("WhatsApp integration cannot be nil")
	}
	if accessToken == "" {
		return errors.New("access token cannot be empty")
	}
	if phoneNumberID == "" {
		return errors.New("phone number ID cannot be empty")
	}

	whatsapp.IntegrationType = WhatsAppIntegrationTypeBusinessAPI
	whatsapp.AccessToken = accessToken
	whatsapp.PhoneNumberID = phoneNumberID
	whatsapp.UserType = WhatsAppUserTypeBusiness
	whatsapp.Status = WhatsAppStatusActive

	if whatsapp.ID == "" {
		_, err := s.CreateIntegration(ctx, whatsapp)
		return err
	} else {
		_, err := s.UpdateIntegration(ctx, whatsapp)
		return err
	}
}

func (s *WhatsAppService) SetupCloudAPIIntegration(ctx context.Context, whatsapp *WhatsAppModel, accessToken string, phoneNumberID string) error {
	if whatsapp == nil {
		return errors.New("WhatsApp integration cannot be nil")
	}
	if accessToken == "" {
		return errors.New("access token cannot be empty")
	}
	if phoneNumberID == "" {
		return errors.New("phone number ID cannot be empty")
	}

	whatsapp.IntegrationType = WhatsAppIntegrationTypeCloudAPI
	whatsapp.AccessToken = accessToken
	whatsapp.PhoneNumberID = phoneNumberID
	whatsapp.Status = WhatsAppStatusActive

	if whatsapp.ID == "" {
		_, err := s.CreateIntegration(ctx, whatsapp)
		return err
	} else {
		_, err := s.UpdateIntegration(ctx, whatsapp)
		return err
	}
}

func (s *WhatsAppService) SetupWebhookIntegration(ctx context.Context, whatsapp *WhatsAppModel, webhookURL string, verifyToken string) error {
	if whatsapp == nil {
		return errors.New("WhatsApp integration cannot be nil")
	}
	if webhookURL == "" {
		return errors.New("webhook URL cannot be empty")
	}
	if verifyToken == "" {
		return errors.New("verify token cannot be empty")
	}

	whatsapp.IntegrationType = WhatsAppIntegrationTypeWebhook
	whatsapp.WebhookURL = webhookURL
	whatsapp.WebhookVerifyToken = verifyToken
	whatsapp.Status = WhatsAppStatusActive

	if whatsapp.ID == "" {
		_, err := s.CreateIntegration(ctx, whatsapp)
		return err
	} else {
		_, err := s.UpdateIntegration(ctx, whatsapp)
		return err
	}
}

func (s *WhatsAppService) SetupGraphAPIIntegration(ctx context.Context, whatsapp *WhatsAppModel, appID string, appSecret string) error {
	if whatsapp == nil {
		return errors.New("WhatsApp integration cannot be nil")
	}
	if appID == "" {
		return errors.New("app ID cannot be empty")
	}
	if appSecret == "" {
		return errors.New("app secret cannot be empty")
	}

	whatsapp.IntegrationType = WhatsAppIntegrationTypeGraph
	whatsapp.AppID = appID
	whatsapp.AppSecret = appSecret
	whatsapp.Status = WhatsAppStatusActive

	if whatsapp.ID == "" {
		_, err := s.CreateIntegration(ctx, whatsapp)
		return err
	} else {
		_, err := s.UpdateIntegration(ctx, whatsapp)
		return err
	}
}

func (s *WhatsAppService) RefreshAccessToken(ctx context.Context, id string, newToken string, expiresAt time.Time) error {
	if id == "" {
		return errors.New("integration ID cannot be empty")
	}
	if newToken == "" {
		return errors.New("new token cannot be empty")
	}

	expiresAtStr := expiresAt.Format(time.RFC3339)
	if err := s.Repo.UpdateAccessToken(ctx, id, newToken, expiresAtStr); err != nil {
		return fmt.Errorf("failed to refresh access token: %w", err)
	}

	gl.Log("info", "WhatsApp integration access token refreshed", id)
	return nil
}

func (s *WhatsAppService) TestConnection(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("integration ID cannot be empty")
	}

	integration, err := s.Repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("integration not found: %w", err)
	}

	// Test connection based on integration type
	switch integration.IntegrationType {
	case WhatsAppIntegrationTypeBusinessAPI, WhatsAppIntegrationTypeCloudAPI:
		if integration.AccessToken == "" {
			return errors.New("access token is required for testing API integration")
		}
		if integration.PhoneNumberID == "" {
			return errors.New("phone number ID is required for testing API integration")
		}
		gl.Log("info", "Testing WhatsApp API connection", integration.ID)

	case WhatsAppIntegrationTypeWebhook:
		if integration.WebhookURL == "" {
			return errors.New("webhook URL is required for testing webhook integration")
		}
		gl.Log("info", "Testing WhatsApp webhook connection", integration.ID)

	case WhatsAppIntegrationTypeGraph:
		if integration.AppID == "" {
			return errors.New("app ID is required for testing Graph API integration")
		}
		if integration.AppSecret == "" {
			return errors.New("app secret is required for testing Graph API integration")
		}
		gl.Log("info", "Testing WhatsApp Graph API connection", integration.ID)

	default:
		return fmt.Errorf("unsupported integration type for testing: %s", integration.IntegrationType)
	}

	// Update last activity on successful test
	if err := s.UpdateLastActivity(ctx, id); err != nil {
		gl.Log("warning", "Failed to update last activity after connection test", err)
	}

	return nil
}

func (s *WhatsAppService) GetSupportedMessageTypes(ctx context.Context, id string) ([]string, error) {
	if id == "" {
		return nil, errors.New("integration ID cannot be empty")
	}

	integration, err := s.Repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("integration not found: %w", err)
	}

	return integration.SupportedMessageTypes, nil
}

func (s *WhatsAppService) UpdateSupportedMessageTypes(ctx context.Context, id string, messageTypes []string) error {
	if id == "" {
		return errors.New("integration ID cannot be empty")
	}
	if len(messageTypes) == 0 {
		return errors.New("message types cannot be empty")
	}

	integration, err := s.Repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("integration not found: %w", err)
	}

	integration.SupportedMessageTypes = messageTypes
	_, err = s.Repo.Update(ctx, integration)
	if err != nil {
		return fmt.Errorf("failed to update supported message types: %w", err)
	}

	gl.Log("info", "WhatsApp integration supported message types updated", id)
	return nil
}
