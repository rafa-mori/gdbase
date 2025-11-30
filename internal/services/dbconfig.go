package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	ci "github.com/kubex-ecosystem/gdbase/internal/interfaces"
	"github.com/kubex-ecosystem/gdbase/internal/module/kbx"
	crp "github.com/kubex-ecosystem/gdbase/internal/security/crypto"
	krs "github.com/kubex-ecosystem/gdbase/internal/security/external"
	"github.com/kubex-ecosystem/logz"
	gl "github.com/kubex-ecosystem/logz"

	ti "github.com/kubex-ecosystem/gdbase/internal/types"
	"gorm.io/gorm"
)

const (
	KeyringService        = "kubex"
	DefaultKubexConfigDir = "$HOME/.kubex"

	DefaultGoBEKeyPath  = "$HOME/.kubex/gobe/gobe-key.pem"
	DefaultGoBECertPath = "$HOME/.kubex/gobe/gobe-cert.pem"

	DefaultConfigDir        = "$HOME/.kubex/gdbase/config"
	DefaultConfigFile       = "$HOME/.kubex/gdbase/config.json"
	DefaultGDBaseConfigPath = "$HOME/.kubex/gdbase/config/config.json"
)

const (
	DefaultVolumesDir     = "$HOME/.kubex/volumes"
	DefaultRedisVolume    = "$HOME/.kubex/volumes/redis"
	DefaultPostgresVolume = "$HOME/.kubex/volumes/postgresql"
	DefaultMongoVolume    = "$HOME/.kubex/volumes/mongo"
	DefaultRabbitMQVolume = "$HOME/.kubex/volumes/rabbitmq"
)

type IDBConfig interface {
	GetDBName() string
	GetDBType() string
	GetEnvironment() string
	GetPostgresConfig() *ti.Database
	GetMySQLConfig() *ti.Database
	GetSQLiteConfig() *ti.Database
	GetMongoDBConfig() *ti.MongoDB
	GetRedisConfig() *ti.Redis
	GetRabbitMQConfig() *ti.RabbitMQ
	IsAutoMigrate() bool
	IsDebug() bool
	GetLogger() any
	GetConfig(context.Context) *DBConfig
	GetConfigMap(context.Context) map[string]any
}

type DBConfig struct {
	// Name is used to configure the name of the database
	Name string `json:"name" yaml:"name" xml:"name" toml:"name" mapstructure:"name"`

	// FilePath is used to configure the file path of the database
	FilePath string `json:"file_path" yaml:"file_path" xml:"file_path" toml:"file_path" mapstructure:"file_path"`

	// Logger is used to configure the logger
	Logger *logz.LoggerZ `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"`

	// Mutexes is used to configure the mutexes, not serialized
	*ti.Mutexes `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"`

	// IsConfidential is used to configure if the environment variables are confidential
	IsConfidential bool `json:"is_confidential,omitempty" yaml:"is_confidential,omitempty" xml:"is_confidential,omitempty" toml:"is_confidential,omitempty" mapstructure:"is_confidential,omitempty"`

	// env is used to configure the environment, not serialized
	env ci.IProperty[*ti.Environment] `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"`

	// Properties is used to configure the properties of the database, not serialized
	properties map[string]interface{} `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"`

	// Debug is used to configure the debug mode
	Debug bool `json:"debug,omitempty" yaml:"debug,omitempty" xml:"debug,omitempty" toml:"debug,omitempty" mapstructure:"debug,omitempty"`

	// AutoMigrate is used to configure the auto migration of the database
	AutoMigrate bool `json:"auto_migrate,omitempty" yaml:"auto_migrate,omitempty" xml:"auto_migrate,omitempty" toml:"auto_migrate,omitempty" mapstructure:"auto_migrate,omitempty"`

	// JWT is used to configure the JWT token settings
	JWT *ti.JWT `json:"jwt,omitempty" yaml:"jwt,omitempty" xml:"jwt,omitempty" toml:"jwt,omitempty" mapstructure:"jwt,omitempty"`

	// Reference is used to configure the reference of the database
	*ti.Reference `json:"reference,omitempty" yaml:"reference,omitempty" xml:"reference,omitempty" toml:"reference,omitempty" mapstructure:"reference,squash,omitempty"`

	// Enabled is used to enable or disable the database
	Enabled bool `json:"enabled" yaml:"enabled" xml:"enabled" toml:"enabled" mapstructure:"enabled"`

	// MongoDB is used to configure the MongoDB database
	MongoDB *ti.MongoDB `json:"mongodb,omitempty" yaml:"mongodb,omitempty" xml:"mongodb,omitempty" toml:"mongodb,omitempty" mapstructure:"mongodb,omitempty"`

	// Databases is used to configure the databases (Postgres, MySQL, SQLite, SQLServer, Oracle)
	Databases map[string]*ti.Database `json:"databases" yaml:"databases" xml:"databases" toml:"databases" mapstructure:"databases"`

	// Messagery is used to configure the messagery database
	Messagery *ti.Messagery `json:"messagery,omitempty" yaml:"messagery,omitempty" xml:"messagery,omitempty" toml:"messagery,omitempty" mapstructure:"messagery,omitempty"`

	// Mapper is used to configure the mapper for serialization and deserialization, not serialized
	Mapper *ti.Mapper[*DBConfig] `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"`
}

