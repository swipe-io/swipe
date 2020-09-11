package generator

import (
	"context"
	"fmt"

	"github.com/swipe-io/swipe/v2/internal/usecase/generator"

	stdtypes "go/types"

	"github.com/swipe-io/swipe/v2/internal/domain/model"
	"github.com/swipe-io/swipe/v2/internal/importer"
	"github.com/swipe-io/swipe/v2/internal/strings"
	"github.com/swipe-io/swipe/v2/internal/types"
	"github.com/swipe-io/swipe/v2/internal/writer"
)

type clientStruct struct {
	writer.GoLangWriter
	i              *importer.Importer
	serviceID      string
	serviceMethods []model.ServiceMethod
	transport      model.TransportOption
}

func (g *clientStruct) Prepare(ctx context.Context) error {
	return nil
}

func (g *clientStruct) Process(ctx context.Context) error {
	var (
		kitHTTPPkg  string
		contextPkg  string
		endpointPkg string
	)

	clientType := fmt.Sprintf("client%s", g.serviceID)
	clientOptionType := fmt.Sprintf("%sClientOption", g.serviceID)

	if len(g.serviceMethods) > 0 {
		contextPkg = g.i.Import("context", "context")
	}

	endpointPkg = g.i.Import("endpoint", "github.com/go-kit/kit/endpoint")

	if g.transport.JsonRPC.Enable {
		if g.transport.FastHTTP {
			kitHTTPPkg = g.i.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/fasthttp/jsonrpc")
		} else {
			kitHTTPPkg = g.i.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/http/jsonrpc")
		}
	} else {
		if g.transport.FastHTTP {
			kitHTTPPkg = g.i.Import("fasthttp", "github.com/l-vitaly/go-kit/transport/fasthttp")
		} else {
			kitHTTPPkg = g.i.Import("http", "github.com/go-kit/kit/transport/http")
		}
	}

	g.W("type %s func(*%s)\n", clientOptionType, clientType)

	g.WriteFunc(
		g.serviceID+"GenericClientOptions",
		"",
		[]string{"opt", "..." + kitHTTPPkg + ".ClientOption"},
		[]string{"", clientOptionType},
		func() {
			g.W("return func(c *%s) { c.genericClientOption = opt }\n", clientType)
		},
	)

	g.WriteFunc(
		g.serviceID+"GenericClientEndpointMiddlewares",
		"",
		[]string{"opt", "..." + endpointPkg + ".Middleware"},
		[]string{"", clientOptionType},
		func() {
			g.W("return func(c *%s) { c.genericEndpointMiddleware = opt }\n", clientType)
		},
	)

	for _, m := range g.serviceMethods {
		g.WriteFunc(g.serviceID+m.Name+"ClientOptions",
			"",
			[]string{"opt", "..." + kitHTTPPkg + ".ClientOption"},
			[]string{"", clientOptionType},
			func() {
				g.W("return func(c *%s) { c.%sClientOption = opt }\n", clientType, m.LcName)
			},
		)

		g.WriteFunc(g.serviceID+m.Name+"ClientEndpointMiddlewares",
			"",
			[]string{"opt", "..." + endpointPkg + ".Middleware"},
			[]string{"", clientOptionType},
			func() {
				g.W("return func(c *%s) { c.%sEndpointMiddleware = opt }\n", clientType, m.LcName)
			},
		)
	}

	g.W("type %s struct {\n", clientType)
	for _, m := range g.serviceMethods {
		g.W("%sEndpoint %s.Endpoint\n", m.LcName, endpointPkg)
		g.W("%sClientOption []%s.ClientOption\n", m.LcName, kitHTTPPkg)
		g.W("%sEndpointMiddleware []%s.Middleware\n", m.LcName, endpointPkg)
	}
	g.W("genericClientOption []%s.ClientOption\n", kitHTTPPkg)
	g.W("genericEndpointMiddleware []%s.Middleware\n", endpointPkg)

	g.W("}\n\n")

	for _, m := range g.serviceMethods {
		var params []string

		if m.ParamCtx != nil {
			params = append(params, m.ParamCtx.Name(), stdtypes.TypeString(m.ParamCtx.Type(), g.i.QualifyPkg))
		}

		params = append(params, types.NameTypeParams(m.Params, g.i.QualifyPkg, nil)...)
		results := types.NameType(m.Results, g.i.QualifyPkg, nil)

		if m.ReturnErr != nil {
			results = append(results, "", "error")
		}

		g.WriteFunc(m.Name, "c *"+clientType, params, results, func() {
			if len(m.Results) > 0 {
				g.W("resp")
			} else {
				g.W("_")
			}
			g.W(", err := ")

			g.W("c.%sEndpoint(", m.LcName)

			if m.ParamCtx != nil {
				g.W("%s,", m.ParamCtx.Name())
			} else {
				g.W("%s.Background(),", contextPkg)
			}

			if len(m.Params) > 0 {
				g.W("%sRequest%s", m.LcName, g.serviceID)
				params := structKeyValue(m.Params, func(p *stdtypes.Var) bool {
					return !types.IsContext(p.Type())
				})
				g.WriteStructAssign(params)
			} else {
				g.W(" nil")
			}

			g.W(")\n")

			if m.ReturnErr != nil {
				g.W("if err != nil {\n")
				g.W("return ")

				if len(m.Results) > 0 {
					for i, r := range m.Results {
						if i > 0 {
							g.W(",")
						}
						g.W(types.ZeroValue(r.Type()))
					}
					g.W(",")
				}

				g.W(" err\n")

				g.W("}\n")
			}

			if len(m.Results) > 0 {
				if m.ResultsNamed {
					g.W("response := resp.(%sResponse%s)\n", m.LcName, g.serviceID)
				} else {
					g.W("response := resp.(%s)\n", stdtypes.TypeString(m.Results[0].Type(), g.i.QualifyPkg))
				}
			}

			g.W("return ")

			if len(m.Results) > 0 {
				if m.ResultsNamed {
					for i, r := range m.Results {
						if i > 0 {
							g.W(",")
						}
						g.W("response.%s", strings.UcFirst(r.Name()))
					}
				} else {
					g.W("response")
				}
				g.W(", ")
			}
			if m.ReturnErr != nil {
				g.W("nil")
			}
			g.W("\n")
		})
	}
	return nil
}

func (g *clientStruct) PkgName() string {
	return ""
}

func (g *clientStruct) OutputDir() string {
	return ""
}

func (g *clientStruct) Filename() string {
	return "client_struct_gen.go"
}

func (g *clientStruct) SetImporter(i *importer.Importer) {
	g.i = i
}

func NewClientStruct(
	serviceID string,
	serviceMethods []model.ServiceMethod,
	transport model.TransportOption,
) generator.Generator {
	return &clientStruct{
		serviceID:      serviceID,
		serviceMethods: serviceMethods,
		transport:      transport,
	}
}
