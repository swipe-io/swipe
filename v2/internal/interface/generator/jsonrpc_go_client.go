package generator

import (
	"context"
	"fmt"
	stdtypes "go/types"
	"strconv"

	"github.com/swipe-io/swipe/v2/internal/domain/model"
	"github.com/swipe-io/swipe/v2/internal/importer"
	"github.com/swipe-io/swipe/v2/internal/usecase/generator"
	"github.com/swipe-io/swipe/v2/internal/writer"
)

type jsonRPCGoClientOptionsGateway interface {
	Prefix() string
	Interfaces() model.Interfaces
	MethodOption(m model.ServiceMethod) model.MethodOption
	UseFast() bool
}

type jsonRPCGoClient struct {
	writer.GoLangWriter
	options jsonRPCGoClientOptionsGateway
	i       *importer.Importer
}

func (g *jsonRPCGoClient) Prepare(ctx context.Context) error {
	return nil
}

func (g *jsonRPCGoClient) Process(ctx context.Context) error {
	for i := 0; i < g.options.Interfaces().Len(); i++ {
		var (
			jsonrpcPkg string
			contextPkg string
			ffJSONPkg  string
			jsonPkg    string
			fmtPkg     string
			urlPkg     string
			netPkg     string
			stringsPkg string
		)

		iface := g.options.Interfaces().At(i)

		clientType := "client" + iface.Name()
		typeStr := stdtypes.TypeString(iface.Type(), g.i.QualifyPkg)

		g.W("// Deprecated\nfunc NewClient%s(tgt string", g.options.Prefix())
		g.W(" ,options ...ClientOption")
		g.W(") (%s, error) {\n", typeStr)
		g.W("return NewClient%s%s(tgt, options...)", g.options.Prefix(), iface.Name())
		g.W("}\n")

		g.W("func NewClient%s%s(tgt string", g.options.Prefix(), iface.Name())
		g.W(" ,options ...ClientOption")
		g.W(") (%s, error) {\n", typeStr)
		g.W("opts := &clientOpts{}\n")
		g.W("c := &%s{}\n", clientType)
		g.W("for _, o := range options {\n")
		g.W("o(opts)\n")
		g.W("}\n")

		if g.options.UseFast() {
			jsonrpcPkg = g.i.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/fasthttp/jsonrpc")
		} else {
			jsonrpcPkg = g.i.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/http/jsonrpc")
		}
		urlPkg = g.i.Import("url", "net/url")
		contextPkg = g.i.Import("context", "context")
		ffJSONPkg = g.i.Import("ffjson", "github.com/pquerna/ffjson/ffjson")
		jsonPkg = g.i.Import("json", "encoding/json")
		fmtPkg = g.i.Import("fmt", "fmt")
		netPkg = g.i.Import("net", "net")
		stringsPkg = g.i.Import("strings", "strings")

		g.W("if %s.HasPrefix(tgt, \"[\") {\n", stringsPkg)
		g.W("host, port, err := %s.SplitHostPort(tgt)\n", netPkg)
		g.WriteCheckErr(func() {
			g.W("return nil, err")
		})
		g.W("tgt = host + \":\" + port\n")
		g.W("}\n")

		g.W("u, err := %s.Parse(tgt)\n", urlPkg)
		g.WriteCheckErr(func() {
			g.W("return nil, err")
		})
		g.W("if u.Scheme == \"\" {\n")
		g.W("u.Scheme = \"https\"")
		g.W("}\n")

		for _, m := range iface.Methods() {
			mopt := g.options.MethodOption(m)

			g.W("opts.%[1]sClientOption = append(\nopts.%[1]sClientOption,\n", m.NameUnExport)

			g.W("%s.ClientRequestEncoder(", jsonrpcPkg)
			g.W("func(_ %s.Context, obj interface{}) (%s.RawMessage, error) {\n", contextPkg, jsonPkg)

			if len(m.Params) > 0 {
				g.W("req, ok := obj.(%s)\n", m.NameRequest)
				g.W("if !ok {\n")
				g.W("return nil, %s.Errorf(\"couldn't assert request as %s, got %%T\", obj)\n", fmtPkg, m.NameRequest)
				g.W("}\n")
				g.W("b, err := %s.Marshal(req)\n", ffJSONPkg)
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
			g.W("return nil, %sErrorDecode(response.Error.Code, response.Error.Message, response.Error.Data)\n", m.NameUnExport)
			g.W("}\n")

			if len(m.Results) > 0 {
				var responseType string
				if m.ResultsNamed {
					responseType = fmt.Sprintf("%s", m.NameResponse)
				} else {
					responseType = stdtypes.TypeString(m.Results[0].Type(), g.i.QualifyPkg)
				}

				if mopt.WrapResponse.Enable {
					g.W("var resp struct {\n Data %s `json:\"%s\"`\n}\n", responseType, mopt.WrapResponse.Name)
				} else {
					g.W("var resp %s\n", responseType)
				}

				g.W("err := %s.Unmarshal(response.Result, &resp)\n", ffJSONPkg)
				g.W("if err != nil {\n")
				g.W("return nil, %s.Errorf(\"couldn't unmarshal body to %s: %%s\", err)\n", fmtPkg, m.NameResponse)
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

			methodName := m.LcName
			if iface.IsNameChange() || g.options.Interfaces().Len() > 1 {
				methodName = iface.NameUnExport() + "." + methodName
			}

			g.W("c.%sEndpoint = %s.NewClient(\n", m.LcName, jsonrpcPkg)
			g.W("u,\n")
			g.W("%s,\n", strconv.Quote(methodName))

			g.W("append(opts.genericClientOption, opts.%sClientOption...)...,\n", m.NameUnExport)

			g.W(").Endpoint()\n")

			g.W(
				"c.%[1]sEndpoint = middlewareChain(append(opts.genericEndpointMiddleware, opts.%[2]sEndpointMiddleware...))(c.%[1]sEndpoint)\n",
				m.LcName,
				m.NameUnExport,
			)
		}

		g.W("return c, nil\n")
		g.W("}\n")
	}
	return nil
}

func (g *jsonRPCGoClient) PkgName() string {
	return ""
}

func (g *jsonRPCGoClient) OutputDir() string {
	return ""
}

func (g *jsonRPCGoClient) Filename() string {
	return "client_gen.go"
}

func (g *jsonRPCGoClient) SetImporter(i *importer.Importer) {
	g.i = i
}

func NewJsonRPCGoClient(options jsonRPCGoClientOptionsGateway) generator.Generator {
	return &jsonRPCGoClient{
		options: options,
	}
}
