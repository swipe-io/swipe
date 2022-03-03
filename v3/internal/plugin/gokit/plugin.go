package gokit

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"

	"github.com/mitchellh/mapstructure"

	"github.com/swipe-io/strcase"
	"github.com/swipe-io/swipe/v3/internal/plugin/gokit/config"
	"github.com/swipe-io/swipe/v3/internal/plugin/gokit/generator"
	"github.com/swipe-io/swipe/v3/option"
	"github.com/swipe-io/swipe/v3/swipe"
)

func init() {
	swipe.RegisterPlugin(&Plugin{})
}

type Plugin struct {
	config  config.Config
	workdir string
}

func (p *Plugin) ID() string {
	return "Gokit"
}

func (p *Plugin) Configure(cfg *swipe.Config, module *option.Module, options map[string]interface{}) []error {
	p.config = config.Config{}
	if err := mapstructure.Decode(options, &p.config); err != nil {
		return []error{err}
	}
	errs := p.validateConfig()
	if len(errs) > 0 {
		return errs
	}
	_, appName := path.Split(module.Path)

	p.workdir = cfg.WorkDir

	p.config.AppName = strcase.ToCamel(appName)

	funcDeclTypes := makeFuncDeclTypes(cfg.Packages)
	funcDeclIfaceTypes := makeFuncIfaceDeclTypes(cfg.Packages, funcDeclTypes)
	funcErrors := findErrors(cfg.Module.Path, funcDeclTypes, cfg.Packages)

	p.config.IfaceErrors = findIfaceErrors(funcDeclTypes, funcDeclIfaceTypes, funcErrors, cfg.Packages, p.config.Interfaces)
	p.config.MethodOptionsMap = map[string]config.MethodOptions{}

	for _, methodOption := range p.config.MethodOptions {
		if sig, ok := methodOption.Signature.Type.(*option.SignType); ok {
			if recvNamed, ok := sig.Recv.(*option.NamedType); ok {
				p.config.MethodOptionsMap[recvNamed.Name.Value+methodOption.Signature.Name.Value] = methodOption.MethodOptions
			}
		}
	}

	for _, iface := range p.config.Interfaces {
		ifaceType := iface.Named.Type.(*option.IfaceType)
		for _, m := range ifaceType.Methods {
			dstMethodOption, _ := p.config.MethodOptionsMap[iface.Named.Name.Value+m.Name.Value]
			dstMethodOption = fillMethodDefaultOptions(dstMethodOption, p.config.MethodDefaultOptions)

			if !p.config.LoggingEnable && dstMethodOption.Logging.Take() {
				p.config.LoggingEnable = true
			}
			if !p.config.InstrumentingEnable && dstMethodOption.Instrumenting.Take() {
				p.config.InstrumentingEnable = true
			}

			if p.config.JSONRPCEnable == nil && dstMethodOption.RESTPath.Value != nil {
				pathVars, err := pathVars(dstMethodOption.RESTPath.Take())
				if err != nil {
					errs = append(errs, err)
					continue
				}
				dstMethodOption.RESTPathVars = pathVars
			}
			p.config.MethodOptionsMap[iface.Named.Name.Value+m.Name.Value] = dstMethodOption
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
	checkErrs, hasExternal := p.checkExternalPackage()
	if len(checkErrs) > 0 {
		errs = append(errs, checkErrs...)
	}

	p.config.HasExternal = hasExternal
	return errs
}

func (p *Plugin) Options() []byte {
	var cfg interface{} = &config.Config{}
	if o, ok := cfg.(interface{ Options() []byte }); ok {
		return o.Options()
	}
	return nil
}

func (p *Plugin) Generators() (result []swipe.Generator, errs []error) {
	goClientEnable := p.config.ClientsEnable.Langs.Contains("go")
	jsClientEnable := p.config.ClientsEnable.Langs.Contains("js")
	jsonRPCEnable := p.config.JSONRPCEnable != nil
	httpServerEnable := p.config.HTTPServer != nil
	useFast := p.config.HTTPFast != nil
	jsonRPCDocEnable := p.config.JSONRPCDocEnable != nil

	if p.config.CURLEnable != nil {
		result = append(result, &generator.CURL{
			Interfaces:    p.config.Interfaces,
			MethodOptions: p.config.MethodOptionsMap,
			JSONRPCEnable: jsonRPCEnable,
			JSONRPCPath:   p.config.JSONRPCPath.Take(),
			Output:        p.config.CURLOutput.Take(),
			URL:           p.config.CURLURL.Take(),
		})
	}
	if httpServerEnable {
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

		if p.config.LoggingEnable {
			result = append(result, &generator.Logging{
				Interfaces:    p.config.Interfaces,
				MethodOptions: p.config.MethodOptionsMap,
			})
		}
		if p.config.InstrumentingEnable {
			result = append(result, &generator.Instrumenting{
				Interfaces:    p.config.Interfaces,
				MethodOptions: p.config.MethodOptionsMap,
				Labels:        p.config.InstrumentingLabels,
			})
		}
		if p.config.OpenapiEnable != nil {
			result = append(result, &generator.Openapi{
				JSONRPCEnable: jsonRPCEnable,
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
		if p.config.HasExternal {
			result = append(result, &generator.GatewayGenerator{
				Interfaces: p.config.Interfaces,
			})
		}
		if jsonRPCEnable {
			result = append(result, &generator.JSONRPCServerGenerator{
				UseFast:             useFast,
				Interfaces:          p.config.Interfaces,
				MethodOptions:       p.config.MethodOptionsMap,
				DefaultErrorEncoder: p.config.ServerErrorEncoder.Value,
				JSONRPCPath:         p.config.JSONRPCPath.Take(),
			})
			if jsClientEnable {
				result = append(result, &generator.JSONRPCJSClientGenerator{
					Interfaces:  p.config.Interfaces,
					IfaceErrors: p.config.IfaceErrors,
				})
			}
			if jsonRPCDocEnable {
				result = append(result, &generator.JSONRPCDocGenerator{
					AppName:         p.config.AppName,
					JSPkgImportPath: p.config.JSPkgImportPath,
					Interfaces:      p.config.Interfaces,
					IfaceErrors:     p.config.IfaceErrors,
					Output:          p.config.JSONRPCDocOutput.Take(),
				})
			}

		} else {
			result = append(result, &generator.RESTServerGenerator{
				UseFast:            useFast,
				JSONRPCEnable:      jsonRPCEnable,
				MethodOptions:      p.config.MethodOptionsMap,
				ServerErrorEncoder: p.config.ServerErrorEncoder.Value,
				Interfaces:         p.config.Interfaces,
			})
		}
	}

	if goClientEnable {
		var pkg string

		output := p.config.ClientOutput.Take()
		if output != "" {
			pkg = strcase.ToSnake(filepath.Base(output))
		}

		result = append(result,
			&generator.Helpers{
				Interfaces:       p.config.Interfaces,
				JSONRPCEnable:    jsonRPCEnable,
				GoClientEnable:   goClientEnable,
				HTTPServerEnable: httpServerEnable,
				UseFast:          useFast,
				IfaceErrors:      p.config.IfaceErrors,
				Pkg:              pkg,
				Output:           output,
			},
			&generator.Endpoint{
				Interfaces:       p.config.Interfaces,
				HTTPServerEnable: httpServerEnable,
				Pkg:              pkg,
				Output:           output,
			},
			&generator.InterfaceGenerator{
				Interfaces: p.config.Interfaces,
				Pkg:        pkg,
				Output:     output,
			},
			&generator.ClientStruct{
				UseFast:       useFast,
				JSONRPCEnable: jsonRPCEnable,
				Interfaces:    p.config.Interfaces,
				Pkg:           pkg,
				Output:        output,
			})
		if jsonRPCEnable {
			result = append(result, &generator.JSONRPCClientGenerator{
				Interfaces: p.config.Interfaces,
				UseFast:    useFast,
				Pkg:        pkg,
				Output:     output,
			})
		} else {
			result = append(result, &generator.RESTClientGenerator{
				Interfaces:    p.config.Interfaces,
				UseFast:       useFast,
				MethodOptions: p.config.MethodOptionsMap,
				Pkg:           pkg,
				Output:        output,
			})
		}
	}
	return
}

func (p *Plugin) checkExternalPackage() (errs []error, hasExternal bool) {
	for _, iface := range p.config.Interfaces {
		if iface.Named.Pkg.Module == nil {
			errs = append(errs, errors.New("not add package for "+iface.Named.Pkg.Path+"."+iface.Named.Name.Value))
			continue
		}
		if iface.Gateway != nil {
			hasExternal = true
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
