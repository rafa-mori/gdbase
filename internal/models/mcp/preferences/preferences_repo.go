package preferences

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	svc "github.com/kubex-ecosystem/gdbase/internal/services"
	t "github.com/kubex-ecosystem/gdbase/types"
	gl "github.com/kubex-ecosystem/logz"
	l "github.com/kubex-ecosystem/logz"
	xtt "github.com/kubex-ecosystem/xtui/types"

	"gorm.io/gorm"
)

type IPreferencesRepo interface {
	// TableName returns the name of the table in the database.
	TableName() string
	Create(p IPreferencesModel) (IPreferencesModel, error)
	FindOne(where ...interface{}) (IPreferencesModel, error)
	FindAll(where ...interface{}) ([]IPreferencesModel, error)
	Update(p IPreferencesModel) (IPreferencesModel, error)
	Delete(id string) error
	Close() error
	List(where ...interface{}) (xtt.TableDataHandler, error)
	GetContextDBService() *t.DBServiceImpl
}

type PreferencesRepo struct {
	// g is the gorm.DB instance used for database operations.
	g *gorm.DB
}

func NewPreferencesRepo(ctx context.Context, dbService *svc.DBServiceImpl) IPreferencesRepo {
	if dbService == nil {
		gl.Log("error", "PreferencesModel repository: dbService is nil")
		return nil
	}
	db, err := svc.GetDB(ctx, dbService)
	if err != nil {
		return nil
	}

	return &PreferencesRepo{g: db}
}

func (pr *PreferencesRepo) TableName() string {
	return "mcp_user_preferences"
}

func (pr *PreferencesRepo) Create(p IPreferencesModel) (IPreferencesModel, error) {
	if p == nil {
		return nil, fmt.Errorf("PreferencesModel repository: PreferencesModel is nil")
	}

	// Get the concrete model object
	pModel, ok := p.(*PreferencesModel)
	if !ok {
		return nil, fmt.Errorf("PreferencesModel repository: model is not of type *PreferencesModel")
	}

	if pModel.GetID() != "" {
		if _, err := uuid.Parse(pModel.GetID()); err != nil {
			return nil, fmt.Errorf("PreferencesModel repository: PreferencesModel ID is not a valid UUID: %w", err)
		}
	} else {
		pModel.SetID(uuid.New().String())
	}

	// Validate the model before creating
	if err := pModel.Validate(); err != nil {
		return nil, fmt.Errorf("PreferencesModel repository: validation failed: %w", err)
	}

	// Sanitize the model
	pModel.Sanitize()

	err := pr.g.Create(pModel).Error
	if err != nil {
		return nil, fmt.Errorf("PreferencesModel repository: failed to create PreferencesModel: %w", err)
	}
	return pModel, nil
}

func (pr *PreferencesRepo) FindOne(where ...interface{}) (IPreferencesModel, error) {
	var pm PreferencesModel
	err := pr.g.Where(where[0], where[1:]...).First(&pm).Error
	if err != nil {
		return nil, fmt.Errorf("PreferencesModel repository: failed to find PreferencesModel: %w", err)
	}
	return &pm, nil
}

func (pr *PreferencesRepo) FindAll(where ...interface{}) ([]IPreferencesModel, error) {
	var pms []PreferencesModel
	err := pr.g.Where(where[0], where[1:]...).Find(&pms).Error
	if err != nil {
		return nil, fmt.Errorf("PreferencesModel repository: failed to find all preferences: %w", err)
	}
	ips := make([]IPreferencesModel, len(pms))
	for i, pref := range pms {
		ips[i] = &pref
	}
	return ips, nil
}

func (pr *PreferencesRepo) Update(p IPreferencesModel) (IPreferencesModel, error) {
	if p == nil {
		return nil, fmt.Errorf("PreferencesModel repository: PreferencesModel is nil")
	}

	pModel, ok := p.(*PreferencesModel)
	if !ok {
		return nil, fmt.Errorf("PreferencesModel repository: model is not of type *PreferencesModel")
	}

	// Validate the model before updating
	if err := pModel.Validate(); err != nil {
		return nil, fmt.Errorf("PreferencesModel repository: validation failed: %w", err)
	}

	// Sanitize the model
	pModel.Sanitize()

	err := pr.g.Save(pModel).Error
	if err != nil {
		return nil, fmt.Errorf("PreferencesModel repository: failed to update PreferencesModel: %w", err)
	}
	return pModel, nil
}

func (pr *PreferencesRepo) Delete(id string) error {
	err := pr.g.Delete(&PreferencesModel{}, id).Error
	if err != nil {
		return fmt.Errorf("PreferencesModel repository: failed to delete PreferencesModel: %w", err)
	}
	return nil
}

func (pr *PreferencesRepo) Close() error {
	sqlDB, err := pr.g.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (pr *PreferencesRepo) List(where ...interface{}) (xtt.TableDataHandler, error) {
	var preferences []PreferencesModel
	err := pr.g.Where(where[0], where[1:]...).Find(&preferences).Error
	if err != nil {
		return nil, fmt.Errorf("PreferencesModel repository: failed to list preferences: %w", err)
	}
	tableHandlerMap := make([][]string, 0)
	for i, pref := range preferences {
		configSize := "0"
		if pref.Config != nil {
			configSize = fmt.Sprintf("%d", len(pref.Config))
		}
		tableHandlerMap = append(tableHandlerMap, []string{
			fmt.Sprintf("%d", i+1),
			pref.GetID(),
			pref.GetScope(),
			configSize + " bytes",
			pref.GetCreatedAt().Format("2006-01-02 15:04:05"),
			pref.GetUpdatedAt().Format("2006-01-02 15:04:05"),
		})
	}

	return xtt.NewTableHandlerFromRows([]string{"#", "ID", "Scope", "Config Size", "Created At", "Updated At"}, tableHandlerMap), nil
}

func (pr *PreferencesRepo) GetContextDBService() *t.DBServiceImpl {
	dbService, dbServiceErr := svc.NewDatabaseService(context.Background(), svc.NewDBConfigWithDBConnection(pr.g), l.GetLogger("GdoBase"))
	if dbServiceErr != nil {
		gl.Log("error", fmt.Sprintf("PreferencesModel repository: failed to get context DB service: %v", dbServiceErr))
		return nil
	}
	return dbService.(*t.DBServiceImpl)
}
