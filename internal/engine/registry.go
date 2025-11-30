package engine

import (
	"github.com/kubex-ecosystem/gdbase/internal/module/kbx"
	"github.com/kubex-ecosystem/gdbase/internal/types"
	logz "github.com/kubex-ecosystem/logz"
)

// Registries
var (
	driverRegistry    = map[kbx.DBType]func(*logz.LoggerZ) types.Driver{}
	validatorRegistry = map[kbx.DBType]types.Validator{}
	migratorRegistry  = map[kbx.DBType]types.Migrator{}
)

// RegisterDriver registra um factory para um tipo de DB.
func RegisterDriver(t kbx.DBType, factory func(*logz.LoggerZ) types.Driver) {
	driverRegistry[t] = factory
}

// RegisterValidator registra um validador para um tipo de DB.
func RegisterValidator(t kbx.DBType, v types.Validator) {
	validatorRegistry[t] = v
}

// RegisterMigrator registra um migrator para um tipo de DB.
func RegisterMigrator(t kbx.DBType, m types.Migrator) {
	migratorRegistry[t] = m
}

func GetDriver(dbType string) (func(*logz.LoggerZ) types.Driver, bool) {
	f, ok := driverRegistry[kbx.DBType(dbType)]
	return f, ok
}

func GetValidator(t kbx.DBType) (types.Validator, bool) {
	v, ok := validatorRegistry[t]
	return v, ok
}

func GetMigrator(t kbx.DBType) (types.Migrator, bool) {
	m, ok := migratorRegistry[t]
	return m, ok
}
