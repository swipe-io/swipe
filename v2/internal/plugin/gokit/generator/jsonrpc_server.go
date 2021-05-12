package generator

import (
	"context"
	"strconv"
	stdstrings "strings"

	"github.com/swipe-io/swipe/v2/internal/option"

	"github.com/swipe-io/strcase"
	"github.com/swipe-io/swipe/v2/internal/plugin/gokit/config"
	"github.com/swipe-io/swipe/v2/internal/swipe"
	"github.com/swipe-io/swipe/v2/internal/writer"
)

type JSONRPCServerGenerator struct {
	w          writer.GoWriter
	UseFast    bool
	Interfaces []*config.Interface
}

func (g *JSONRPCServerGenerator) Generate(ctx context.Context) []byte {

	var (
		routerPkg  string
		jsonrpcPkg string
	)

	importer := ctx.Value(swipe.ImporterKey).(swipe.Importer)

	ffJSONPkg := importer.Import("ffjson", "github.com/pquerna/ffjson/ffjson")
	jsonPkg := importer.Import("json", "encoding/json")
	contextPkg := importer.Import("context", "context")

	if g.UseFast {
		jsonrpcPkg = importer.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/fasthttp/jsonrpc")
		routerPkg = importer.Import("routing", "github.com/qiangxue/fasthttp-routing")
	} else {
		jsonrpcPkg = importer.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/http/jsonrpc")
		routerPkg = importer.Import("mux", "github.com/gorilla/mux")
	}

	g.w.W("func MergeEndpointCodecMaps(ecms ...jsonrpc.EndpointCodecMap) jsonrpc.EndpointCodecMap {\n")
	g.w.W("mergedECM := make(jsonrpc.EndpointCodecMap, 512)\n")
	g.w.W("for _, ecm := range ecms {\nfor key, codec := range ecm {\nmergedECM[key] = codec\n}\n}\n")
	g.w.W("return mergedECM\n}\n")

	g.w.W("func encodeResponseJSONRPC(_ %s.Context, result interface{}) (%s.RawMessage, error) {\n", contextPkg, jsonPkg)
	g.w.W("b, err := %s.Marshal(result)\n", ffJSONPkg)
	g.w.W("if err != nil {\n")
	g.w.W("return nil, err\n")
	g.w.W("}\n")
	g.w.W("return b, nil\n")
	g.w.W("}\n\n")

	stringsPkg := importer.Import("strings", "strings")

	for _, iface := range g.Interfaces {
		ifaceType := iface.Named.Type.(*option.IfaceType)

		if !iface.Named.Pkg.Module.External {
			g.w.W("func Make%[1]sEndpointCodecMap(ep %[1]sEndpointSet", iface.Named.Name.UpperCase)
			g.w.W(",ns ...string) %s.EndpointCodecMap {\n", jsonrpcPkg)

			g.w.W("var namespace string\n")

			g.w.W("if len(ns) > 0 {\n")
			g.w.W("namespace = %s.Join(ns, \".\") + \".\"\n", stringsPkg)
			g.w.W("}\n")

			g.w.W("ecm := %[1]s.EndpointCodecMap{}\n", jsonrpcPkg)

			for _, m := range ifaceType.Methods {
				mopt := g.options.MethodOption(m)

				g.w.W("if ep.%sEndpoint != nil {\n", m.Name)

				g.w.W("ecm[namespace+\"%s\"] = %s.EndpointCodec{\n", strcase.ToLowerCamel(m.Name), jsonrpcPkg)
				g.w.W("Endpoint: ep.%sEndpoint,\n", m.Name)
				g.w.W("Decode: ")

				if mopt.ServerRequestFunc.Expr != nil {
					writer.WriteAST(g, g.i, mopt.ServerRequestFunc.Expr)
				} else {
					g.w.W("func(_ %s.Context, msg %s.RawMessage) (interface{}, error) {\n", contextPkg, jsonPkg)

					if len(m.Sig.Params) > 0 {
						fmtPkg := importer.Import("fmt", "fmt")
						g.w.W("var req %s\n", m.NameRequest)
						g.w.W("err := %s.Unmarshal(msg, &req)\n", ffJSONPkg)
						g.w.W("if err != nil {\n")
						g.w.W("return nil, %s.Errorf(\"couldn't unmarshal body to %s: %%s\", err)\n", fmtPkg, m.NameRequest)
						g.w.W("}\n")
						g.w.W("return req, nil\n")
					} else {
						g.w.W("return nil, nil\n")
					}
					g.w.W("}")
				}

				g.w.W(",\n")

				g.w.W("Encode:")

				if mopt.WrapResponse.Enable && len(m.Results) > 0 {
					jsonPkg := importer.Import("json", "encoding/json")
					g.w.W("func (ctx context.Context, response interface{}) (%s.RawMessage, error) {\n", jsonPkg)
					g.w.W("return encodeResponseJSONRPC(ctx, map[string]interface{}{\"%s\": response})\n", mopt.WrapResponse.Name)
					g.w.W("},\n")
				} else {
					g.w.W("encodeResponseJSONRPC,\n")
				}
				g.w.W("}\n}\n")
			}

			g.w.W("return ecm\n")

			g.w.W("}\n\n")
		}

	}

	g.w.W("// HTTP Helpers\n")
	g.w.W("func MakeHandlerJSONRPC(")

	var external bool

	for i := 0; i < g.options.Interfaces().Len(); i++ {
		ifaceType := g.options.Interfaces().At(i)
		typeStr := ifaceType.LcNameWithPrefix() + "Interface"
		if i > 0 {
			g.w.W(",")
		}

		if ifaceType.External() {
			external = true
			g.w.W("%s %sOption", ifaceType.LcNameWithPrefix(), ifaceType.UcNameWithPrefix())
		} else {
			g.w.W("svc%s %s", ifaceType.UcName(), typeStr)
		}
	}

	if external {
		g.w.W(", logger %s.Logger", importer.Import("log", "github.com/go-kit/kit/log"))
	}

	g.w.W(", options ...ServerOption")
	g.w.W(") (")
	if g.UseFast {
		g.w.W("%s.RequestHandler", importer.Import("fasthttp", "github.com/valyala/fasthttp"))
	} else {
		g.w.W("%s.Handler", importer.Import("http", "net/http"))
	}

	g.w.W(", error) {\n")

	g.w.W("opts := &serverOpts{}\n")
	g.w.W("for _, o := range options {\n o(opts)\n }\n")

	for i := 0; i < g.options.Interfaces().Len(); i++ {
		ifaceType := g.options.Interfaces().At(i)

		epSetName := makeEpSetName(ifaceType)

		if ifaceType.External() {
			pkgExtTransport := importer.Import(ifaceType.ExternalSwipePkg().Name, ifaceType.ExternalSwipePkg().PkgPath)
			sdPkg := importer.Import("sd", "github.com/go-kit/kit/sd")
			lbPkg := importer.Import("sd", "github.com/go-kit/kit/sd/lb")

			if ifaceType.External() {
				g.w.W("%s := %s.%sEndpointSet{}\n", epSetName, pkgExtTransport, ifaceType.UcName())
			} else {
				g.w.W("%s := %sEndpointSet{}\n", epSetName, ifaceType.UcName())
			}

			for _, m := range ifaceType.Methods() {
				optName := ifaceType.LcNameWithPrefix()
				epFactoryName := "endpointFactory"
				kitEndpointPkg := importer.Import("endpoint", "github.com/go-kit/kit/endpoint")
				transportExtPkg := importer.Import(ifaceType.ExternalSwipePkg().Name, ifaceType.ExternalSwipePkg().PkgPath)
				ioPkg := importer.Import("io", "io")

				g.w.W("{\n")

				g.w.W("if %s.%s.Balancer == nil {\n", optName, m.Name)
				g.w.W("%s.%s.Balancer = %s.NewRoundRobin\n", optName, m.Name, lbPkg)
				g.w.W("}\n")

				g.w.W("if %s.%s.RetryMax == 0 {\n", optName, m.Name)
				g.w.W("%s.%s.RetryMax = DefaultRetryMax\n", optName, m.Name)
				g.w.W("}\n")

				g.w.W("if %s.%s.RetryTimeout == 0 {\n", optName, m.Name)
				g.w.W("%s.%s.RetryTimeout = DefaultRetryTimeout\n", optName, m.Name)
				g.w.W("}\n")

				g.w.W("%s := func (instance string) (%s.Endpoint, %s.Closer, error) {\n", epFactoryName, kitEndpointPkg, ioPkg)
				g.w.W("if %s.Instance != \"\"{\n", optName)
				g.w.W("instance = %[1]s.TrimRight(instance, \"/\") + \"/\" + %[1]s.TrimLeft(%[2]s.Instance, \"/\")", stringsPkg, optName)
				g.w.W("}\n")

				g.w.W("c, err := %s.NewClientJSONRPC%s(instance, %s.ClientOptions...)\n", transportExtPkg, ifaceType.UcName(), optName)

				g.w.WriteCheckErr(func() {
					g.w.W("return nil, nil, err\n")
				})
				g.w.W("return ")

				g.w.W("%s.Make%sEndpoint(c), nil, nil\n", transportExtPkg, m.IfaceUcName)

				g.w.W("\n}\n\n")

				g.w.W("endpointer := %s.NewEndpointer(%s.Instancer, %s, logger)\n", sdPkg, optName, epFactoryName)
				g.w.W(
					"%[4]s.%[3]sEndpoint = %[1]s.RetryWithCallback(%[2]s.%[3]s.RetryTimeout, %[2]s.%[3]s.Balancer(endpointer), retryMax(%[2]s.%[3]s.RetryMax))\n",
					lbPkg, optName, m.Name, epSetName,
				)
				g.w.W(
					"%[2]s.%[1]sEndpoint = RetryErrorExtractor()(%[2]s.%[1]sEndpoint)\n",
					m.Name, epSetName,
				)
				g.w.W(
					"%[3]s.%[2]sEndpoint = middlewareChain(append(opts.genericEndpointMiddleware, opts.%[1]sEndpointMiddleware...))(%[3]s.%[2]sEndpoint)\n",
					ifaceType.LcNameWithPrefix()+m.Name, m.Name, epSetName,
				)
				g.w.W("}\n")
			}
		} else {
			g.w.W("%[1]s := Make%[2]sEndpointSet(svc%[2]s)\n", makeEpSetName(ifaceType), ifaceType.UcName())
			for _, m := range ifaceType.Methods() {
				g.w.W(
					"%[3]s.%[2]sEndpoint = middlewareChain(append(opts.genericEndpointMiddleware, opts.%[1]sEndpointMiddleware...))(%[3]s.%[2]sEndpoint)\n",
					m.IfaceLcName, m.Name, epSetName,
				)
			}
		}
	}

	if g.UseFast {
		g.w.W("r := %s.New()\n", routerPkg)
	} else {
		g.w.W("r := %s.NewRouter()\n", routerPkg)
	}

	g.w.W("handler := %s.NewServer(", jsonrpcPkg)

	if g.options.Interfaces().Len() > 1 {
		g.w.W("MergeEndpointCodecMaps(")
	}

	for i := 0; i < g.options.Interfaces().Len(); i++ {
		ifaceType := g.options.Interfaces().At(i)

		epSetName := makeEpSetName(ifaceType)

		if i > 0 {
			g.w.W(",")
		}
		if ifaceType.External() {
			pkgExtTransport := importer.Import(ifaceType.ExternalSwipePkg().Name, ifaceType.ExternalSwipePkg().PkgPath)
			g.w.W("%s.Make%sEndpointCodecMap(%s", pkgExtTransport, ifaceType.UcName(), epSetName)
		} else {
			g.w.W("Make%sEndpointCodecMap(%s", ifaceType.UcName(), epSetName)
		}
		if ifaceType.Namespace() != "" {
			g.w.W(",%s", strconv.Quote(ifaceType.Namespace()))
		}
		g.w.W(")")
	}

	if g.options.Interfaces().Len() > 1 {
		g.w.W(")")
	}

	g.w.W(", opts.genericServerOption...)\n")

	jsonRPCPath := g.options.JSONRPCPath()
	if g.UseFast {
		r := stdstrings.NewReplacer("{", "<", "}", ">")
		jsonRPCPath = r.Replace(jsonRPCPath)

		g.w.W("r.Post(\"%s\", func(c *routing.Context) error {\nhandler.ServeFastHTTP(c.RequestCtx)\nreturn nil\n})\n", jsonRPCPath)
	} else {
		g.w.W("r.Methods(\"POST\").Path(\"%s\").Handler(handler)\n", jsonRPCPath)
	}
	if g.UseFast {
		g.w.W("return r.HandleRequest, nil")
	} else {
		g.w.W("return r, nil")
	}
	g.w.W("}\n\n")
	return g.w.Bytes()
}

func (g *JSONRPCServerGenerator) OutputDir() string {
	return ""
}

func (g *JSONRPCServerGenerator) Filename() string {
	return "jsonrpc.go"
}
