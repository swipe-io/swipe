package generator

import (
	"context"
	stdtypes "go/types"
	stdstrings "strings"

	"github.com/swipe-io/swipe/v2/internal/types"

	"github.com/swipe-io/swipe/v2/internal/domain/model"

	"github.com/swipe-io/strcase"
	"github.com/swipe-io/swipe/v2/internal/importer"
	"github.com/swipe-io/swipe/v2/internal/strings"
	"github.com/swipe-io/swipe/v2/internal/usecase/generator"
	"github.com/swipe-io/swipe/v2/internal/writer"
)

type endpointOptionsGateway interface {
	Interfaces() model.Interfaces
}

type EndpointOption struct {
}

type endpoint struct {
	writer.GoLangWriter
	options endpointOptionsGateway
	i       *importer.Importer
}

func (g *endpoint) Prepare(ctx context.Context) error {
	return nil
}

func (g *endpoint) Process(ctx context.Context) error {
	g.writeEndpointMake()

	for i := 0; i < g.options.Interfaces().Len(); i++ {
		iface := g.options.Interfaces().At(i)
		typeStr := stdtypes.TypeString(iface.Type(), g.i.QualifyPkg)
		epSetName := iface.NameExport() + "EndpointSet"

		g.W("type %s struct {\n", epSetName)

		if len(iface.Methods()) > 0 {
			kitEndpointPkg := g.i.Import("endpoint", "github.com/go-kit/kit/endpoint")
			for _, m := range iface.Methods() {
				g.W("%sEndpoint %s.Endpoint\n", m.Name, kitEndpointPkg)
			}
		}

		g.W("}\n")

		g.W("func Make%[1]s(svc %[2]s) %[1]s {\n", epSetName, typeStr)
		g.W("return %s{\n", epSetName)
		for _, m := range iface.Methods() {
			g.W("%sEndpoint: make%sEndpoint(svc),\n", m.Name, m.NameExport)

		}
		g.W("}\n")
		g.W("}\n")

		for _, m := range iface.Methods() {
			if len(m.Params) > 0 {
				g.W("type %s struct {\n", m.NameRequest)
				for _, p := range m.Params {
					g.W("%s %s `json:\"%s\"`\n", strings.UcFirst(p.Name()), stdtypes.TypeString(p.Type(), g.i.QualifyPkg), strcase.ToLowerCamel(p.Name()))
				}
				g.W("}\n")
			}

			if m.ResultsNamed {
				g.W("type %s struct {\n", m.NameResponse)
				for _, p := range m.Results {
					name := p.Name()
					g.W("%s %s `json:\"%s\"`\n", strings.UcFirst(name), stdtypes.TypeString(p.Type(), g.i.QualifyPkg), strcase.ToLowerCamel(name))
				}
				g.W("}\n")
			}
		}
	}
	return nil
}

func (g *endpoint) Filename() string {
	return "endpoint_gen.go"
}

func (g *endpoint) OutputDir() string {
	return ""
}

func (g *endpoint) PkgName() string {
	return ""
}

func (g *endpoint) SetImporter(i *importer.Importer) {
	g.i = i
}

func (g *endpoint) writeEndpointMake() {
	for i := 0; i < g.options.Interfaces().Len(); i++ {
		iface := g.options.Interfaces().At(i)

		if len(iface.Methods()) > 0 {
			contextPkg := g.i.Import("context", "context")
			kitEndpointPkg := g.i.Import("endpoint", "github.com/go-kit/kit/endpoint")
			typeStr := stdtypes.TypeString(iface.Type(), g.i.QualifyPkg)

			for _, m := range iface.Methods() {
				g.W("func make%sEndpoint(s %s) %s.Endpoint {\n", m.NameExport, typeStr, kitEndpointPkg)
				g.W("return func (ctx %s.Context, request interface{}) (interface{}, error) {\n", contextPkg)

				var callParams []string
				if m.ParamCtx != nil {
					callParams = append(callParams, "ctx")
				}
				callParams = append(callParams, types.Params(m.Params, func(p *stdtypes.Var) []string {
					name := p.Name()
					name = stdstrings.ToUpper(name[:1]) + name[1:]
					return []string{"req." + name}
				}, nil)...)

				if len(m.Params) > 0 {
					g.W("req := request.(%s)\n", m.NameRequest)
				}

				if len(m.Results) > 0 {
					if m.ResultsNamed {
						for i, p := range m.Results {
							if i > 0 {
								g.W(", ")
							}
							g.W(p.Name())
						}
					} else {
						g.W("result")
					}
					g.W(", ")
				}
				if m.ReturnErr != nil {
					g.W("err")
				}
				if len(m.Results) > 0 || m.ReturnErr != nil {
					g.W(" := ")
				}
				g.WriteFuncCall("s", m.Name, callParams)
				if m.ReturnErr != nil {
					g.WriteCheckErr(func() {
						g.W("return nil, err\n")
					})
				}
				g.W("return ")
				if len(m.Results) > 0 {
					if m.ResultsNamed {
						g.W("%s", m.NameResponse)
						g.WriteStructAssign(structKeyValue(m.Results, nil))
					} else {
						g.W("result")
					}
				} else {
					g.W("nil")
				}
				g.W(" ,nil\n")
				g.W("}\n\n")
				g.W("}\n\n")
			}
		}
	}
}

func NewEndpoint(options endpointOptionsGateway) generator.Generator {
	return &endpoint{
		options: options,
	}
}
