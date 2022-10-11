package generator

import (
	"context"

	"github.com/swipe-io/swipe/v3/internal/plugin/echo/config"
	"github.com/swipe-io/swipe/v3/swipe"
	"github.com/swipe-io/swipe/v3/writer"
)

type ClientGenerator struct {
	w          writer.GoWriter
	Interfaces []*config.Interface
	Output     string
	Pkg        string
}

func (g *ClientGenerator) Generate(ctx context.Context) []byte {
	importer := ctx.Value(swipe.ImporterKey).(swipe.Importer)

	//fmtPkg := importer.Import("fmt", "fmt")
	//contextPkg := importer.Import("context", "context")
	urlPkg := importer.Import("url", "net/url")
	netPkg := importer.Import("net", "net")
	stringsPkg := importer.Import("strings", "strings")
	//jsonPkg := importer.Import("json", "encoding/json")
	//httpPkg = importer.Import("http", "net/http")

	//g.w.W("type clientErrorWrapper struct {\n")
	//g.w.W("Error string `json:\"error\"`\n")
	//g.w.W("Code string `json:\"code,omitempty\"`\n")
	//g.w.W("Data interface{} `json:\"data,omitempty\"`\n")
	//g.w.W("}\n")

	for _, iface := range g.Interfaces {
		//ifaceType := iface.Named.Type.(*option.IfaceType)
		clientType := ClientType(iface)

		constructPostfix := UcNameWithAppPrefix(iface)
		if len(g.Interfaces) == 1 {
			constructPostfix = ""
		}

		g.w.W("func NewClientREST%s(tgt string", constructPostfix)
		g.w.W(" ,options ...ClientOption")
		g.w.W(") (*%s, error) {\n", clientType)
		g.w.W("opts := &clientOpts{}\n")

		g.w.W("for _, o := range options {\n")
		g.w.W("o(opts)\n")
		g.w.W("}\n")

		g.w.W("if %s.HasPrefix(tgt, \"[\") {\n", stringsPkg)
		g.w.W("host, port, err := %s.SplitHostPort(tgt)\n", netPkg)
		g.w.WriteCheckErr("err", func() {
			g.w.W("return nil, err\n")
		})
		g.w.W("tgt = host + \":\" + port\n")
		g.w.W("}\n")

		g.w.W("u, err := %s.Parse(tgt)\n", urlPkg)

		g.w.WriteCheckErr("err", func() {
			g.w.W("return nil, err")
		})

		g.w.W("if u.Scheme == \"\" {\n")
		g.w.W("u.Scheme = \"https\"\n")
		g.w.W("}\n")

		g.w.W("return &%s{u: u}, nil\n}\n\n", clientType)
	}
	return g.w.Bytes()
}

func (g *ClientGenerator) Package() string {
	return g.Pkg
}

func (g *ClientGenerator) OutputPath() string {
	return g.Output
}

func (g *ClientGenerator) Filename() string {
	return "rest_client.go"
}
