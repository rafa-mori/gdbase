// Package factory fornece uma fábrica para criar serviços Docker.
package factory

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/client"
	dksk "github.com/kubex-ecosystem/gdbase/internal/backends/dockerstack"
	ci "github.com/kubex-ecosystem/gdbase/internal/interfaces"
	dkrs "github.com/kubex-ecosystem/gdbase/internal/services/docker"
	"github.com/kubex-ecosystem/logz"
)

type DockerSrv = ci.IDockerService

func NewDockerService(logger *logz.LoggerZ) (DockerSrv, error) {
	return dkrs.NewDockerService(logger)
}

type TunnelMode string

const (
	TunnelQuick TunnelMode = "quick" // HTTP efêmero (URL dinâmica)
	TunnelNamed TunnelMode = "named" // HTTP+TCP fixo (Access)
)

type CloudflaredOpts struct {
	Mode        TunnelMode
	NetworkName string
	TargetDNS   string // quick: service DNS a expor
	TargetPort  int    // quick: porta HTTP do alvo
	Token       string // named: TUNNEL_TOKEN
}

type TunnelHandle interface {
	Stop(ctx context.Context) error
}

func (o CloudflaredOpts) Start(ctx context.Context, cli *client.Client) (TunnelHandle, string /*URL ou hostname*/, error) {
	switch o.Mode {
	case TunnelQuick:
		h, err := dkrs.StartQuickTunnel(ctx, cli, o.NetworkName, o.TargetDNS, o.TargetPort, 10*time.Second)
		if err != nil {
			return nil, "", err
		}
		return tunnelStopFunc(func(ctx context.Context) error { return dkrs.StopQuickTunnel(ctx, cli, h) }), h.PublicURL, nil
	case TunnelNamed:
		h, err := dkrs.StartNamedTunnel(ctx, cli, o.NetworkName, o.Token)
		if err != nil {
			return nil, "", err
		}
		// hostnames são os que você criou no dashboard (exibir na UI)
		return tunnelStopFunc(func(ctx context.Context) error { return dkrs.StopNamedTunnel(ctx, cli, h) }), "(use seus hostnames do tunnel)", nil
	default:
		return nil, "", fmt.Errorf("modo inválido")
	}
}

type tunnelStopFunc func(ctx context.Context) error

func (f tunnelStopFunc) Stop(ctx context.Context) error { return f(ctx) }

func NewCloudflaredOpts(mode TunnelMode, networkName, targetDNS string, targetPort int, token string) CloudflaredOpts {
	return CloudflaredOpts{
		Mode:        mode,
		NetworkName: networkName,
		TargetDNS:   targetDNS,
		TargetPort:  targetPort,
		Token:       token,
	}
}

type DockerStackProvider = dksk.Provider
type MigrationManager = dksk.MigrationManager
type MigrationResult = dksk.MigrationResult
type SQLStatement = dksk.SQLStatement
type StatementError = dksk.StatementError

func NewDockerStackProvider() *DockerStackProvider {
	dockerService, dockerServiceErr := dkrs.NewDockerService(nil)
	if dockerServiceErr != nil {
		return nil
	}
	return dksk.New(dockerService)
}

func NewMigrationManager(dsn string, logger *logz.LoggerZ) *MigrationManager {
	return dksk.NewMigrationManager(dsn, logger)
}

func CreateMigrationManager(dsn string, logger *logz.LoggerZ) *MigrationManager {
	return dksk.NewMigrationManager(dsn, logger)
}
