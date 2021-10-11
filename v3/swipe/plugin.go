package swipe

import (
	"log"

	"github.com/swipe-io/swipe/v3/option"
)

type Plugin interface {
	ID() string
	Configure(cfg *Config, module *option.Module, options map[string]interface{}) []error
	Generators() ([]Generator, []error)
	Options() []byte
}

var registeredPlugins = map[string]Plugin{}

func RegisterPlugin(p Plugin) {
	if _, found := registeredPlugins[p.ID()]; found {
		log.Fatalf("plugin %q already registered", p.ID())
	}
	registeredPlugins[p.ID()] = p
}
