package generator

import (
	"context"
	"fmt"

	"github.com/swipe-io/swipe/v2/internal/option"
	"github.com/swipe-io/swipe/v2/internal/plugin/gokit/config"
	"github.com/swipe-io/swipe/v2/internal/swipe"
	"github.com/swipe-io/swipe/v2/internal/writer"
)

type Helpers struct {
	w              writer.GoWriter
	Interfaces     []*config.Interface
	JSONRPCEnable  bool
	GoClientEnable bool
	UseFast        bool
}

func (g *Helpers) Generate(ctx context.Context) []byte {
	var (
		kitHTTPPkg string
	)
	importer := ctx.Value(swipe.ImporterKey).(swipe.Importer)

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
	endpointPkg := importer.Import("endpoint", "github.com/go-kit/kit/endpoint")

	g.writeFuncMiddlewareChain(endpointPkg)

	serverOptType := "serverOpts"
	serverOptionType := "ServerOption"
	kitHTTPServerOption := fmt.Sprintf("%s.ServerOption", kitHTTPPkg)
	endpointMiddlewareOption := fmt.Sprintf("%s.Middleware", endpointPkg)

	g.w.W("type %s func (*%s)\n", serverOptionType, serverOptType)

	g.w.W("type %s struct {\n", serverOptType)
	g.w.W("genericServerOption []%s\n", kitHTTPServerOption)
	g.w.W("genericEndpointMiddleware []%s\n", endpointMiddlewareOption)
	for _, iface := range g.Interfaces {
		ifaceType := iface.Named.Type.(*option.IfaceType)
		for _, m := range ifaceType.Methods {
			name := LcNameWithAppPrefix(iface.Named) + m.Name.Origin
			g.w.W("%sServerOption []%s\n", name, kitHTTPServerOption)
			g.w.W("%sEndpointMiddleware []%s\n", name, endpointMiddlewareOption)
		}
	}
	g.w.W("}\n")

	for _, iface := range g.Interfaces {
		ifaceType := iface.Named.Type.(*option.IfaceType)

		for _, m := range ifaceType.Methods {
			fnPrefix := UcNameWithAppPrefix(iface.Named) + m.Name.Origin
			paramPrefix := LcNameWithAppPrefix(iface.Named) + m.Name.Origin

			g.w.W("func %sServerOptions(opt ...%s) %s {\n", fnPrefix, kitHTTPServerOption, serverOptionType)
			g.w.W("return func(c *%s) { c.%sServerOption = opt }\n", serverOptType, paramPrefix)
			g.w.W("}\n")

			g.w.W("func %sServerEndpointMiddlewares(opt ...%s) %s {\n", fnPrefix, endpointMiddlewareOption, serverOptionType)
			g.w.W("return func(c *%s) { c.%sEndpointMiddleware = opt }\n", serverOptType, paramPrefix)
			g.w.W("}\n")
		}
	}

	g.w.W("func GenericServerOptions(v ...%s) %s {\n", kitHTTPServerOption, serverOptionType)
	g.w.W("return func(o *%s) { o.genericServerOption = v }\n", serverOptType)
	g.w.W("}\n")

	g.w.W("func GenericServerEndpointMiddlewares(v ...%s) %s {\n", endpointMiddlewareOption, serverOptionType)
	g.w.W("return func(o *%s) { o.genericEndpointMiddleware = v }\n", serverOptType)
	g.w.W("}\n")

	return g.w.Bytes()
}

func (g *Helpers) writeFuncMiddlewareChain(endpointPkg string) {
	g.w.W("func middlewareChain(middlewares []%[1]s.Middleware) %[1]s.Middleware {\n", endpointPkg)
	g.w.W("return func(next %[1]s.Endpoint) %[1]s.Endpoint {\n", endpointPkg)
	g.w.W("if len(middlewares) == 0 {\n")
	g.w.W("return next\n")
	g.w.W("}\n")
	g.w.W("outer := middlewares[0]\n")
	g.w.W("others := middlewares[1:]\n")
	g.w.W("for i := len(others) - 1; i >= 0; i-- {\n")
	g.w.W("next = others[i](next)\n")
	g.w.W("}\n")
	g.w.W("return outer(next)\n")
	g.w.W("}\n")
	g.w.W("}\n")
}

func (g *Helpers) OutputDir() string {
	return ""
}

func (g *Helpers) Filename() string {
	return "helpers.go"
}
