package flavors

import (
	"github.com/kubex-ecosystem/gdbase/internal/engine"
	"github.com/kubex-ecosystem/gdbase/internal/provider"
	logz "github.com/kubex-ecosystem/logz"

	"github.com/kubex-ecosystem/gdbase/internal/types"
)

var registry = map[string]provider.Provider{}

func Register(p provider.Provider)              { registry[p.Name()] = p }
func Get(name string) (provider.Provider, bool) { p, ok := registry[name]; return p, ok }
func All() []provider.Provider {
	out := make([]provider.Provider, 0, len(registry))
	for _, p := range registry {
		out = append(out, p)
	}
	return out
}

// init com todos os flavors
func init() {
	engine.RegisterDriver("postgres", func(logger *logz.LoggerZ) types.Driver { return NewPostgresDriver(logger) })
	engine.RegisterDriver("mongodb", func(logger *logz.LoggerZ) types.Driver { return NewMongoDriver(logger) })
	engine.RegisterDriver("redis", func(logger *logz.LoggerZ) types.Driver { return NewRedisDriver(logger) })
	engine.RegisterDriver("rabbitmq", func(logger *logz.LoggerZ) types.Driver { return NewRabbitDriver(logger) })
}
