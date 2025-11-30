package data

import (
	"github.com/kubex-ecosystem/gdbase/internal/services/adaptive"
	"github.com/kubex-ecosystem/gdbase/internal/services/adaptive/backend"
)

// AdaptiveRepo atua como ponte genérica entre múltiplos backends.
// Ele é drop-in compatível com qualquer Repository padrão dos models Kubex.
type AdaptiveRepo struct {
	sql   backend.Backend
	mongo backend.Backend
}

func NewAdaptiveRepo(sql, mongo backend.Backend) *AdaptiveRepo {
	return &AdaptiveRepo{sql: sql, mongo: mongo}
}

func (r *AdaptiveRepo) backendFor(value any) backend.Backend {
	if adaptive.Mongo(value) {
		return r.mongo
	}
	return r.sql
}

func (r *AdaptiveRepo) Create(value any) error {
	return r.backendFor(value).Create(value)
}

func (r *AdaptiveRepo) Update(value any) error {
	return r.backendFor(value).Update(value)
}

func (r *AdaptiveRepo) Delete(id string) error {
	// Em teoria poderia rotear pelo tipo, mas no CRUD genérico basta SQL.
	return r.sql.Delete(id)
}

func (r *AdaptiveRepo) GetAll(out any) error {
	// GetAll raramente depende de tipo de model, pode ficar default.
	return r.sql.GetAll(out)
}

func (r *AdaptiveRepo) GetByID(id string, out any) error {
	return r.sql.GetByID(id, out)
}
