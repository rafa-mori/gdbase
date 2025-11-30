// Package discord provides the repository for managing Discord integrations in the MCP system.
package discord

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

type IDiscordRepo interface {
	// TableName returns the name of the table in the database.
	TableName() string
	Create(d IDiscordModel) (IDiscordModel, error)
	FindOne(where ...interface{}) (IDiscordModel, error)
	FindAll(where ...interface{}) ([]IDiscordModel, error)
	Update(d IDiscordModel) (IDiscordModel, error)
	Delete(id string) error
	Close() error
	List(where ...interface{}) (xtt.TableDataHandler, error)
	GetContextDBService() *svc.DBServiceImpl
}

type DiscordRepo struct {
	// g is the gorm.DB instance used for database operations.
	g *gorm.DB
}

func NewDiscordRepo(ctx context.Context, dbService *svc.DBServiceImpl) IDiscordRepo {
	if dbService == nil {
		gl.Log("error", "DiscordModel repository: gorm DB is nil")
		return nil
	}
	db, err := svc.GetDB(ctx, dbService)
	if err != nil {
		gl.Log("error", "DiscordModel repository: error getting gorm DB: "+err.Error())
		return nil
	}
	return &DiscordRepo{g: db}
}

func (dr *DiscordRepo) TableName() string {
	return "mcp_discord_integrations"
}

func (dr *DiscordRepo) Create(d IDiscordModel) (IDiscordModel, error) {
	if d == nil {
		return nil, fmt.Errorf("DiscordModel repository: DiscordModel is nil")
	}

	// Get the concrete model object
	dModel, ok := d.(*DiscordModel)
	if !ok {
		return nil, fmt.Errorf("DiscordModel repository: model is not of type *DiscordModel")
	}

	if dModel.GetID() != "" {
		gl.Log("warn", "DiscordModel repository: ID already set, using existing ID: "+dModel.GetID())
	} else {
		dModel.SetID(uuid.New().String())
	}

	// Validate the model before creating
	if err := dModel.Validate(); err != nil {
		return nil, fmt.Errorf("DiscordModel repository: validation error: %w", err)
	}

	// Sanitize the model
	dModel.Sanitize()

	err := dr.g.Create(dModel).Error
	if err != nil {
		return nil, fmt.Errorf("DiscordModel repository: error creating Discord integration: %w", err)
	}
	return dModel, nil
}

func (dr *DiscordRepo) FindOne(where ...interface{}) (IDiscordModel, error) {
	var dm DiscordModel
	err := dr.g.Where(where[0], where[1:]...).First(&dm).Error
	if err != nil {
		return nil, fmt.Errorf("DiscordModel repository: error finding Discord integration: %w", err)
	}
	return &dm, nil
}

func (dr *DiscordRepo) FindAll(where ...interface{}) ([]IDiscordModel, error) {
	var dms []DiscordModel
	err := dr.g.Where(where[0], where[1:]...).Find(&dms).Error
	if err != nil {
		return nil, fmt.Errorf("DiscordModel repository: error finding Discord integrations: %w", err)
	}
	ids := make([]IDiscordModel, len(dms))
	for i, discord := range dms {
		ids[i] = &discord
	}
	return ids, nil
}

func (dr *DiscordRepo) Update(d IDiscordModel) (IDiscordModel, error) {
	if d == nil {
		return nil, fmt.Errorf("DiscordModel repository: DiscordModel is nil")
	}

	dModel, ok := d.(*DiscordModel)
	if !ok {
		return nil, fmt.Errorf("DiscordModel repository: model is not of type *DiscordModel")
	}

	// Validate the model before updating
	if err := dModel.Validate(); err != nil {
		return nil, fmt.Errorf("DiscordModel repository: validation error: %w", err)
	}

	// Sanitize the model
	dModel.Sanitize()

	err := dr.g.Save(dModel).Error
	if err != nil {
		return nil, fmt.Errorf("DiscordModel repository: error updating Discord integration: %w", err)
	}
	return dModel, nil
}

func (dr *DiscordRepo) Delete(id string) error {
	err := dr.g.Delete(&DiscordModel{}, id).Error
	if err != nil {
		return fmt.Errorf("DiscordModel repository: error deleting Discord integration: %w", err)
	}
	return nil
}

func (dr *DiscordRepo) Close() error {
	sqlDB, err := dr.g.DB()
	if err != nil {
		return fmt.Errorf("DiscordModel repository: error getting database instance: %w", err)
	}
	return sqlDB.Close()
}

func (dr *DiscordRepo) List(where ...interface{}) (xtt.TableDataHandler, error) {
	var discords []DiscordModel
	err := dr.g.Where(where[0], where[1:]...).Find(&discords).Error
	if err != nil {
		return nil, fmt.Errorf("DiscordModel repository: error listing Discord integrations: %w", err)
	}

	tableHandlerMap := make([][]string, 0)
	for i, discord := range discords {
		row := []string{
			fmt.Sprintf("%d", i+1),
			discord.GetID(),
			discord.GetDiscordUserID(),
			discord.GetUsername(),
			discord.GetDisplayName(),
			string(discord.GetUserType()),
			string(discord.GetStatus()),
			string(discord.GetIntegrationType()),
			discord.GetGuildID(),
			discord.GetChannelID(),
			discord.GetCreatedAt().Format("2006-01-02 15:04:05"),
			discord.GetUpdatedAt().Format("2006-01-02 15:04:05"),
		}
		tableHandlerMap = append(tableHandlerMap, row)
	}

	return xtt.NewTableHandlerFromRows([]string{
		"#", "ID", "Discord User ID", "Username", "Display Name", "User Type",
		"Status", "Integration Type", "Guild ID", "Channel ID", "Created At", "Updated At",
	}, tableHandlerMap), nil
}

func (dr *DiscordRepo) GetContextDBService() *svc.DBServiceImpl {
	dbService, dbServiceErr := svc.NewDatabaseService(context.Background(), svc.NewDBConfigWithDBConnection(dr.g), l.GetLogger("GdoBase"))
	if dbServiceErr != nil {
		gl.Log("error", "DiscordModel repository: error creating database service: "+dbServiceErr.Error())
		return nil
	}
	return dbService.(*svc.DBServiceImpl)
}