func newDBConfig(name, filePath string, enabled bool, logger *logz.LoggerZ, debug bool) *DBConfig {
	if logger == nil {
		logger = gl.NewLogger("GDBase")
	}

	if debug {
		gl.SetDebugMode(debug)
	}

	if name == "" {
		name = "kubex_db"
	}
	if filePath == "" {
		filePath = os.ExpandEnv(DefaultGDBaseConfigPath)
	}

	dbConfig := &DBConfig{FilePath: filePath}
	mapper := ti.NewMapper(&dbConfig, filePath)
	obj, err := mapper.DeserializeFromFile("json")
	if err != nil {
		gl.Log("warn", fmt.Sprintf("Error deserializing file: %v", err))
		if obj == nil {
			pgPass, pgPassErr := getPasswordFromKeyring(name + "_Postgres")
			if pgPassErr != nil {
				gl.Log("error", fmt.Sprintf("Error getting password from keyring: %v", pgPassErr))
				return nil
			}
			redisPass, redisPassErr := getPasswordFromKeyring(name + "_Redis")
			if redisPassErr != nil {
				gl.Log("error", fmt.Sprintf("Error getting password from keyring: %v", redisPassErr))
				return nil
			}
			rabbitPass, rabbitPassErr := getPasswordFromKeyring(name + "_RabbitMQ")
			if rabbitPassErr != nil {
				gl.Log("error", fmt.Sprintf("Error getting password from keyring: %v", rabbitPassErr))
				return nil
			}
			dsn := fmt.Sprintf(
				"postgres://%s:%s@localhost:5432/%s?sslmode=disable",
				kbx.GetEnvOrDefault("POSTGRES_USER", "kubex_adm"),
				pgPass,
				kbx.GetEnvOrDefault("POSTGRES_DB", "kubex_db"),
			)
			dbConfigDefault := &DBConfig{
				Databases: map[string]*ti.Database{
					"kubex_db": {
						Enabled:          (kbx.GetEnvOrDefault("GDBASE_POSTGRES_ENABLED", "true") != "false"),
						Reference:        ti.NewReference(name).GetReference(),
						Type:             "postgresql",
						Driver:           "postgres",
						ConnectionString: kbx.GetEnvOrDefault("GDBASE_POSTGRES_CONNECTION_STRING", dsn),
						Dsn:              kbx.GetEnvOrDefault("GDBASE_POSTGRES_DSN", dsn),
						Path:             kbx.GetEnvOrDefault("GDBASE_POSTGRES_PATH", os.ExpandEnv(DefaultPostgresVolume)),
						Host:             kbx.GetEnvOrDefault("GDBASE_POSTGRES_HOST", "localhost"),
						Port:             kbx.GetEnvOrDefault("GDBASE_POSTGRES_PORT", "5432"),
						Username:         kbx.GetEnvOrDefault("GDBASE_POSTGRES_USER", "kubex_adm"),
						Password:         kbx.GetEnvOrDefault("GDBASE_POSTGRES_PASSWORD", pgPass),
						Volume:           kbx.GetEnvOrDefault("GDBASE_POSTGRES_VOLUME", os.ExpandEnv(DefaultPostgresVolume)),
						Name:             kbx.GetEnvOrDefault("GDBASE_POSTGRES_NAME", "kubex_db"),
						IsDefault:        kbx.GetEnvOrDefault("GDBASE_POSTGRES_IS_DEFAULT", "true") != "false",
					},
				},
				Messagery: &ti.Messagery{
					Redis: &ti.Redis{
						Enabled:   (kbx.GetEnvOrDefault("GDBASE_REDIS_ENABLED", "true") != "false"),
						Reference: ti.NewReference(name + "_Redis").GetReference(),
						FilePath:  kbx.GetEnvOrDefault("GDBASE_REDIS_FILE_PATH", filePath),
						Addr:      kbx.GetEnvOrDefault("GDBASE_REDIS_ADDR", "localhost"),
						Port:      kbx.GetEnvOrDefault("GDBASE_REDIS_PORT", "6379"),
						Username:  kbx.GetEnvOrDefault("GDBASE_REDIS_USER", "default"),
						Password:  kbx.GetEnvOrDefault("GDBASE_REDIS_PASSWORD", redisPass),
						DB:        0,
					},
					RabbitMQ: &ti.RabbitMQ{
						Enabled:        (kbx.GetEnvOrDefault("GDBASE_RABBITMQ_ENABLED", "true") != "false"),
						Reference:      ti.NewReference(name + "_RabbitMQ").GetReference(),
						FilePath:       kbx.GetEnvOrDefault("GDBASE_RABBITMQ_FILE_PATH", filePath),
						Username:       kbx.GetEnvOrDefault("GDBASE_RABBITMQ_USER", "gobe"),
						Password:       kbx.GetEnvOrDefault("GDBASE_RABBITMQ_PASSWORD", rabbitPass),
						Port:           kbx.GetEnvOrDefault("GDBASE_RABBITMQ_PORT", "5672"),
						ManagementPort: kbx.GetEnvOrDefault("GDBASE_RABBITMQ_MANAGEMENT_PORT", "15672"),
						Vhost:          kbx.GetEnvOrDefault("GDBASE_RABBITMQ_VHOST", "gobe"),
					},
				},
			}
			dbConfig = dbConfigDefault
			mapper = ti.NewMapper(&dbConfig, filePath)
			mapper.SerializeToFile("json")
			if _, statErr := os.Stat(filepath.Dir(filePath)); os.IsNotExist(statErr) {
				gl.Log("fatal", fmt.Sprintf("Error creating directory: %v", statErr))
			}
			gl.Log("info", fmt.Sprintf("File %s created with default values", filePath))

			if data, dataErr := mapper.Serialize("json"); dataErr != nil {
				gl.Log("fatal", fmt.Sprintf("Error serializing file: %v", dataErr))
			} else {
				if err := os.WriteFile(filePath, data, 0644); err != nil {
					gl.Log("fatal", fmt.Sprintf("Error writing file: %v", err))
				}
			}
		}
	}
	if dbConfig.Logger == nil {
		dbConfig.Logger = logger
	}
	if dbConfig.Databases == nil {
		dbConfig.Databases = map[string]*ti.Database{}
	}
	if dbConfig.Messagery == nil {
		dbConfig.Messagery = &ti.Messagery{}
	}

	return dbConfig
}
func NewDBConfigWithArgs(ctx context.Context, name, filePath string, enabled bool, logger *logz.LoggerZ, debug bool) *DBConfig {
	return newDBConfig(name, filePath, enabled, logger, debug)
}
func NewDBConfig(dbConfig *DBConfig) *DBConfig {
	if dbConfig.Logger == nil {
		dbConfig.Logger = gl.NewLogger("GDBase")
	}
	if dbConfig.Name == "" {
		dbConfig.Name = "default"
	}
	if dbConfig.Mutexes == nil {
		dbConfig.Mutexes = ti.NewMutexesType()
	}
	if dbConfig.FilePath == "" {
		dbConfig.FilePath = os.ExpandEnv(DefaultGDBaseConfigPath)
	}
	willWrite := false
	if dbConfig.Mapper == nil {
		dbConfig.Mapper = ti.NewMapperType(&dbConfig, dbConfig.FilePath)
		if _, statErr := os.Stat(dbConfig.FilePath); os.IsNotExist(statErr) {
			if err := os.MkdirAll(filepath.Dir(dbConfig.FilePath), os.ModePerm); err != nil {
				gl.Log("error", fmt.Sprintf("Error creating directory: %v", err))
			} else {
				willWrite = true
			}
		} else {
			_, err := dbConfig.Mapper.DeserializeFromFile("json")
			if err != nil {
				gl.Log("error", fmt.Sprintf("Error deserializing file: %v", err))
			}
		}
	}
	if willWrite {
		dbConfig.Mapper.SerializeToFile("json")
	}
	return dbConfig
}
func NewDBConfigWithDBConnection(db *gorm.DB) *DBConfig {
	return newDBConfig("default", "", true, nil, false)
}
func NewDBConfigWithFilePath(name, filePath string) *DBConfig {
	return newDBConfig(name, filePath, true, nil, false)
}
func NewDBConfigFromFile(ctx context.Context, dbConfigFilePath string, autoMigrate bool, logger *logz.LoggerZ, debug bool) (*DBConfig, error) {
	var dbConfig *DBConfig
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return dbConfig, nil
}

