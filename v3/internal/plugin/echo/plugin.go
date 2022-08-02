package echo

import (
	"github.com/mitchellh/mapstructure"
	"github.com/swipe-io/swipe/v3/internal/finder"
	"github.com/swipe-io/swipe/v3/internal/plugin"

	"github.com/swipe-io/swipe/v3/internal/plugin/echo/config"
	"github.com/swipe-io/swipe/v3/internal/plugin/echo/generator"
	"github.com/swipe-io/swipe/v3/option"
	"github.com/swipe-io/swipe/v3/swipe"
)

func init() {
	swipe.RegisterPlugin(new(Plugin).ID(), func() swipe.Plugin {
		return &Plugin{}
	})
}

type Plugin struct {
	config config.Config
}

func (p *Plugin) ID() string {
	return "Echo"
}

func (p *Plugin) Configure(cfg *swipe.Config, module *option.Module, options map[string]interface{}) (errs []error) {
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

	p.config.OpenapiMethodTags = map[string][]string{}

	for _, o := range p.config.OpenapiTags {
		for _, m := range o.Methods {
			sig := m.Type.(*option.SignType)
			recv := sig.Recv.(*option.NamedType)
			p.config.OpenapiMethodTags[recv.Name.Value+m.Name.Value] = o.Tags
		}
	}

	for _, iface := range p.config.Interfaces {
		ifaceType := iface.Named.Type.(*option.IfaceType)
		for _, m := range ifaceType.Methods {
			dstMethodOption, _ := p.config.MethodOptionsMap[iface.Named.Name.Value+m.Name.Value]
			//dstMethodOption = fillMethodDefaultOptions(dstMethodOption, p.config.MethodDefaultOptions)

			pathVars, err := plugin.PathVars(dstMethodOption.RESTPath.Take())
			if err != nil {
				errs = append(errs, err)
				continue
			}
			dstMethodOption.RESTPathVars = pathVars

			p.config.MethodOptionsMap[iface.Named.Name.Value+m.Name.Value] = dstMethodOption
		}
	}

	var interfaces []*option.NamedType
	for _, iface := range p.config.Interfaces {
		interfaces = append(interfaces, iface.Named)
	}
	f := finder.NewFinder(cfg.Packages, cfg.Module.Path)
	p.config.IfaceErrors = f.FindIfaceErrors(interfaces)
	return
}

func (p *Plugin) Generators() ([]swipe.Generator, []error) {
	generators := []swipe.Generator{
		&generator.RoutesGenerator{
			Interfaces:    p.config.Interfaces,
			MethodOptions: p.config.MethodOptionsMap,
		},
	}
	if p.config.OpenapiEnable != nil {
		generators = append(generators, &generator.Openapi{
			Contact:       p.config.OpenapiContact,
			Info:          p.config.OpenapiInfo,
			MethodTags:    p.config.OpenapiMethodTags,
			Licence:       p.config.OpenapiLicence,
			Servers:       p.config.OpenapiServers,
			Output:        p.config.OpenapiOutput.Take(),
			Interfaces:    p.config.Interfaces,
			MethodOptions: p.config.MethodOptionsMap,
			IfaceErrors:   p.config.IfaceErrors,
		})
	}
	return generators, nil
}

func (p *Plugin) Options() []byte {
	return (&config.Config{}).Options()
}
