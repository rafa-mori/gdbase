// Package provider defines interfaces and types for managing service providers like Postgres, MongoDB, Redis, and RabbitMQ.
package provider

import (
	"context"

	"github.com/kubex-ecosystem/gdbase/internal/module/kbx"
	"github.com/kubex-ecosystem/gdbase/internal/types"
)

type Engine string

const (
	EnginePostgres Engine = "postgres"
	EngineMongo    Engine = "mongo"
	EngineRedis    Engine = "redis"
	EngineRabbit   Engine = "rabbitmq"
)

type ServiceRef struct {
	Name   string // "pg", "mongo", "redis", "rabbit"
	Engine Engine
}

type Endpoint struct {
	DSN      string // ex: postgres://user:pass@host:port/db?sslmode=disable
	Redacted string
	Host     string
	Port     string
}

type Capabilities struct {
	Managed  bool
	Notes    []string
	Features map[string]bool // ex: "extensions.pgcrypto": true
}

type StartSpec struct {
	Services      []ServiceRef // quais serviços subir/anexar
	PreferredPort map[string]int
	Secrets       map[string]string // senhas já geradas pelo GoBE
	Labels        map[string]string // rastreabilidade
}

// Provider is the base interface that all service providers must implement.
// It provides lifecycle management (start, health, stop) without assuming
// any specific database migration capabilities.
type Provider interface {
	// Name returns the provider identifier (e.g., "dockerstack", "aws-rds").
	Name() string

	// Capabilities returns what this provider can do (managed services, features).
	Capabilities(ctx context.Context) (Capabilities, error)

	// Start provisions or attaches services and returns ready endpoints.
	// Does NOT handle migrations - use MigratableProvider for that.
	Start(ctx context.Context, spec StartSpec) (map[string]Endpoint, error)

	// Health verifies connectivity to each service (e.g., SELECT 1, PING, AMQP open/close).
	Health(ctx context.Context, eps map[string]Endpoint) error

	// Stop tears down managed resources (optional for managed providers).
	Stop(ctx context.Context, refs []ServiceRef) error
}

// MigratableProvider extends Provider with schema migration capabilities.
// Only database providers with schema management (Postgres, MySQL, MongoDB)
// should implement this interface. Redis, RabbitMQ, and other non-schema
// services only need to implement the base Provider interface.
//
// By embedding Provider, this interface INHERITS all Provider methods:
// - Name(), Capabilities(), Start(), Health(), Stop()
// AND adds migration-specific methods:
// - PrepareMigrations(), RunMigrations()
//
// Any type implementing MigratableProvider must implement ALL 7 methods.
type MigratableProvider interface {
	Provider

	// PrepareMigrations validates connection and readiness for migrations.
	PrepareMigrations(ctx context.Context, conn *types.DBConnection) error

	// RunMigrations executes schema migrations for the given database connection.
	RunMigrations(ctx context.Context, conn *types.DBConnection, migrationInfo *kbx.MigrationInfo) error
}

// RootConfigProvider is a convenience interface for providers that can
// start services directly from a RootConfig (high-level orchestration).
// This is typically used by CLI commands that want to start all services
// defined in a configuration file in one call.
//
// By embedding Provider, this interface INHERITS all Provider methods:
// - Name(), Capabilities(), Start(), Health(), Stop()
// AND adds high-level orchestration:
// - StartServices()
//
// Any type implementing RootConfigProvider must implement ALL 6 methods.
type RootConfigProvider interface {
	Provider

	// StartServices starts all services defined in RootConfig.
	// This includes container startup, readiness checks, and migrations
	// (if the provider implements MigratableProvider).
	StartServices(ctx context.Context, rootConfig *kbx.RootConfig) error
}
