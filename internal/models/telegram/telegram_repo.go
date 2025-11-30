package telegram

import (
	"context"
	"fmt"

	svc "github.com/kubex-ecosystem/gdbase/internal/services"
	gl "github.com/kubex-ecosystem/logz"

	"gorm.io/gorm"
)

type ITelegramRepo interface {
	Create(ctx context.Context, telegram *TelegramModel) (*TelegramModel, error)
	Update(ctx context.Context, telegram *TelegramModel) (*TelegramModel, error)
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*TelegramModel, error)
	FindByTelegramUserID(ctx context.Context, telegramUserID string) (*TelegramModel, error)
	FindByUsername(ctx context.Context, username string) (*TelegramModel, error)
	FindByChatID(ctx context.Context, chatID string) ([]*TelegramModel, error)
	FindByUserID(ctx context.Context, userID string) ([]*TelegramModel, error)
	FindByStatus(ctx context.Context, status TelegramStatus) ([]*TelegramModel, error)
	FindByIntegrationType(ctx context.Context, integrationType TelegramIntegrationType) ([]*TelegramModel, error)
	FindActiveIntegrations(ctx context.Context) ([]*TelegramModel, error)
	FindAll(ctx context.Context) ([]*TelegramModel, error)
	Count(ctx context.Context) (int64, error)
	UpdateStatus(ctx context.Context, id string, status TelegramStatus) error
	UpdateLastActivity(ctx context.Context, id string) error
	FindByTargetTaskID(ctx context.Context, targetTaskID string) ([]*TelegramModel, error)
}

type TelegramRepository struct {
	db *gorm.DB
}

func NewTelegramRepository(ctx context.Context, dbService *svc.DBServiceImpl) ITelegramRepo {
	db, err := svc.GetDB(ctx, dbService)
	if err != nil {
		gl.Log("error", "Failed to get DB from DBService", err)
		return nil
	}
	return &TelegramRepository{db: db}
}

func (r *TelegramRepository) Create(ctx context.Context, telegram *TelegramModel) (*TelegramModel, error) {
	if err := telegram.Validate(); err != nil {
		gl.Log("error", "Telegram validation failed", err)
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	telegram.Sanitize()

	if err := r.db.WithContext(ctx).Create(telegram).Error; err != nil {
		gl.Log("error", "Failed to create telegram integration", err)
		return nil, fmt.Errorf("failed to create telegram integration: %w", err)
	}

	gl.Log("info", "Telegram integration created successfully", telegram.ID)
	return telegram, nil
}

func (r *TelegramRepository) Update(ctx context.Context, telegram *TelegramModel) (*TelegramModel, error) {
	if err := telegram.Validate(); err != nil {
		gl.Log("error", "Telegram validation failed", err)
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	telegram.Sanitize()

	if err := r.db.WithContext(ctx).Save(telegram).Error; err != nil {
		gl.Log("error", "Failed to update telegram integration", err)
		return nil, fmt.Errorf("failed to update telegram integration: %w", err)
	}

	gl.Log("info", "Telegram integration updated successfully", telegram.ID)
	return telegram, nil
}

func (r *TelegramRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&TelegramModel{}, "id = ?", id)
	if result.Error != nil {
		gl.Log("error", "Failed to delete telegram integration", result.Error)
		return fmt.Errorf("failed to delete telegram integration: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		gl.Log("warning", "Telegram integration not found for deletion", id)
		return fmt.Errorf("telegram integration not found")
	}

	gl.Log("info", "Telegram integration deleted successfully", id)
	return nil
}

func (r *TelegramRepository) FindByID(ctx context.Context, id string) (*TelegramModel, error) {
	var telegram TelegramModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&telegram).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("telegram integration not found")
		}
		gl.Log("error", "Failed to find telegram integration by ID", err)
		return nil, fmt.Errorf("failed to find telegram integration: %w", err)
	}

	return &telegram, nil
}

func (r *TelegramRepository) FindByTelegramUserID(ctx context.Context, telegramUserID string) (*TelegramModel, error) {
	var telegram TelegramModel
	if err := r.db.WithContext(ctx).Where("telegram_user_id = ?", telegramUserID).First(&telegram).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("telegram integration not found for user ID: %s", telegramUserID)
		}
		gl.Log("error", "Failed to find telegram integration by user ID", err)
		return nil, fmt.Errorf("failed to find telegram integration: %w", err)
	}

	return &telegram, nil
}

