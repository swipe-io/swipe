package gokit

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"

	"github.com/imdario/mergo"
	"github.com/mitchellh/mapstructure"

	"github.com/swipe-io/strcase"
	"github.com/swipe-io/swipe/v2/internal/plugin/gokit/config"
	"github.com/swipe-io/swipe/v2/internal/plugin/gokit/generator"
	"github.com/swipe-io/swipe/v2/option"
	"github.com/swipe-io/swipe/v2/swipe"
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
	_, appName := path.Split(module.Path)
	p.config.AppName = strcase.ToCamel(appName)

	funcDeclTypes := makeFuncDeclTypes(cfg.Packages)

	p.config.IfaceErrors = findIfaceErrors(funcDeclTypes, cfg.Packages, p.config.Interfaces)
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
		p.config.MethodOptionsMap[recvNamed.Name.Value+methodOption.Signature.Name.Value] = methodOption
	}

	p.config.OpenapiMethodTags = map[string][]string{}

	for _, o := range p.config.OpenapiTags {
		for _, m := range o.Methods {
			sig := m.Type.(*option.SignType)
			recv := sig.Recv.(*option.NamedType)
			p.config.OpenapiMethodTags[recv.Name.Value+m.Name.Value] = o.Tags
		}
	}

	pkgJsonFilepath := filepath.Join(cfg.WorkDir, "package.json")
	data, err := ioutil.ReadFile(pkgJsonFilepath)
	if err == nil {
		var packageJSON map[string]interface{}
		err := json.Unmarshal(data, &packageJSON)
		if err == nil {
			if name, ok := packageJSON["name"].(string); ok {
				p.config.JSPkgImportPath = name
			}
		} else {
			errs = append(errs, err)
		}
	}
	errs = append(errs, p.fillInterfacesByInternal(cfg)...)
	return errs
}

func (p *Plugin) fillInterfacesByInternal(cfg *swipe.Config) (errs []error) {
	for _, iface := range p.config.Interfaces {
		if iface.Named.Pkg.Module == nil {
			errs = append(errs, errors.New("not add package for "+iface.Named.Pkg.Path+"."+iface.Named.Name.Value))
			continue
		}

		configCache := map[string]*config.Config{}

		if iface.Named.Pkg.Module.External {
			p.config.HasExternal = true
			cfg.WalkBuilds(func(module *option.Module, build *option.Build) bool {
				if !module.External {
					return true
				}
				if options, ok := build.Option["Gokit"]; ok {
					c := configCache[build.Pkg.Path]
					if c == nil {
						if err := mapstructure.Decode(options, &c); err != nil {
							errs = append(errs, err)
							return true
						}
					}
					for _, iface := range p.config.Interfaces {
						for _, extIface := range c.Interfaces {
							if iface.Named.ID() == extIface.Named.ID() {
								iface.External = &config.ExternalInterface{
									Iface:  extIface,
									Config: c,
									Build:  build,
								}
							}
						}
					}
				}
				return true
			})
		}
	}
	return
}

func (p *Plugin) validateConfig() (errs []error) {
	for _, iface := range p.config.Interfaces {
		if _, ok := iface.Named.Type.(*option.IfaceType); !ok {
			errs = append(errs, fmt.Errorf("type is not an interface"))
		}
	}
	return
}

func (p *Plugin) Generators() (result []swipe.Generator, errs []error) {
	goClientEnable := p.config.ClientsEnable.Langs.Contains("go")
	jsClientEnable := p.config.ClientsEnable.Langs.Contains("js")
	jsonRPCEnable := p.config.JSONRPCEnable != nil
	httpServerEnable := p.config.HTTPServer != nil
	useFast := p.config.HTTPFast != nil
	jsonrpcDocEnable := p.config.JSONRPCDocEnable != nil

	result = append(result,
		&generator.Helpers{
			Interfaces:       p.config.Interfaces,
			JSONRPCEnable:    jsonRPCEnable,
			GoClientEnable:   goClientEnable,
			HTTPServerEnable: httpServerEnable,
			UseFast:          useFast,
			IfaceErrors:      p.config.IfaceErrors,
		},
		&generator.Endpoint{
			Interfaces:       p.config.Interfaces,
			HTTPServerEnable: httpServerEnable,
		},
		&generator.InterfaceGenerator{
			Interfaces: p.config.Interfaces,
		},
	)

	if httpServerEnable {
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
		if p.config.OpenapiEnable != nil {
			result = append(result, &generator.Openapi{
				JSONRPCEnable:        jsonRPCEnable,
				Contact:              p.config.OpenapiContact,
				Info:                 p.config.OpenapiInfo,
				MethodTags:           p.config.OpenapiMethodTags,
				Licence:              p.config.OpenapiLicence,
				Servers:              p.config.OpenapiServers,
				Output:               p.config.OpenapiOutput.Value,
				Interfaces:           p.config.Interfaces,
				MethodOptions:        p.config.MethodOptionsMap,
				DefaultMethodOptions: p.config.MethodDefaultOptions,
				IfaceErrors:          p.config.IfaceErrors,
			})
		}
		if jsonRPCEnable {
			result = append(result, &generator.JSONRPCServerGenerator{
				UseFast:              useFast,
				Interfaces:           p.config.Interfaces,
				MethodOptions:        p.config.MethodOptionsMap,
				DefaultMethodOptions: p.config.MethodDefaultOptions,
				DefaultErrorEncoder:  p.config.DefaultErrorEncoder.Value,
				JSONRPCPath:          p.config.JSONRPCPath.Value,
			})
			if jsClientEnable {
				result = append(result, &generator.JSONRPCJSClientGenerator{
					Interfaces:  p.config.Interfaces,
					IfaceErrors: p.config.IfaceErrors,
				})
			}
			if jsonrpcDocEnable {
				result = append(result, &generator.JSONRPCDocGenerator{
					AppName:         p.config.AppName,
					JSPkgImportPath: p.config.JSPkgImportPath,
					Interfaces:      p.config.Interfaces,
					IfaceErrors:     p.config.IfaceErrors,
				})
			}
			if p.config.HasExternal {
				result = append(result, &generator.GatewayGenerator{
					Interfaces: p.config.Interfaces,
				})
			}
		} else {
			result = append(result, &generator.RESTServerGenerator{
				UseFast:              useFast,
				JSONRPCEnable:        jsonRPCEnable,
				MethodOptions:        p.config.MethodOptionsMap,
				DefaultMethodOptions: p.config.MethodDefaultOptions,
				DefaultErrorEncoder:  p.config.DefaultErrorEncoder.Value,
				Interfaces:           p.config.Interfaces,
			})
		}
	}

	if goClientEnable {
		result = append(result, &generator.ClientStruct{
			UseFast:       useFast,
			JSONRPCEnable: jsonRPCEnable,
			Interfaces:    p.config.Interfaces,
		})
		if jsonRPCEnable {
			result = append(result, &generator.JSONRPCClientGenerator{
				Interfaces:           p.config.Interfaces,
				UseFast:              useFast,
				MethodOptions:        p.config.MethodOptionsMap,
				DefaultMethodOptions: p.config.MethodDefaultOptions,
			})
		} else {
			result = append(result, &generator.RESTClientGenerator{
				Interfaces:           p.config.Interfaces,
				UseFast:              useFast,
				MethodOptions:        p.config.MethodOptionsMap,
				DefaultMethodOptions: p.config.MethodDefaultOptions,
			})
		}
	}
	return
}
