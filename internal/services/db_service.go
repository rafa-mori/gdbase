package services

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"os"
	"time"

	// gl "github.com/kubex-ecosystem/gdbase/logger"
	"sync"

	"github.com/kubex-ecosystem/gdbase/internal/module/kbx"
	crp "github.com/kubex-ecosystem/gdbase/internal/security/crypto"
	krs "github.com/kubex-ecosystem/gdbase/internal/security/external"
	"github.com/kubex-ecosystem/logz"

	ci "github.com/kubex-ecosystem/gdbase/internal/interfaces"
	ti "github.com/kubex-ecosystem/gdbase/internal/types"
	gl "github.com/kubex-ecosystem/logz"
	l "github.com/kubex-ecosystem/logz"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

type DirectDatabase interface {
	Query(query string, args ...interface{}) (Rows, error)
}
type Rows interface {
	Next() bool
	Scan(dest ...interface{}) error
	Close() error
	Err() error
}

type DBService interface {
	Initialize(ctx context.Context) error
	InitializeFromEnv(ctx context.Context, env ci.IEnvironment) error
	CloseDBConnection(ctx context.Context) error
	CheckDatabaseHealth(ctx context.Context) error
	// GetConnection(ctx context.Context, timeout time.Duration) (*sql.Conn, error)
	IsConnected(ctx context.Context) error
	IsReady(ctx context.Context) bool
	Reconnect(ctx context.Context) error
	GetName(ctx context.Context) (string, error)
	GetHost(ctx context.Context) (string, error)
	GetConfig(ctx context.Context) IDBConfig
	RunMigrations(ctx context.Context, files map[string]string) (int, int, error)
}
type DBServiceImpl struct {
	Logger    *logz.LoggerZ
	reference ci.IReference
	mutexes   ci.IMutexes

	db   map[string]*gorm.DB
	pool *sync.Pool

	// config holds the database configuration
	config *DBConfig

	// properties are used to store database settings and configurations
	properties map[string]any
}

func NewDatabaseServiceImpl(_ context.Context, config *DBConfig, logger *logz.LoggerZ) (*DBServiceImpl, error) {
	if logger == nil {
		logger = l.GetLogger("GDBase")
	}

	if config == nil {
		return nil, fmt.Errorf("❌ Configuração do banco de dados não pode ser nula")
	}
	if len(config.Databases) == 0 {
		return nil, fmt.Errorf("❌ Configuração de banco de dados não pode ser vazia")
	}

	dbService := &DBServiceImpl{
		Logger:     logger,
		reference:  ti.NewReference("DBServiceImpl"),
		mutexes:    ti.NewMutexesType(),
		properties: make(map[string]any),
		pool:       &sync.Pool{},
		db:         make(map[string]*gorm.DB),
	}
	dbService.config = config
	dbService.properties["config"] = ti.NewProperty("config", &config, true, nil)

	return dbService, nil
}

func NewDatabaseService(ctx context.Context, config *DBConfig, logger *logz.LoggerZ) (DBService, error) {
	return NewDatabaseServiceImpl(ctx, config, logger)
}

