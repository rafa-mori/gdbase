// Package backend provides a common interface for different backend implementations.
package backend

import (
	"context"
)

// Backend é o contrato mínimo que ambos os motores (GORM e Mongo)
// precisam cumprir pra coexistir sem refactor nos repos/services.
type Backend interface {
	Create(value any) error
	Update(value any) error
	Delete(id string) error
	GetAll(out any) error
	GetByID(id string, out any) error
}

type ContextualBackend interface {
	Backend
	WithContext(ctx context.Context) Backend
}
