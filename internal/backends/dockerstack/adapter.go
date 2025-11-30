// Package dockerstack provides Docker-based backend implementation
package dockerstack

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/kubex-ecosystem/gdbase/internal/module/kbx"
	"github.com/kubex-ecosystem/gdbase/internal/provider"
	"github.com/kubex-ecosystem/logz"

	svc "github.com/kubex-ecosystem/gdbase/internal/services"
	"github.com/kubex-ecosystem/gdbase/internal/types"
	gl "github.com/kubex-ecosystem/logz"
)

// DockerStackProvider wraps legacy Docker services into new Provider interface
type DockerStackProvider struct {
	logger        *logz.LoggerZ
	dockerService *svc.DockerService
}

// NewDockerStackProvider creates a new Docker-based provider
func NewDockerStackProvider() *DockerStackProvider {
	return &DockerStackProvider{
		logger: gl.GetLogger("DockerStackProvider"),
	}
}

// Name returns the provider name
func (p *DockerStackProvider) Name() string {
	return "dockerstack"
}

// Capabilities returns what this provider can do
func (p *DockerStackProvider) Capabilities(ctx context.Context) (provider.Capabilities, error) {
	return provider.Capabilities{
		Managed: true, // Docker managed containers
		Notes: []string{
			"Zero-config local stack using Docker",
			"Supports PostgreSQL, MongoDB, Redis, RabbitMQ",
			"Auto-generates credentials via keyring",
		},
		Features: map[string]bool{
			"network.internal": true,
			"publish.ports":    true,
			"volumes.persist":  true,
		},
	}, nil
}

// Start initializes database services using legacy SetupDatabaseServices
func (p *DockerStackProvider) Start(ctx context.Context, spec provider.StartSpec) (map[string]provider.Endpoint, error) {
	// 1. Convert provider.StartSpec to legacy DBConfig format
	dbConfig := p.ConvertSpecToDBConfig(spec)
	var ok bool
	// 2. Initialize Docker service (legacy)
	dockerService, err := svc.NewDockerService(dbConfig, p.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker service: %w", err)
	}

	p.dockerService, ok = dockerService.(*svc.DockerService)
	if !ok {
		return nil, fmt.Errorf("failed to assert docker service type")
	}

	// 3. Initialize services (calls legacy SetupDatabaseServices)
	if err := p.dockerService.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize docker services: %w", err)
	}

	// 4. Extract endpoints from running containers
	endpoints, err := p.ExtractEndpoints(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to extract endpoints: %w", err)
	}

	// 5. 🎯 NOVA LÓGICA: Executar migrations SEMPRE após container subir
	if pgEndpoint, exists := endpoints["pg"]; exists {
		if err := p.RunPostgresMigrations(ctx, pgEndpoint.DSN); err != nil {
			return nil, fmt.Errorf("failed to run PostgreSQL migrations: %w", err)
		}
	}

	return endpoints, nil
}

// ConvertSpecToDBConfig converts new StartSpec to legacy DBConfig
func (p *DockerStackProvider) ConvertSpecToDBConfig(spec provider.StartSpec) *svc.DBConfig {
	dbConfig := &svc.DBConfig{
		Databases: make(map[string]*types.Database),
	}

	for _, svc := range spec.Services {
		db := &types.Database{
			Enabled: true,
		}

		var key string
		switch svc.Engine {
		case provider.EnginePostgres:
			key = "kubex_db"
			db.Enabled = true
			db.Volume = kbx.GetEnvOrDefault("GOBE_POSTGRES_VOLUME", os.ExpandEnv(kbx.DefaultPostgresVolume))
			db.Type = "postgresql"
			db.Name = "kubex_db"
			db.Username = "kubex_adm"
			db.Password = spec.Secrets["pg_admin"]
			db.Host = "127.0.0.1"
			if port, ok := spec.PreferredPort["pg"]; ok {
				db.Port = port
			} else {
				db.Port = 5432
			}

		case provider.EngineMongo:
			key = "mongodb"
			db.Enabled = true
			db.Type = "mongodb"
			db.Name = "gdbase"
			db.Username = "root"
			db.Password = spec.Secrets["mongo_root"]
			db.Host = "127.0.0.1"
			if port, ok := spec.PreferredPort["mongo"]; ok {
				db.Port = port
			} else {
				db.Port = 27017
			}

		case provider.EngineRedis:
			key = "redis"
			db.Enabled = true
			db.Type = "redis"
			db.Password = spec.Secrets["redis_pass"]
			db.Host = "127.0.0.1"
			if port, ok := spec.PreferredPort["redis"]; ok {
				db.Port = port
			} else {
				db.Port = 6379
			}

		case provider.EngineRabbit:
			key = "rabbitmq"
			db.Enabled = true
			db.Type = "rabbitmq"
			db.Username = "admin"
			db.Password = spec.Secrets["rabbit_pass"]
			db.Host = "127.0.0.1"
			if port, ok := spec.PreferredPort["rabbit"]; ok {
				db.Port = port
			} else {
				db.Port = 5672
			}
		}

		if key != "" {
			dbConfig.Databases[key] = db
		}
	}

	return dbConfig
}

