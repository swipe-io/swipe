package generator

import (
	"context"
	"fmt"
	stdtypes "go/types"
	"strconv"

	"github.com/swipe-io/swipe/pkg/domain/model"
	"github.com/swipe-io/swipe/pkg/importer"
	"github.com/swipe-io/swipe/pkg/writer"
)

type jsonRPCGoClient struct {
	*writer.GoLangWriter

	info model.GenerateInfo
	o    model.ServiceOption
	i    *importer.Importer
}

func (g *jsonRPCGoClient) Process(ctx context.Context) error {
	clientType := "client" + g.o.ID
	typeStr := stdtypes.TypeString(g.o.Type, g.i.QualifyPkg)

	g.W("func NewClient%s%s(tgt string", g.o.Transport.Prefix, g.o.ID)

	g.W(" ,opts ...%sClientOption", g.o.ID)

	g.W(") (%s, error) {\n", typeStr)

	g.W("c := &%s{}\n", clientType)

	g.W("for _, o := range opts {\n")
	g.W("o(c)\n")
	g.W("}\n")

	transportOpt := g.o.Transport

	var (
		jsonrpcPkg string
	)
	if transportOpt.FastHTTP {
		jsonrpcPkg = g.i.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/fasthttp/jsonrpc")
	} else {
		jsonrpcPkg = g.i.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/http/jsonrpc")
	}

	urlPkg := g.i.Import("url", "net/url")
	contextPkg := g.i.Import("context", "context")
	ffjsonPkg := g.i.Import("ffjson", "github.com/pquerna/ffjson/ffjson")
	jsonPkg := g.i.Import("json", "encoding/json")
	fmtPkg := g.i.Import("fmt", "fmt")

	g.W("u, err := %s.Parse(tgt)\n", urlPkg)

	g.WriteCheckErr(func() {
		g.W("return nil, err")
	})

	for _, m := range g.o.Methods {
		mopt := transportOpt.MethodOptions[m.Name]

		g.W("c.%[1]sClientOption = append(\nc.%[1]sClientOption,\n", m.LcName)

		g.W("%s.ClientRequestEncoder(", jsonrpcPkg)
		g.W("func(_ %s.Context, obj interface{}) (%s.RawMessage, error) {\n", contextPkg, jsonPkg)

		if len(m.Params) > 0 {
			g.W("req, ok := obj.(%sRequest%s)\n", m.LcName, g.o.ID)
			g.W("if !ok {\n")
			g.W("return nil, %s.Errorf(\"couldn't assert request as %sRequest%s, got %%T\", obj)\n", fmtPkg, m.LcName, g.o.ID)
			g.W("}\n")
			g.W("b, err := %s.Marshal(req)\n", ffjsonPkg)
			g.W("if err != nil {\n")
			g.W("return nil, %s.Errorf(\"couldn't marshal request %%T: %%s\", obj, err)\n", fmtPkg)
			g.W("}\n")
			g.W("return b, nil\n")
		} else {
			g.W("return nil, nil\n")
		}
		g.W("}),\n")

		g.W("%s.ClientResponseDecoder(", jsonrpcPkg)
		g.W("func(_ %s.Context, response %s.Response) (interface{}, error) {\n", contextPkg, jsonrpcPkg)
		g.W("if response.Error != nil {\n")
		g.W("return nil, ErrorDecode(response.Error.Code)\n")
		g.W("}\n")

		if len(m.Results) > 0 {
			var responseType string
			if m.ResultsNamed {
				responseType = fmt.Sprintf("%sResponse%s", m.LcName, g.o.ID)
			} else {
				responseType = stdtypes.TypeString(m.Results[0].Type(), g.i.QualifyPkg)
			}

			if mopt.WrapResponse.Enable {
				g.W("var resp struct {\n Data %s `json:\"%s\"`\n}\n", responseType, mopt.WrapResponse.Name)
			} else {
				g.W("var resp %s\n", responseType)
			}

			g.W("err := %s.Unmarshal(response.Result, &resp)\n", ffjsonPkg)
			g.W("if err != nil {\n")
			g.W("return nil, %s.Errorf(\"couldn't unmarshal body to %sResponse%s: %%s\", err)\n", fmtPkg, m.LcName, g.o.ID)
			g.W("}\n")

			if mopt.WrapResponse.Enable {
				g.W("return resp.Data, nil\n")
			} else {
				g.W("return resp, nil\n")
			}
		} else {
			g.W("return nil, nil\n")
		}

		g.W("}),\n")

		g.W(")\n")

		g.W("c.%sEndpoint = %s.NewClient(\n", m.LcName, jsonrpcPkg)
		g.W("u,\n")
		g.W("%s,\n", strconv.Quote(m.LcName))

		g.W("append(c.genericClientOption, c.%sClientOption...)...,\n", m.LcName)

		g.W(").Endpoint()\n")

		g.W(
			"c.%[1]sEndpoint = middlewareChain(append(c.genericEndpointMiddleware, c.%[1]sEndpointMiddleware...))(c.%[1]sEndpoint)\n",
			m.LcName,
		)
	}

	g.W("return c, nil\n")
	g.W("}\n")
	return nil
}

func (g *jsonRPCGoClient) PkgName() string {
	return ""
}

func (g *jsonRPCGoClient) OutputDir() string {
	return ""
}

func (g *jsonRPCGoClient) Filename() string {
	return "client_jsonrpc_gen.go"
}

func (g *jsonRPCGoClient) Imports() []string {
	return g.i.SortedImports()
}

func NewJsonRPCGoClient(info model.GenerateInfo, o model.ServiceOption, i *importer.Importer) Generator {
	return &jsonRPCGoClient{info: info, o: o, i: i, GoLangWriter: writer.NewGoLangWriter(i)}
}
