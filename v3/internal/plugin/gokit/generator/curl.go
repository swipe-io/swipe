package generator

import (
	"context"
	"strings"

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
	URL           string
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

func (g *CURL) buildBody(vars option.VarsType) map[string]interface{} {
	body := map[string]interface{}{}
	for _, v := range vars {
		if v.IsContext {
			continue
		}
		body[v.Name.Value] = curlTypeValue(v.Type)
	}
	return body
}

func (g *CURL) writeCURLJSONRPC(iface *config.Interface) {
	ifaceType := iface.Named.Type.(*option.IfaceType)
	for _, m := range ifaceType.Methods {
		methodName := m.Name.Lower()
		if iface.Namespace != "" {
			methodName = iface.Namespace + "." + methodName
		}

		body := g.buildBody(m.Sig.Params)
		result := curlbuilder.New().
			SetMethod("POST").
			SetURL(strings.TrimRight(g.URL, "/") + "/" + strings.TrimLeft(g.JSONRPCPath, "/")).
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
	ifaceType := iface.Named.Type.(*option.IfaceType)
	for _, m := range ifaceType.Methods {
		mopt := g.MethodOptions[iface.Named.Name.Value+m.Name.Value]
		body := g.buildBody(m.Sig.Params)
		methodName := m.Name.Lower()
		if iface.Namespace != "" {
			methodName = iface.Namespace + "." + methodName
		}
		result := curlbuilder.New().
			SetMethod(mopt.RESTMethod.Take()).
			SetURL(strings.TrimRight(g.URL, "/") + "/" + strings.TrimLeft(mopt.RESTPath.Take(), "/")).
			SetBody(body).
			String()

		g.w.W("## %s\n\n", methodName)
		g.w.W("```\n%s\n```\n", result)
	}
}

func (g *CURL) OutputDir() string {
	return g.Output
}

func (g *CURL) Filename() string {
	return "curl.md"
}
