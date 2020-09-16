package generator

import (
	"context"
	stdtypes "go/types"
	stdstrings "strings"

	"github.com/swipe-io/swipe/v2/internal/usecase/generator"

	"github.com/swipe-io/strcase"

	"github.com/swipe-io/swipe/v2/internal/domain/model"
	"github.com/swipe-io/swipe/v2/internal/importer"
	"github.com/swipe-io/swipe/v2/internal/strings"
	"github.com/swipe-io/swipe/v2/internal/types"
	"github.com/swipe-io/swipe/v2/internal/writer"
)

type EndpointOption struct {
}

type endpoint struct {
	writer.GoLangWriter
	i              *importer.Importer
	serviceID      string
	serviceType    stdtypes.Type
	serviceMethods []model.ServiceMethod
}

func (g *endpoint) Prepare(ctx context.Context) error {
	return nil
}

func (g *endpoint) Process(ctx context.Context) error {
	var (
		contextPkg     string
		kitEndpointPkg string
	)
	typeStr := stdtypes.TypeString(g.serviceType, g.i.QualifyPkg)
	if len(g.serviceMethods) > 0 {
		contextPkg = g.i.Import("context", "context")
		kitEndpointPkg = g.i.Import("endpoint", "github.com/go-kit/kit/endpoint")
	}

	g.W("type EndpointSet struct {\n")

	for _, m := range g.serviceMethods {
		g.W("%sEndpoint %s.Endpoint\n", m.Name, kitEndpointPkg)
	}

	g.W("}\n")

	g.W("func MakeEndpointSet(s %s) EndpointSet {\n", typeStr)
	g.W("return EndpointSet{\n")
	for _, m := range g.serviceMethods {
		g.W("%[1]sEndpoint: make%[1]sEndpoint(s),\n", m.Name)
	}
	g.W("}\n")
	g.W("}\n")

	for _, m := range g.serviceMethods {
		if len(m.Params) > 0 {
			g.W("type %sRequest%s struct {\n", m.LcName, g.serviceID)
			for _, p := range m.Params {
				g.W("%s %s `json:\"%s\"`\n", strings.UcFirst(p.Name()), stdtypes.TypeString(p.Type(), g.i.QualifyPkg), strcase.ToLowerCamel(p.Name()))
			}
			g.W("}\n")
		}

		if m.ResultsNamed {
			g.W("type %sResponse%s struct {\n", m.LcName, g.serviceID)
			for _, p := range m.Results {
				name := p.Name()
				g.W("%s %s `json:\"%s\"`\n", strings.UcFirst(name), stdtypes.TypeString(p.Type(), g.i.QualifyPkg), strcase.ToLowerCamel(name))
			}
			g.W("}\n")
		}

		g.W("func make%sEndpoint(s %s", m.Name, typeStr)
		g.W(") %s.Endpoint {\n", kitEndpointPkg)
		g.W("w := func(ctx %s.Context, request interface{}) (interface{}, error) {\n", contextPkg)

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
			g.W("req := request.(%sRequest%s)\n", m.LcName, g.serviceID)
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
				g.W("%sResponse%s", m.LcName, g.serviceID)
				g.WriteStructAssign(structKeyValue(m.Results, nil))
			} else {
				g.W("result")
			}
		} else {
			g.W("nil")
		}
		g.W(" ,nil\n")
		g.W("}\n")
		g.W("return w\n")
		g.W("}\n\n")
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

func NewEndpoint(
	serviceID string,
	serviceType stdtypes.Type,
	serviceMethods []model.ServiceMethod,
) generator.Generator {
	return &endpoint{
		serviceID:      serviceID,
		serviceType:    serviceType,
		serviceMethods: serviceMethods,
	}
}
