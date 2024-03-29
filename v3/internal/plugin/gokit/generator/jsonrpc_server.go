package generator

import (
	"context"
	"strconv"
	stdstrings "strings"

	"github.com/swipe-io/swipe/v3/internal/plugin/gokit/config"
	"github.com/swipe-io/swipe/v3/option"
	"github.com/swipe-io/swipe/v3/swipe"
	"github.com/swipe-io/swipe/v3/writer"
)

type JSONRPCServerGenerator struct {
	w           writer.GoWriter
	UseFast     bool
	Interfaces  []*config.Interface
	JSONRPCPath string
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

		epSetName := NameEndpointSetName(iface)

		g.w.W("func Make%sEndpointCodecMap(ep %s", UcNameWithAppPrefix(iface), epSetName)
		g.w.W(",ns ...string) %s.EndpointCodecMap {\n", jsonrpcPkg)

		g.w.W("var namespace string\n")

		g.w.W("if len(ns) > 0 {\n")
		g.w.W("namespace = %s.Join(ns, \".\") + \".\"\n", stringsPkg)
		g.w.W("}\n")

		g.w.W("ecm := %[1]s.EndpointCodecMap{}\n", jsonrpcPkg)

		for _, m := range ifaceType.Methods {
			nameRequest := NameRequest(m, iface)
			g.w.W("if ep.%sEndpoint != nil {\n", m.Name)
			g.w.W("ecm[namespace+\"%s\"] = %s.EndpointCodec{\n", m.Name.Lower(), jsonrpcPkg)
			g.w.W("Endpoint: ep.%sEndpoint,\n", m.Name)
			g.w.W("Decode: ")
			g.w.W("func(_ %s.Context, msg %s.RawMessage) (interface{}, error) {\n", contextPkg, jsonPkg)
			if len(m.Sig.Params) > 0 {
				fmtPkg := importer.Import("fmt", "fmt")
				g.w.W("var req %s\n", nameRequest)
				g.w.W("err := %s.Unmarshal(msg, &req)\n", ffJSONPkg)
				g.w.W("if err != nil {\n")
				g.w.W("return nil, %s.Errorf(\"couldn't unmarshal body to %s: %%s\", err)\n", fmtPkg, nameRequest)
				g.w.W("}\n")
				g.w.W("return req, nil\n")
			} else {
				g.w.W("return nil, nil\n")
			}
			g.w.W("},\n")
			g.w.W("Encode: ")

			g.w.W("encodeResponseJSONRPC,\n")
			g.w.W("}\n}\n")
		}
		g.w.W("return ecm\n")
		g.w.W("}\n\n")
	}

	g.w.W("// MakeHandlerJSONRPC make HTTP JSONRPC handler.\n")
	g.w.W("func MakeHandlerJSONRPC(")

	var external bool

	for i, iface := range g.Interfaces {
		typeStr := NameInterface(iface)

		if i > 0 {
			g.w.W(",")
		}
		if iface.Gateway != nil {
			external = true
			g.w.W("%s %sOption", LcNameWithAppPrefix(iface, true), UcNameWithAppPrefix(iface, true))
		} else {
			g.w.W("%s %s", ServicePropName(iface), typeStr)
		}
	}

	if external {
		g.w.W(", logger %s.Logger", importer.Import("log", "github.com/go-kit/log"))
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

	for _, iface := range g.Interfaces {
		optName := LcNameWithAppPrefix(iface, iface.Gateway != nil)
		ifaceType := iface.Named.Type.(*option.IfaceType)

		epSetName := NameEndpointSetNameVar(iface)

		if iface.Gateway != nil {
			epEndpointSetName := NameEndpointSetName(iface)

			sdPkg := importer.Import("sd", "github.com/go-kit/kit/sd")
			lbPkg := importer.Import("sd", "github.com/go-kit/kit/sd/lb")

			g.w.W("%s := %s{}\n", epSetName, epEndpointSetName)

			for _, m := range ifaceType.Methods {

				epFactoryName := "endpointFactory"
				kitEndpointPkg := importer.Import("endpoint", "github.com/go-kit/kit/endpoint")
				stdLogPkg := importer.Import("log", "log")

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

				g.w.W("if %s.Factory == nil {\n", optName)
				g.w.W("%s.Panic(\"%s.Factory is not set\")\n", stdLogPkg, optName)
				g.w.W("}\n")

				g.w.W("%s := func (instance string) (%s.Endpoint, %s.Closer, error) {\n", epFactoryName, kitEndpointPkg, ioPkg)
				g.w.W("c, err := %s.Factory(instance)\n", optName)

				g.w.WriteCheckErr("err", func() {
					g.w.W("return nil, nil, err\n")
				})

				g.w.W("return ")
				g.w.W("Make%sEndpoint(c), nil, nil\n", UcNameWithAppPrefix(iface)+m.Name.Upper())
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
					"%[3]s.%[2]sEndpoint = middlewareChain(append(opts.genericOpts.endpointMiddleware, opts.%[1]sOpts.endpointMiddleware...))(%[3]s.%[2]sEndpoint)\n",
					LcNameIfaceMethod(iface, m), m.Name, epSetName,
				)
				g.w.W("}\n")
			}
		} else {
			g.w.W("%s := Make%s(%s)\n", NameEndpointSetNameVar(iface), NameEndpointSetName(iface), ServicePropName(iface))
			for _, m := range ifaceType.Methods {
				g.w.W(
					"%[3]s.%[2]sEndpoint = middlewareChain(append(opts.genericOpts.endpointMiddleware, opts.%[1]sOpts.endpointMiddleware...))(%[3]s.%[2]sEndpoint)\n",
					LcNameIfaceMethod(iface, m), m.Name, epSetName,
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

	if len(g.Interfaces) > 1 {
		g.w.W("MergeEndpointCodecMaps(")
	}

	for i, iface := range g.Interfaces {
		epSetName := NameEndpointSetNameVar(iface)

		if i > 0 {
			g.w.W(",")
		}

		g.w.W("Make%sEndpointCodecMap(%s", UcNameWithAppPrefix(iface), epSetName)

		if iface.Namespace != "" {
			g.w.W(",%s", strconv.Quote(iface.Namespace))
		}
		g.w.W(")")
	}

	if len(g.Interfaces) > 1 {
		g.w.W(")")
	}

	g.w.W(", opts.genericOpts.serverOption...)\n")

	jsonRPCPath := g.JSONRPCPath
	if g.UseFast {
		r := stdstrings.NewReplacer("{", "<", "}", ">")
		jsonRPCPath = r.Replace(jsonRPCPath)

		g.w.W("r.Post(\"%s\", func(c *routing.Context) error {\nhandler.ServeFastHTTP(c.RequestCtx)\nreturn nil\n})\n", jsonRPCPath)
	} else {
		g.w.W("r.Methods(\"POST\").")
		if jsonRPCPath != "" {
			g.w.W("Path(\"%s\").", jsonRPCPath)
		}
		g.w.W("Handler(handler)\n")
	}
	if g.UseFast {
		g.w.W("return r.HandleRequest, nil")
	} else {
		g.w.W("return r, nil")
	}
	g.w.W("}\n\n")
	return g.w.Bytes()
}

func (g *JSONRPCServerGenerator) OutputPath() string {
	return ""
}

func (g *JSONRPCServerGenerator) Filename() string {
	return "jsonrpc_server.go"
}
