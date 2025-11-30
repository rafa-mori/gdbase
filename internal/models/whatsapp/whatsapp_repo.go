package whatsapp

import (
	"context"
	"fmt"

	svc "github.com/kubex-ecosystem/gdbase/internal/services"
	gl "github.com/kubex-ecosystem/logz"

	"gorm.io/gorm"
)

type IWhatsAppRepo interface {
	Create(ctx context.Context, whatsapp *WhatsAppModel) (*WhatsAppModel, error)
	Update(ctx context.Context, whatsapp *WhatsAppModel) (*WhatsAppModel, error)
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*WhatsAppModel, error)
	FindByPhoneNumber(ctx context.Context, phoneNumber string) (*WhatsAppModel, error)
	FindByPhoneNumberID(ctx context.Context, phoneNumberID string) (*WhatsAppModel, error)
	FindByBusinessID(ctx context.Context, businessID string) ([]*WhatsAppModel, error)
	FindByUserID(ctx context.Context, userID string) ([]*WhatsAppModel, error)
	FindByStatus(ctx context.Context, status WhatsAppStatus) ([]*WhatsAppModel, error)
	FindByIntegrationType(ctx context.Context, integrationType WhatsAppIntegrationType) ([]*WhatsAppModel, error)
	FindActiveIntegrations(ctx context.Context) ([]*WhatsAppModel, error)
	FindBusinessIntegrations(ctx context.Context) ([]*WhatsAppModel, error)
	FindAll(ctx context.Context) ([]*WhatsAppModel, error)
	Count(ctx context.Context) (int64, error)
	UpdateStatus(ctx context.Context, id string, status WhatsAppStatus) error
	UpdateLastActivity(ctx context.Context, id string) error
	FindByTargetTaskID(ctx context.Context, targetTaskID string) ([]*WhatsAppModel, error)
	FindByAppID(ctx context.Context, appID string) ([]*WhatsAppModel, error)
	UpdateAccessToken(ctx context.Context, id string, accessToken string, expiresAt string) error
}

type WhatsAppRepository struct {
	db *gorm.DB
}

func NewWhatsAppRepository(ctx context.Context, dbService *svc.DBServiceImpl) IWhatsAppRepo {
	db, err := svc.GetDB(ctx, dbService)
	if err != nil {
		gl.Log("error", "Failed to get DB from DBService", err)
		return nil
	}
	return &WhatsAppRepository{db: db}
}

func (r *WhatsAppRepository) Create(ctx context.Context, whatsapp *WhatsAppModel) (*WhatsAppModel, error) {
	if err := whatsapp.Validate(); err != nil {
		gl.Log("error", "WhatsApp validation failed", err)
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	whatsapp.Sanitize()

	if err := r.db.WithContext(ctx).Create(whatsapp).Error; err != nil {
		gl.Log("error", "Failed to create WhatsApp integration", err)
		return nil, fmt.Errorf("failed to create WhatsApp integration: %w", err)
	}

	gl.Log("info", "WhatsApp integration created successfully", whatsapp.ID)
	return whatsapp, nil
}

func (r *WhatsAppRepository) Update(ctx context.Context, whatsapp *WhatsAppModel) (*WhatsAppModel, error) {
	if err := whatsapp.Validate(); err != nil {
		gl.Log("error", "WhatsApp validation failed", err)
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	whatsapp.Sanitize()

	if err := r.db.WithContext(ctx).Save(whatsapp).Error; err != nil {
		gl.Log("error", "Failed to update WhatsApp integration", err)
		return nil, fmt.Errorf("failed to update WhatsApp integration: %w", err)
	}

	gl.Log("info", "WhatsApp integration updated successfully", whatsapp.ID)
	return whatsapp, nil
}

func (r *WhatsAppRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&WhatsAppModel{}, "id = ?", id)
	if result.Error != nil {
		gl.Log("error", "Failed to delete WhatsApp integration", result.Error)
		return fmt.Errorf("failed to delete WhatsApp integration: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		gl.Log("warning", "WhatsApp integration not found for deletion", id)
		return fmt.Errorf("WhatsApp integration not found")
	}

	gl.Log("info", "WhatsApp integration deleted successfully", id)
	return nil
}

func (r *WhatsAppRepository) FindByID(ctx context.Context, id string) (*WhatsAppModel, error) {
	var whatsapp WhatsAppModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&whatsapp).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("WhatsApp integration not found")
		}
		gl.Log("error", "Failed to find WhatsApp integration by ID", err)
		return nil, fmt.Errorf("failed to find WhatsApp integration: %w", err)
	}

	return &whatsapp, nil
}

