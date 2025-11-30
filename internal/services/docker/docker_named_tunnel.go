// Package docker fornece funcionalidades para gerenciar túneis Docker.
package docker

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

type NamedTunnelHandle struct{ ContainerID string }

func StartNamedTunnel(
	ctx context.Context,
	cli *client.Client,
	networkName string,
	tunnelToken string, // CF Zero Trust -> Tunnel -> Token
) (*NamedTunnelHandle, error) {

	img := "cloudflare/cloudflared:latest"
	args := []string{"tunnel", "--no-autoupdate", "run"} // ingress vem do dashboard

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: img,
		Env:   []string{"TUNNEL_TOKEN=" + tunnelToken},
		Cmd:   args,
	}, &container.HostConfig{
		NetworkMode: container.NetworkMode(networkName),
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
		Resources: container.Resources{
			NanoCPUs: 250_000_000,
			Memory:   128 << 20,
		},
	}, &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			networkName: {},
		},
	}, nil, "cf-named")
	if err != nil {
		return nil, err
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return nil, err
	}
	return &NamedTunnelHandle{ContainerID: resp.ID}, nil
}

func StopNamedTunnel(ctx context.Context, cli *client.Client, h *NamedTunnelHandle) error {
	timeout := 2
	_ = cli.ContainerStop(ctx, h.ContainerID, container.StopOptions{Timeout: &timeout})
	return cli.ContainerRemove(ctx, h.ContainerID, container.RemoveOptions{Force: true})
}

// cli, _ := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

// // Garanta que "pgadmin" e "cloudflared" estejam na mesma rede Docker (crie se necessário).
// h, err := StartQuickTunnel(ctx, cli, "CanalizeDS_net", "pgadmin", 80, 10*time.Second)
// if err != nil { /* lidar erro */ }
// fmt.Println("Acesse:", h.PublicURL)

// // ... depois:
// _ = StopQuickTunnel(ctx, cli, h)