func (d *DBServiceImpl) Initialize(ctx context.Context) error {
	if d == nil {
		return fmt.Errorf("❌ Serviço de banco de dados não inicializado")
	}
	if d.db != nil {
		d.db = make(map[string]*gorm.DB)
	}
	config := d.config
	if config == nil {
		cfgT := d.properties["config"].(*ti.Property[*DBConfig])
		if cfgT == nil {
			return fmt.Errorf("❌ Erro ao recuperar a configuração do banco de dados (propriedade nula)")
		}
		config = cfgT.GetValue()
		if config == nil {
			return fmt.Errorf("❌ Erro ao recuperar a configuração do banco de dados (valor nulo)")
		}
	}

	//driver = db.Driver // Pro futuro.. rs
	var dbHost, dbPort, dbUser, dbPass, dbName, dsn string
	for _, dbConfig := range config.Databases {
		if dsn == "" {
			dsn = dbConfig.ConnectionString
		}
		if dsn == "" {
			dbHost = dbConfig.Host
			dbUser = dbConfig.Username
			if dbConfig.Type != "postgresql" {
				if dbConfig.Port == nil {
					dbConfig.Port = "5432" // Padrão PostgreSQL
				}
				dbPort = dbConfig.Port.(string)
				dbPass = dbConfig.Password
				if dbPass == "" {
					dbPassKey, dbPassErr := kbx.GetOrGenPasswordKeyringPass("pgpass")
					if dbPassErr != nil {
						gl.Log("error", fmt.Sprintf("❌ Erro ao recuperar senha do banco de dados: %v", dbPassErr))
						continue
					}
					dbConfig.Password = string(dbPassKey)
					dbPass = dbConfig.Password
				}

				dbName = dbConfig.Name
				dsn = fmt.Sprintf(
					"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=America/Sao_Paulo",
					dbHost, dbPort, dbUser, dbPass, dbName,
				)
			} else {
				// dbPass = dbConfig.Password
				dbName = dbConfig.Name
				dsn = fmt.Sprintf(
					"host=%s port=%s user=%s dbname=%s sslmode=disable TimeZone=America/Sao_Paulo",
					dbHost, dbPort, dbUser, dbName,
				)
			}
			dbConfig.ConnectionString = dsn
			break
		}
	}

	// Conecta (Databases habilitados)
	for _, dbConfig := range config.Databases {
		if dbConfig.Enabled {
			db, _, err := connectDatabase(ctx, dbConfig)
			if err != nil {
				gl.Log("error", fmt.Sprintf("❌ Erro ao conectar ao banco de dados '%s': %v", dbConfig.Name, err))
				continue
			}
			d.db[dbConfig.Name] = db
			continue
		}
	}

	// Conecta (Messagery habilitados)
	if config.Messagery != nil {
		if config.Messagery.RabbitMQ != nil {
			if config.Messagery.RabbitMQ.Enabled {
				// Implementar conexão com RabbitMQ se necessário
				gl.Log("info", "RabbitMQ habilitado")
			}
		}
		if config.Messagery.Redis != nil {
			if config.Messagery.Redis.Enabled {
				// Implementar conexão com Redis se necessário
				gl.Log("info", "Redis habilitado")
			}
		}
	}

	return nil
}

func (d *DBServiceImpl) InitializeFromEnv(ctx context.Context, env ci.IEnvironment) error {
	if d.db != nil {
		return nil
	}
	if env == nil {
		return fmt.Errorf("❌ Serviço de ambiente não pode ser nulo")
	}
	dbType := env.Getenv("DB_TYPE")
	dbHost := env.Getenv("DB_HOST")
	dbPort := env.Getenv("DB_PORT")
	dbUser := env.Getenv("DB_USER")
	dbPass := env.Getenv("DB_PASS")
	dbName := env.Getenv("DB_NAME")
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=America/Sao_Paulo",
		dbHost, dbPort, dbUser, dbPass, dbName,
	)
	databaseConfig := &ti.Database{
		Type:             dbType,
		Host:             dbHost,
		Port:             dbPort,
		Username:         dbUser,
		Password:         dbPass,
		Name:             dbName,
		ConnectionString: dsn,
		Enabled:          true,
	}
	dbConfig := d.properties["config"].(*ti.Property[*DBConfig]).GetValue()

	if dbConfig != nil {
		if _, exists := dbConfig.Databases[databaseConfig.Name]; exists {
			gl.Log("info", fmt.Sprintf("Configuração do banco de dados '%s' já existe, pulando criação", databaseConfig.Name))
			return nil
		}
	}
	if dbConfig == nil {
		dbConfig := &DBConfig{
			Databases: map[string]*ti.Database{
				databaseConfig.Name: databaseConfig,
			},
		}
		d.properties["config"] = ti.NewProperty("config", &dbConfig, true, nil)
	}
	// Aguarda o banco de dados ficar pronto e conecta
	// Timeout de 1 minuto para aguardar o banco de dados ficar pronto
	db, conn, err := waitAndConnect(context.Background(), databaseConfig, 1*time.Minute)
	if err != nil {
		return fmt.Errorf("❌ Erro ao conectar ao banco de dados: %v", err)
	}
	defer conn.Close()
	d.db[databaseConfig.Name] = db
	return nil
}

