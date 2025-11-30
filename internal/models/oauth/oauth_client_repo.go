package oauth

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	svc "github.com/kubex-ecosystem/gdbase/internal/services"
	gl "github.com/kubex-ecosystem/logz"
	"gorm.io/gorm"
)

// IOAuthClientRepo interface for OAuth client repository operations
type IOAuthClientRepo interface {
	TableName() string
	Create(client IOAuthClient) (IOAuthClient, error)
	FindByID(id string) (IOAuthClient, error)
	FindByClientID(clientID string) (IOAuthClient, error)
	FindAll(where ...interface{}) ([]IOAuthClient, error)
	Update(client IOAuthClient) (IOAuthClient, error)
	Delete(id string) error
	Close() error
}

// OAuthClientRepo implements IOAuthClientRepo
type OAuthClientRepo struct {
	db *gorm.DB
}

// NewOAuthClientRepo creates a new OAuth client repository
func NewOAuthClientRepo(ctx context.Context, dbService *svc.DBServiceImpl) IOAuthClientRepo {
	if dbService == nil {
		gl.Log("error", "OAuthClientRepo: dbService is nil")
		return nil
	}
	db, err := svc.GetDB(ctx, dbService)
	if err != nil {
		gl.Log("error", fmt.Sprintf("OAuthClientRepo: failed to get db: %v", err))
		return nil
	}
	if db == nil {
		gl.Log("error", "OAuthClientRepo: db is nil")
		return nil
	}
	return &OAuthClientRepo{db}
}

func (r *OAuthClientRepo) TableName() string {
	return "oauth_clients"
}

func (r *OAuthClientRepo) Create(client IOAuthClient) (IOAuthClient, error) {
	if client == nil {
		return nil, fmt.Errorf("OAuthClientRepo: client is nil")
	}

	model, ok := client.(*OAuthClientModel)
	if !ok {
		return nil, fmt.Errorf("OAuthClientRepo: client is not of type *OAuthClientModel")
	}

	if model.GetID() == "" {
		model.SetID(uuid.New().String())
	} else {
		if _, err := uuid.Parse(model.GetID()); err != nil {
			return nil, fmt.Errorf("OAuthClientRepo: invalid UUID: %w", err)
		}
	}

	if err := model.Validate(); err != nil {
		return nil, fmt.Errorf("OAuthClientRepo: validation failed: %w", err)
	}

	model.Sanitize()

	if err := r.db.Create(model).Error; err != nil {
		return nil, fmt.Errorf("OAuthClientRepo: failed to create client: %w", err)
	}

	return model, nil
}

func (r *OAuthClientRepo) FindByID(id string) (IOAuthClient, error) {
	if id == "" {
		return nil, fmt.Errorf("OAuthClientRepo: id is required")
	}

	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("OAuthClientRepo: invalid UUID: %w", err)
	}

	var model OAuthClientModel
	if err := r.db.Where("id = ?", id).First(&model).Error; err != nil {
		return nil, fmt.Errorf("OAuthClientRepo: failed to find client: %w", err)
	}

	model.Sanitize()
	return &model, nil
}

func (r *OAuthClientRepo) FindByClientID(clientID string) (IOAuthClient, error) {
	if clientID == "" {
		return nil, fmt.Errorf("OAuthClientRepo: client_id is required")
	}

	var model OAuthClientModel
	if err := r.db.Where("client_id = ?", clientID).First(&model).Error; err != nil {
		return nil, fmt.Errorf("OAuthClientRepo: failed to find client by client_id: %w", err)
	}

	model.Sanitize()
	return &model, nil
}

func (r *OAuthClientRepo) FindAll(where ...interface{}) ([]IOAuthClient, error) {
	var models []OAuthClientModel
	query := r.db

	if len(where) > 0 {
		query = query.Where(where[0], where[1:]...)
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("OAuthClientRepo: failed to find all clients: %w", err)
	}

	clients := make([]IOAuthClient, len(models))
	for i := range models {
		models[i].Sanitize()
		clients[i] = &models[i]
	}

	return clients, nil
}

func (r *OAuthClientRepo) Update(client IOAuthClient) (IOAuthClient, error) {
	if client == nil {
		return nil, fmt.Errorf("OAuthClientRepo: client is nil")
	}

	model, ok := client.(*OAuthClientModel)
	if !ok {
		return nil, fmt.Errorf("OAuthClientRepo: client is not of type *OAuthClientModel")
	}

	if err := model.Validate(); err != nil {
		return nil, fmt.Errorf("OAuthClientRepo: validation failed: %w", err)
	}

	model.Sanitize()

	if err := r.db.Save(model).Error; err != nil {
		return nil, fmt.Errorf("OAuthClientRepo: failed to update client: %w", err)
	}

	return model, nil
}

func (r *OAuthClientRepo) Delete(id string) error {
	if id == "" {
		return fmt.Errorf("OAuthClientRepo: id is required")
	}

	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("OAuthClientRepo: invalid UUID: %w", err)
	}

	if err := r.db.Delete(&OAuthClientModel{}, id).Error; err != nil {
		return fmt.Errorf("OAuthClientRepo: failed to delete client: %w", err)
	}

	return nil
}

func (r *OAuthClientRepo) Close() error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
