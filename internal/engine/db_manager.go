package engine

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/kubex-ecosystem/gdbase/internal/module/kbx"
	"github.com/kubex-ecosystem/gdbase/internal/types"

	logz "github.com/kubex-ecosystem/logz"
)

// DatabaseManager é o orquestrador das conexões ativas.
type DatabaseManager struct {
	Logger *logz.LoggerZ

	// Mapa de conexões ativas, chaveadas pelo ID da DB.
	Conns map[string]*types.DBConnection

	DefaultID string

	*types.Mutexes

	rootConfig *kbx.RootConfig
}

// NewDatabaseManager cria um manager vazio com logger.

func NewDatabaseManagerType(logger *logz.LoggerZ) *DatabaseManager {
	return &DatabaseManager{
		Logger:  logger,
		Conns:   make(map[string]*types.DBConnection),
		Mutexes: types.NewMutexesType(),
	}
}

// NewDatabaseManager cria um manager vazio com logger.
func NewDatabaseManager(logger *logz.LoggerZ) *DatabaseManager {
	return NewDatabaseManagerType(logger)
}

// LoadOrBootstrap carrega uma config existente ou gera uma default se não existir nada.
// 1) Se cfgPath != "" e existe → carrega
// 2) Senão, usa path padrão; se existir → carrega
// 3) Se não existir → gera default, salva, carrega
func (m *DatabaseManager) LoadOrBootstrap(cfgPath string) (*kbx.RootConfig, error) {
	var err error

	if m.Mutexes == nil {
		m.Mutexes = types.NewMutexesType()
	}

	if cfgPath == "" {
		cfgPath, err = GetDefaultConfigPath()
		if err != nil {
			return nil, err
		}
	}

	// Tenta carregar
	if _, err := os.Stat(cfgPath); errors.Is(err, os.ErrNotExist) {
		// não existe → gerar default
		root := GenerateDefaultPostgresConfig()
		root.FilePath = cfgPath
		if err := SaveRootConfig(root); err != nil {
			return nil, fmt.Errorf("failed to save default config: %w", err)
		}
		return root, nil
	} else if err != nil {
		return nil, err
	}

	root, err := LoadRootConfig(cfgPath)
	if err != nil {
		return nil, err
	}

	m.rootConfig = root
	return root, nil
}

func (m *DatabaseManager) LoadDBConfig(dbConfig *kbx.DBConfig) (*types.DBConnection, error) {
	if m.Mutexes == nil {
		m.Mutexes = types.NewMutexesType()
	}
	if !kbx.DefaultTrue(dbConfig.Enabled) {
		return nil, errors.New("db config is disabled")
	}

	// Validação por tipo
	if v, ok := validatorRegistry[kbx.DBType(dbConfig.Type)]; ok {
		if err := v.Validate(dbConfig); err != nil {
			return nil, fmt.Errorf("db %s failed validation: %w", dbConfig.Name, err)
		}
	}

	// Factory de driver
	factory, ok := driverRegistry[kbx.DBType(dbConfig.Type)]
	if !ok {
		return nil, fmt.Errorf("no driver registered for type=%s, skipping db=%s", dbConfig.Type, dbConfig.Name)
	}

	driver := factory(m.Logger)

	// Setup context com timeout mínimo
	cctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	err := driver.Connect(cctx, dbConfig)
	cancel()
	if err != nil {
		return nil, fmt.Errorf("failed to connect db=%s: %w", dbConfig.Name, err)
	}

	rt := &types.DBConfigRT{
		DBConfig: dbConfig,
		Mapper:   types.NewMapperType(&dbConfig, ""),
		Mutexes:  types.NewMutexesType(),
	}

	conn := &types.DBConnection{
		Config: rt,
		Driver: driver,
	}

	return conn, nil
}