func (r *WhatsAppRepository) FindByPhoneNumber(ctx context.Context, phoneNumber string) (*WhatsAppModel, error) {
	var whatsapp WhatsAppModel
	if err := r.db.WithContext(ctx).Where("phone_number = ?", phoneNumber).First(&whatsapp).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("WhatsApp integration not found for phone number: %s", phoneNumber)
		}
		gl.Log("error", "Failed to find WhatsApp integration by phone number", err)
		return nil, fmt.Errorf("failed to find WhatsApp integration: %w", err)
	}

	return &whatsapp, nil
}

func (r *WhatsAppRepository) FindByPhoneNumberID(ctx context.Context, phoneNumberID string) (*WhatsAppModel, error) {
	var whatsapp WhatsAppModel
	if err := r.db.WithContext(ctx).Where("phone_number_id = ?", phoneNumberID).First(&whatsapp).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("WhatsApp integration not found for phone number ID: %s", phoneNumberID)
		}
		gl.Log("error", "Failed to find WhatsApp integration by phone number ID", err)
		return nil, fmt.Errorf("failed to find WhatsApp integration: %w", err)
	}

	return &whatsapp, nil
}

func (r *WhatsAppRepository) FindByBusinessID(ctx context.Context, businessID string) ([]*WhatsAppModel, error) {
	var whatsapps []*WhatsAppModel
	if err := r.db.WithContext(ctx).Where("whatsapp_business_id = ?", businessID).Find(&whatsapps).Error; err != nil {
		gl.Log("error", "Failed to find WhatsApp integrations by business ID", err)
		return nil, fmt.Errorf("failed to find WhatsApp integrations: %w", err)
	}

	return whatsapps, nil
}

func (r *WhatsAppRepository) FindByUserID(ctx context.Context, userID string) ([]*WhatsAppModel, error) {
	var whatsapps []*WhatsAppModel
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&whatsapps).Error; err != nil {
		gl.Log("error", "Failed to find WhatsApp integrations by user ID", err)
		return nil, fmt.Errorf("failed to find WhatsApp integrations: %w", err)
	}

	return whatsapps, nil
}

func (r *WhatsAppRepository) FindByStatus(ctx context.Context, status WhatsAppStatus) ([]*WhatsAppModel, error) {
	var whatsapps []*WhatsAppModel
	if err := r.db.WithContext(ctx).Where("status = ?", status).Find(&whatsapps).Error; err != nil {
		gl.Log("error", "Failed to find WhatsApp integrations by status", err)
		return nil, fmt.Errorf("failed to find WhatsApp integrations: %w", err)
	}

	return whatsapps, nil
}

func (r *WhatsAppRepository) FindByIntegrationType(ctx context.Context, integrationType WhatsAppIntegrationType) ([]*WhatsAppModel, error) {
	var whatsapps []*WhatsAppModel
	if err := r.db.WithContext(ctx).Where("integration_type = ?", integrationType).Find(&whatsapps).Error; err != nil {
		gl.Log("error", "Failed to find WhatsApp integrations by integration type", err)
		return nil, fmt.Errorf("failed to find WhatsApp integrations: %w", err)
	}

	return whatsapps, nil
}

func (r *WhatsAppRepository) FindActiveIntegrations(ctx context.Context) ([]*WhatsAppModel, error) {
	var whatsapps []*WhatsAppModel
	if err := r.db.WithContext(ctx).Where("status = ?", WhatsAppStatusActive).Find(&whatsapps).Error; err != nil {
		gl.Log("error", "Failed to find active WhatsApp integrations", err)
		return nil, fmt.Errorf("failed to find active WhatsApp integrations: %w", err)
	}

	return whatsapps, nil
}

