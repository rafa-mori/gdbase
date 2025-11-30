package backend

import (
	"gorm.io/gorm"
)

// GormBackend is a backend implementation for GORM.
type GormBackend struct {
	db *gorm.DB
}

// NewGormBackend creates a new GormBackend.
func NewGormBackend(db *gorm.DB) *GormBackend {
	return &GormBackend{db: db}
}

// Create creates a new record in the database.
func (b *GormBackend) Create(value any) error {
	return b.db.Create(value).Error
}

// Update updates a record in the database.
func (b *GormBackend) Update(value any) error {
	return b.db.Save(value).Error
}

// Delete deletes a record from the database.
func (b *GormBackend) Delete(id string) error {
	// This is a generic implementation. You might need to adjust it
	// based on your actual data models.
	return b.db.Delete(id).Error
}

// GetAll retrieves all records of a certain type from the database.
func (b *GormBackend) GetAll(out any) error {
	return b.db.Find(out).Error
}

// GetByID retrieves a record by its ID.
func (b *GormBackend) GetByID(id string, out any) error {
	return b.db.First(out, id).Error
}