func (d *DBServiceImpl) CloseDBConnection(ctx context.Context) error {
	db, err := GetDB(ctx, d)
	if err != nil {
		return fmt.Errorf("❌ Erro ao obter banco de dados: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("❌ Erro ao obter conexão SQL: %v", err)
	}
	return sqlDB.Close()
}

func (d *DBServiceImpl) CheckDatabaseHealth(ctx context.Context) error {
	db, err := GetDB(ctx, d)
	if err != nil {
		return fmt.Errorf("❌ Erro ao obter banco de dados: %v", err)
	}
	if err := db.Raw("SELECT 1").Error; err != nil {
		return fmt.Errorf("❌ Banco de dados offline: %v", err)
	}
	return nil
}

func (d *DBServiceImpl) GetConnection(ctx context.Context, timeout time.Duration) (*sql.Conn, error) {
	if d.db == nil {
		return nil, fmt.Errorf("❌ Banco de dados não inicializado")
	}
	// Pega a configuração do banco de dados
	cfgT := d.properties["config"].(*ti.Property[*DBConfig])
	if cfgT == nil {
		return nil, fmt.Errorf("❌ Erro ao recuperar a configuração do banco de dados")
	}
	config := cfgT.GetValue()
	if config == nil {
		return nil, fmt.Errorf("❌ Erro ao recuperar a configuração do banco de dados")
	}
	dbName, ok := d.GetDefaultDBName()
	if !ok {
		return nil, fmt.Errorf("❌ Erro ao recuperar o nome do banco de dados padrão")
	}
	var dbConfig *ti.Database
	for _, dbConf := range config.Databases {
		if dbConf.Name == dbName && dbConf.Enabled {
			dbConfig = dbConf
			break
		}
	}
	if dbConfig == nil {
		return nil, fmt.Errorf("❌ Configuração do banco de dados não encontrada ou desabilitada")
	}

	// Aguarda o banco de dados ficar pronto e retorna a conexão
	// Timeout de 1 minuto para aguardar o banco de dados ficar pronto
	if timeout <= 0 {
		timeout = 1 * time.Minute
	}
	_, conn, err := waitAndConnect(context.Background(), dbConfig, timeout)
	if err != nil {
		return nil, fmt.Errorf("❌ Erro ao conectar ao banco de dados: %v", err)
	}
	return conn, nil
}

func (d *DBServiceImpl) GetDefaultDBName() (string, bool) {
	if d == nil {
		return "", false
	}

	if len(d.config.Databases) > 0 {
		for name, dbConfig := range d.config.Databases {
			if dbConfig.Enabled && dbConfig.IsDefault {
				return name, true
			}
		}
	}

	return "", false
}

func (d *DBServiceImpl) IsConnected(ctx context.Context) error {

	db, err := GetDB(ctx, d)
	if err != nil {
		return fmt.Errorf("❌ Erro ao obter banco de dados: %v", err)
	}
	if err := db.Raw("SELECT 1").Error; err != nil {
		return fmt.Errorf("❌ Banco de dados offline: %v", err)
	}
	return nil
}

// IsReady checks if the database service is fully initialized and ready for use
// Returns true only if all prerequisites are met:
// - DBService is not nil
// - Database map is initialized
// - Configuration is loaded
// - At least one database connection exists
// - Default database is configured and connected
func (d *DBServiceImpl) IsReady(ctx context.Context) bool {
	// Check basic prerequisites
	if d == nil {
		gl.Log("info", "IsReady: DBService is nil")
		return false
	}
	if d.db == nil {
		gl.Log("info", "IsReady: Database map is nil")
		return false
	}
	if d.config == nil {
		gl.Log("info", "IsReady: Config is nil")
		return false
	}

	// Check if we have a default database configured
	dbName, hasDefault := d.GetDefaultDBName()

	// Check if the configured default actually exists in the map
	if hasDefault {
		if _, exists := d.db[dbName]; !exists {
			gl.Log("info", fmt.Sprintf("IsReady: Configured default '%s' not found in map, will use fallback", dbName))
			hasDefault = false
		}
	}

	if !hasDefault {
		gl.Log("info", "IsReady: No default database configured, checking if any database exists")
		// Fallback: if there's only one database, use it
		if len(d.db) == 1 {
			for name := range d.db {
				dbName = name
				hasDefault = true
				gl.Log("info", fmt.Sprintf("IsReady: Using single available database '%s' as default", dbName))
				break
			}
		}
		if !hasDefault {
			gl.Log("info", "IsReady: No database available")
			return false
		}
	}

	// Check if the default database connection exists
	db, exists := d.db[dbName]
	if !exists {
		// Debug: show what keys are available
		keys := make([]string, 0, len(d.db))
		for k := range d.db {
			keys = append(keys, k)
		}
		gl.Log("info", fmt.Sprintf("IsReady: Database '%s' not found. Available databases: %v", dbName, keys))
		return false
	}
	if db == nil {
		gl.Log("info", fmt.Sprintf("IsReady: Database '%s' connection is nil", dbName))
		return false
	}

	// Optional: Verify connection is actually alive
	sqlDB, err := db.DB()
	if err != nil {
		gl.Log("info", fmt.Sprintf("IsReady: Failed to get SQL DB: %v", err))
		return false
	}

	// Quick ping to ensure connection is live
	if err := sqlDB.PingContext(ctx); err != nil {
		gl.Log("info", fmt.Sprintf("IsReady: Ping failed: %v", err))
		return false
	}

	gl.Log("info", fmt.Sprintf("IsReady: ✅ Database '%s' is ready", dbName))
	return true
}

func (d *DBServiceImpl) Reconnect(ctx context.Context) error {
	var err error
	db, err := GetDB(ctx, d)
	if err != nil {
		return fmt.Errorf("❌ Erro ao obter banco de dados: %v", err)
	}
	if db != nil {
		sqlDB, err := db.DB()
		if err != nil {
			return fmt.Errorf("❌ Erro ao obter conexão SQL: %v", err)
		}
		if err := sqlDB.Close(); err != nil {
			return fmt.Errorf("❌ Erro ao fechar conexão SQL: %v", err)
		}
		if err := sqlDB.PingContext(ctx); err == nil {
			// A conexão está ativa, não é necessário reconectar
			return nil
		}
	}
	// Limpa conexões antigas e reconecta
	d.db = make(map[string]*gorm.DB) // Limpa conexões antigas
	return d.Initialize(ctx)
}

func (d *DBServiceImpl) GetName(ctx context.Context) (string, error) {
	if d.db == nil {
		return "", fmt.Errorf("❌ Banco de dados não inicializado")
	}
	name, ok := d.properties["name"].(*ti.Property[string])
	if !ok {
		return "", fmt.Errorf("❌ Erro ao obter nome do banco de dados")
	}
	vl := name.GetValue()
	if vl == "" {
		return "", fmt.Errorf("❌ Nome do banco de dados não encontrado")
	}
	return vl, nil
}

func (d *DBServiceImpl) GetHost(ctx context.Context) (string, error) {
	if d.db == nil {
		return "", fmt.Errorf("❌ Banco de dados não inicializado")
	}
	host, ok := d.properties["host"].(*ti.Property[string])
	if !ok {
		return "", fmt.Errorf("❌ Erro ao obter host do banco de dados")
	}
	vl := host.GetValue()
	if vl == "" {
		return "", fmt.Errorf("❌ Host do banco de dados não encontrado")
	}
	return vl, nil
}

func (d *DBServiceImpl) GetConfig(ctx context.Context) IDBConfig {
	if d == nil {
		return nil
	}
	return d.config
}

func (d *DBServiceImpl) RunMigrations(ctx context.Context, files map[string]string) (int, int, error) {
	if d.db == nil {
		return 0, 0, fmt.Errorf("❌ Banco de dados não inicializado")
	}
	dbName, ok := d.GetDefaultDBName()
	if !ok {
		return 0, 0, fmt.Errorf("❌ Banco de dados padrão não encontrado")
	}
	sqlDB, err := d.db[dbName].DB()
	if err != nil {
		return 0, 0, fmt.Errorf("❌ Erro ao obter conexão SQL: %v", err)
	}
	conn, err := sqlDB.Conn(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("❌ Erro ao obter conexão SQL: %v", err)
	}
	defer conn.Close()

	return 0, 0, nil

}

func (d *DBServiceImpl) GetProperties(ctx context.Context) map[string]any {
	return d.properties
}

func (d *DBServiceImpl) Query(ctx context.Context, query string, args ...interface{}) (any, error) {
	db, err := GetDB(ctx, d)
	if err != nil {
		return nil, fmt.Errorf("❌ Erro ao obter banco de dados: %v", err)
	}
	if db == nil {
		return nil, fmt.Errorf("❌ Banco de dados não inicializado")
	}
	dbSQL, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("❌ Erro ao executar consulta: %v", err)
	}
	rows, err := dbSQL.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("❌ Erro ao executar consulta: %v", err)
	}
	return rows, nil
}

