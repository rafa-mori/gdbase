package llm

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	is "github.com/kubex-ecosystem/gdbase/internal/services"
	gl "github.com/kubex-ecosystem/logz"

	// xtt "github.com/kubex-ecosystem/xtui/types"

	"gorm.io/gorm"
)

type ILLMRepo interface {
	// TableName returns the name of the table in the database.
	TableName() string
	Create(m ILLMModel) (ILLMModel, error)
	FindOne(where ...interface{}) (ILLMModel, error)
	FindAll(where ...interface{}) ([]ILLMModel, error)
	Update(m ILLMModel) (ILLMModel, error)
	Delete(id string) error
	Close() error
// 	List(where ...interface{}) (xtt.TableDataHandler, error)
	GetContextDBService() is.DBService
}

type LLMRepo struct {
	// g is the gorm.DB instance used for database operations.
	g *gorm.DB
}

func NewLLMRepo(db *gorm.DB) ILLMRepo {
	if db == nil {
		gl.Log("error", "LLMModel repository: gorm DB is nil")
		return nil
	}
	return &LLMRepo{db}
}

func (lr *LLMRepo) TableName() string {
	return "mcp_llm_models"
}

func (lr *LLMRepo) Create(m ILLMModel) (ILLMModel, error) {
	if m == nil {
		return nil, fmt.Errorf("LLMModel repository: LLMModel is nil")
	}

	// Get the concrete model object
	lModel, ok := m.(*LLMModel)
	if !ok {
		return nil, fmt.Errorf("LLMModel repository: model is not of type *LLMModel")
	}

	if lModel.GetID() != "" {
		if _, err := uuid.Parse(lModel.GetID()); err != nil {
			return nil, fmt.Errorf("LLMModel repository: LLMModel ID is not a valid UUID: %w", err)
		}
	} else {
		lModel.SetID(uuid.New().String())
	}

	// Validate the model before creating
	if err := lModel.Validate(); err != nil {
		return nil, fmt.Errorf("LLMModel repository: validation failed: %w", err)
	}

	// Sanitize the model
	lModel.Sanitize()

	err := lr.g.Create(lModel).Error
	if err != nil {
		return nil, fmt.Errorf("LLMModel repository: failed to create LLMModel: %w", err)
	}
	return lModel, nil
}

func (lr *LLMRepo) FindOne(where ...interface{}) (ILLMModel, error) {
	var lm LLMModel
	err := lr.g.Where(where[0], where[1:]...).First(&lm).Error
	if err != nil {
		return nil, fmt.Errorf("LLMModel repository: failed to find LLMModel: %w", err)
	}
	return &lm, nil
}

func (lr *LLMRepo) FindAll(where ...interface{}) ([]ILLMModel, error) {
	var lms []LLMModel
	err := lr.g.Where(where[0], where[1:]...).Find(&lms).Error
	if err != nil {
		return nil, fmt.Errorf("LLMModel repository: failed to find all LLM models: %w", err)
	}
	ils := make([]ILLMModel, len(lms))
	for i, llm := range lms {
		ils[i] = &llm
	}
	return ils, nil
}

func (lr *LLMRepo) Update(m ILLMModel) (ILLMModel, error) {
	if m == nil {
		return nil, fmt.Errorf("LLMModel repository: LLMModel is nil")
	}

	lModel, ok := m.(*LLMModel)
	if !ok {
		return nil, fmt.Errorf("LLMModel repository: model is not of type *LLMModel")
	}

	// Validate the model before updating
	if err := lModel.Validate(); err != nil {
		return nil, fmt.Errorf("LLMModel repository: validation failed: %w", err)
	}

	// Sanitize the model
	lModel.Sanitize()

	err := lr.g.Save(lModel).Error
	if err != nil {
		return nil, fmt.Errorf("LLMModel repository: failed to update LLMModel: %w", err)
	}
	return lModel, nil
}

func (lr *LLMRepo) Delete(id string) error {
	err := lr.g.Delete(&LLMModel{}, id).Error
	if err != nil {
		return fmt.Errorf("LLMModel repository: failed to delete LLMModel: %w", err)
	}
	return nil
}

func (lr *LLMRepo) Close() error {
	sqlDB, err := lr.g.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// func (lr *LLMRepo) List(where ...interface{}) (xtt.TableDataHandler, error) {
// 	var models []LLMModel
// 	err := lr.g.Where(where[0], where[1:]...).Find(&models).Error
// 	if err != nil {
// 		return nil, fmt.Errorf("LLMModel repository: failed to list LLM models: %w", err)
// 	}
// 	tableHandlerMap := make([][]string, 0)
// 	for i, model := range models {
// 		tableHandlerMap = append(tableHandlerMap, []string{
// 			fmt.Sprintf("%d", i+1),
// 			model.GetID(),
// 			model.GetProvider(),
// 			model.GetModel(),
// 			fmt.Sprintf("%.2f", model.GetTemperature()),
// 			fmt.Sprintf("%d", model.GetMaxTokens()),
// 			fmt.Sprintf("%.2f", model.GetTopP()),
// 			fmt.Sprintf("%t", model.Enabled),
// 		})
// 	}

// 	return xtt.NewTableHandlerFromRows([]string{"#", "ID", "Provider", "Model", "Temperature", "Max Tokens", "Top P", "Enabled"}, tableHandlerMap), nil
// }

func (lr *LLMRepo) GetContextDBService() is.DBService {
	dbService, dbServiceErr := is.NewDatabaseService(context.Background(), is.NewDBConfigWithDBConnection(lr.g), gl.GetLoggerZ("GdoBase"))
	if dbServiceErr != nil {
		gl.Log("error", fmt.Sprintf("LLMModel repository: failed to get context DB service: %v", dbServiceErr))
		return nil
	}
	return dbService
}
