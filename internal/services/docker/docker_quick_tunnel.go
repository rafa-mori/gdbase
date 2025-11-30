// Package docker provides functionality for managing Docker tunnels.
package docker

import (
	"bufio"
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

var cfURL = regexp.MustCompile(`https://[a-z0-9-]+\.trycloudflare\.com`)

type QuickTunnelHandle struct {
	ContainerID string
	PublicURL   string
}

func StartQuickTunnel(
	ctx context.Context,
	cli *client.Client,
	networkName string, // rede onde estão os serviços
	targetServiceDNS string, // ex: "pgadmin" (nome DNS do container)
	targetPort int, // ex: 80
	timeout time.Duration, // ex: 10 * time.Second
) (*QuickTunnelHandle, error) {

	img := "cloudflare/cloudflared:latest"
	// puxe a imagem se quiser (opcional)
	// _, _ = cli.ImagePull(ctx, img, types.ImagePullOptions{})

	args := []string{
		"tunnel", "--no-autoupdate",
		"--url", fmt.Sprintf("http://%s:%d", targetServiceDNS, targetPort),
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: img,
		Cmd:   args,
		// importante: mesma rede do serviço, pra resolver o DNS "targetServiceDNS"
	}, &container.HostConfig{
		NetworkMode: container.NetworkMode(networkName),
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
		Resources: container.Resources{
			NanoCPUs: 250_000_000, // 0.25 CPU
			Memory:   128 << 20,   // 128MB
		},
	}, &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			networkName: {Aliases: []string{"cloudflared"}},
		},
	}, nil, "cf-quick-%d")

	if err != nil {
		return nil, err
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return nil, err
	}

	// Captura a URL dos logs
	logs, err := cli.ContainerLogs(ctx, resp.ID, container.LogsOptions{
		ShowStdout: true, ShowStderr: true, Follow: true, Tail: "50",
	})
	if err != nil {
		return nil, err
	}

	done := make(chan string, 1)
	go func() {
		defer logs.Close()
		sc := bufio.NewScanner(logs)
		for sc.Scan() {
			if m := cfURL.Find(sc.Bytes()); m != nil {
				done <- string(m)
				return
			}
		}
	}()

	select {
	case u := <-done:
		return &QuickTunnelHandle{ContainerID: resp.ID, PublicURL: u}, nil
	case <-time.After(timeout):
		_ = cli.ContainerStop(context.Background(), resp.ID, container.StopOptions{})
		return nil, fmt.Errorf("timeout aguardando URL do cloudflared")
	}
}

func StopQuickTunnel(ctx context.Context, cli *client.Client, h *QuickTunnelHandle) error {
	sec := (2 * time.Second).Seconds()
	if timeout := int(sec); timeout > 0 {
		_ = cli.ContainerStop(ctx, h.ContainerID, container.StopOptions{Timeout: &timeout})
	}
	return cli.ContainerRemove(ctx, h.ContainerID, container.RemoveOptions{Force: true})
}

// cli, _ := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

// // Garantir que "pgadmin" e "cloudflared" estejam na mesma rede Docker.
// h, err := StartQuickTunnel(ctx, cli, "KubexDS_net", "pgadmin", 80, 10*time.Second)
// if err != nil { /* lidar erro */ }
// logz.Log("info", "Cloudflare Tunnel URL: "+h.PublicURL)

// // Para parar:
// _ = StopQuickTunnel(ctx, cli, h)