// connectDatabase conecta ao banco de dados e retorna a instância do GORM
// Retorna também um booleano indicando se a conexão foi bem-sucedida
// e um erro caso ocorra algum problema durante a conexão
// Retorna:
// - *gorm.DB: instância do GORM conectada ao banco de dados
// - bool: Indica se é uma conexão válida
// - error: erro caso ocorra algum problema durante a conexão
func connectDatabase(_ context.Context, config *ti.Database) (*gorm.DB, bool, error) {
	// var dialector *sql.DB
	var dialector *sql.DB
	var err error
	// Abre a conexão SQL padrão
	switch config.Type {
	case "mysql":
		dialector, err = sql.Open("mysql", GetConnectionString(config))
	case "postgres", "postgresql":
		dialector, err = sql.Open("postgres", GetConnectionString(config))
	case "sqlite":
		dialector, err = sql.Open("sqlite", GetConnectionString(config))
	case "mariadb":
		dialector, err = sql.Open("mariadb", GetConnectionString(config))
	case "sqlserver":
		dialector, err = sql.Open("sqlserver", GetConnectionString(config))
	case "oracle":
		// dialector = oracle.Open(dsn) // Implementar quando necessário
		return nil, false, fmt.Errorf("banco de dados Oracle não suportado no momento")
	case "mongodb":
		return nil, false, fmt.Errorf("banco de dados MongoDB não suportado pelo GORM")
	case "redis":
		return nil, false, fmt.Errorf("banco de dados Redis não suportado pelo GORM")
	case "rabbitmq":
		return nil, false, fmt.Errorf("RabbitMQ não é um banco de dados suportado pelo GORM")
	default:
		return nil, false, fmt.Errorf("banco de dados não suportado: %s", config.Type)
	}
	if err != nil {
		return nil, true, fmt.Errorf("erro ao abrir conexão SQL: %v", err)
	}

	var gormDialector gorm.Dialector
	switch config.Type {
	case "mysql":
		gormDialector = mysql.New(mysql.Config{
			Conn:                      dialector,
			SkipInitializeWithVersion: false,
		})
	case "postgres", "postgresql":
		gormDialector = postgres.New(postgres.Config{
			Conn:                 dialector,
			PreferSimpleProtocol: true, // Recomendado para evitar problemas com tipos complexos
		})
	case "sqlite":
		gormDialector = sqlite.New(sqlite.Config{
			Conn: dialector,
		})
	case "mariadb":
		gormDialector = mysql.New(mysql.Config{
			Conn:                      dialector,
			SkipInitializeWithVersion: false,
		})
	case "sqlserver":
		gormDialector = sqlserver.New(sqlserver.Config{
			Conn: dialector,
		})
	default:
		return nil, false, fmt.Errorf("banco de dados não suportado: %s", config.Type)
	}

	db, err := gorm.Open(gormDialector, &gorm.Config{})
	if err != nil {
		return nil, true, fmt.Errorf("❌ Erro ao conectar ao banco de dados: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, true, fmt.Errorf("❌ Erro ao obter conexão SQL: %v", err)
	}

	// Testa a conexão
	if err := sqlDB.Ping(); err != nil {
		return nil, true, fmt.Errorf("❌ Erro ao pingar o banco de dados: %v", err)
	}

	return db, true, nil
}

// waitAndConnect aguarda PostgreSQL estar pronto e retorna conexão
func waitAndConnect(ctx context.Context, cfg *ti.Database, maxWait time.Duration) (*gorm.DB, *sql.Conn, error) {
	// Configuração inteligente de retry
	baseRetryInterval := 2 * time.Second // Base menor, mais responsivo
	maxRetryInterval := 10 * time.Second // Limite máximo
	maxAttempts := 5                     // Padrão conservador

	if maxWait > 0 {
		// Calcula tentativas baseado no tempo total disponível
		maxAttempts = int(maxWait / baseRetryInterval)
		if maxAttempts < 3 {
			maxAttempts = 3 // Mínimo de 3 tentativas
		}
		if maxAttempts > 15 {
			maxAttempts = 15 // Máximo de 15 tentativas
		}
	}

	gl.Log("debug", fmt.Sprintf("⏳ Aguardando PostgreSQL ficar pronto (até %d tentativas)...", maxAttempts))

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		db, _, err := connectDatabase(ctx, cfg)
		if err != nil {
			// Exponential backoff com jitter
			retryDelay := calculateBackoff(attempt, baseRetryInterval, maxRetryInterval)
			gl.Log("debug", fmt.Sprintf("Tentativa %d/%d: falha ao conectar: %v (retry em %v)", attempt, maxAttempts, err, retryDelay))
			time.Sleep(retryDelay)
			continue
		}

		sqlDB, err := db.DB()
		if err != nil {
			retryDelay := calculateBackoff(attempt, baseRetryInterval, maxRetryInterval)
			gl.Log("debug", fmt.Sprintf("Tentativa %d/%d: falha ao obter DB: %v (retry em %v)", attempt, maxAttempts, err, retryDelay))
			time.Sleep(retryDelay)
			continue
		}

		conn, err := sqlDB.Conn(ctx)
		if err != nil {
			sqlDB.Close()
			retryDelay := calculateBackoff(attempt, baseRetryInterval, maxRetryInterval)
			gl.Log("debug", fmt.Sprintf("Tentativa %d/%d: falha ao obter conexão: %v (retry em %v)", attempt, maxAttempts, err, retryDelay))
			time.Sleep(retryDelay)
			continue
		}

		// Beauty sleep pra PostgreSQL acordar completamente
		time.Sleep(500 * time.Millisecond)

		// Ping com retry interno curto (2 tentativas)
		pingSuccess := false
		for pingAttempt := 1; pingAttempt <= 2; pingAttempt++ {
			pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
			if err := conn.PingContext(pingCtx); err != nil {
				cancel()
				if pingAttempt == 2 {
					// Última tentativa de ping falhou
					conn.Close()
					sqlDB.Close()
					retryDelay := calculateBackoff(attempt, baseRetryInterval, maxRetryInterval)
					gl.Log("debug", fmt.Sprintf("Tentativa %d/%d: falha ao pingar conexão após %d tentativas: %v (retry em %v)",
						attempt, maxAttempts, pingAttempt, err, retryDelay))
					time.Sleep(retryDelay)
					break
				}
				cancel()
				time.Sleep(500 * time.Millisecond) // Pequeno delay entre pings
				continue
			}
			cancel()
			pingSuccess = true
			break
		}

		if pingSuccess {
			gl.Log("info", fmt.Sprintf("✅ PostgreSQL pronto (tentativa %d/%d)", attempt, maxAttempts))
			return db, conn, nil
		}
	}

	return nil, nil, fmt.Errorf("PostgreSQL não ficou pronto após %d tentativas", maxAttempts)
}

