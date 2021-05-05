package swipe

import (
	"log"

	"github.com/swipe-io/swipe/v2/internal/option"
)

type Plugin interface {
	ID() string
	Configure(cfg *Config, module *option.Module, build *option.Build, options map[string]interface{}) []error
	Generators() ([]Generator, []error)
}

var registeredPlugins = map[string]Plugin{}

func RegisterPlugin(p Plugin) {
	if _, found := registeredPlugins[p.ID()]; found {
		log.Fatalf("plugin %q already registered", p.ID())
	}
	registeredPlugins[p.ID()] = p
}