// ExtractEndpoints converts legacy DBConfig to new Endpoint format
func (p *DockerStackProvider) ExtractEndpoints(cfg *svc.DBConfig) (map[string]provider.Endpoint, error) {
	endpoints := make(map[string]provider.Endpoint)

	for _, db := range cfg.Databases {
		if db == nil || !db.Enabled {
			continue
		}

		var ep provider.Endpoint
		var name string

		switch db.Type {
		case "postgresql", "postgres":
			name = "kubex_db"
			port := db.Port
			if portInt, ok := port.(int); ok {
				ep.Port = portInt
			} else if portStr, ok := port.(string); ok {
				fmt.Sscanf(portStr, "%d", &ep.Port)
			}
			ep.Host = db.Host
			ep.DSN = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
				db.Username, db.Password, db.Host, ep.Port, db.Name)
			ep.Redacted = fmt.Sprintf("postgres://%s:***@%s:%d/%s", db.Username, db.Host, ep.Port, db.Name)

		case "mongodb":
			name = "mongo"
			port := db.Port
			if portInt, ok := port.(int); ok {
				ep.Port = portInt
			} else if portStr, ok := port.(string); ok {
				fmt.Sscanf(portStr, "%d", &ep.Port)
			}
			ep.Host = db.Host
			ep.DSN = fmt.Sprintf("mongodb://%s:%s@%s:%d", db.Username, db.Password, db.Host, ep.Port)
			ep.Redacted = fmt.Sprintf("mongodb://%s:***@%s:%d", db.Username, db.Host, ep.Port)

		case "redis":
			name = "redis"
			port := db.Port
			if portInt, ok := port.(int); ok {
				ep.Port = portInt
			} else if portStr, ok := port.(string); ok {
				fmt.Sscanf(portStr, "%d", &ep.Port)
			}
			ep.Host = db.Host
			ep.DSN = fmt.Sprintf("redis://:%s@%s:%d", db.Password, db.Host, ep.Port)
			ep.Redacted = fmt.Sprintf("redis://:***@%s:%d", db.Host, ep.Port)

		case "rabbitmq":
			name = "rabbit"
			port := db.Port
			if portInt, ok := port.(int); ok {
				ep.Port = portInt
			} else if portStr, ok := port.(string); ok {
				fmt.Sscanf(portStr, "%d", &ep.Port)
			}
			ep.Host = db.Host
			ep.DSN = fmt.Sprintf("amqp://%s:%s@%s:%d/", db.Username, db.Password, db.Host, ep.Port)
			ep.Redacted = fmt.Sprintf("amqp://%s:***@%s:%d/", db.Username, db.Host, ep.Port)
		}

		if name != "" {
			endpoints[name] = ep
		}
	}

	return endpoints, nil
}

// RunPostgresMigrations executes SQL migrations programmatically
func (p *DockerStackProvider) RunPostgresMigrations(ctx context.Context, dsn string) error {
	migrationManager := NewMigrationManager(dsn, p.logger)

	// Wait for PostgreSQL to be ready
	if err := migrationManager.WaitForPostgres(ctx, 30*time.Second); err != nil {
		return err
	}

	// Check if schema already exists (avoid re-running)
	exists, err := migrationManager.SchemaExists()
	if err != nil {
		gl.Log("warn", fmt.Sprintf("Could not check schema existence: %v", err))
	}

	if exists {
		gl.Log("info", "PostgreSQL schema already exists, skipping migrations")
		return nil
	}

	// Run migrations with error recovery
	gl.Log("info", "Running PostgreSQL migrations...")
	results, err := migrationManager.RunMigrations(ctx)
	if err != nil {
		return err
	}

	// Log final summary
	totalSuccess := 0
	totalFailed := 0
	for _, r := range results {
		totalSuccess += r.SuccessfulStmts
		totalFailed += r.FailedStmts
	}

	if totalFailed > 0 {
		gl.Log("warn", fmt.Sprintf("Migration completed with partial success: %d succeeded, %d failed", totalSuccess, totalFailed))
		// Don't return error for partial failures - let the service continue
	} else {
		gl.Log("info", fmt.Sprintf("All migrations completed successfully! (%d statements)", totalSuccess))
	}

	return nil
}

// Health verifies connectivity to all services
func (p *DockerStackProvider) Health(ctx context.Context, eps map[string]provider.Endpoint) error {
	// TODO: Implement real health checks
	// For now, just verify Docker service is initialized
	if p.dockerService == nil {
		return fmt.Errorf("docker service not initialized")
	}
	return nil
}

// Stop stops all managed containers
func (p *DockerStackProvider) Stop(ctx context.Context, refs []provider.ServiceRef) error {
	// TODO: Call docker service stop methods
	return nil
}