// calculateBackoff calcula delay com exponential backoff + jitter
func calculateBackoff(attempt int, baseInterval, maxInterval time.Duration) time.Duration {
	// Exponential: 2s, 4s, 8s, 16s...
	backoff := baseInterval * time.Duration(1<<uint(attempt-1))

	// Limita ao máximo
	if backoff > maxInterval {
		backoff = maxInterval
	}

	// Adiciona jitter (aleatoriedade de ±25%)
	jitter := time.Duration(rand.Int63n(int64(backoff) / 2)) // 0 a 50% do backoff
	if rand.Intn(2) == 0 {
		backoff += jitter / 2 // +0 a 25%
	} else {
		backoff -= jitter / 2 // -0 a 25%
	}

	return backoff
}

// GetOrGenPasswordKeyringPass retrieves the password from the keyring or generates a new one if it doesn't exist
// It uses the keyring service name to store and retrieve the password
// These methods aren't exposed to the outside world, only accessible through the package main logic
func GetOrGenPasswordKeyringPass(name string) (string, error) {
	// Try to retrieve the password from the keyring
	krPass, krPassErr := krs.NewKeyringService(KeyringService, fmt.Sprintf("gobe-%s", name)).RetrievePassword()
	if krPassErr != nil && krPassErr == os.ErrNotExist {
		gl.Log("debug", fmt.Sprintf("Key found for %s", name))
		// If the error is "keyring: item not found", generate a new key
		gl.Log("debug", fmt.Sprintf("Key not found, generating new key for %s", name))
		krPassKey, krPassKeyErr := crp.NewCryptoServiceType().GenerateKey()
		if krPassKeyErr != nil {
			gl.Log("error", fmt.Sprintf("Error generating key: %v", krPassKeyErr))
			return "", krPassKeyErr
		}
		krPass = string(krPassKey)

		// Store the password in the keyring and return the encoded password
		return storeKeyringPassword(name, []byte(krPass))
	} else if krPassErr != nil {
		gl.Log("error", fmt.Sprintf("Error retrieving key: %v", krPassErr))
		return "", krPassErr
	}

	if !crp.IsBase64String(krPass) {
		krPass = crp.NewCryptoService().EncodeBase64([]byte(krPass))
	}

	return krPass, nil
}

