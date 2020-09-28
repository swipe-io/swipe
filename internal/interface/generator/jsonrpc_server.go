package generator

import (
	"context"
	stdtypes "go/types"
	stdstrings "strings"

	"github.com/swipe-io/swipe/v2/internal/usecase/generator"

	"github.com/swipe-io/swipe/v2/internal/domain/model"
	"github.com/swipe-io/swipe/v2/internal/importer"
	"github.com/swipe-io/swipe/v2/internal/writer"
)

type jsonRPCServer struct {
	writer.GoLangWriter
	serviceID      string
	serviceType    stdtypes.Type
	serviceMethods []model.ServiceMethod
	transport      model.TransportOption
	i              *importer.Importer
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
	typeStr := stdtypes.TypeString(g.serviceType, g.i.QualifyPkg)

	if g.transport.FastHTTP {
		jsonrpcPkg = g.i.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/fasthttp/jsonrpc")
		routerPkg = g.i.Import("routing", "github.com/qiangxue/fasthttp-routing")
	} else {
		jsonrpcPkg = g.i.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/http/jsonrpc")
		routerPkg = g.i.Import("mux", "github.com/gorilla/mux")
	}

	g.W("func encodeResponseJSONRPC%s(_ %s.Context, result interface{}) (%s.RawMessage, error) {\n", g.serviceID, contextPkg, jsonPkg)
	g.W("b, err := %s.Marshal(result)\n", ffJSONPkg)
	g.W("if err != nil {\n")
	g.W("return nil, err\n")
	g.W("}\n")
	g.W("return b, nil\n")
	g.W("}\n\n")

	stringsPkg := g.i.Import("strings", "strings")

	g.W("func Make%sEndpointCodecMap(ep EndpointSet, ns ...string) %s.EndpointCodecMap {\n", g.serviceID, jsonrpcPkg)

	g.W("var namespace = %s.Join(ns, \".\")\n", stringsPkg)
	g.W("if len(ns) > 0 {\n")
	g.W("namespace += \".\"\n")
	g.W("}\n")

	g.W("ecm := %[1]s.EndpointCodecMap{}\n", jsonrpcPkg)

	//g.w("return %[1]s.EndpointCodecMap{\n", jsonrpcPkg)

	for _, m := range g.serviceMethods {
		mopt := g.transport.MethodOptions[m.Name]

		g.W("if ep.%sEndpoint != nil {\n", m.Name)

		g.W("ecm[namespace+\"%s\"] = %s.EndpointCodec{\n", m.LcName, jsonrpcPkg)
		g.W("Endpoint: ep.%sEndpoint,\n", m.Name)
		g.W("Decode: ")

		if mopt.ServerRequestFunc.Expr != nil {
			writer.WriteAST(g, g.i, mopt.ServerRequestFunc.Expr)
		} else {
			fmtPkg := g.i.Import("fmt", "fmt")

			g.W("func(_ %s.Context, msg %s.RawMessage) (interface{}, error) {\n", contextPkg, jsonPkg)

			if len(m.Params) > 0 {
				g.W("var req %sRequest%s\n", m.LcName, g.serviceID)
				g.W("err := %s.Unmarshal(msg, &req)\n", ffJSONPkg)
				g.W("if err != nil {\n")
				g.W("return nil, %s.Errorf(\"couldn'tpl unmarshal body to %sRequest%s: %%s\", err)\n", fmtPkg, m.LcName, g.serviceID)
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
			g.W("return encodeResponseJSONRPC%s(ctx, map[string]interface{}{\"%s\": response})\n", g.serviceID, mopt.WrapResponse.Name)
			g.W("},\n")
		} else {
			g.W("encodeResponseJSONRPC%s,\n", g.serviceID)
		}
		g.W("}\n}\n")
	}

	g.W("return ecm\n")

	g.W("}\n")

	g.W("// HTTP %s Transport\n", g.transport.Prefix)
	g.W("func MakeHandler%s%s(s %s", g.transport.Prefix, g.serviceID, typeStr)

	g.W(", opts ...%sServerOption", g.serviceID)
	g.W(") (")
	if g.transport.FastHTTP {
		g.W("%s.RequestHandler", g.i.Import("fasthttp", "github.com/valyala/fasthttp"))
	} else {
		g.W("%s.Handler", g.i.Import("http", "net/http"))
	}

	g.W(", error) {\n")

	g.W("sopt := &server%sOpts{}\n", g.serviceID)

	g.W("for _, o := range opts {\n o(sopt)\n }\n")

	g.W("ep := MakeEndpointSet(s)\n")

	for _, m := range g.serviceMethods {
		g.W(
			"ep.%[1]sEndpoint = middlewareChain(append(sopt.genericEndpointMiddleware, sopt.%[2]sEndpointMiddleware...))(ep.%[1]sEndpoint)\n",
			m.Name, m.LcName,
		)
	}

	if g.transport.FastHTTP {
		g.W("r := %s.New()\n", routerPkg)
	} else {
		g.W("r := %s.NewRouter()\n", routerPkg)
	}
	g.W("handler := %[1]s.NewServer(Make%sEndpointCodecMap(ep), sopt.genericServerOption...)\n", jsonrpcPkg, g.serviceID)
	jsonRPCPath := g.transport.JsonRPC.Path
	if g.transport.FastHTTP {
		r := stdstrings.NewReplacer("{", "<", "}", ">")
		jsonRPCPath = r.Replace(jsonRPCPath)

		g.W("r.Post(\"%s\", func(c *routing.Context) error {\nhandler.ServeFastHTTP(c.RequestCtx)\nreturn nil\n})\n", jsonRPCPath)
	} else {
		g.W("r.Methods(\"POST\").Path(\"%s\").Handler(handler)\n", jsonRPCPath)
	}
	if g.transport.FastHTTP {
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
	serviceID string,
	serviceType stdtypes.Type,
	serviceMethods []model.ServiceMethod,
	transport model.TransportOption,
) generator.Generator {
	return &jsonRPCServer{
		serviceID:      serviceID,
		serviceType:    serviceType,
		serviceMethods: serviceMethods,
		transport:      transport,
	}
}
