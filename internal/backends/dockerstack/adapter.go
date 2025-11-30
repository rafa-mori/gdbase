// Package dockerstack provides Docker-based backend implementation
package dockerstack

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/kubex-ecosystem/gdbase/internal/engine"
	"github.com/kubex-ecosystem/gdbase/internal/module/kbx"
	"github.com/kubex-ecosystem/gdbase/internal/provider"
	"github.com/kubex-ecosystem/gdbase/internal/types"

	ci "github.com/kubex-ecosystem/gdbase/internal/interfaces"
	logz "github.com/kubex-ecosystem/logz"
)

// DockerStackProvider wraps legacy Docker services into new Provider interface.
// It implements Provider, MigratableProvider, and RootConfigProvider interfaces.
type DockerStackProvider struct {
	logger        *logz.LoggerZ
	dockerService ci.IDockerService
}

// NewDockerStackProvider creates a new Docker-based provider with constructor injection.
// The dockerService parameter is required and must not be nil.
func NewDockerStackProvider(dockerService ci.IDockerService) *DockerStackProvider {
	return &DockerStackProvider{
		logger:        logz.NewLogger("GoLife"),
		dockerService: dockerService,
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
			"migrations":       true,
		},
	}, nil
}

// Start provisions or attaches services and returns ready endpoints.
// This implements the Provider interface without handling migrations.
// Use StartServices() for complete orchestration including migrations.
func (p *DockerStackProvider) Start(ctx context.Context, spec provider.StartSpec) (map[string]provider.Endpoint, error) {
	// Validate dockerService was injected
	if p.dockerService == nil {
		return nil, fmt.Errorf("dockerService not initialized (use NewDockerStackProvider with service injection)")
	}

	// 1. Convert provider.StartSpec to legacy DBConfig format
	cfg := p.ConvertSpecToDBConfig(spec)

	// 2. Initialize services (calls legacy SetupDatabaseServices)
	if err := p.dockerService.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize docker services: %w", err)
	}

	// 3. Extract endpoints from running containers
	endpoints, err := p.ExtractEndpoints(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to extract endpoints: %w", err)
	}

	return endpoints, nil
}

// ConvertSpecToDBConfig converts new StartSpec to legacy DBConfig
func (p *DockerStackProvider) ConvertSpecToDBConfig(spec provider.StartSpec) *engine.DatabaseManager {
	dbConfig := &engine.DatabaseManager{
		Conns: make(map[string]*types.DBConnection),
	}

	for _, svc := range spec.Services {
		db := &types.DBConfig{
			Enabled: kbx.BoolPtr(true),
		}

		var key string
		switch svc.Engine {
		case provider.EnginePostgres:
			vol, ok := db.Options["volume"].(string)
			if !ok || vol == "" {
				vol = "kubex_pgdata"
			}
			db.Options = map[string]interface{}{
				"volume": vol,
			}
			key = "kubex_db"
			db.Enabled = kbx.BoolPtr(true)
			db.Type = "postgresql"
			db.Name = "kubex_db"
			db.User = "kubex_adm"
			db.Pass = spec.Secrets["pg_admin"]
			db.Host = "127.0.0.1"
			if port, ok := spec.PreferredPort["pg"]; ok {
				db.Port = strconv.Itoa(port)
			} else {
				db.Port = "5432"
			}

		case provider.EngineMongo:
			key = "mongodb"
			db.Enabled = kbx.BoolPtr(true)
			db.Type = "mongodb"
			db.Name = "kubexdb"
			db.User = "root"
			db.Pass = spec.Secrets["mongo_root"]
			db.Host = "127.0.0.1"
			if port, ok := spec.PreferredPort["mongo"]; ok {
				db.Port = strconv.Itoa(port)
			} else {
				db.Port = "27017"
			}

		case provider.EngineRedis:
			key = "redis"
			db.Enabled = kbx.BoolPtr(true)
			db.Type = "redis"
			db.Pass = spec.Secrets["redis_pass"]
			db.Host = "127.0.0.1"
			if port, ok := spec.PreferredPort["redis"]; ok {
				db.Port = strconv.Itoa(port)
			} else {
				db.Port = "6379"
			}

		case provider.EngineRabbit:
			key = "rabbitmq"
			db.Enabled = kbx.BoolPtr(true)
			db.Type = "rabbitmq"
			db.User = "admin"
			db.Pass = spec.Secrets["rabbit_pass"]
			db.Host = "127.0.0.1"
			if port, ok := spec.PreferredPort["rabbit"]; ok {
				db.Port = strconv.Itoa(port)
			} else {
				db.Port = "5672"
			}
		}

		if key != "" {
			d, ok := engine.GetDriver(key)
			if !ok {
				return nil
			}

			dbConfig.Conns[key] = &types.DBConnection{
				Config: &types.DBConfigRT{
					DBConfig: db,
					Mutexes:  types.NewMutexesType(),
				},
				Driver: d(logz.GetLoggerZ("GoLife")),
			}
		}
	}

	return dbConfig
}