// storeKeyringPassword stores the password in the keyring
// It will check if data is encoded, if so, will decode, store and then
// encode again or encode for the first time, returning always a portable data for
// the caller/logic outside this package be able to use it better and safer
// This method is not exposed to the outside world, only accessible through the package main logic
func storeKeyringPassword(name string, pass []byte) (string, error) {
	cryptoService := crp.NewCryptoServiceType()
	// Will decode if encoded, but only if the password is not empty, not nil and not ENCODED
	copyPass := make([]byte, len(pass))
	copy(copyPass, pass)

	var decodedPass []byte
	if crp.IsBase64String(string(copyPass)) {
		var decodeErr error
		// Will decode if encoded, but only if the password is not empty, not nil and not ENCODED
		decodedPass, decodeErr = cryptoService.DecodeIfEncoded(copyPass)
		if decodeErr != nil {
			gl.Log("error", fmt.Sprintf("Error decoding password: %v", decodeErr))
			return "", decodeErr
		}
	} else {
		decodedPass = copyPass
	}

	// Store the password in the keyring decoded to avoid storing the encoded password
	// locally are much better for security keep binary static and encoded to handle with transport
	// integration and other utilities
	storeErr := krs.NewKeyringService(KeyringService, fmt.Sprintf("gobe-%s", name)).StorePassword(string(decodedPass))
	if storeErr != nil {
		gl.Log("error", fmt.Sprintf("Error storing key: %v", storeErr))
		return "", storeErr
	}

	// Handle with logging here for getOrGenPasswordKeyringPass output
	encodedPass, encodeErr := cryptoService.EncodeIfDecoded(decodedPass)
	if encodeErr != nil {
		gl.Log("error", fmt.Sprintf("Error encoding password: %v", encodeErr))
		return "", encodeErr
	}

	// Return the encoded password to be used by the caller/logic outside this package
	return encodedPass, nil
}