func (r *WhatsAppRepository) FindBusinessIntegrations(ctx context.Context) ([]*WhatsAppModel, error) {
	var whatsapps []*WhatsAppModel
	if err := r.db.WithContext(ctx).Where("user_type = ?", WhatsAppUserTypeBusiness).Find(&whatsapps).Error; err != nil {
		gl.Log("error", "Failed to find business WhatsApp integrations", err)
		return nil, fmt.Errorf("failed to find business WhatsApp integrations: %w", err)
	}

	return whatsapps, nil
}

func (r *WhatsAppRepository) FindAll(ctx context.Context) ([]*WhatsAppModel, error) {
	var whatsapps []*WhatsAppModel
	if err := r.db.WithContext(ctx).Find(&whatsapps).Error; err != nil {
		gl.Log("error", "Failed to find all WhatsApp integrations", err)
		return nil, fmt.Errorf("failed to find WhatsApp integrations: %w", err)
	}

	return whatsapps, nil
}

func (r *WhatsAppRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&WhatsAppModel{}).Count(&count).Error; err != nil {
		gl.Log("error", "Failed to count WhatsApp integrations", err)
		return 0, fmt.Errorf("failed to count WhatsApp integrations: %w", err)
	}

	return count, nil
}

func (r *WhatsAppRepository) UpdateStatus(ctx context.Context, id string, status WhatsAppStatus) error {
	result := r.db.WithContext(ctx).Model(&WhatsAppModel{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		gl.Log("error", "Failed to update WhatsApp integration status", result.Error)
		return fmt.Errorf("failed to update WhatsApp integration status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("WhatsApp integration not found")
	}

	gl.Log("info", "WhatsApp integration status updated", id, status)
	return nil
}

func (r *WhatsAppRepository) UpdateLastActivity(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Model(&WhatsAppModel{}).Where("id = ?", id).Update("last_activity", "NOW()")
	if result.Error != nil {
		gl.Log("error", "Failed to update WhatsApp integration last activity", result.Error)
		return fmt.Errorf("failed to update WhatsApp integration last activity: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("WhatsApp integration not found")
	}

	return nil
}

func (r *WhatsAppRepository) FindByTargetTaskID(ctx context.Context, targetTaskID string) ([]*WhatsAppModel, error) {
	var whatsapps []*WhatsAppModel
	if err := r.db.WithContext(ctx).Where("target_task_id = ?", targetTaskID).Find(&whatsapps).Error; err != nil {
		gl.Log("error", "Failed to find WhatsApp integrations by target task ID", err)
		return nil, fmt.Errorf("failed to find WhatsApp integrations: %w", err)
	}

	return whatsapps, nil
}

func (r *WhatsAppRepository) FindByAppID(ctx context.Context, appID string) ([]*WhatsAppModel, error) {
	var whatsapps []*WhatsAppModel
	if err := r.db.WithContext(ctx).Where("app_id = ?", appID).Find(&whatsapps).Error; err != nil {
		gl.Log("error", "Failed to find WhatsApp integrations by app ID", err)
		return nil, fmt.Errorf("failed to find WhatsApp integrations: %w", err)
	}

	return whatsapps, nil
}

func (r *WhatsAppRepository) UpdateAccessToken(ctx context.Context, id string, accessToken string, expiresAt string) error {
	updates := map[string]interface{}{
		"access_token":     accessToken,
		"token_expires_at": expiresAt,
		"updated_at":       "NOW()",
	}

	result := r.db.WithContext(ctx).Model(&WhatsAppModel{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		gl.Log("error", "Failed to update WhatsApp integration access token", result.Error)
		return fmt.Errorf("failed to update WhatsApp integration access token: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("WhatsApp integration not found")
	}

	gl.Log("info", "WhatsApp integration access token updated", id)
	return nil
}
