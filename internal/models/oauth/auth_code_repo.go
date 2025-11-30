package oauth

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	gl "github.com/kubex-ecosystem/logz"
	"gorm.io/gorm"
)

// IAuthCodeRepo interface for authorization code repository operations
type IAuthCodeRepo interface {
	TableName() string
	Create(code IAuthCode) (IAuthCode, error)
	FindByCode(code string) (IAuthCode, error)
	FindByUserID(userID string) ([]IAuthCode, error)
	MarkAsUsed(code string) error
	DeleteExpired() error
	Delete(code string) error
	Close() error
}

// AuthCodeRepo implements IAuthCodeRepo
type AuthCodeRepo struct {
	db *gorm.DB
}

// NewAuthCodeRepo creates a new authorization code repository
func NewAuthCodeRepo(db *gorm.DB) IAuthCodeRepo {
	if db == nil {
		gl.Log("error", "AuthCodeRepo: gorm DB is nil")
		return nil
	}
	return &AuthCodeRepo{db: db}
}

func (r *AuthCodeRepo) TableName() string {
	return "oauth_authorization_codes"
}

func (r *AuthCodeRepo) Create(code IAuthCode) (IAuthCode, error) {
	if code == nil {
		return nil, fmt.Errorf("AuthCodeRepo: code is nil")
	}

	model, ok := code.(*AuthCodeModel)
	if !ok {
		return nil, fmt.Errorf("AuthCodeRepo: code is not of type *AuthCodeModel")
	}

	if model.GetID() == "" {
		model.SetID(uuid.New().String())
	} else {
		if _, err := uuid.Parse(model.GetID()); err != nil {
			return nil, fmt.Errorf("AuthCodeRepo: invalid UUID: %w", err)
		}
	}

	if err := model.Validate(); err != nil {
		return nil, fmt.Errorf("AuthCodeRepo: validation failed: %w", err)
	}

	if err := r.db.Create(model).Error; err != nil {
		return nil, fmt.Errorf("AuthCodeRepo: failed to create code: %w", err)
	}

	return model, nil
}

func (r *AuthCodeRepo) FindByCode(code string) (IAuthCode, error) {
	if code == "" {
		return nil, fmt.Errorf("AuthCodeRepo: code is required")
	}

	var model AuthCodeModel
	if err := r.db.Where("code = ?", code).First(&model).Error; err != nil {
		return nil, fmt.Errorf("AuthCodeRepo: failed to find code: %w", err)
	}

	return &model, nil
}

func (r *AuthCodeRepo) FindByUserID(userID string) ([]IAuthCode, error) {
	if userID == "" {
		return nil, fmt.Errorf("AuthCodeRepo: user_id is required")
	}

	if _, err := uuid.Parse(userID); err != nil {
		return nil, fmt.Errorf("AuthCodeRepo: invalid user_id UUID: %w", err)
	}

	var models []AuthCodeModel
	if err := r.db.Where("user_id = ?", userID).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("AuthCodeRepo: failed to find codes by user_id: %w", err)
	}

	codes := make([]IAuthCode, len(models))
	for i := range models {
		codes[i] = &models[i]
	}

	return codes, nil
}

func (r *AuthCodeRepo) MarkAsUsed(code string) error {
	if code == "" {
		return fmt.Errorf("AuthCodeRepo: code is required")
	}

	result := r.db.Model(&AuthCodeModel{}).
		Where("code = ?", code).
		Update("used", true)

	if result.Error != nil {
		return fmt.Errorf("AuthCodeRepo: failed to mark code as used: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("AuthCodeRepo: code not found")
	}

	return nil
}

func (r *AuthCodeRepo) DeleteExpired() error {
	result := r.db.Where("expires_at < ?", time.Now()).Delete(&AuthCodeModel{})
	if result.Error != nil {
		return fmt.Errorf("AuthCodeRepo: failed to delete expired codes: %w", result.Error)
	}

	gl.Log("info", fmt.Sprintf("AuthCodeRepo: deleted %d expired codes", result.RowsAffected))
	return nil
}

func (r *AuthCodeRepo) Delete(code string) error {
	if code == "" {
		return fmt.Errorf("AuthCodeRepo: code is required")
	}

	if err := r.db.Where("code = ?", code).Delete(&AuthCodeModel{}).Error; err != nil {
		return fmt.Errorf("AuthCodeRepo: failed to delete code: %w", err)
	}

	return nil
}

func (r *AuthCodeRepo) Close() error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