func GetConnectionString(dbConfig *ti.Database) string {
	if dbConfig.ConnectionString != "" {
		return dbConfig.ConnectionString
	}
	if dbConfig.Host != "" && dbConfig.Port != nil && dbConfig.Username != "" && dbConfig.Name != "" {
		dbPass := dbConfig.Password
		if dbPass == "" {
			dbPassKey, dbPassErr := getPasswordFromKeyring("pgpass")
			if dbPassErr != nil {
				gl.Log("error", fmt.Sprintf("❌ Erro ao recuperar senha do banco de dados: %v", dbPassErr))
			} else {
				dbConfig.Password = string(dbPassKey)
				dbPass = dbConfig.Password
			}
		}
		return fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=America/Sao_Paulo",
			// "host=%s port=%s user=%s dbname=%s sslmode=disable TimeZone=America/Sao_Paulo",
			dbConfig.Host, dbConfig.Port.(string), dbConfig.Username, dbPass, dbConfig.Name,
			// dbConfig.Host, dbConfig.Port.(string), dbConfig.Username /* dbPass, */, dbConfig.Name,
		)
	}
	return ""
}

func GetDB(ctx context.Context, d *DBServiceImpl) (*gorm.DB, error) {
	if d == nil {
		return nil, fmt.Errorf("❌ Serviço de banco de dados não inicializado")
	}
	if d.db == nil {
		return nil, fmt.Errorf("❌ Banco de dados não inicializado")
	}
	if d.config == nil {
		return nil, fmt.Errorf("❌ Database Service não configurado")
	}
	dbName, hasDefault := d.GetDefaultDBName()

	// Check if the configured default actually exists in the map
	if hasDefault {
		if _, exists := d.db[dbName]; !exists {
			hasDefault = false
		}
	}

	if !hasDefault {
		// Fallback: if there's only one database, use it
		dbLength := len(d.db)
		if dbLength == 0 {
			d.Initialize(ctx)
			dbLength = len(d.db)
		}
		// If there's exactly one DB, use it as default
		if dbLength > 0 {
			if dbLength == 1 {
				gl.Log("notice", "No default DB configured, using the only available one")
				for name := range d.db {
					dbName = name
					hasDefault = true
					break
				}
			} else {
				gl.Log("error", "No default DB configured and multiple DBs available, cannot decide which to use")
				return nil, fmt.Errorf("❌ No default DB configured and multiple DBs available (%d), cannot decide which to use", dbLength)
			}
		} else {
			if len(d.config.Databases) > 0 {
				for _, dbConf := range d.config.Databases {
					if dbConf.Enabled {
						gl.Log("notice", fmt.Sprintf("No DB connections available, attempting to connect to '%s'", dbConf.Name))
						db, _, err := connectDatabase(ctx, dbConf)
						if err != nil {
							gl.Log("error", fmt.Sprintf("Error connecting to DB '%s': %v", dbConf.Name, err))
							continue
						}
						d.db[dbConf.Name] = db
						dbName = dbConf.Name
						hasDefault = true
						break
					}
				}
			}
		}

		if !hasDefault {
			return nil, fmt.Errorf("%s", fmt.Sprintf("❌ No default DB configured (%d available)", dbLength))
		}
	}

	db := d.db[dbName]
	if db == nil {
		d.Reconnect(ctx)
		db = d.db[dbName]
		if db == nil {
			return nil, fmt.Errorf("❌ Default DB connection is down")
		}
	}

	return db, nil
}
