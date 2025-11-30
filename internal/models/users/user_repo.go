package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	svc "github.com/kubex-ecosystem/gdbase/internal/services"
	gl "github.com/kubex-ecosystem/logz"

	// xtt "github.com/kubex-ecosystem/xtui/types"
	"gorm.io/gorm"
)

type IUserRepo interface {
	// TableName returns the name of the table in the database.
	TableName() string
	Create(u IUser) (IUser, error)
	FindOne(where ...interface{}) (IUser, error)
	FindAll(where ...interface{}) ([]IUser, error)
	Update(u IUser) (IUser, error)
	Delete(id string) error
	Close() error
	// 	List(where ...interface{}) (xtt.TableDataHandler, error)
	GetContextDBService() *svc.DBServiceImpl
}

type UserRepo struct {
	// g is the gorm.DB instance used for database operations.
	g *gorm.DB
}

func NewUserRepo(ctx context.Context, dbService *svc.DBServiceImpl) IUserRepo {
	if dbService == nil {
		gl.Log("error", "UserModel repository: dbService is nil")
		return nil
	}
	db, err := svc.GetDB(ctx, dbService)
	if err != nil {
		gl.Log("error", fmt.Sprintf("UserModel repository: failed to get DB from dbService: %v", err))
		return nil
	}
	return &UserRepo{g: db}
}

func (ur *UserRepo) TableName() string {
	return "users"
}

func (ur *UserRepo) Create(um IUser) (IUser, error) {
	if um == nil {
		return nil, fmt.Errorf("UserModel repository: UserModel is nil")
	}
	if uModel := um.GetUserObj(); uModel == nil {
		return nil, fmt.Errorf("UserModel repository: UserModel is not of type *UserModel")
	} else {
		if uModel.GetID() != "" {
			if _, err := uuid.Parse(uModel.GetID()); err != nil {
				return nil, fmt.Errorf("UserModel repository: UserModel ID is not a valid UUID: %w", err)
			}
		} else {
			uModel.SetID(uuid.New().String())
		}

		err := ur.g.Create(uModel.GetUserObj()).Error
		if err != nil {
			return nil, fmt.Errorf("UserModel repository: failed to create UserModel: %w", err)
		}
		return uModel, nil
	}
}
func (ur *UserRepo) FindOne(where ...interface{}) (IUser, error) {
	var um UserModel

	ur.g.AutoMigrate(&UserModel{})

	err := ur.g.Where(where[0], where[1:]...).First(&um).Error
	if err != nil {
		return nil, fmt.Errorf("UserModel repository: failed to find UserModel: %w", err)
	}
	return &um, nil
}
func (ur *UserRepo) FindAll(where ...interface{}) ([]IUser, error) {
	var ums []UserModel
	err := ur.g.Where(where[0], where[1:]...).Find(&ums).Error
	if err != nil {
		return nil, fmt.Errorf("UserModel repository: failed to find all users: %w", err)
	}
	ius := make([]IUser, len(ums))
	for i, usr := range ums {
		ius[i] = &usr
	}
	return ius, nil
}
func (ur *UserRepo) Update(um IUser) (IUser, error) {
	if um == nil {
		return nil, fmt.Errorf("UserModel repository: UserModel is nil")
	}
	uModel := um.GetUserObj()
	if uModel == nil {
		return nil, fmt.Errorf("UserModel repository: UserModel is not of type *UserModel")
	}
	err := ur.g.Save(uModel).Error
	if err != nil {
		return nil, fmt.Errorf("UserModel repository: failed to update UserModel: %w", err)
	}
	return uModel, nil
}
func (ur *UserRepo) Delete(id string) error {
	err := ur.g.Delete(&UserModel{}, id).Error
	if err != nil {
		return fmt.Errorf("UserModel repository: failed to delete UserModel: %w", err)
	}
	return nil
}
func (ur *UserRepo) Close() error {
	sqlDB, err := ur.g.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// func (ur *UserRepo) List(where ...interface{}) (xtt.TableDataHandler, error) {
// var users []UserModel
// err := ur.g.Where(where[0], where[1:]...).Find(&users).Error
// if err != nil {
// 	return nil, fmt.Errorf("UserModel repository: failed to list users: %w", err)
// }
// tableHandlerMap := make([][]string, 0)
// for i, usr := range users {
// 	tableHandlerMap = append(tableHandlerMap, []string{
// 		fmt.Sprintf("%d", i+1),
// 		usr.GetID(),
// 		usr.GetName(),
// 		usr.GetUsername(),
// 		usr.GetEmail(),
// 		usr.GetPhone(),
// 		fmt.Sprintf("%t", usr.GetActive()),
// 	})
// }

//		return xtt.NewTableHandlerFromRows([]string{"#", "ID", "Name", "Username", "Email", "Phone", "Active"}, tableHandlerMap), nil
//	}
func (ur *UserRepo) GetContextDBService() *svc.DBServiceImpl {
	dbService, dbServiceErr := svc.NewDatabaseService(context.Background(), svc.NewDBConfigWithDBConnection(ur.g), gl.GetLoggerZ("GDBase"))
	if dbServiceErr != nil {
		gl.Log("error", fmt.Sprintf("UserModel repository: failed to get context DB service (%s)", dbServiceErr))
		return nil
	}
	return dbService.(*svc.DBServiceImpl)
}
