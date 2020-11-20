package generator

import (
	"context"
	stdtypes "go/types"
	stdstrings "strings"

	"github.com/swipe-io/swipe/v2/internal/domain/model"
	"github.com/swipe-io/swipe/v2/internal/importer"
	"github.com/swipe-io/swipe/v2/internal/usecase/generator"
	"github.com/swipe-io/swipe/v2/internal/writer"
)

type jsonRPCServerOptionsGateway interface {
	AppID() string
	UseFast() bool
	JSONRPCEnable() bool
	JSONRPCPath() string
	MethodOption(m model.ServiceMethod) model.MethodOption
	Interfaces() model.Interfaces
	Prefix() string
}

type jsonRPCServer struct {
	writer.GoLangWriter
	options jsonRPCServerOptionsGateway
	i       *importer.Importer
}

func (g *jsonRPCServer) Prepare(ctx context.Context) error {
	return nil
}

func (g *jsonRPCServer) Process(ctx context.Context) error {
	var (
		routerPkg  string
		jsonrpcPkg string
	)
	ffJSONPkg := g.i.Import("ffjson", "github.com/pquerna/ffjson/ffjson")
	jsonPkg := g.i.Import("json", "encoding/json")
	contextPkg := g.i.Import("context", "context")

	if g.options.UseFast() {
		jsonrpcPkg = g.i.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/fasthttp/jsonrpc")
		routerPkg = g.i.Import("routing", "github.com/qiangxue/fasthttp-routing")
	} else {
		jsonrpcPkg = g.i.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/http/jsonrpc")
		routerPkg = g.i.Import("mux", "github.com/gorilla/mux")
	}

	g.W("func MergeEndpointCodecMaps(ecms ...jsonrpc.EndpointCodecMap) jsonrpc.EndpointCodecMap {\n")
	g.W("mergedECM := make(jsonrpc.EndpointCodecMap, 512)\n")
	g.W("for _, ecm := range ecms {\nfor key, codec := range ecm {\nmergedECM[key] = codec\n}\n}\n")
	g.W("return mergedECM\n}\n")

	g.W("func encodeResponseJSONRPC(_ %s.Context, result interface{}) (%s.RawMessage, error) {\n", contextPkg, jsonPkg)
	g.W("b, err := %s.Marshal(result)\n", ffJSONPkg)
	g.W("if err != nil {\n")
	g.W("return nil, err\n")
	g.W("}\n")
	g.W("return b, nil\n")
	g.W("}\n\n")

	stringsPkg := g.i.Import("strings", "strings")

	for i := 0; i < g.options.Interfaces().Len(); i++ {
		iface := g.options.Interfaces().At(i)

		g.W("func Make%[1]sEndpointCodecMap(ep %[1]sEndpointSet", iface.Name())
		g.W(",ns ...string) %s.EndpointCodecMap {\n", jsonrpcPkg)
		g.W("var namespace string\n")
		if g.options.Interfaces().Len() > 1 {
			prefix := iface.NameUnExport()
			if iface.NameUnExport() != "" {
				prefix = iface.NameUnExport()
			}
			g.W("namespace = \"%s.\"\n", prefix)
		}
		g.W("if len(ns) > 0 {\n")
		g.W("namespace = %s.Join(ns, \".\") + \".\"\n", stringsPkg)
		g.W("}\n")

		g.W("ecm := %[1]s.EndpointCodecMap{}\n", jsonrpcPkg)

		for _, m := range iface.Methods() {
			mopt := g.options.MethodOption(m)

			g.W("if ep.%sEndpoint != nil {\n", m.Name)

			g.W("ecm[namespace+\"%s\"] = %s.EndpointCodec{\n", m.LcName, jsonrpcPkg)
			g.W("Endpoint: ep.%sEndpoint,\n", m.Name)
			g.W("Decode: ")

			if mopt.ServerRequestFunc.Expr != nil {
				writer.WriteAST(g, g.i, mopt.ServerRequestFunc.Expr)
			} else {
				g.W("func(_ %s.Context, msg %s.RawMessage) (interface{}, error) {\n", contextPkg, jsonPkg)

				if len(m.Params) > 0 {
					fmtPkg := g.i.Import("fmt", "fmt")
					g.W("var req %s\n", m.NameRequest)
					g.W("err := %s.Unmarshal(msg, &req)\n", ffJSONPkg)
					g.W("if err != nil {\n")
					g.W("return nil, %s.Errorf(\"couldn't unmarshal body to %s: %%s\", err)\n", fmtPkg, m.NameRequest)
					g.W("}\n")
					g.W("return req, nil\n")

				} else {
					g.W("return nil, nil\n")
				}
				g.W("}")
			}

			g.W(",\n")

			g.W("Encode:")

			if mopt.WrapResponse.Enable && len(m.Results) > 0 {
				jsonPkg := g.i.Import("json", "encoding/json")
				g.W("func (ctx context.Context, response interface{}) (%s.RawMessage, error) {\n", jsonPkg)
				g.W("return encodeResponseJSONRPC(ctx, map[string]interface{}{\"%s\": response})\n", mopt.WrapResponse.Name)
				g.W("},\n")
			} else {
				g.W("encodeResponseJSONRPC,\n")
			}
			g.W("}\n}\n")
		}

		g.W("return ecm\n")

		g.W("}\n\n")
	}

	g.W("// HTTP %s Transport\n", g.options.Prefix())
	g.W("func MakeHandler%s(", g.options.Prefix())
	for i := 0; i < g.options.Interfaces().Len(); i++ {
		iface := g.options.Interfaces().At(i)
		typeStr := stdtypes.TypeString(iface.Type(), g.i.QualifyPkg)
		if i > 0 {
			g.W(",")
		}
		g.W("svc%s %s", iface.Name(), typeStr)
	}
	g.W(", options ...ServerOption")
	g.W(") (")
	if g.options.UseFast() {
		g.W("%s.RequestHandler", g.i.Import("fasthttp", "github.com/valyala/fasthttp"))
	} else {
		g.W("%s.Handler", g.i.Import("http", "net/http"))
	}

	g.W(", error) {\n")

	g.W("opts := &serverOpts{}\n")

	g.W("for _, o := range options {\n o(opts)\n }\n")

	for i := 0; i < g.options.Interfaces().Len(); i++ {
		iface := g.options.Interfaces().At(i)
		g.W("%[1]s := Make%[2]sEndpointSet(svc%[2]s)\n", makeEpSetName(iface, g.options.Interfaces().Len()), iface.Name())
	}

	for i := 0; i < g.options.Interfaces().Len(); i++ {
		iface := g.options.Interfaces().At(i)
		epSetName := makeEpSetName(iface, g.options.Interfaces().Len())
		for _, m := range iface.Methods() {
			g.W(
				"%[3]s.%[2]sEndpoint = middlewareChain(append(opts.genericEndpointMiddleware, opts.%[1]sEndpointMiddleware...))(%[3]s.%[2]sEndpoint)\n",
				m.NameUnExport, m.Name, epSetName,
			)
		}
	}
	if g.options.UseFast() {
		g.W("r := %s.New()\n", routerPkg)
	} else {
		g.W("r := %s.NewRouter()\n", routerPkg)
	}

	g.W("handler := %s.NewServer(", jsonrpcPkg)

	if g.options.Interfaces().Len() > 1 {
		g.W("MergeEndpointCodecMaps(")
	}

	for i := 0; i < g.options.Interfaces().Len(); i++ {
		iface := g.options.Interfaces().At(i)
		if i > 0 {
			g.W(",")
		}
		g.W("Make%[1]sEndpointCodecMap(%[2]s)", iface.Name(), makeEpSetName(iface, g.options.Interfaces().Len()))
	}

	if g.options.Interfaces().Len() > 1 {
		g.W(")")
	}

	g.W(", opts.genericServerOption...)\n")

	jsonRPCPath := g.options.JSONRPCPath()
	if g.options.UseFast() {
		r := stdstrings.NewReplacer("{", "<", "}", ">")
		jsonRPCPath = r.Replace(jsonRPCPath)

		g.W("r.Post(\"%s\", func(c *routing.Context) error {\nhandler.ServeFastHTTP(c.RequestCtx)\nreturn nil\n})\n", jsonRPCPath)
	} else {
		g.W("r.Methods(\"POST\").Path(\"%s\").Handler(handler)\n", jsonRPCPath)
	}
	if g.options.UseFast() {
		g.W("return r.HandleRequest, nil")
	} else {
		g.W("return r, nil")
	}
	g.W("}\n\n")
	return nil
}

func (g *jsonRPCServer) PkgName() string {
	return ""
}

func (g *jsonRPCServer) OutputDir() string {
	return ""
}

func (g *jsonRPCServer) Filename() string {
	return "server_gen.go"
}

func (g *jsonRPCServer) SetImporter(i *importer.Importer) {
	g.i = i
}

func NewJsonRPCServer(
	options jsonRPCServerOptionsGateway,
) generator.Generator {
	return &jsonRPCServer{
		options: options,
	}
}