func getPasswordFromKeyring(name string) (string, error) {
	krPass, pgPassErr := krs.NewKeyringService(KeyringService, fmt.Sprintf("gdbase-%s", name)).RetrievePassword()
	if pgPassErr != nil && pgPassErr.Error() != "keyring: item not found" {
		krPassKey, krPassKeyErr := crp.NewCryptoServiceType().GenerateKey()
		if krPassKeyErr != nil {
			gl.Log("error", fmt.Sprintf("Error generating key: %v", krPassKeyErr))
			return "", krPassKeyErr
		}
		storeErr := krs.NewKeyringService(KeyringService, fmt.Sprintf("gdbase-%s", name)).StorePassword(string(krPassKey))
		if storeErr != nil {
			gl.Log("error", fmt.Sprintf("Error storing key: %v", storeErr))
			return "", storeErr
		}
		krPass = string(krPassKey)
	}
	return base64.URLEncoding.EncodeToString([]byte(krPass)), nil
}

func (d *DBConfig) GetDBName() string {
	if d == nil {
		return ""
	}
	for _, db := range d.Databases {
		if db.Enabled {
			return db.Name
		}
	}
	return ""
}
func (d *DBConfig) GetEnvironment() string {
	if d == nil {
		return ""
	}
	var env *ti.Environment
	if d.env == nil {
		e, err := ti.NewEnvironmentType(kbx.GetEnvOrDefault("GDBASE_ENV_FILE", ".env"), d.IsConfidential, gl.GetLoggerZ("GDBase"))
		if err != nil {
			gl.Log("error", fmt.Sprintf("Error creating environment: %v", err))
			return "development"
		}
		d.env = ti.NewProperty(
			"env",
			&e,
			false,
			nil,
		)
		env = e
	} else {
		env = d.env.GetValue()
	}
	dotEnvMode, _ := kbx.GetValueOrDefault(env.Getenv("GDBASE_ENV"), "development")
	envs := os.Environ()
	for _, env := range envs {
		pair := os.ExpandEnv(env)
		if pairParts := filepath.SplitList(pair); len(pairParts) == 2 {
			if pairParts[0] == "GDBASE_ENV" {
				return pairParts[1]
			}

		}
	}
	return dotEnvMode
}
func (d *DBConfig) GetDBType() string {
	if d == nil {
		return ""
	}
	for _, db := range d.Databases {
		if db.Enabled {
			return db.Type
		}
	}
	return ""
}
func (d *DBConfig) GetPostgresConfig() *ti.Database {
	if d == nil {
		return nil
	}
	postgres, ok := d.Databases["kubex_db"]
	if !ok {
		return nil
	}
	return postgres
}
func (d *DBConfig) GetMySQLConfig() *ti.Database {
	if d == nil {
		return nil
	}
	mysql, ok := d.Databases["mysql"]
	if !ok {
		return nil
	}
	return mysql
}

