package dockerstack

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	logz "github.com/kubex-ecosystem/logz"
)

type MongoProvider struct{}

func NewMongoProvider() *MongoProvider {
	return &MongoProvider{}
}

func (m *MongoProvider) Name() string {
	return "dockerstack-mongo"
}

func (m *MongoProvider) Description() string {
	return "DockerStack backend provider for MongoDB"
}

func (m *MongoProvider) DefaultEnv() map[string]string {
	return GetDefaultMongoEnv()
}

func (m *MongoProvider) DefaultURI() string {
	return GetDefaultMongoURI()
}

func (m *MongoProvider) Start(ctx context.Context, env map[string]string, composeFile string, logger *logz.LoggerZ, wait time.Duration) (stop func(context.Context) error, public string, err error) {
	return StartMongoCompose(ctx, env, composeFile, logger, wait)
}

func (m *MongoProvider) Stop(ctx context.Context, stopFunc func(context.Context) error) error {
	return stopFunc(ctx)
}

func (m *MongoProvider) HealthCheck(ctx context.Context, public string) error {
	d := net.Dialer{Timeout: 2 * time.Second}
	conn, err := d.DialContext(ctx, "tcp", public)
	if err != nil {
		return fmt.Errorf("mongo health check failed: %w", err)
	}
	_ = conn.Close()
	return nil
}

func (m *MongoProvider) Info() string {
	return "MongoDB running via Docker Stack"
}

// StartMongoCompose brings MongoDB up using docker compose and waits for readiness.
// It returns a Stop function (docker compose down -v) and the public host:port.
func StartMongoCompose(ctx context.Context, env map[string]string, composeFile string, logger *logz.LoggerZ, wait time.Duration) (stop func(context.Context) error, public string, err error) {
	if composeFile == "" {
		composeFile = filepath.FromSlash("support/docker/mongo.yaml")
	}
	abs, _ := filepath.Abs(composeFile)
	if _, statErr := os.Stat(abs); statErr != nil {
		return nil, "", fmt.Errorf("compose file not found: %s", abs)
	}

	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", abs, "up", "-d")
	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	if out, runErr := cmd.CombinedOutput(); runErr != nil {
		if logger != nil {
			logger.Error("docker compose up failed", "err", runErr, "out", string(out))
		}
		return nil, "", fmt.Errorf("compose up failed: %w", runErr)
	}
	if logger != nil {
		logger.Info("Mongo compose up dispatched", "file", abs)
	}

	// Wait readiness – TCP dial to host:port
	host := "127.0.0.1"
	port := env["MONGO_PORT"]
	if port == "" {
		port = "27017"
	}
	public = net.JoinHostPort(host, port)
	deadline := time.Now().Add(wait)
	for {
		d := net.Dialer{Timeout: 800 * time.Millisecond}
		conn, derr := d.DialContext(ctx, "tcp", public)
		if derr == nil {
			_ = conn.Close()
			break
		}
		if time.Now().After(deadline) {
			return nil, "", fmt.Errorf("mongo readiness timeout at %s", public)
		}
		time.Sleep(600 * time.Millisecond)
	}

	if logger != nil {
		logger.Info("Mongo is ready", "addr", public)
	}

	stop = func(c2 context.Context) error {
		down := exec.CommandContext(c2, "docker", "compose", "-f", abs, "down", "-v")
		down.Env = os.Environ()
		for k, v := range env {
			down.Env = append(down.Env, fmt.Sprintf("%s=%s", k, v))
		}
		if out, derr := down.CombinedOutput(); derr != nil {
			if logger != nil {
				logger.Error("docker compose down failed", "err", derr, "out", string(out))
			}
			return derr
		}
		if logger != nil {
			logger.Info("Mongo compose down done", "file", abs)
		}
		return nil
	}
	return stop, public, nil
}

func GetDefaultMongoEnv() map[string]string {
	return map[string]string{
		"MONGO_INITDB_ROOT_USERNAME": "admin",
		"MONGO_INITDB_ROOT_PASSWORD": "password",
		"MONGO_PORT":                 "27017",
	}
}

func GetDefaultMongoURI() string {
	return "mongodb://admin:password@127.0.0.1:27017" // pragma: allowlist secret
}
