package gokit

import (
	"fmt"

	"github.com/imdario/mergo"
	"github.com/mitchellh/mapstructure"

	"github.com/swipe-io/swipe/v2/internal/option"
	"github.com/swipe-io/swipe/v2/internal/plugin/gokit/config"
	"github.com/swipe-io/swipe/v2/internal/plugin/gokit/generator"
	"github.com/swipe-io/swipe/v2/internal/swipe"
)

func init() {
	swipe.RegisterPlugin(&Plugin{})
}

type Plugin struct {
	config config.Config
}

func (p *Plugin) ID() string {
	return "Gokit"
}

func (p *Plugin) Configure(cfg *swipe.Config, module *option.Module, build *option.Build, options map[string]interface{}) []error {
	if err := mapstructure.Decode(options, &p.config); err != nil {
		return []error{err}
	}
	errs := p.validateConfig()
	if len(errs) > 0 {
		return errs
	}

	p.config.MethodOptionsMap = map[string]*config.MethodOption{}

	for _, methodOption := range p.config.MethodOptions {
		if err := mergo.Merge(methodOption, p.config.MethodDefaultOptions); err != nil {
			errs = append(errs, err)
			continue
		}
		if !p.config.LoggingEnable && methodOption.Logging.Value {
			p.config.LoggingEnable = true
		}
		if !p.config.InstrumentingEnable && methodOption.Instrumenting.Value {
			p.config.InstrumentingEnable = true
		}

		sig := methodOption.Signature.Type.(*option.SignType)
		recvNamed := sig.Recv.(*option.NamedType)

		if p.config.JSONRPCEnable == nil && methodOption.RESTPath.Value != "" {
			pathVars, err := pathVars(methodOption.RESTPath.Value)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			methodOption.RESTPathVars = pathVars
		}
		p.config.MethodOptionsMap[recvNamed.Name.Origin+methodOption.Signature.Name.Origin] = methodOption
	}

	p.config.OpenapiMethodTags = map[string][]string{}

	for _, o := range p.config.OpenapiTags {
		for _, m := range o.Methods {
			sig := m.Type.(*option.SignType)
			recv := sig.Recv.(*option.NamedType)
			p.config.OpenapiMethodTags[recv.Name.Origin+m.Name.Origin] = o.Tags
		}
	}
	return nil
}

func (p *Plugin) validateConfig() (errs []error) {
	for _, named := range p.config.Interfaces {
		if _, ok := named.Named.Type.(*option.IfaceType); !ok {
			errs = append(errs, fmt.Errorf("type is not an interface"))
		}
	}
	return
}

func (p *Plugin) Generators() (result []swipe.Generator, errs []error) {
	goClientEnable := p.config.ClientsEnable.Langs.Contains("go")
	result = append(result,
		&generator.Helpers{
			Interfaces:     p.config.Interfaces,
			JSONRPCEnable:  p.config.JSONRPCEnable != nil,
			GoClientEnable: goClientEnable,
			UseFast:        p.config.HTTPFast != nil,
		},
		&generator.Endpoint{
			Interfaces: p.config.Interfaces,
		},
		&generator.InterfaceGenerator{
			Interfaces: p.config.Interfaces,
		},
	)
	if p.config.LoggingEnable {
		result = append(result, &generator.Logging{
			Interfaces:           p.config.Interfaces,
			MethodOptions:        p.config.MethodOptionsMap,
			DefaultMethodOptions: p.config.MethodDefaultOptions,
		})
	}
	if p.config.InstrumentingEnable {
		result = append(result, &generator.Instrumenting{
			Interfaces:           p.config.Interfaces,
			MethodOptions:        p.config.MethodOptionsMap,
			DefaultMethodOptions: p.config.MethodDefaultOptions,
		})
	}
	if goClientEnable {
		result = append(result, &generator.ClientStruct{
			UseFast:       p.config.HTTPFast != nil,
			JSONRPCEnable: p.config.JSONRPCEnable != nil,
			Interfaces:    p.config.Interfaces,
		})
	}
	if p.config.OpenapiEnable != nil {

		result = append(result, &generator.Openapi{
			JSONRPCEnable:        p.config.JSONRPCEnable != nil,
			Contact:              p.config.OpenapiContact,
			Info:                 p.config.OpenapiInfo,
			MethodTags:           p.config.OpenapiMethodTags,
			Licence:              p.config.OpenapiLicence,
			Servers:              p.config.OpenapiServers,
			Output:               p.config.OpenapiOutput.Value,
			Interfaces:           p.config.Interfaces,
			MethodOptions:        p.config.MethodOptionsMap,
			DefaultMethodOptions: p.config.MethodDefaultOptions,
		})
	}

	return
}
