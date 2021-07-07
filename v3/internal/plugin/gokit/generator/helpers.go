package generator

import (
	"context"
	"fmt"

	"github.com/swipe-io/swipe/v3/internal/plugin/gokit/config"
	"github.com/swipe-io/swipe/v3/option"
	"github.com/swipe-io/swipe/v3/swipe"
	"github.com/swipe-io/swipe/v3/writer"
)

type Helpers struct {
	w                writer.GoWriter
	Interfaces       []*config.Interface
	JSONRPCEnable    bool
	GoClientEnable   bool
	HTTPServerEnable bool
	UseFast          bool
	IfaceErrors      map[string]map[string][]config.Error
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

	if g.HTTPServerEnable {
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
				name := LcNameWithAppPrefix(iface) + m.Name.Value
				g.w.W("%sServerOption []%s\n", name, kitHTTPServerOption)
				g.w.W("%sEndpointMiddleware []%s\n", name, endpointMiddlewareOption)
			}
		}
		g.w.W("}\n")

		g.w.W("func GenericServerOptions(v ...%s) %s {\n", kitHTTPServerOption, serverOptionType)
		g.w.W("return func(o *%s) { o.genericServerOption = v }\n", serverOptType)
		g.w.W("}\n")

		g.w.W("func GenericServerEndpointMiddlewares(v ...%s) %s {\n", endpointMiddlewareOption, serverOptionType)
		g.w.W("return func(o *%s) { o.genericEndpointMiddleware = v }\n", serverOptType)
		g.w.W("}\n")

		for _, iface := range g.Interfaces {
			ifaceType := iface.Named.Type.(*option.IfaceType)
			for _, m := range ifaceType.Methods {
				fnPrefix := UcNameWithAppPrefix(iface) + m.Name.Value
				paramPrefix := LcNameWithAppPrefix(iface) + m.Name.Value

				g.w.W("func %sServerOptions(opt ...%s) %s {\n", fnPrefix, kitHTTPServerOption, serverOptionType)
				g.w.W("return func(c *%s) { c.%sServerOption = opt }\n", serverOptType, paramPrefix)
				g.w.W("}\n")

				g.w.W("func %sServerEndpointMiddlewares(opt ...%s) %s {\n", fnPrefix, endpointMiddlewareOption, serverOptionType)
				g.w.W("return func(c *%s) { c.%sEndpointMiddleware = opt }\n", serverOptType, paramPrefix)
				g.w.W("}\n")
			}
		}
	}

	if g.GoClientEnable {
		g.w.W("type httpError struct {\n")
		g.w.W("code int\n")
		if g.JSONRPCEnable {
			g.w.W("data interface{}\n")
			g.w.W("message string\n")
		}
		g.w.W("}\n")

		if g.JSONRPCEnable {
			g.w.W("func (e *httpError) Error() string {\nreturn e.message\n}\n")
		} else {
			if g.UseFast {
				httpPkg := importer.Import("fasthttp", "github.com/valyala/fasthttp")
				g.w.W("func (e *httpError) Error() string {\nreturn %s.StatusMessage(e.code)\n}\n", httpPkg)
			} else {
				httpPkg := importer.Import("http", "net/http")
				g.w.W("func (e *httpError) Error() string {\nreturn %s.StatusText(e.code)\n}\n", httpPkg)
			}
		}

		g.w.W("func (e *httpError) StatusCode() int {\nreturn e.code\n}\n")

		errorDecodeParams := []string{"code", "int"}
		if g.JSONRPCEnable {
			g.w.W("func (e *httpError) ErrorData() interface{} {\nreturn e.data\n}\n")
			g.w.W("func (e *httpError) SetErrorData(data interface{}) {\ne.data = data\n}\n")
			g.w.W("func (e *httpError) SetErrorMessage(message string) {\ne.message = message\n}\n")

			errorDecodeParams = append(errorDecodeParams, "message", "string", "data", "interface{}")
		}

		for _, iface := range g.Interfaces {
			ifaceType := iface.Named.Type.(*option.IfaceType)
			ifaceErrors := g.IfaceErrors[iface.Named.Name.Value]

			for _, m := range ifaceType.Methods {
				methodErrors := ifaceErrors[m.Name.Value]

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
				if g.JSONRPCEnable {
					for _, e := range methodErrors {
						g.w.W("case %d:\n", e.Code)
						pkgName := importer.Import(e.PkgName, e.PkgPath)
						if pkgName != "" {
							pkgName += "."
						}
						newPrefix := ""
						if e.IsPointer {
							newPrefix = "&"
						}
						g.w.W("err = %s%s%s{}\n", newPrefix, pkgName, e.Name)
					}
				}
				g.w.W("}\n")
				if g.JSONRPCEnable {
					g.w.W("if err, ok := err.(interface{ SetErrorData(data interface{}) }); ok {\n")
					g.w.W("err.SetErrorData(data)\n")
					g.w.W("}\n")

					g.w.W("if err, ok := err.(interface{ SetErrorMessage(message string) }); ok {\n")
					g.w.W("err.SetErrorMessage(message)\n")
					g.w.W("}\n")
				}
				g.w.W("return\n")
				g.w.W("}\n")
			}
		}
	}

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