func (d *DBConfig) GetSQLiteConfig() *ti.Database {
	if d == nil {
		return nil
	}
	sqlite, ok := d.Databases["sqlite"]
	if !ok {
		return nil
	}
	return sqlite
}
func (d *DBConfig) GetMongoDBConfig() *ti.MongoDB {
	if d == nil {
		return nil
	}
	return d.MongoDB
}
func (d *DBConfig) GetRedisConfig() *ti.Redis {
	if d == nil {
		return nil
	}
	return d.Messagery.Redis
}
func (d *DBConfig) GetRabbitMQConfig() *ti.RabbitMQ {
	if d == nil {
		return nil
	}
	return d.Messagery.RabbitMQ
}
func (d *DBConfig) IsAutoMigrate() bool {
	if d == nil {
		return false
	}

	return d.Enabled
}
func (d *DBConfig) IsDebug() bool {
	if d == nil {
		return false
	}
	config, ok := d.properties["config"].(*ti.Property[*DBConfig])
	if !ok {
		return false
	}
	dbConfig := config.GetValue()
	if dbConfig == nil {
		return false
	}
	return dbConfig.Debug
}
func (d *DBConfig) GetLogger() interface{} {
	if d == nil {
		return nil
	}
	return d.Logger
}

func (d *DBConfig) GetConfig(ctx context.Context) *DBConfig {
	if d == nil {
		return nil
	}
	return d
}