func (p *DockerStackProvider) ConvertDBConfigToSpec(dbConfig *kbx.DBConfig) (*provider.StartSpec, error) {
	spec := &provider.StartSpec{
		Services: []provider.ServiceRef{
			{
				Name:   dbConfig.Name,
				Engine: provider.Engine(dbConfig.Type),
			},
		},
		PreferredPort: map[string]int{
			dbConfig.Name: func() int {
				port, err := strconv.Atoi(dbConfig.Port)
				if err != nil {
					return 0
				}
				return port
			}(),
		},
		Secrets: map[string]string{
			"pg_admin": dbConfig.Pass,
		},
		Labels: map[string]string{
			"app": dbConfig.Name,
		},
	}

	return spec, nil
}

// ExtractEndpoints converts legacy DBConfig to new Endpoint format
func (p *DockerStackProvider) ExtractEndpoints(cfg *engine.DatabaseManager) (map[string]provider.Endpoint, error) {
	endpoints := make(map[string]provider.Endpoint)

	if cfg == nil {
		return nil, nil
	}
	if len(cfg.Conns) == 0 {
		cn, ok := cfg.GetDefault()
		if !ok {
			return nil, fmt.Errorf("no default database connection found")
		}
		cfg.Conns[cn.Config.Name] = cn
	}

	for _, db := range cfg.Conns {
		if db == nil || !kbx.DefaultTrue(db.Config.Enabled) {
			continue
		}

		var ep provider.Endpoint
		var name string

		switch db.Config.Type {
		case "postgresql", "postgres":
			name = "kubex_db"
			ep.Port = db.Config.Port
			ep.Host = db.Config.Host
			ep.DSN = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
				db.Config.User, db.Config.Pass, db.Config.Host, ep.Port, db.Config.Name)
			ep.Redacted = fmt.Sprintf("postgres://%s:***@%s:%s/%s", db.Config.User, db.Config.Host, ep.Port, db.Config.Name)
			// endpoints["pg"] = ep

		case "mongodb":
			name = "mongo"
			port := db.Config.Port
			// if portInt, ok := port.(int); ok {
			ep.Port = port
			// } else if portStr, ok := port.(string); ok {
			// fmt.Sscanf(portStr, "%d", &ep.Port)
			// }
			ep.Host = db.Config.Host
			ep.DSN = fmt.Sprintf("mongodb://%s:%s@%s:%s", db.Config.User, db.Config.Pass, db.Config.Host, ep.Port)
			ep.Redacted = fmt.Sprintf("mongodb://%s:***@%s:%s", db.Config.User, db.Config.Host, ep.Port)

		case "redis":
			name = "redis"
			port := db.Config.Port
			// if portInt, ok := port.(int); ok {
			ep.Port = port
			// } else if portStr, ok := port.(string); ok {
			// fmt.Sscanf(portStr, "%d", &ep.Port)
			// }
			ep.Host = db.Config.Host
			ep.DSN = fmt.Sprintf("redis://:%s@%s:%s", db.Config.Pass, db.Config.Host, ep.Port)
			ep.Redacted = fmt.Sprintf("redis://:***@%s:%s", db.Config.Host, ep.Port)

		case "rabbitmq":
			name = "rabbit"
			port := db.Config.Port
			// if portInt, ok := port.(int); ok {
			ep.Port = port
			// } else if portStr, ok := port.(string); ok {
			// fmt.Sscanf(portStr, "%d", &ep.Port)
			// }
			ep.Host = db.Config.Host
			ep.DSN = fmt.Sprintf("amqp://%s:%s@%s:%s/", db.Config.User, db.Config.Pass, db.Config.Host, ep.Port)
			ep.Redacted = fmt.Sprintf("amqp://%s:***@%s:%s/", db.Config.User, db.Config.Host, ep.Port)
		}

		if name != "" {
			endpoints[name] = ep
		}
	}

	return endpoints, nil
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

