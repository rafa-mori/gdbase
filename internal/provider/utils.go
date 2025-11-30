// Package provider utilities for DSN manipulation and endpoint management.
package provider

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/kubex-ecosystem/gdbase/internal/module/kbx"
)

// RedactDSN removes sensitive information (passwords) from a DSN string.
// Useful for safe logging of connection strings.
//
// Example:
//
//	postgres://user:secretpass@localhost:5432/db // pragma: allowlist secret
//	-> postgres://user:***@localhost:5432/db
func RedactDSN(dsn string) string {
	// Regex: match protocol://username:password@host // pragma: allowlist secret
	re := regexp.MustCompile(`://([^:]+):([^@]+)@`)
	return re.ReplaceAllString(dsn, "://$1:***@")
}

// BuildDSN constructs a connection string from database configuration components.
// Supports: PostgreSQL, MongoDB, Redis, RabbitMQ.
func BuildDSN(dbConfig *kbx.DBConfig) string {
	if dbConfig == nil {
		return ""
	}

	switch dbConfig.Type {
	case "postgresql", "postgres":
		return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", // pragma: allowlist secret
			dbConfig.User, dbConfig.Pass, dbConfig.Host, dbConfig.Port, dbConfig.Name)

	case "mongodb", "mongo":
		return fmt.Sprintf("mongodb://%s:%s@%s:%s", // pragma: allowlist secret
			dbConfig.User, dbConfig.Pass, dbConfig.Host, dbConfig.Port)

	case "redis":
		return fmt.Sprintf("redis://:%s@%s:%s", // pragma: allowlist secret
			dbConfig.Pass, dbConfig.Host, dbConfig.Port)

	case "rabbitmq":
		return fmt.Sprintf("amqp://%s:%s@%s:%s/", // pragma: allowlist secret
			dbConfig.User, dbConfig.Pass, dbConfig.Host, dbConfig.Port)

	default:
		return ""
	}
}

// BuildEndpoint creates an Endpoint from a DBConfig with DSN generation and redaction.
func BuildEndpoint(dbConfig *kbx.DBConfig) Endpoint {
	dsn := BuildDSN(dbConfig)
	return Endpoint{
		DSN:      dsn,
		Redacted: RedactDSN(dsn),
		Host:     dbConfig.Host,
		Port:     dbConfig.Port,
	}
}

// ConvertRootConfigToStartSpec translates a RootConfig into a StartSpec.
// This helper is useful for CLI commands that need to convert high-level
// configuration into provider-specific startup specifications.
func ConvertRootConfigToStartSpec(rootConfig *kbx.RootConfig) StartSpec {
	spec := StartSpec{
		Services:      []ServiceRef{},
		PreferredPort: map[string]int{},
		Secrets:       map[string]string{},
		Labels:        map[string]string{},
	}

	if rootConfig == nil {
		return spec
	}

	for _, db := range rootConfig.Databases {
		// Skip disabled databases
		if !kbx.DefaultFalse(db.Enabled) {
			continue
		}

		// Map database type to engine
		var engine Engine
		switch db.Type {
		case "postgresql", "postgres":
			engine = EnginePostgres
		case "mongodb", "mongo":
			engine = EngineMongo
		case "redis":
			engine = EngineRedis
		case "rabbitmq":
			engine = EngineRabbit
		default:
			continue // Skip unknown types
		}

		// Add service reference
		spec.Services = append(spec.Services, ServiceRef{
			Name:   db.Name,
			Engine: engine,
		})

		// Parse and add preferred port
		if port, err := strconv.Atoi(db.Port); err == nil {
			spec.PreferredPort[db.Name] = port
		}

		// Add secrets (passwords)
		if db.Pass != "" {
			spec.Secrets[db.Name+"_pass"] = db.Pass
		}

		// Add labels
		if db.Name != "" {
			spec.Labels["db_"+db.Name] = string(db.Type)
		}
	}

	return spec
}
