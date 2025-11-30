// Package auth provides authentication repository implementations
package auth

import (
	"context"
	"fmt"
	"time"

	svc "github.com/kubex-ecosystem/gdbase/internal/services"
	gl "github.com/kubex-ecosystem/logz"
	"gorm.io/gorm"
)

// ITokenRepo defines the interface for token repository operations
type ITokenRepo interface {
	SetRefreshToken(ctx context.Context, userID string, tokenID string, expiresIn time.Duration) error
	GetRefreshToken(ctx context.Context, tokenID string) (*RefreshTokenModel, error)
	DeleteRefreshToken(ctx context.Context, userID string, tokenID string) error
	DeleteUserRefreshTokens(ctx context.Context, userID string) error
	CleanupExpiredTokens(ctx context.Context) error
}

// TokenRepoImpl implements ITokenRepo using GORM
type TokenRepoImpl struct {
	db *gorm.DB
}

// NewTokenRepo creates a new token repository instance
func NewTokenRepo(ctx context.Context, dbService svc.DBService) ITokenRepo {
	if dbService == nil {
		gl.Log("error", "TokenRepo: DBService cannot be nil")
		return nil
	}
	db, err := svc.GetDB(ctx, dbService.(*svc.DBServiceImpl))
	if err != nil {
		gl.Log("error", fmt.Sprintf("TokenRepo: failed to get DB: %v", err))
		return nil
	}

	if db == nil {
		gl.Log("error", "TokenRepo: DB connection is nil")
		return nil
	}

	// // Auto-migrate the refresh_tokens table
	// // Note: We ignore migration errors since the table might already exist
	// if err := db.AutoMigrate(&RefreshTokenModel{}); err != nil {
	// 	gl.Log("warn", fmt.Sprintf("TokenRepo: auto-migrate warning (table may already exist): %v", err))
	// 	// Don't return nil - continue even if migration has issues
	// 	// The table probably exists from previous runs
	// }

	return &TokenRepoImpl{db: db}
}

// SetRefreshToken stores a new refresh token in the database
func (r *TokenRepoImpl) SetRefreshToken(ctx context.Context, userID string, tokenID string, expiresIn time.Duration) error {
	if r.db == nil {
		return fmt.Errorf("database connection is nil")
	}

	expirationTime := time.Now().Add(expiresIn)
	token := NewRefreshTokenModel(userID, tokenID, expirationTime)

	if err := token.Validate(); err != nil {
		return fmt.Errorf("token validation failed: %w", err)
	}

	if err := r.db.WithContext(ctx).Create(token).Error; err != nil {
		return fmt.Errorf("failed to insert refresh token: %w", err)
	}

	return nil
}

// GetRefreshToken retrieves a refresh token by its token ID
func (r *TokenRepoImpl) GetRefreshToken(ctx context.Context, tokenID string) (*RefreshTokenModel, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var token RefreshTokenModel
	if err := r.db.WithContext(ctx).Where("token_id = ?", tokenID).First(&token).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to fetch refresh token: %w", err)
	}

	return &token, nil
}

// DeleteRefreshToken removes a specific refresh token
func (r *TokenRepoImpl) DeleteRefreshToken(ctx context.Context, userID string, tokenID string) error {
	if r.db == nil {
		return fmt.Errorf("database connection is nil")
	}

	result := r.db.WithContext(ctx).
		Where("user_id = ? AND token_id = ?", userID, tokenID).
		Delete(&RefreshTokenModel{})

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to delete refresh token: %w", result.Error)
	}

	return nil
}

// DeleteUserRefreshTokens removes all refresh tokens for a specific user
func (r *TokenRepoImpl) DeleteUserRefreshTokens(ctx context.Context, userID string) error {
	if r.db == nil {
		return fmt.Errorf("database connection is nil")
	}

	result := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&RefreshTokenModel{})

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to delete user refresh tokens: %w", result.Error)
	}

	return nil
}

// CleanupExpiredTokens removes all expired tokens from the database
func (r *TokenRepoImpl) CleanupExpiredTokens(ctx context.Context) error {
	if r.db == nil {
		return fmt.Errorf("database connection is nil")
	}

	result := r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&RefreshTokenModel{})

	if result.Error != nil {
		return fmt.Errorf("failed to cleanup expired tokens: %w", result.Error)
	}

	return nil
}