func (p *DockerStackProvider) PrepareMigrations(ctx context.Context, conn *types.DBConnection) error {
	if conn == nil || conn.Config == nil {
		return fmt.Errorf("invalid database connection")
	}
	if conn.Config.Type != "postgresql" && conn.Config.Type != "postgres" {
		return fmt.Errorf("migrations only supported for PostgreSQL")
	}
	if !conn.Driver.Ping(ctx) {
		return fmt.Errorf("database is not reachable")
	}
	if err := conn.Driver.Connect(ctx, conn.Config.DBConfig); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	migrationManager := NewMigrationManager(conn.Config.DSN, p.logger)

	// Wait for PostgreSQL to be ready
	if err := migrationManager.WaitForPostgres(ctx, 30*time.Second); err != nil {
		return err
	}

	logz.Log("info", "Validating PostgreSQL connection for migrations...")
	if err := migrationManager.ValidateConnection(); err != nil {
		return fmt.Errorf("failed to validate connection: %w", err)
	}

	logz.Log("info", "PostgreSQL migrations ready to be executed.")
	return nil
}

func (p *DockerStackProvider) RunMigrations(ctx context.Context, conn *types.DBConnection, migrationInfo *kbx.MigrationInfo) error {
	if conn == nil || conn.Config == nil {
		return fmt.Errorf("invalid database connection")
	}
	if conn.Config.Type != "postgresql" && conn.Config.Type != "postgres" {
		return fmt.Errorf("migrations only supported for PostgreSQL")
	}

	migrationManager := NewMigrationManager(conn.Config.DSN, p.logger)

	// Wait for PostgreSQL to be ready
	if err := migrationManager.WaitForPostgres(ctx, 30*time.Second); err != nil {
		return err
	}

	// logz.Log("info", "Running PostgreSQL migrations...")
	// err := any(*migrationManager).(provider.MigratableProvider).RunMigrations(ctx, conn, migrationInfo)
	// if err != nil {
	// 	return fmt.Errorf("no migration results returned")
	// }

	results, err := migrationManager.RunMigrations(ctx, migrationInfo)
	if err != nil {
		return fmt.Errorf("migrations failed: %w", err)
	}

	// Log final summary
	totalSuccess := 0
	totalFailed := 0
	for _, r := range results {
		totalSuccess += r.SuccessfulStmts
		totalFailed += r.FailedStmts
	}

	if totalFailed > 0 {
		logz.Log("warn", fmt.Sprintf("Migration completed with partial success: %d succeeded, %d failed", totalSuccess, totalFailed))
		// Don't return error for partial failures - let the service continue
	} else {
		logz.Log("info", fmt.Sprintf("All migrations completed successfully! (%d statements)", totalSuccess))
	}

	return nil
}

