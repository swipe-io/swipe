package es

import (
	"github.com/mitchellh/mapstructure"

	"github.com/swipe-io/swipe/v3/internal/plugin/es/config"
	"github.com/swipe-io/swipe/v3/internal/plugin/es/generator"
	"github.com/swipe-io/swipe/v3/option"
	"github.com/swipe-io/swipe/v3/swipe"
)

//func init() {
//swipe.RegisterPlugin(&Plugin{})
//}

type Plugin struct {
	config config.Config
}

func (p *Plugin) ID() string {
	return "EventSourcing"
}

func (p *Plugin) Configure(cfg *swipe.Config, module *option.Module, options map[string]interface{}) []error {
	p.config = config.Config{}
	if err := mapstructure.Decode(options, &p.config); err != nil {
		return []error{err}
	}
	return nil
}

func (p *Plugin) Generators() ([]swipe.Generator, []error) {
	generators := []swipe.Generator{
		&generator.UpdateGenerator{
			Entity: p.config.Entity.Value,
		},
	}
	return generators, nil
}

func (p *Plugin) Options() []byte {
	return (&config.Config{}).Options()
}