func (d *DBConfig) GetConfigMap(ctx context.Context) map[string]any {
	// Quero percorrer toda a struct de DBConfig independente da profundidade e das propriedades
	// Criar um map disso e retornar
	if d == nil {
		return nil
	}
	arrForConfig := make(map[string]any)

	for key, val := range d.properties {
		arrForConfig[key] = val
	}

	v := reflect.ValueOf(*d)
	reflect.VisibleFields(v.Type())
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		value := v.Field(i).Interface()
		if field.Type.Kind() == reflect.Ptr && !v.Field(i).IsNil() {
			elem := v.Field(i).Elem()
			if elem.Kind() == reflect.Struct {
				nestedConfig := make(map[string]any)
				for j := 0; j < elem.NumField(); j++ {
					nestedField := elem.Type().Field(j)
					nestedValue := elem.Field(j).Interface()
					nestedConfig[nestedField.Name] = nestedValue
				}
				arrForConfig[field.Name] = nestedConfig
			} else {
				arrForConfig[field.Name] = value
			}
		} else if field.Type.Kind() == reflect.Struct {
			nestedConfig := make(map[string]any)
			elem := v.Field(i)
			for j := 0; j < elem.NumField(); j++ {
				nestedField := elem.Type().Field(j)
				nestedValue := elem.Field(j).Interface()
				nestedConfig[nestedField.Name] = nestedValue
			}
			arrForConfig[field.Name] = nestedConfig
		} else {
			arrForConfig[field.Name] = value
		}
	}

	return arrForConfig
}
