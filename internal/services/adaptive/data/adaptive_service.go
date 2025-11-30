// Package data provides a service that adapts between SQL and MongoDB backends.
package data

import (
	"reflect"

	"github.com/kubex-ecosystem/gdbase/internal/services/adaptive"
)

// AdaptiveService reflete métodos idênticos entre services SQL e Mongo
// e despacha automaticamente para o backend correto.
type AdaptiveService[T any] struct {
	sql   any
	mongo any
}

func NewAdaptiveService[T any](sqlImpl, mongoImpl any) *AdaptiveService[T] {
	return &AdaptiveService[T]{sql: sqlImpl, mongo: mongoImpl}
}

func (a *AdaptiveService[T]) call(method string, args ...any) []reflect.Value {
	target := reflect.ValueOf(a.sql)
	if len(args) > 0 && adaptive.Mongo(args[0]) {
		target = reflect.ValueOf(a.mongo)
	}
	m := target.MethodByName(method)
	if !m.IsValid() {
		panic("adaptive: method " + method + " not found")
	}
	in := make([]reflect.Value, len(args))
	for i, v := range args {
		in[i] = reflect.ValueOf(v)
	}
	return m.Call(in)
}

// CRUD genérico — compatível com todos os services Canalize

func (a *AdaptiveService[T]) Create(v any) error {
	return a.call("Create", v)[0].Interface().(error)
}
func (a *AdaptiveService[T]) Update(v any) error {
	return a.call("Update", v)[0].Interface().(error)
}
func (a *AdaptiveService[T]) Delete(id string) error {
	return a.call("Delete", id)[0].Interface().(error)
}
func (a *AdaptiveService[T]) GetAll() (any, error) {
	res := a.call("GetAll")
	return res[0].Interface(), res[1].Interface().(error)
}
func (a *AdaptiveService[T]) GetByID(id string) (any, error) {
	res := a.call("GetByID", id)
	return res[0].Interface(), res[1].Interface().(error)
}
