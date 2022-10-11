package generator

import (
	"context"

	"github.com/swipe-io/swipe/v3/internal/plugin/echo/config"
	"github.com/swipe-io/swipe/v3/option"
	"github.com/swipe-io/swipe/v3/swipe"
	"github.com/swipe-io/swipe/v3/writer"
)

type ClientHelpers struct {
	w          writer.GoWriter
	Interfaces []*config.Interface
	Output     string
	Pkg        string
}

func (g *ClientHelpers) Generate(ctx context.Context) []byte {
	importer := ctx.Value(swipe.ImporterKey).(swipe.Importer)

	clientOptionType := "ClientOption"

	g.w.W("type Option func (*opts)\n\n")

	g.w.W("type opts struct {\n")

	g.w.W("}\n\n")

	for _, iface := range g.Interfaces {
		ifaceType := iface.Named.Type.(*option.IfaceType)

		for _, m := range ifaceType.Methods {
			name := LcNameIfaceMethod(iface, m)
			clientOptType := name + "Opts"

			g.w.W("type %s struct { opts }\n\n", clientOptType)
		}
	}

	g.w.W("type %s func(*clientOpts)\n", clientOptionType)

	g.w.W("func GenericClientOptions(opt ...Option) %s {\nreturn func(c *clientOpts) {\nfor _, o := range opt {\no(&c.genericOpts)\n}\n}\n}\n\n", clientOptionType)

	g.w.W("type clientOpts struct {\n")
	g.w.W("genericOpts opts\n")

	for _, iface := range g.Interfaces {
		ifaceType := iface.Named.Type.(*option.IfaceType)
		for _, m := range ifaceType.Methods {
			g.w.W("%[1]s %[1]s\n", LcNameWithAppPrefix(iface)+m.Name.Value+"Opts")
		}
	}

	g.w.W("}\n\n")

	for _, iface := range g.Interfaces {
		ifaceType := iface.Named.Type.(*option.IfaceType)
		for _, m := range ifaceType.Methods {
			clientOptName := LcNameIfaceMethod(iface, m) + "Opts"
			clientOptFuncName := UcNameIfaceMethod(iface, m)
			g.w.W("func %sOptions(opt ...Option) ClientOption {\nreturn func(c *clientOpts) {\nfor _, o := range opt {\no(&c.%s.opts)\n}\n}\n}\n\n", clientOptFuncName, clientOptName)
		}
	}

	g.w.W("type httpError struct {\n")
	g.w.W("code int\n")
	g.w.W("}\n")

	httpPkg := importer.Import("http", "net/http")
	g.w.W("func (e *httpError) Error() string {\nreturn %s.StatusText(e.code)\n}\n", httpPkg)

	g.w.W("func (e *httpError) StatusCode() int {\nreturn e.code\n}\n")

	errorDecodeParams := []string{"code", "int", "errCode", "string"}

	for _, iface := range g.Interfaces {
		ifaceType := iface.Named.Type.(*option.IfaceType)

		for _, m := range ifaceType.Methods {

			g.w.W("func %sErrorDecode(", LcNameIfaceMethod(iface, m))

			for i := 0; i < len(errorDecodeParams); i += 2 {
				if i > 0 {
					g.w.W(",")
				}
				g.w.W("%s %s", errorDecodeParams[i], errorDecodeParams[i+1])
			}

			g.w.W(") (err error) {\n")

			g.w.W("switch code {\n")
			g.w.W("default:\nerr = &httpError{code: code}\n")

			g.w.W("}\n")
			g.w.W("return\n")
			g.w.W("}\n")
		}
	}

	return g.w.Bytes()
}

func (g *ClientHelpers) Package() string {
	return g.Pkg
}

func (g *ClientHelpers) OutputPath() string {
	return g.Output
}

func (g *ClientHelpers) Filename() string {
	return "client_helpers.go"
}
