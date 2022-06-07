package generator

import (
	"context"
	"fmt"

	"github.com/swipe-io/swipe/v3/internal/plugin/gokit/config"
	"github.com/swipe-io/swipe/v3/option"
	"github.com/swipe-io/swipe/v3/swipe"
	"github.com/swipe-io/swipe/v3/writer"
)

type ServerHelpers struct {
	w                writer.GoWriter
	Interfaces       []*config.Interface
	JSONRPCEnable    bool
	HTTPServerEnable bool
	UseFast          bool
	Output           string
	Pkg              string
}

func (g *ServerHelpers) Generate(ctx context.Context) []byte {
	importer := ctx.Value(swipe.ImporterKey).(swipe.Importer)
	endpointPkg := importer.Import("endpoint", "github.com/go-kit/kit/endpoint")

	if g.HTTPServerEnable {
		var (
			kitHTTPPkg string
		)
		if g.JSONRPCEnable {
			if g.UseFast {
				kitHTTPPkg = importer.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/fasthttp/jsonrpc")
			} else {
				kitHTTPPkg = importer.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/http/jsonrpc")
			}
		} else {
			if g.UseFast {
				kitHTTPPkg = importer.Import("fasthttp", "github.com/l-vitaly/go-kit/transport/fasthttp")
			} else {
				kitHTTPPkg = importer.Import("http", "github.com/go-kit/kit/transport/http")
			}
		}

		kitHTTPServerOption := fmt.Sprintf("%s.ServerOption", kitHTTPPkg)
		endpointMiddlewareOption := fmt.Sprintf("%s.Middleware", endpointPkg)
		endpointOption := fmt.Sprintf("%s.Endpoint", endpointPkg)
		serverOptType := "serverOpts"

		g.w.W("type Option func (*opts)\n\n")

		g.w.W("func ServerOptions(opt ...%s) Option {\n", kitHTTPServerOption)
		g.w.W("return func(c *opts) { c.serverOption = opt }\n")
		g.w.W("}\n")

		g.w.W("func MiddlewareOption(opt ...%s) Option {\n", endpointMiddlewareOption)
		g.w.W("return func(c *opts) { c.endpointMiddleware = opt }\n")
		g.w.W("}\n")

		g.w.W("type opts struct {\n")
		g.w.W("serverOption []%s\n", kitHTTPServerOption)
		g.w.W("endpoint %s\n", endpointOption)
		g.w.W("endpointMiddleware []%s\n", endpointMiddlewareOption)

		g.w.W("}\n\n")

		for _, iface := range g.Interfaces {
			ifaceType := iface.Named.Type.(*option.IfaceType)

			for _, m := range ifaceType.Methods {
				g.w.W("type %s struct { opts }\n\n", LcNameIfaceMethod(iface, m)+"Opts")
			}
		}

		g.w.W("type ServerOption func(*%s)\n\n", serverOptType)

		g.w.W("func GenericServerOptions(opt ...Option) ServerOption {\nreturn func(c *serverOpts) {\nfor _, o := range opt {\no(&c.genericOpts)\n}\n}\n}\n\n")
		g.w.W("func ErrorEncoderOption(opt %s.ErrorEncoder) ServerOption {\nreturn func(c *serverOpts) {\n c.errorEncoder = opt\n}\n}\n\n", kitHTTPPkg)

		g.w.W("type %s struct {\n", serverOptType)
		g.w.W("errorEncoder %s.ErrorEncoder\n", kitHTTPPkg)
		g.w.W("genericOpts opts\n")
		for _, iface := range g.Interfaces {
			ifaceType := iface.Named.Type.(*option.IfaceType)
			for _, m := range ifaceType.Methods {
				serverOpt := LcNameWithAppPrefix(iface) + m.Name.Value + "Opts"
				g.w.W("%[1]s %[1]s\n", serverOpt)
			}
		}
		g.w.W("}\n")

		for _, iface := range g.Interfaces {
			ifaceType := iface.Named.Type.(*option.IfaceType)
			for _, m := range ifaceType.Methods {
				serverOptName := LcNameWithAppPrefix(iface) + m.Name.Value + "Opts"
				serverOptFuncName := UcNameWithAppPrefix(iface) + m.Name.Value
				g.w.W("func %sOptions(opt ...Option) ServerOption {\nreturn func(c *serverOpts) {\nfor _, o := range opt {\no(&c.%s.opts)\n}\n}\n}\n\n", serverOptFuncName, serverOptName)
			}
		}
	}

	return g.w.Bytes()
}

func (g *ServerHelpers) Package() string {
	return g.Pkg
}

func (g *ServerHelpers) OutputDir() string {
	return g.Output
}

func (g *ServerHelpers) Filename() string {
	return "server_helpers.go"
}
