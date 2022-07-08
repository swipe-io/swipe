package generator

import (
	"context"
	"strconv"

	"github.com/swipe-io/swipe/v3/internal/plugin/gokit/config"
	"github.com/swipe-io/swipe/v3/option"
	"github.com/swipe-io/swipe/v3/swipe"
	"github.com/swipe-io/swipe/v3/writer"
)

type JSONRPCClientGenerator struct {
	w          writer.GoWriter
	Interfaces []*config.Interface
	UseFast    bool
	Output     string
	Pkg        string
}

func (g *JSONRPCClientGenerator) Package() string {
	return g.Pkg
}

func (g *JSONRPCClientGenerator) Generate(ctx context.Context) []byte {
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

	importer := ctx.Value(swipe.ImporterKey).(swipe.Importer)

	for _, iface := range g.Interfaces {
		ifaceType := iface.Named.Type.(*option.IfaceType)

		clientType := ClientType(iface)

		constructPostfix := UcNameWithAppPrefix(iface)
		if len(g.Interfaces) == 1 {
			constructPostfix = ""
		}

		g.w.W("func NewClientJSONRPC%s(tgt string", constructPostfix)
		g.w.W(" ,options ...ClientOption")
		g.w.W(") (*%s, error) {\n", clientType)
		g.w.W("opts := &clientOpts{}\n")
		g.w.W("c := &%s{}\n", clientType)
		g.w.W("for _, o := range options {\n")
		g.w.W("o(opts)\n")
		g.w.W("}\n")

		if g.UseFast {
			jsonrpcPkg = importer.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/fasthttp/jsonrpc")
		} else {
			jsonrpcPkg = importer.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/http/jsonrpc")
		}
		urlPkg = importer.Import("url", "net/url")
		contextPkg = importer.Import("context", "context")
		ffJSONPkg = importer.Import("ffjson", "github.com/pquerna/ffjson/ffjson")
		jsonPkg = importer.Import("json", "encoding/json")
		fmtPkg = importer.Import("fmt", "fmt")
		netPkg = importer.Import("net", "net")
		stringsPkg = importer.Import("strings", "strings")

		g.w.W("if %s.HasPrefix(tgt, \"[\") {\n", stringsPkg)
		g.w.W("host, port, err := %s.SplitHostPort(tgt)\n", netPkg)
		g.w.WriteCheckErr("err", func() {
			g.w.W("return nil, err")
		})
		g.w.W("tgt = host + \":\" + port\n")
		g.w.W("}\n")

		g.w.W("u, err := %s.Parse(tgt)\n", urlPkg)
		g.w.WriteCheckErr("err", func() {
			g.w.W("return nil, err")
		})
		g.w.W("if u.Scheme == \"\" {\n")
		g.w.W("u.Scheme = \"https\"")
		g.w.W("}\n")

		for _, m := range ifaceType.Methods {
			g.w.W("opts.%[1]sOpts.clientOption = append(\nopts.%[1]sOpts.clientOption,\n", LcNameIfaceMethod(iface, m))
			g.w.W("%s.ClientRequestEncoder(", jsonrpcPkg)
			g.w.W("func(_ %s.Context, obj interface{}) (%s.RawMessage, error) {\n", contextPkg, jsonPkg)

			requestName := NameRequest(m, iface)
			paramsLen := LenWithoutContexts(m.Sig.Params)
			resultsLen := LenWithoutErrors(m.Sig.Results)

			if paramsLen > 0 {
				g.w.W("req, ok := obj.(%s)\n", requestName)
				g.w.W("if !ok {\n")
				g.w.W("return nil, %s.Errorf(\"couldn't assert request as %s, got %%T\", obj)\n", fmtPkg, requestName)
				g.w.W("}\n")
				g.w.W("b, err := %s.Marshal(req)\n", ffJSONPkg)
				g.w.W("if err != nil {\n")
				g.w.W("return nil, %s.Errorf(\"couldn't marshal request %%T: %%s\", obj, err)\n", fmtPkg)
				g.w.W("}\n")
				g.w.W("return b, nil\n")
			} else {
				g.w.W("return nil, nil\n")
			}
			g.w.W("}),\n")

			g.w.W("%s.ClientResponseDecoder(", jsonrpcPkg)
			g.w.W("func(_ %s.Context, response %s.Response) (interface{}, error) {\n", contextPkg, jsonrpcPkg)
			g.w.W("if response.Error != nil {\n")
			g.w.W("return nil, %sErrorDecode(response.Error.Code, response.Error.Message, response.Error.Data)\n", LcNameIfaceMethod(iface, m))
			g.w.W("}\n")

			if resultsLen > 0 {
				var responseType string
				responseName := NameResponse(m, iface)
				if m.Sig.IsNamed && resultsLen > 1 {
					responseType = responseName
				} else {
					responseType = swipe.TypeString(m.Sig.Results[0].Type, false, importer)
				}
				g.w.W("var resp %s\n", responseType)
				g.w.W("err := %s.Unmarshal(response.Result, &resp)\n", ffJSONPkg)
				g.w.W("if err != nil {\n")
				g.w.W("return nil, %s.Errorf(\"couldn't unmarshal body to %s: %%s\", err)\n", fmtPkg, responseName)
				g.w.W("}\n")
				g.w.W("return resp, nil\n")
			} else {
				g.w.W("return nil, nil\n")
			}
			g.w.W("}),\n")
			g.w.W(")\n")

			methodName := m.Name.Lower()
			if iface.Namespace != "" {
				methodName = iface.Namespace + "." + methodName
			}

			g.w.W("c.%sEndpoint = %s.NewClient(\n", LcNameIfaceMethod(iface, m), jsonrpcPkg)
			g.w.W("u,\n")
			g.w.W("%s,\n", strconv.Quote(methodName))
			g.w.W("append(opts.genericOpts.clientOption, opts.%sOpts.clientOption...)...,\n", LcNameIfaceMethod(iface, m))
			g.w.W(").Endpoint()\n")

			g.w.W(
				"c.%[1]sEndpoint = middlewareChain(append(opts.genericOpts.endpointMiddleware, opts.%[1]sOpts.endpointMiddleware...))(c.%[1]sEndpoint)\n",
				LcNameIfaceMethod(iface, m),
			)
		}

		g.w.W("return c, nil\n")
		g.w.W("}\n")
	}
	return g.w.Bytes()
}

func (g *JSONRPCClientGenerator) OutputPath() string {
	return g.Output
}

func (g *JSONRPCClientGenerator) Filename() string {
	return "jsonrpc_client.go"
}
