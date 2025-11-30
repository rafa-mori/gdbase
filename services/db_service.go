// Package services provides an interface for database service operations.
package services

import (
	"context"

	ci "github.com/kubex-ecosystem/gdbase/internal/interfaces"
	"github.com/kubex-ecosystem/gdbase/internal/module/kbx"
	svc "github.com/kubex-ecosystem/gdbase/internal/services"
	t "github.com/kubex-ecosystem/gdbase/internal/types"
	"github.com/kubex-ecosystem/logz"
)

type DBConfig = svc.DBConfig
type IDBConfig interface {
	svc.DBConfig
}

type IDBService interface{ svc.DBService }
type DBServiceImpl = svc.DBServiceImpl
type DBService = svc.DBService

type IDockerService = ci.IDockerService
type DockerService = svc.Services

func NewDatabaseService(ctx context.Context, config *svc.DBConfig, logger *logz.LoggerZ) (svc.DBService, error) {
	return svc.NewDatabaseService(ctx, config, logger)
}

func SetupDatabaseServices(ctx context.Context, d ci.IDockerService, config *kbx.RootConfig) error {
	// return svc.
	return d.InitializeWithConfig(ctx, config)
}

func NewDBConfigWithArgs(ctx context.Context, dbName, dbConfigFilePath string, autoMigrate bool, logger *logz.LoggerZ, debug bool) *svc.DBConfig {
	return svc.NewDBConfigWithArgs(ctx, dbName, dbConfigFilePath, autoMigrate, logger, debug)
}
func NewDBConfigFromFile(ctx context.Context, dbConfigFilePath string, autoMigrate bool, logger *logz.LoggerZ, debug bool) (*svc.DBConfig, error) {
	return svc.NewDBConfigFromFile(ctx, dbConfigFilePath, autoMigrate, logger, debug)
}

type DatabaseType = t.Database
type IDatabase interface {
	t.Database
}
type DatabaseObj = *t.Database

type RabbitMQ = t.RabbitMQ
type IRabbitMQ interface {
	t.RabbitMQ
}
type RabbitMQObj = *t.RabbitMQ

type Postgres = t.Database
type IPostgres interface {
	t.Database
}
type PostgresObj = *t.Database

type MySQL = t.Database
type IMySQL interface {
	t.Database
}
type MySQLObj = *t.Database

type SQLite = t.Database
type ISQLite interface {
	t.Database
}
type SQLiteObj = *t.Database

type MongoDB = t.MongoDB
type IMongoDB interface {
	t.MongoDB
}
type MongoDBObject = *t.MongoDB

type Redis = t.Redis
type IRedis interface {
	t.Redis
}
type RedisObj = *t.Redis

type EnvironmentType = t.Environment

type IJSONB interface {
	ci.IJSONB
}
type JSONBObj = ci.IJSONB
type JSONBImpl = t.JSONBImpl
