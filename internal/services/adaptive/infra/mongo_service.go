// Package infra provides infrastructure-related services.
package infra

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/kubex-ecosystem/gdbase/internal/module/kbx"
	"github.com/kubex-ecosystem/gdbase/internal/services/docker"
	"github.com/kubex-ecosystem/gdbase/internal/types"
	logz "github.com/kubex-ecosystem/logz"
)

// MongoBootstrapOptions controls how EnsureMongoUp behaves.
type MongoBootstrapOptions struct {
	// Database is the database name to create/use. Default: kubexds
	Database string
	// ComposeFile is the docker compose file path. Default: support/docker/mongo.yaml
	ComposeFile string
	// WaitTimeout is the readiness timeout. Default: 45s
	WaitTimeout time.Duration
	// AppUser / AppPass are the application credentials created by init.sh
	AppUser string // default: kubexds
	AppPass string // default: kubexds_pass
}

// EnsureMongoUp starts MongoDB (via docker compose) if needed and returns an application DSN.
// It requires cfg.Enabled=true. It does NOT persist any state beyond the running container.
func EnsureMongoUp(ctx context.Context, cfg *types.DBConfig, logger *logz.LoggerZ, opts *MongoBootstrapOptions) (appURI string, err error) {
	if cfg == nil {
		return "", errors.New("mongo cfg is nil")
	}
	if !kbx.DefaultTrue(cfg.Enabled) {
		return "", errors.New("mongo not enabled")
	}

	// Defaults
	if opts == nil {
		opts = &MongoBootstrapOptions{}
	}
	if opts.ComposeFile == "" {
		opts.ComposeFile = filepath.FromSlash("support/docker/mongo.yaml")
	}
	if opts.WaitTimeout <= 0 {
		opts.WaitTimeout = 45 * time.Second
	}
	if cfg.DBName == "" {
		cfg.DBName = "kubexds"
	}
	// if cfg.Volume == "" {
	// 	cfg.Volume = ".data/mongo" // relative to repo root
	// }
	appUser := opts.AppUser
	appPass := opts.AppPass
	if appUser == "" {
		appUser = "kubexds"
	}
	if appPass == "" {
		appPass = "kubexds_pass"
	}

	// Resolve admin credentials
	username := cfg.User
	if username == "" {
		username = "root"
	}
	password := cfg.Pass // pragma: allowlist secret
	if password == "" {
		password = "example" // pragma: allowlist secret
	}

	var vol string
	if volume, ok := cfg.Options["volume"]; ok {
		vol = volume.(string)
	}

	// ENV for compose
	env := map[string]string{
		"MONGO_INITDB_ROOT_USERNAME": username,
		"MONGO_INITDB_ROOT_PASSWORD": password,
		"MONGO_INITDB_DATABASE":      cfg.DBName,
		"MONGO_APP_USER":             appUser,
		"MONGO_APP_PASSWORD":         appPass,
		"MONGO_PORT":                 cfg.Port,
		"MONGO_DATA_VOLUME":          vol,
	}

	// Docker service

	ds, err := docker.NewDockerService(logger)
	if err != nil {
		logz.Log("error", fmt.Sprintf("failed to get docker service: %v", err))
		return "", err
	}

	if err := ds.StartContainer(cfg.DBName, "mongo:latest", env, nil, nil); err != nil {
		return "", err
	}

	// Build application URI (not admin)
	hostport := fmt.Sprintf("localhost:%d", cfg.Port)
	return fmt.Sprintf("mongodb://%s:%s@%s/%s", appUser, appPass, hostport, cfg.DBName), nil // pragma: allowlist secret
}
