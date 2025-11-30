package products

import (
	"context"
	"fmt"

	svc "github.com/kubex-ecosystem/gdbase/internal/services"
	gl "github.com/kubex-ecosystem/logz"

	"gorm.io/gorm"
)

type IProductRepo interface {
	Create(p *Product) (*Product, error)
	FindOne(where ...interface{}) (*Product, error)
	FindAll(where ...interface{}) ([]*Product, error)
	Update(p *Product) (*Product, error)
	Delete(id string) error
	Close() error
}

type ProductRepo struct {
	g *gorm.DB
}

func NewProductRepo(ctx context.Context, dbService *svc.DBServiceImpl) IProductRepo {
	if dbService == nil {
		return nil
	}
	db, err := svc.GetDB(ctx, dbService)
	if err != nil {
		gl.Log("error", fmt.Sprintf("ProductRepo: failed to get DB: %v", err))
		return nil
	}
	return &ProductRepo{db}
}

func (pr *ProductRepo) Create(p *Product) (*Product, error) {
	if p == nil {
		return nil, fmt.Errorf("ProductRepo: Product is nil")
	}
	err := pr.g.Create(p).Error
	if err != nil {
		return nil, fmt.Errorf("ProductRepo: failed to create Product: %w", err)
	}
	return p, nil
}

func (pr *ProductRepo) FindOne(where ...interface{}) (*Product, error) {
	var p Product
	err := pr.g.Where(where[0], where[1:]...).First(&p).Error
	if err != nil {
		return nil, fmt.Errorf("ProductRepo: failed to find Product: %w", err)
	}
	return &p, nil
}

func (pr *ProductRepo) FindAll(where ...interface{}) ([]*Product, error) {
	var ps []*Product
	err := pr.g.Where(where[0], where[1:]...).Find(&ps).Error
	if err != nil {
		return nil, fmt.Errorf("ProductRepo: failed to find all products: %w", err)
	}
	return ps, nil
}

func (pr *ProductRepo) Update(p *Product) (*Product, error) {
	if p == nil {
		return nil, fmt.Errorf("ProductRepo: Product is nil")
	}
	err := pr.g.Save(p).Error
	if err != nil {
		return nil, fmt.Errorf("ProductRepo: failed to update Product: %w", err)
	}
	return p, nil
}

func (pr *ProductRepo) Delete(id string) error {
	err := pr.g.Delete(&Product{}, id).Error
	if err != nil {
		return fmt.Errorf("ProductRepo: failed to delete Product: %w", err)
	}
	return nil
}

func (pr *ProductRepo) Close() error {
	sqlDB, err := pr.g.DB()
	if err != nil {
		gl.Log("error", fmt.Sprintf("ProductRepo: failed to get DB from gorm.DB: %v", err))
		return err
	}
	return sqlDB.Close()
}
