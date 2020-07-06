package generator

import (
	"context"
	stdtypes "go/types"
	stdstrings "strings"

	"github.com/swipe-io/swipe/pkg/domain/model"
	"github.com/swipe-io/swipe/pkg/importer"
	"github.com/swipe-io/swipe/pkg/writer"
)

type jsonRPCServer struct {
	*writer.GoLangWriter
	filename string
	info     model.GenerateInfo
	o        model.ServiceOption
	i        *importer.Importer
}

func (g *jsonRPCServer) Prepare(ctx context.Context) error {
	return nil
}

func (g *jsonRPCServer) Process(ctx context.Context) error {
	var (
		routerPkg  string
		jsonrpcPkg string
	)

	ffjsonPkg := g.i.Import("ffjson", "github.com/pquerna/ffjson/ffjson")
	jsonPkg := g.i.Import("json", "encoding/json")
	contextPkg := g.i.Import("context", "context")
	typeStr := stdtypes.TypeString(g.o.Type, g.i.QualifyPkg)

	transportOpt := g.o.Transport

	if transportOpt.FastHTTP {
		jsonrpcPkg = g.i.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/fasthttp/jsonrpc")
		routerPkg = g.i.Import("routing", "github.com/qiangxue/fasthttp-routing")
	} else {
		jsonrpcPkg = g.i.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/http/jsonrpc")
		routerPkg = g.i.Import("mux", "github.com/gorilla/mux")
	}

	g.W("func encodeResponseJSONRPC%s(_ %s.Context, result interface{}) (%s.RawMessage, error) {\n", g.o.ID, contextPkg, jsonPkg)
	g.W("b, err := %s.Marshal(result)\n", ffjsonPkg)
	g.W("if err != nil {\n")
	g.W("return nil, err\n")
	g.W("}\n")
	g.W("return b, nil\n")
	g.W("}\n\n")

	stringsPkg := g.i.Import("strings", "strings")

	g.W("func Make%sEndpointCodecMap(ep EndpointSet, ns ...string) %s.EndpointCodecMap {\n", g.o.ID, jsonrpcPkg)

	g.W("var namespace = %s.Join(ns, \".\")\n", stringsPkg)
	g.W("if len(ns) > 0 {\n")
	g.W("namespace += \".\"\n")
	g.W("}\n")

	g.W("ecm := %[1]s.EndpointCodecMap{}\n", jsonrpcPkg)

	//g.W("return %[1]s.EndpointCodecMap{\n", jsonrpcPkg)

	for _, m := range g.o.Methods {
		mopt := transportOpt.MethodOptions[m.Name]

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
				g.W("var req %sRequest%s\n", m.LcName, g.o.ID)
				g.W("err := %s.Unmarshal(msg, &req)\n", ffjsonPkg)
				g.W("if err != nil {\n")
				g.W("return nil, %s.Errorf(\"couldn't unmarshal body to %sRequest%s: %%s\", err)\n", fmtPkg, m.LcName, g.o.ID)
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
			g.W("return encodeResponseJSONRPC%s(ctx, map[string]interface{}{\"%s\": response})\n", g.o.ID, mopt.WrapResponse.Name)
			g.W("},\n")
		} else {
			g.W("encodeResponseJSONRPC%s,\n", g.o.ID)
		}
		g.W("}\n}\n")
	}

	g.W("return ecm\n")

	g.W("}\n")

	g.W("// HTTP %s Transport\n", transportOpt.Prefix)
	g.W("func MakeHandler%s%s(s %s", g.o.Transport.Prefix, g.o.ID, typeStr)

	g.W(", opts ...%sServerOption", g.o.ID)
	g.W(") (")
	if transportOpt.FastHTTP {
		g.W("%s.RequestHandler", g.i.Import("fasthttp", "github.com/valyala/fasthttp"))
	} else {
		g.W("%s.Handler", g.i.Import("http", "net/http"))
	}

	g.W(", error) {\n")

	g.W("sopt := &server%sOpts{}\n", g.o.ID)

	g.W("for _, o := range opts {\n o(sopt)\n }\n")

	g.W("ep := MakeEndpointSet(s)\n")

	for _, m := range g.o.Methods {
		g.W(
			"ep.%[1]sEndpoint = middlewareChain(append(sopt.genericEndpointMiddleware, sopt.%[2]sEndpointMiddleware...))(ep.%[1]sEndpoint)\n",
			m.Name, m.LcName,
		)
	}

	if transportOpt.FastHTTP {
		g.W("r := %s.New()\n", routerPkg)
	} else {
		g.W("r := %s.NewRouter()\n", routerPkg)
	}
	g.W("handler := %[1]s.NewServer(Make%sEndpointCodecMap(ep), sopt.genericServerOption...)\n", jsonrpcPkg, g.o.ID)
	jsonRPCPath := transportOpt.JsonRPC.Path
	if transportOpt.FastHTTP {
		r := stdstrings.NewReplacer("{", "<", "}", ">")
		jsonRPCPath = r.Replace(jsonRPCPath)

		g.W("r.Post(\"%s\", func(c *routing.Context) error {\nhandler.ServeFastHTTP(c.RequestCtx)\nreturn nil\n})\n", jsonRPCPath)
	} else {
		g.W("r.Methods(\"POST\").Path(\"%s\").Handler(handler)\n", jsonRPCPath)
	}
	if transportOpt.FastHTTP {
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
	return g.filename
}

func (g *jsonRPCServer) SetImporter(i *importer.Importer) {
	g.i = i
}

func NewJsonRPCServer(filename string, info model.GenerateInfo, o model.ServiceOption) Generator {
	return &jsonRPCServer{GoLangWriter: writer.NewGoLangWriter(), filename: filename, info: info, o: o}
}