// StartServices implements RootConfigProvider interface.
// This is the complete orchestration flow that:
// 1. Starts Docker containers for all enabled databases
// 2. Waits for database readiness
// 3. Runs migrations (if auto-migrate is enabled)
// 4. Returns only when everything is ready
func (p *DockerStackProvider) StartServices(ctx context.Context, rootConfig *kbx.RootConfig) error {
	// Validate inputs
	if p.dockerService == nil {
		return fmt.Errorf("dockerService not initialized (use NewDockerStackProvider with service injection)")
	}
	if rootConfig == nil {
		return fmt.Errorf("rootConfig cannot be nil")
	}

	// ========== STEP 1: START DOCKER CONTAINERS ==========
	logz.Log("info", "🐳 Starting Docker containers...")
	if err := p.dockerService.InitializeWithConfig(ctx, rootConfig); err != nil {
		return fmt.Errorf("failed to start containers: %w", err)
	}

	// ========== STEP 2-6: WAIT + MIGRATE FOR EACH DATABASE ==========
	for _, dbConf := range rootConfig.Databases {
		// Skip disabled databases
		if !kbx.DefaultFalse(dbConf.Enabled) {
			logz.Log("debug", "⏭️  Skipping disabled database:", dbConf.Name)
			continue
		}

		// Build DSN if missing
		if dbConf.DSN == "" {
			dbConf.DSN = p.buildDSN(dbConf)
		}

		// Log database processing
		logz.Log("info", fmt.Sprintf("📦 Processing database: %s (%s)", dbConf.Name, dbConf.Type))

		// Wait for database readiness
		logz.Log("info", "⏳ Waiting for database readiness:", dbConf.Name)
		mm := NewMigrationManager(dbConf.DSN, p.logger)
		if err := mm.WaitForPostgres(ctx, 30*time.Second); err != nil {
			return fmt.Errorf("database %s not ready: %w", dbConf.Name, err)
		}

		// Run migrations if enabled
		if dbConf.Migration != nil && kbx.DefaultFalse(dbConf.Migration.Auto) {
			logz.Log("info", "🔄 Running migrations for database:", dbConf.Name)

			// Check if schema already exists (skip if so)
			exists, err := mm.SchemaExists()
			if err != nil {
				logz.Log("warn", fmt.Sprintf("Could not check schema existence: %v", err))
			}

			if exists {
				logz.Log("info", fmt.Sprintf("✅ Schema already exists for %s, skipping migrations", dbConf.Name))
				continue
			}

			// Run migrations with error recovery
			results, err := mm.RunMigrations(ctx, dbConf.Migration)
			if err != nil {
				return fmt.Errorf("migrations failed for %s: %w", dbConf.Name, err)
			}

			// Summary logging
			totalSuccess := 0
			totalFailed := 0
			for _, r := range results {
				totalSuccess += r.SuccessfulStmts
				totalFailed += r.FailedStmts
			}

			if totalFailed > 0 {
				logz.Log("warn", fmt.Sprintf("⚠️  %s: %d succeeded, %d failed",
					dbConf.Name, totalSuccess, totalFailed))
				// Don't return error - allow partial success (resilience)
			} else {
				logz.Log("success", fmt.Sprintf("✅ %s: all %d statements executed successfully",
					dbConf.Name, totalSuccess))
			}
		} else {
			logz.Log("info", "⏭️  Skipping migrations for:", dbConf.Name, "(auto-migrate disabled)")
		}
	}

	logz.Log("success", "🎉 All services started and migrated successfully")
	return nil
}

// buildDSN constructs a connection string from DBConfig.
// Helper method to avoid repeating DSN logic.
func (p *DockerStackProvider) buildDSN(db *kbx.DBConfig) string {
	switch db.Type {
	case "postgresql", "postgres":
		return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			db.User, db.Pass, db.Host, db.Port, db.Name)
	case "mongodb", "mongo":
		return fmt.Sprintf("mongodb://%s:%s@%s:%s",
			db.User, db.Pass, db.Host, db.Port)
	case "redis":
		return fmt.Sprintf("redis://:%s@%s:%s",
			db.Pass, db.Host, db.Port)
	case "rabbitmq":
		return fmt.Sprintf("amqp://%s:%s@%s:%s/",
			db.User, db.Pass, db.Host, db.Port)
	default:
		return ""
	}
}

// Note: EndpointRedacted is now available as provider.RedactDSN(dsn) utility function.

func init() {
	// Provider registration is now handled by CLI/main initialization
	// to allow proper dependency injection of dockerService.
	// See cmd/cli/migrate.go for usage pattern.
}
