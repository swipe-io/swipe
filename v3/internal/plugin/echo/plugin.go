package echo

import (
	"github.com/mitchellh/mapstructure"

	"github.com/swipe-io/swipe/v3/internal/plugin/echo/config"
	"github.com/swipe-io/swipe/v3/internal/plugin/echo/generator"
	"github.com/swipe-io/swipe/v3/option"
	"github.com/swipe-io/swipe/v3/swipe"
)

func init() {
	swipe.RegisterPlugin(&Plugin{})
}

type Plugin struct {
	config config.Config
}

func (p *Plugin) ID() string {
	return "Echo"
}

func (p *Plugin) Configure(cfg *swipe.Config, module *option.Module, options map[string]interface{}) []error {
	p.config = config.Config{}
	if err := mapstructure.Decode(options, &p.config); err != nil {
		return []error{err}
	}

	p.config.MethodOptionsMap = map[string]config.MethodOptions{}

	for _, methodOption := range p.config.MethodOptions {
		if sig, ok := methodOption.Signature.Type.(*option.SignType); ok {
			if recvNamed, ok := sig.Recv.(*option.NamedType); ok {
				p.config.MethodOptionsMap[recvNamed.Name.Value+methodOption.Signature.Name.Value] = methodOption.MethodOptions
			}
		}
	}

	return nil
}

func (p *Plugin) Generators() ([]swipe.Generator, []error) {
	generators := []swipe.Generator{
		&generator.RoutesGenerator{
			Interfaces:    p.config.Interfaces,
			MethodOptions: p.config.MethodOptionsMap,
		},
	}
	return generators, nil
}

func (p *Plugin) Options() []byte {
	return (&config.Config{}).Options()
}