func (r *TelegramRepository) FindByUsername(ctx context.Context, username string) (*TelegramModel, error) {
	var telegram TelegramModel
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&telegram).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("telegram integration not found for username: %s", username)
		}
		gl.Log("error", "Failed to find telegram integration by username", err)
		return nil, fmt.Errorf("failed to find telegram integration: %w", err)
	}

	return &telegram, nil
}

func (r *TelegramRepository) FindByChatID(ctx context.Context, chatID string) ([]*TelegramModel, error) {
	var telegrams []*TelegramModel
	if err := r.db.WithContext(ctx).Where("chat_id = ?", chatID).Find(&telegrams).Error; err != nil {
		gl.Log("error", "Failed to find telegram integrations by chat ID", err)
		return nil, fmt.Errorf("failed to find telegram integrations: %w", err)
	}

	return telegrams, nil
}

func (r *TelegramRepository) FindByUserID(ctx context.Context, userID string) ([]*TelegramModel, error) {
	var telegrams []*TelegramModel
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&telegrams).Error; err != nil {
		gl.Log("error", "Failed to find telegram integrations by user ID", err)
		return nil, fmt.Errorf("failed to find telegram integrations: %w", err)
	}

	return telegrams, nil
}

func (r *TelegramRepository) FindByStatus(ctx context.Context, status TelegramStatus) ([]*TelegramModel, error) {
	var telegrams []*TelegramModel
	if err := r.db.WithContext(ctx).Where("status = ?", status).Find(&telegrams).Error; err != nil {
		gl.Log("error", "Failed to find telegram integrations by status", err)
		return nil, fmt.Errorf("failed to find telegram integrations: %w", err)
	}

	return telegrams, nil
}

func (r *TelegramRepository) FindByIntegrationType(ctx context.Context, integrationType TelegramIntegrationType) ([]*TelegramModel, error) {
	var telegrams []*TelegramModel
	if err := r.db.WithContext(ctx).Where("integration_type = ?", integrationType).Find(&telegrams).Error; err != nil {
		gl.Log("error", "Failed to find telegram integrations by integration type", err)
		return nil, fmt.Errorf("failed to find telegram integrations: %w", err)
	}

	return telegrams, nil
}

func (r *TelegramRepository) FindActiveIntegrations(ctx context.Context) ([]*TelegramModel, error) {
	var telegrams []*TelegramModel
	if err := r.db.WithContext(ctx).Where("status = ?", TelegramStatusActive).Find(&telegrams).Error; err != nil {
		gl.Log("error", "Failed to find active telegram integrations", err)
		return nil, fmt.Errorf("failed to find active telegram integrations: %w", err)
	}

	return telegrams, nil
}

func (r *TelegramRepository) FindAll(ctx context.Context) ([]*TelegramModel, error) {
	var telegrams []*TelegramModel
	if err := r.db.WithContext(ctx).Find(&telegrams).Error; err != nil {
		gl.Log("error", "Failed to find all telegram integrations", err)
		return nil, fmt.Errorf("failed to find telegram integrations: %w", err)
	}

	return telegrams, nil
}

func (r *TelegramRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&TelegramModel{}).Count(&count).Error; err != nil {
		gl.Log("error", "Failed to count telegram integrations", err)
		return 0, fmt.Errorf("failed to count telegram integrations: %w", err)
	}

	return count, nil
}

func (r *TelegramRepository) UpdateStatus(ctx context.Context, id string, status TelegramStatus) error {
	result := r.db.WithContext(ctx).Model(&TelegramModel{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		gl.Log("error", "Failed to update telegram integration status", result.Error)
		return fmt.Errorf("failed to update telegram integration status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("telegram integration not found")
	}

	gl.Log("info", "Telegram integration status updated", id, status)
	return nil
}

func (r *TelegramRepository) UpdateLastActivity(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Model(&TelegramModel{}).Where("id = ?", id).Update("last_activity", "NOW()")
	if result.Error != nil {
		gl.Log("error", "Failed to update telegram integration last activity", result.Error)
		return fmt.Errorf("failed to update telegram integration last activity: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("telegram integration not found")
	}

	return nil
}

func (r *TelegramRepository) FindByTargetTaskID(ctx context.Context, targetTaskID string) ([]*TelegramModel, error) {
	var telegrams []*TelegramModel
	if err := r.db.WithContext(ctx).Where("target_task_id = ?", targetTaskID).Find(&telegrams).Error; err != nil {
		gl.Log("error", "Failed to find telegram integrations by target task ID", err)
		return nil, fmt.Errorf("failed to find telegram integrations: %w", err)
	}

	return telegrams, nil
}
