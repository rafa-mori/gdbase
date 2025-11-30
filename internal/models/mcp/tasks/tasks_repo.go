package tasks

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

type ITasksRepo interface {
	// TableName returns the name of the table in the database.
	TableName() string
	Create(task ITasksModel) (ITasksModel, error)
	FindOne(where ...any) (ITasksModel, error)
	FindAll(where ...any) ([]ITasksModel, error)
	Update(task ITasksModel) (ITasksModel, error)
	Delete(id string) error
	Close() error
	List(where ...any) (xtt.TableDataHandler, error)
	GetContextDBService() *svc.DBServiceImpl
}

type TasksRepo struct {
	// g is the gorm.DB instance used for database operations.
	g *gorm.DB
}

func NewTasksRepo(ctx context.Context, dbService *svc.DBServiceImpl) *TasksRepo {
	if dbService == nil {
		gl.Log("error", "TasksModel repository: dbService is nil")
		return nil
	}
	db, err := svc.GetDB(ctx, dbService)
	if err != nil {
		gl.Log("error", fmt.Sprintf("TasksModel repository: failed to get DB from dbService: %v", err))
		return nil
	}
	return &TasksRepo{g: db}
}

func (tr *TasksRepo) TableName() string {
	return "mcp_sync_tasks"
}

func (tr *TasksRepo) Create(task ITasksModel) (ITasksModel, error) {
	if task == nil {
		return nil, fmt.Errorf("TasksModel repository: TasksModel is nil")
	}

	// Get the concrete model object
	tModel, ok := task.(*TasksModel)
	if !ok {
		return nil, fmt.Errorf("TasksModel repository: model is not of type *TasksModel")
	}

	if tModel.GetID() != "" {
		if _, err := uuid.Parse(tModel.GetID()); err != nil {
			return nil, fmt.Errorf("TasksModel repository: TasksModel ID is not a valid UUID: %w", err)
		}
	} else {
		tModel.SetID(uuid.New().String())
	}

	// Validate the model before creating
	if err := tModel.Validate(); err != nil {
		return nil, fmt.Errorf("TasksModel repository: validation failed: %w", err)
	}

	// Sanitize the model
	tModel.Sanitize()

	err := tr.g.Create(tModel).Error
	if err != nil {
		return nil, fmt.Errorf("TasksModel repository: failed to create TasksModel: %w", err)
	}
	return tModel, nil
}

func (tr *TasksRepo) FindOne(where ...any) (ITasksModel, error) {
	var tm TasksModel
	err := tr.g.Where(where[0], where[1:]...).First(&tm).Error
	if err != nil {
		return nil, fmt.Errorf("TasksModel repository: failed to find TasksModel: %w", err)
	}
	return &tm, nil
}

func (tr *TasksRepo) FindAll(where ...any) ([]ITasksModel, error) {
	var err error
	var tms []TasksModel

	if len(where) >= 1 {
		// Só vou jogar considerando que passou pelo serviço, então tá relativamente seguro
		err = tr.g.
			Where(where).
			Find(&tms).
			Error

	} else {
		// Aqui vou deixar assim pra trazer TUDO literalmente caso não seja passado nada de condição
		err = tr.g.Find(&tms).Error
	}

	if err != nil {
		return nil, fmt.Errorf("TasksModel repository: failed to find all tasks: %w", err)
	}
	its := make([]ITasksModel, len(tms))
	for i, task := range tms {
		its[i] = &task
	}
	return its, nil
}

func (tr *TasksRepo) Update(task ITasksModel) (ITasksModel, error) {
	if task == nil {
		return nil, fmt.Errorf("TasksModel repository: TasksModel is nil")
	}

	tModel, ok := task.(*TasksModel)
	if !ok {
		return nil, fmt.Errorf("TasksModel repository: model is not of type *TasksModel")
	}

	// Validate the model before updating
	if err := tModel.Validate(); err != nil {
		return nil, fmt.Errorf("TasksModel repository: validation failed: %w", err)
	}

	// Sanitize the model
	tModel.Sanitize()

	err := tr.g.Save(tModel).Error
	if err != nil {
		return nil, fmt.Errorf("TasksModel repository: failed to update TasksModel: %w", err)
	}
	return tModel, nil
}

func (tr *TasksRepo) Delete(id string) error {
	err := tr.g.Delete(&TasksModel{}, id).Error
	if err != nil {
		return fmt.Errorf("TasksModel repository: failed to delete TasksModel: %w", err)
	}
	return nil
}

func (tr *TasksRepo) Close() error {
	sqlDB, err := tr.g.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (tr *TasksRepo) List(where ...any) (xtt.TableDataHandler, error) {
	var tasks []TasksModel
	err := tr.g.Where(where[0], where[1:]...).Find(&tasks).Error
	if err != nil {
		return nil, fmt.Errorf("TasksModel repository: failed to list tasks: %w", err)
	}
	tableHandlerMap := make([][]string, 0)
	for i, task := range tasks {
		nextRun := "N/A"
		if task.TaskNextRun != nil {
			nextRun = task.TaskNextRun.Format("2006-01-02 15:04:05")
		}
		lastRun := "Never"
		if task.TaskLastRun != nil {
			lastRun = task.TaskLastRun.Format("2006-01-02 15:04:05")
		}

		tableHandlerMap = append(tableHandlerMap, []string{
			fmt.Sprintf("%d", i+1),
			task.GetID(),
			task.GetMCPProvider(),
			task.GetTargetTask(),
			string(task.GetTaskType()),
			task.GetTaskExpression(),
			string(task.TaskStatus),
			fmt.Sprintf("%t", task.GetActive()),
			nextRun,
			lastRun,
			task.TaskLastRunStatus,
		})
	}

	return xtt.NewTableHandlerFromRows([]string{"#", "ID", "Provider", "Target", "Type", "Cron", "Status", "Active", "Next Run", "Last Run", "Last Status"}, tableHandlerMap), nil
}

func (tr *TasksRepo) GetContextDBService() *svc.DBServiceImpl {
	dbService, dbServiceErr := svc.NewDatabaseService(context.Background(), svc.NewDBConfigWithDBConnection(tr.g), l.GetLogger("GdoBase"))
	if dbServiceErr != nil {
		gl.Log("error", fmt.Sprintf("TasksModel repository: failed to get context DB service: %v", dbServiceErr))
		return nil
	}
	return dbService.(*svc.DBServiceImpl)
}
