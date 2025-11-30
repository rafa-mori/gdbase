// Package adaptive provides configurations for adaptive services in CanalizeDB.
package adaptive

import (
	"reflect"

	"github.com/kubex-ecosystem/gdbase/internal/services/adaptive/data"
	"github.com/kubex-ecosystem/gdbase/internal/types"
)

type AdaptiveConfig[T any, A data.AdaptiveService[T]] = ConfigImpl[T, A]
type IAdaptiveConfig[T any, A data.AdaptiveService[T]] interface {
	ConfigImpl[T, A]
}

type General struct {
	// Identification fields
	Reference *types.Reference `json:"reference" yaml:"reference" xml:"reference" toml:"reference" mapstructure:"reference,squash"`

	// General fields
	Enabled     bool         `json:"enabled" yaml:"enabled" xml:"enabled" toml:"enabled" mapstructure:"enabled"`
	Debug       bool         `json:"debug" yaml:"debug" xml:"debug" toml:"debug" mapstructure:"debug"`
	AutoMigrate bool         `json:"auto_migrate" yaml:"auto_migrate" xml:"auto_migrate" toml:"auto_migrate" mapstructure:"auto_migrate"`
	DBType      reflect.Type `json:"type" yaml:"type" xml:"type" toml:"type" mapstructure:"type"`
}

func NewGeneral(name string, enabled bool, debug bool, autoMigrate bool, dbType reflect.Type) *General {
	return &General{
		Reference:   types.NewReference(name).GetReference(),
		Enabled:     enabled,
		Debug:       debug,
		AutoMigrate: autoMigrate,
		DBType:      dbType,
	}
}

type Authentication struct {
	// Authentication fields
	Username string `json:"username" yaml:"username" xml:"username" toml:"username" mapstructure:"username"`
	CertPath string `json:"cert_path" yaml:"cert_path" xml:"cert_path" toml:"cert_path" mapstructure:"cert_path"`
	KeyPath  string `json:"key_path" yaml:"key_path" xml:"key_path" toml:"key_path" mapstructure:"key_path"`
	Password string `json:"password" yaml:"password" xml:"password" toml:"password" mapstructure:"password"`
}

func NewAuthentication(username, certPath, keyPath, password string) *Authentication {
	return &Authentication{
		Username: username,
		CertPath: certPath,
		KeyPath:  keyPath,
		Password: password,
	}
}

type Connection struct {
	// Connection fields
	ConfigPath string `json:"config_path" yaml:"config_path" xml:"config_path" toml:"config_path" mapstructure:"config_path"`
	Database   string `json:"database" yaml:"database" xml:"database" toml:"database" mapstructure:"database"`
	Proxy      string `json:"proxy" yaml:"proxy" xml:"proxy" toml:"proxy" mapstructure:"proxy"`
	Host       string `json:"host" yaml:"host" xml:"host" toml:"host" mapstructure:"host"`
	Port       any    `json:"port" yaml:"port" xml:"port" toml:"port" mapstructure:"port"`
	Dsn        string `json:"dsn" yaml:"dsn" xml:"dsn" toml:"dsn" mapstructure:"dsn"`
}

func NewConnection(configPath, database, proxy, host string, port any, dsn string) *Connection {
	return &Connection{
		ConfigPath: configPath,
		Database:   database,
		Proxy:      proxy,
		Host:       host,
		Port:       port,
		Dsn:        dsn,
	}
}

type Containerization struct {
	// Containerization fields
	Image  string   `json:"image" yaml:"image" xml:"image" toml:"image" mapstructure:"image"`
	Volume []string `json:"volume" yaml:"volume" xml:"volume" toml:"volume" mapstructure:"volume"`
}

func NewContainerization(image string, volume []string) *Containerization {
	return &Containerization{
		Image:  image,
		Volume: volume,
	}
}

type Metadata struct {
	env      map[string]any
	options  map[string]any
	metadata map[string]any
}

func NewMetadata(env, options, metadata map[string]any) *Metadata {
	return &Metadata{
		env:      env,
		options:  options,
		metadata: metadata,
	}
}

type ConfigImpl[T any, A data.AdaptiveService[T]] struct {
	// General configuration
	*General

	// Authentication configuration
	*Authentication

	// Connection configuration
	*Connection

	// Containerization configuration
	*Containerization

	// Metadata
	*Metadata

	// Internal mapper
	Mapper *types.Mapper[*T] `json:"-" yaml:"-" xml:"-" toml:"-" mapstructure:"-"`
}

func newConfigImpl[T any, A data.AdaptiveService[T]](general *General, authentication *Authentication, connection *Connection, containerization *Containerization, metadata *Metadata) *ConfigImpl[T, A] {
	return &ConfigImpl[T, A]{
		General:          general,
		Authentication:   authentication,
		Connection:       connection,
		Containerization: containerization,
		Metadata:         metadata,
	}
}

func NewConfigImpl[T any, A data.AdaptiveService[T]](name string, enabled bool, debug bool, autoMigrate bool, dbType reflect.Type, username, certPath, keyPath, password, configPath, database, proxy, host string, port any, dsn, image string, volume []string, env, options, metadata map[string]any) *ConfigImpl[T, A] {
	general := NewGeneral(name, enabled, debug, autoMigrate, dbType)
	authentication := NewAuthentication(username, certPath, keyPath, password)
	connection := NewConnection(configPath, database, proxy, host, port, dsn)
	containerization := NewContainerization(image, volume)
	meta := NewMetadata(env, options, metadata)

	return newConfigImpl[T, A](general, authentication, connection, containerization, meta)
}
