// Package backends registers all available backend providers
package backends

import (
	"github.com/kubex-ecosystem/gdbase/internal/backends/dockerstack"
	"github.com/kubex-ecosystem/gdbase/internal/interfaces"
	"github.com/kubex-ecosystem/gdbase/internal/provider"
	"github.com/kubex-ecosystem/gdbase/internal/provider/flavors"
)

func init() {
	var dockerService interfaces.IDockerService
	// TODO: Initialize dockerService with a concrete implementation if needed
	registerProviders(
		dockerstack.New(dockerService),
	)
}

func registerProviders(providers ...provider.Provider) {
	for _, p := range providers {
		if p == nil || p.Name() == "" {
			continue
		}
		if _, exists := flavors.Get(p.Name()); exists {
			continue
		}
		flavors.Register(p)
	}
}

func GetProvider(name string) (provider.Provider, bool) {
	return flavors.Get(name)
}

func ListProviders() []provider.Provider {
	return flavors.All()
}

func DefaultProvider() provider.Provider {
	if p, exists := flavors.Get("dockerstack"); exists {
		return p
	}
	return nil
}

func DefaultProviderName() string {
	return "dockerstack"
}

func IsDefaultProvider(name string) bool {
	return name == DefaultProviderName()
}
