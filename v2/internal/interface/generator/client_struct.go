package generator

import (
	"context"
	"fmt"
	stdtypes "go/types"

	"github.com/swipe-io/strcase"
	"github.com/swipe-io/swipe/v2/internal/domain/model"
	"github.com/swipe-io/swipe/v2/internal/importer"
	"github.com/swipe-io/swipe/v2/internal/types"
	"github.com/swipe-io/swipe/v2/internal/usecase/generator"
	"github.com/swipe-io/swipe/v2/internal/writer"
)

type clientStructOptionsGateway interface {
	UseFast() bool
	JSONRPCEnable() bool
	Interfaces() model.Interfaces
}

type clientStruct struct {
	writer.GoLangWriter
	options clientStructOptionsGateway
	i       *importer.Importer
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

	if g.options.JSONRPCEnable() {
		if g.options.UseFast() {
			kitHTTPPkg = g.i.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/fasthttp/jsonrpc")
		} else {
			kitHTTPPkg = g.i.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/http/jsonrpc")
		}
	} else {
		if g.options.UseFast() {
			kitHTTPPkg = g.i.Import("fasthttp", "github.com/l-vitaly/go-kit/transport/fasthttp")
		} else {
			kitHTTPPkg = g.i.Import("http", "github.com/go-kit/kit/transport/http")
		}
	}

	endpointPkg = g.i.Import("endpoint", "github.com/go-kit/kit/endpoint")
	clientOptionType := "ClientOption"

	if g.options.Interfaces().Len() > 1 {
		g.W("type AppClient struct {\n")
		for i := 0; i < g.options.Interfaces().Len(); i++ {
			iface := g.options.Interfaces().At(i)

			name := iface.UcName()
			if iface.Namespace() != "" {
				name = strcase.ToCamel(iface.Namespace())
			}

			clientType := name + "Client"

			g.W("%s *%s\n", name, clientType)
		}
		g.W("}\n\n")

		if g.options.JSONRPCEnable() {
			g.W("func NewClientJSONRPC(tgt string")
		} else {
			g.W("func NewClientREST(tgt string")
		}

		g.W(" ,opts ...ClientOption")
		g.W(") (*AppClient, error) {\n")

		for i := 0; i < g.options.Interfaces().Len(); i++ {
			iface := g.options.Interfaces().At(i)

			name := iface.UcName()
			if iface.Namespace() != "" {
				name = strcase.ToCamel(iface.Namespace())
			}
			lcName := strcase.ToLowerCamel(name)

			if g.options.JSONRPCEnable() {
				g.W("%s, err := NewClientJSONRPC%s(tgt, opts...)\n", lcName, name)
			} else {
				g.W("%s, err := NewClientREST%s(tgt, opts...)\n", lcName, name)
			}

			g.WriteCheckErr(func() {
				g.W("return nil, err")
			})
		}

		g.W("return &AppClient{\n")
		for i := 0; i < g.options.Interfaces().Len(); i++ {
			iface := g.options.Interfaces().At(i)
			name := iface.UcName()
			if iface.Namespace() != "" {
				name = strcase.ToCamel(iface.Namespace())
			}
			lcName := strcase.ToLowerCamel(name)

			g.W("%[1]s: %[2]s,\n", name, lcName)
		}
		g.W("}, nil\n")
		g.W("}\n\n")
	}

	g.W("type %s func(*clientOpts)\n", clientOptionType)
	g.W("type clientOpts struct {\n")
	for i := 0; i < g.options.Interfaces().Len(); i++ {
		iface := g.options.Interfaces().At(i)
		for _, m := range iface.Methods() {
			g.W("%sClientOption []%s.ClientOption\n", m.IfaceLcName, kitHTTPPkg)
			g.W("%sEndpointMiddleware []%s.Middleware\n", m.IfaceLcName, endpointPkg)
		}
	}
	g.W("genericClientOption []%s.ClientOption\n", kitHTTPPkg)
	g.W("genericEndpointMiddleware []%s.Middleware\n", endpointPkg)
	g.W("}\n\n")

	g.WriteFunc(
		"GenericClientOptions",
		"",
		[]string{"opt", "..." + kitHTTPPkg + ".ClientOption"},
		[]string{"", clientOptionType},
		func() {
			g.W("return func(c *clientOpts) { c.genericClientOption = opt }\n")
		},
	)

	g.WriteFunc(
		"GenericClientEndpointMiddlewares",
		"",
		[]string{"opt", "..." + endpointPkg + ".Middleware"},
		[]string{"", clientOptionType},
		func() {
			g.W("return func(c *clientOpts) { c.genericEndpointMiddleware = opt }\n")
		},
	)

	for i := 0; i < g.options.Interfaces().Len(); i++ {
		iface := g.options.Interfaces().At(i)
		for _, m := range iface.Methods() {
			g.WriteFunc(m.IfaceUcName+"ClientOptions",
				"",
				[]string{"opt", "..." + kitHTTPPkg + ".ClientOption"},
				[]string{"", clientOptionType},
				func() {
					g.W("return func(c *clientOpts) { c.%sClientOption = opt }\n", m.IfaceLcName)
				},
			)
			g.WriteFunc(m.IfaceUcName+"ClientEndpointMiddlewares",
				"",
				[]string{"opt", "..." + endpointPkg + ".Middleware"},
				[]string{"", clientOptionType},
				func() {
					g.W("return func(c *clientOpts) { c.%sEndpointMiddleware = opt }\n", m.IfaceLcName)
				},
			)
		}
	}

	for i := 0; i < g.options.Interfaces().Len(); i++ {
		iface := g.options.Interfaces().At(i)

		name := iface.UcName()
		if iface.Namespace() != "" {
			name = strcase.ToCamel(iface.Namespace())
		}

		clientType := fmt.Sprintf("%sClient", name)

		contextPkg = g.i.Import("context", "context")

		g.W("type %s struct {\n", clientType)
		for _, m := range iface.Methods() {
			g.W("%sEndpoint %s.Endpoint\n", m.IfaceLcName, endpointPkg)

		}

		g.W("}\n\n")

		for _, m := range iface.Methods() {
			var params []string

			if m.ParamCtx != nil {
				params = append(params, m.ParamCtx.Name(), stdtypes.TypeString(m.ParamCtx.Type(), g.i.QualifyPkg))
			}

			params = append(params, types.NameTypeParams(m.Params, g.i.QualifyPkg, nil)...)

			if m.ParamVariadic != nil {
				pt := m.ParamVariadic.Type()
				if t, ok := pt.(*stdtypes.Slice); ok {
					pt = t.Elem()
				}
				params = append(params, m.ParamVariadic.Name(), "..."+stdtypes.TypeString(pt, g.i.QualifyPkg))
			}

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
				if m.ReturnErr != nil {
					g.W(", err")
				} else {
					g.W(", _")
				}
				if len(m.Results) == 0 && m.ReturnErr == nil {
					g.W(" = ")
				} else {
					g.W(" := ")
				}

				g.W("c.%sEndpoint(", m.IfaceLcName)

				if m.ParamCtx != nil {
					g.W("%s,", m.ParamCtx.Name())
				} else {
					g.W("%s.Background(),", contextPkg)
				}

				if len(m.Params) > 0 {
					g.W("%s", m.NameRequest)
					params := structKeyValue(m.Params, func(p *stdtypes.Var) bool {
						return !types.IsContext(p.Type())
					})
					if m.ParamVariadic != nil {
						params = append(params, strcase.ToCamel(m.ParamVariadic.Name()), m.ParamVariadic.Name())
					}
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
							g.W(types.ZeroValue(r.Type(), g.i.QualifyPkg))
						}
						g.W(",")
					}

					g.W(" err\n")

					g.W("}\n")
				}

				if len(m.Results) > 0 {
					if m.ResultsNamed {
						g.W("response := resp.(%s)\n", m.NameResponse)
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
							g.W("response.%s", strcase.ToCamel(r.Name()))
						}
					} else {
						g.W("response")
					}
				}
				if len(m.Results) > 0 && m.ReturnErr != nil {
					g.W(", ")
				}
				if m.ReturnErr != nil {
					g.W("nil")
				}
				g.W("\n")
			})
		}
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
	options clientStructOptionsGateway,
) generator.Generator {
	return &clientStruct{
		options: options,
	}
}
