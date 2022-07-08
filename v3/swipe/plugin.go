package swipe

import (
	"log"
	"sync"

	"github.com/swipe-io/swipe/v3/option"
)

type Plugin interface {
	ID() string
	Configure(cfg *Config, module *option.Module, options map[string]interface{}) []error
	Generators() ([]Generator, []error)
	Options() []byte
}

var registeredPlugins = sync.Map{}

func RegisterPlugin(id string, cb func() Plugin) {
	_, loaded := registeredPlugins.LoadOrStore(id, cb)
	if loaded {
		log.Fatalf("plugin %q already registered", id)
	}
}
