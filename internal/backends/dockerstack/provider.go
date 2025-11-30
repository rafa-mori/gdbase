// Package dockerstack provides a local Docker-based stack implementation
package dockerstack

import (
	ci "github.com/kubex-ecosystem/gdbase/internal/interfaces"
)

// Provider is an alias to the real implementation in adapter.go
// This file exists for backward compatibility
type Provider = DockerStackProvider

// New creates a new dockerstack provider instance with injected dockerService.
// The dockerService parameter is required for container orchestration.
func New(dockerService ci.IDockerService) *Provider {
	return NewDockerStackProvider(dockerService)
}
