// Package gdbase provides a set of interfaces and types for creating properties, channels, validations, and other components in a Go application.
package gdbase

import (
	. "github.com/kubex-ecosystem/gdbase/internal/interfaces"
	t "github.com/kubex-ecosystem/gdbase/internal/types"
	"github.com/kubex-ecosystem/logz"
)

type PropertyValBase[T any] interface{ IPropertyValBase[T] }
type Property[T any] interface{ IProperty[T] }

func NewProperty[T any](name string, v *T, withMetrics bool, cb func(any) (bool, error)) IProperty[T] {
	return t.NewProperty(name, v, withMetrics, cb)
}

type Channel[T any] interface{ IChannelCtl[T] }
type ChannelBase[T any] interface{ IChannelBase[T] }

func NewChannel[T any](name string, logger *logz.LoggerZ) IChannelCtl[T] {
	return t.NewChannelCtl[T](name, logger)
}
func NewChannelCtlWithProperty[T any, P IProperty[T]](name string, buffers *int, property P, withMetrics bool, logger *logz.LoggerZ) IChannelCtl[T] {
	return t.NewChannelCtlWithProperty[T, P](name, buffers, property, withMetrics, logger)
}
func NewChannelBase[T any](name string, buffers int, logger *logz.LoggerZ) IChannelBase[T] {
	return t.NewChannelBase[T](name, buffers, logger)
}

type Validation[T any] interface{ IValidation[T] }

func NewValidation[T any](name string, v *T, withMetrics bool, cb func(any) (bool, error)) IValidation[T] {
	return t.NewValidation[T]()
}

type ValidationFunc[T any] interface{ IValidationFunc[T] }

func NewValidationFunc[T any](priority int, f func(value *T, args ...any) IValidationResult) IValidationFunc[T] {
	return t.NewValidationFunc[T](priority, f)
}

type ValidationResult interface{ IValidationResult }

func NewValidationResult(isValid bool, message string, metadata map[string]any, err error) IValidationResult {
	return t.NewValidationResult(isValid, message, metadata, err)
}

type Environment interface{ IEnvironment }

func NewEnvironment(envFile string, isConfidential bool, logger *logz.LoggerZ) (IEnvironment, error) {
	return t.NewEnvironment(envFile, isConfidential, logger)
}

type Mapper[T any] interface{ IMapper[T] }

func NewMapper[T any](object *T, filePath string) IMapper[T] {
	return t.NewMapper[T](object, filePath)
}

type Mutexes interface{ IMutexes }

func NewMutexes() IMutexes { return t.NewMutexes() }

func NewMutexesType() Mutexes { return t.NewMutexesType() }

type Reference interface{ IReference }

func NewReference(name string) IReference { return t.NewReference(name) }

type SignalManager[T chan string] interface{ ISignalManager[T] }

func NewSignalManager[T chan string](signalChan T, logger *logz.LoggerZ) ISignalManager[T] {
	return t.NewSignalManager[T](signalChan, logger)
}
