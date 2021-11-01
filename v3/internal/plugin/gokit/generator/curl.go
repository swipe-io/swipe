package generator

import (
	"context"

	"github.com/555f/curlbuilder"
	"github.com/swipe-io/swipe/v3/internal/plugin/gokit/config"
	"github.com/swipe-io/swipe/v3/option"
	"github.com/swipe-io/swipe/v3/writer"
)

type CURL struct {
	w             writer.TextWriter
	Interfaces    []*config.Interface
	MethodOptions map[string]config.MethodOptions
	JSONRPCEnable bool
	JSONRPCPath   string
	Output        string
}

func (g *CURL) Generate(ctx context.Context) []byte {
	for _, iface := range g.Interfaces {

		name := iface.Named.Name.Value
		if iface.ClientName.IsValid() {
			name = iface.ClientName.Take()
		}

		g.w.W("# %s\n\n", name)

		if g.JSONRPCEnable {
			g.writeCURLJSONRPC(iface)
		} else {
			g.writeCURLREST(iface)
		}
	}
	return g.w.Bytes()
}

func (g *CURL) buildBody() {

}

func (g *CURL) writeCURLJSONRPC(iface *config.Interface) {
	ifaceType := iface.Named.Type.(*option.IfaceType)
	for _, m := range ifaceType.Methods {
		//mopt := g.MethodOptions[iface.Named.Name.Value+m.Name.Value]
		body := map[string]interface{}{}
		for _, p := range m.Sig.Params {
			if p.IsContext {
				continue
			}
			body[p.Name.Value] = ""
		}
		methodName := m.Name.Lower()
		if iface.Namespace != "" {
			methodName = iface.Namespace + "." + methodName
		}
		result := curlbuilder.New().
			SetMethod("POST").
			SetURL(g.JSONRPCPath).
			SetBody(map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      1,
				"method":  methodName,
				"params":  body,
			}).
			String()

		g.w.W("## %s\n\n", methodName)
		g.w.W("```\n%s\n```\n", result)
	}
}

func (g *CURL) writeCURLREST(iface *config.Interface) {
	//curlbuilder.New().SetMethod()
}

func (g *CURL) OutputDir() string {
	return g.Output
}

func (g *CURL) Filename() string {
	return "curl.md"
}
