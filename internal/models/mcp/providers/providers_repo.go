// Package providers provides the implementation of the providers service.
package providers

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	svc "github.com/kubex-ecosystem/gdbase/internal/services"
	gl "github.com/kubex-ecosystem/logz"
	l "github.com/kubex-ecosystem/logz"
	xtt "github.com/kubex-ecosystem/xtui/types"

	"gorm.io/gorm"
)

type IProvidersRepo interface {
	// TableName returns the name of the table in the database.
	TableName() string
	Create(p IProvidersModel) (IProvidersModel, error)
	FindOne(where ...interface{}) (IProvidersModel, error)
	FindAll(where ...interface{}) ([]IProvidersModel, error)
	Update(p IProvidersModel) (IProvidersModel, error)
	Delete(id string) error
	Close() error
	List(where ...interface{}) (xtt.TableDataHandler, error)
	GetContextDBService() *svc.DBServiceImpl
}

type ProvidersRepo struct {
	// g is the gorm.DB instance used for database operations.
	g *gorm.DB
}

func NewProvidersRepo(ctx context.Context, dbService *svc.DBServiceImpl) IProvidersRepo {
	if dbService == nil {
		gl.Log("error", "ProvidersModel repository: dbService is nil")
		return nil
	}
	db, err := svc.GetDB(ctx, dbService)
	if err != nil {
		gl.Log("error", fmt.Sprintf("ProvidersModel repository: failed to get DB from dbService: %v", err))
		return nil
	}
	return &ProvidersRepo{db}
}

func (pr *ProvidersRepo) TableName() string {
	return "mcp_provider_configs"
}

func (pr *ProvidersRepo) Create(p IProvidersModel) (IProvidersModel, error) {
	if p == nil {
		return nil, fmt.Errorf("ProvidersModel repository: ProvidersModel is nil")
	}

	// Get the concrete model object
	pModel, ok := p.(*ProvidersModel)
	if !ok {
		return nil, fmt.Errorf("ProvidersModel repository: model is not of type *ProvidersModel")
	}

	if pModel.GetID() != "" {
		if _, err := uuid.Parse(pModel.GetID()); err != nil {
			return nil, fmt.Errorf("ProvidersModel repository: ProvidersModel ID is not a valid UUID: %w", err)
		}
	} else {
		pModel.SetID(uuid.New().String())
	}

	// Validate the model before creating
	if err := pModel.Validate(); err != nil {
		return nil, fmt.Errorf("ProvidersModel repository: validation failed: %w", err)
	}

	// Sanitize the model
	pModel.Sanitize()

	err := pr.g.Create(pModel).Error
	if err != nil {
		return nil, fmt.Errorf("ProvidersModel repository: failed to create ProvidersModel: %w", err)
	}
	return pModel, nil
}

func (pr *ProvidersRepo) FindOne(where ...interface{}) (IProvidersModel, error) {
	var pm ProvidersModel
	err := pr.g.Where(where[0], where[1:]...).First(&pm).Error
	if err != nil {
		return nil, fmt.Errorf("ProvidersModel repository: failed to find ProvidersModel: %w", err)
	}
	return &pm, nil
}

func (pr *ProvidersRepo) FindAll(where ...interface{}) ([]IProvidersModel, error) {
	var pms []ProvidersModel
	err := pr.g.Where(where[0], where[1:]...).Find(&pms).Error
	if err != nil {
		return nil, fmt.Errorf("ProvidersModel repository: failed to find all providers: %w", err)
	}
	ips := make([]IProvidersModel, len(pms))
	for i, provider := range pms {
		ips[i] = &provider
	}
	return ips, nil
}

func (pr *ProvidersRepo) Update(p IProvidersModel) (IProvidersModel, error) {
	if p == nil {
		return nil, fmt.Errorf("ProvidersModel repository: ProvidersModel is nil")
	}

	pModel, ok := p.(*ProvidersModel)
	if !ok {
		return nil, fmt.Errorf("ProvidersModel repository: model is not of type *ProvidersModel")
	}

	// Validate the model before updating
	if err := pModel.Validate(); err != nil {
		return nil, fmt.Errorf("ProvidersModel repository: validation failed: %w", err)
	}

	// Sanitize the model
	pModel.Sanitize()

	err := pr.g.Save(pModel).Error
	if err != nil {
		return nil, fmt.Errorf("ProvidersModel repository: failed to update ProvidersModel: %w", err)
	}
	return pModel, nil
}

func (pr *ProvidersRepo) Delete(id string) error {
	err := pr.g.Delete(&ProvidersModel{}, id).Error
	if err != nil {
		return fmt.Errorf("ProvidersModel repository: failed to delete ProvidersModel: %w", err)
	}
	return nil
}

func (pr *ProvidersRepo) Close() error {
	sqlDB, err := pr.g.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (pr *ProvidersRepo) List(where ...interface{}) (xtt.TableDataHandler, error) {
	var providers []ProvidersModel
	err := pr.g.Where(where[0], where[1:]...).Find(&providers).Error
	if err != nil {
		return nil, fmt.Errorf("ProvidersModel repository: failed to list providers: %w", err)
	}
	tableHandlerMap := make([][]string, 0)
	for i, provider := range providers {
		configSize := "0"
		if provider.Config != nil {
			configSize = fmt.Sprintf("%d", len(provider.Config))
		}
		tableHandlerMap = append(tableHandlerMap, []string{
			fmt.Sprintf("%d", i+1),
			provider.GetID(),
			provider.GetProvider(),
			provider.GetOrgOrGroup(),
			configSize + " bytes",
			provider.GetCreatedAt().Format("2006-01-02 15:04:05"),
			provider.GetUpdatedAt().Format("2006-01-02 15:04:05"),
		})
	}

	return xtt.NewTableHandlerFromRows([]string{"#", "ID", "Provider", "Org/Group", "Config Size", "Created At", "Updated At"}, tableHandlerMap), nil
}

func (pr *ProvidersRepo) GetContextDBService() *svc.DBServiceImpl {
	dbService, dbServiceErr := svc.NewDatabaseService(context.Background(), svc.NewDBConfigWithDBConnection(pr.g), l.GetLogger("GdoBase"))
	if dbServiceErr != nil {
		gl.Log("error", fmt.Sprintf("ProvidersModel repository: failed to get context DB service: %v", dbServiceErr))
		return nil
	}
	return dbService.(*svc.DBServiceImpl)
}