// InitFromRootConfig valida e conecta os bancos dessa
func (m *DatabaseManager) InitFromRootConfig(ctx context.Context, root *kbx.RootConfig) error {
	if m.Mutexes == nil {
		m.Mutexes = types.NewMutexesType()
	}
	if !kbx.DefaultTrue(root.Enabled) {
		return errors.New("root config is disabled")
	}

	m.MuLock()
	defer m.MuUnlock()

	for _, db := range root.Databases {
		if !kbx.DefaultTrue(db.Enabled) {
			continue
		}

		// Validação por tipo
		if v, ok := validatorRegistry[kbx.DBType(db.Type)]; ok {
			if err := v.Validate(db); err != nil {
				logz.Error("db %s failed validation: %v", db.Name, err)
				continue
			}
		}

		// Factory de driver
		factory, ok := driverRegistry[kbx.DBType(db.Type)]
		if !ok {
			logz.Warn("no driver registered for type=%s, skipping db=%s", db.Type, db.Name)
			continue
		}

		driver := factory(logz.NewLogger("GoLife"))

		// Setup context com timeout mínimo
		cctx, cancel := context.WithTimeout(ctx, 15*time.Second)
		err := driver.Connect(cctx, db)
		cancel()
		if err != nil {
			logz.Error("failed to connect db=%s: %v", db.Name, err)
			continue
		}

		rt := &types.DBConfigRT{
			DBConfig: db,
			Mapper:   types.NewMapperType(&db, root.FilePath),
			Mutexes:  types.NewMutexesType(),
		}

		m.Conns[db.ID] = &types.DBConnection{
			Config: rt,
			Driver: driver,
		}

		if kbx.DefaultTrue(db.Enabled) && m.DefaultID == "" {
			m.DefaultID = db.ID
		}
	}

	if len(m.Conns) == 0 {
		return errors.New("no database connections available after init")
	}
	return nil
}

// GetDefault retorna a conexão default.
func (m *DatabaseManager) GetDefault() (*types.DBConnection, bool) {
	if m.Mutexes == nil {
		m.Mutexes = types.NewMutexesType()
	}
	m.MuRLock()
	defer m.MuRUnlock()
	if m.DefaultID != "" {
		return m.Conns[m.DefaultID], true
	}
	conn, ok := m.Conns[m.DefaultID]
	return conn, ok
}

// GetByID retorna a conexão por ID.
func (m *DatabaseManager) GetByID(id string) (*types.DBConnection, bool) {
	if m.Mutexes == nil {
		m.Mutexes = types.NewMutexesType()
	}
	m.MuRLock()
	defer m.MuRUnlock()
	conn, ok := m.Conns[id]
	return conn, ok
}

// HealthCheck verifica se todas as conexões estão vivas.
func (m *DatabaseManager) HealthCheck(ctx context.Context) error {
	if m.Mutexes == nil {
		m.Mutexes = types.NewMutexesType()
	}
	m.MuRLock()
	defer m.MuRUnlock()

	for id, conn := range m.Conns {
		cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		ok := conn.Driver.Ping(cctx)
		cancel()
		if !ok {
			return fmt.Errorf("health check failed for db id=%s", id)
		}
	}
	return nil
}

func (m *DatabaseManager) SecureConn(ctx context.Context, dbName string) (*types.DBConnection, error) {
	if m.Mutexes == nil {
		m.Mutexes = types.NewMutexesType()
	}
	m.MuRLock()
	defer m.MuRUnlock()

	conn, ok := m.Conns[dbName]
	if !ok {
		return nil, fmt.Errorf("no connection found for db name=%s", dbName)
	}

	// Ping + reconexão simples
	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	ok = conn.Driver.Ping(cctx)
	cancel()
	if !ok {
		logz.Info("reconnecting to db name=%s", dbName)
		cctx, cancel := context.WithTimeout(ctx, 15*time.Second)
		err := conn.Driver.Connect(cctx, conn.Config.DBConfig)
		cancel()
		if err != nil {
			return nil, fmt.Errorf("failed to reconnect to db name=%s: %w", dbName, err)
		}
	}

	return conn, nil
}

// Shutdown fecha todas as conexões ativas.
func (m *DatabaseManager) Shutdown(ctx context.Context) error {
	if m.Mutexes == nil {
		m.Mutexes = types.NewMutexesType()
	}
	m.MuLock()
	defer m.MuUnlock()

	for id, conn := range m.Conns {
		if err := conn.Driver.Close(); err != nil {
			logz.Error("failed to close db id=%s: %v", id, err)
		}
	}
	return nil
}
