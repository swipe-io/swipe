package generator

import (
	"context"
	"fmt"

	"github.com/swipe-io/swipe/v2/internal/domain/model"
	"github.com/swipe-io/swipe/v2/internal/importer"
	"github.com/swipe-io/swipe/v2/internal/usecase/generator"
	"github.com/swipe-io/swipe/v2/internal/writer"
)

type httpTransportOptionsGateway interface {
	AppID() string
	Interfaces() model.Interfaces
	JSONRPCEnable() bool
	GoClientEnable() bool
	UseFast() bool
	Error(uint32) *model.HTTPError
	ErrorKeys() []uint32
}

type httpTransport struct {
	writer.GoLangWriter
	options httpTransportOptionsGateway
	i       *importer.Importer
}

func (g *httpTransport) Prepare(ctx context.Context) error {
	return nil
}

func (g *httpTransport) Process(ctx context.Context) error {
	var (
		kitHTTPPkg string
	)
	if g.options.JSONRPCEnable() {
		if g.options.UseFast() {
			kitHTTPPkg = g.i.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/fasthttp/jsonrpc")
		} else {
			kitHTTPPkg = g.i.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/http/jsonrpc")
		}
	} else {
		if g.options.UseFast() {
			kitHTTPPkg = g.i.Import("fasthttp", "github.com/l-vitaly/go-kit/transport/fasthttp")
		} else {
			kitHTTPPkg = g.i.Import("http", "github.com/go-kit/kit/transport/http")
		}
	}

	endpointPkg := g.i.Import("endpoint", "github.com/go-kit/kit/endpoint")

	if g.options.GoClientEnable() {
		g.W("type httpError struct {\n")
		g.W("code int\n")
		if g.options.JSONRPCEnable() {
			g.W("data interface{}\n")
			g.W("message string\n")
		}
		g.W("}\n")

		if g.options.JSONRPCEnable() {
			g.W("func (e *httpError) Error() string {\nreturn e.message\n}\n")
		} else {
			if g.options.UseFast() {
				httpPkg := g.i.Import("fasthttp", "github.com/valyala/fasthttp")
				g.W("func (e *httpError) Error() string {\nreturn %s.StatusMessage(e.code)\n}\n", httpPkg)
			} else {
				httpPkg := g.i.Import("http", "net/http")
				g.W("func (e *httpError) Error() string {\nreturn %s.StatusText(e.code)\n}\n", httpPkg)
			}
		}

		g.W("func (e *httpError) StatusCode() int {\nreturn e.code\n}\n")

		errorDecodeParams := []string{"code", "int"}
		if g.options.JSONRPCEnable() {
			g.W("func (e *httpError) ErrorData() interface{} {\nreturn e.data\n}\n")
			g.W("func (e *httpError) SetErrorData(data interface{}) {\ne.data = data\n}\n")
			g.W("func (e *httpError) SetErrorMessage(message string) {\ne.message = message\n}\n")

			errorDecodeParams = append(errorDecodeParams, "message", "string", "data", "interface{}")
		}

		g.WriteFunc("ErrorDecode", "", errorDecodeParams, []string{"err", "error"}, func() {
			g.W("switch code {\n")
			g.W("default:\nerr = &httpError{code: code}\n")
			if g.options.JSONRPCEnable() {
				for _, key := range g.options.ErrorKeys() {
					e := g.options.Error(key)
					g.W("case %d:\n", e.Code)
					pkgName := g.i.Import(e.Named.Obj().Pkg().Name(), e.Named.Obj().Pkg().Path())
					if pkgName != "" {
						pkgName += "."
					}
					newPrefix := ""
					if e.IsPointer {
						newPrefix = "&"
					}
					g.W("err = %s%s%s{}\n", newPrefix, pkgName, e.Named.Obj().Name())
				}
			}

			g.W("}\n")
			if g.options.JSONRPCEnable() {
				g.W("if err, ok := err.(interface{ SetErrorData(data interface{}) }); ok {\n")
				g.W("err.SetErrorData(data)\n")
				g.W("}\n")

				g.W("if err, ok := err.(interface{ SetErrorMessage(message string) }); ok {\n")
				g.W("err.SetErrorMessage(message)\n")
				g.W("}\n")
			}
			g.W("return")
		})
	}

	g.W("func middlewareChain(middlewares []%[1]s.Middleware) %[1]s.Middleware {\n", endpointPkg)
	g.W("return func(next %[1]s.Endpoint) %[1]s.Endpoint {\n", endpointPkg)
	g.W("if len(middlewares) == 0 {\n")
	g.W("return next\n")
	g.W("}\n")
	g.W("outer := middlewares[0]\n")
	g.W("others := middlewares[1:]\n")
	g.W("for i := len(others) - 1; i >= 0; i-- {\n")
	g.W("next = others[i](next)\n")
	g.W("}\n")
	g.W("return outer(next)\n")
	g.W("}\n")
	g.W("}\n")

	serverOptType := fmt.Sprintf("serverOpts")
	serverOptionType := fmt.Sprintf("ServerOption")
	kithttpServerOption := fmt.Sprintf("%s.ServerOption", kitHTTPPkg)
	endpointMiddlewareOption := fmt.Sprintf("%s.Middleware", endpointPkg)

	g.WriteFunc(
		"GenericServerOptions",
		"",
		[]string{"v", "..." + kithttpServerOption},
		[]string{"", serverOptionType},
		func() {
			g.W("return func(o *%s) { o.genericServerOption = v }\n", serverOptType)
		},
	)

	g.WriteFunc(
		"GenericServerEndpointMiddlewares",
		"",
		[]string{"v", "..." + endpointMiddlewareOption},
		[]string{"", serverOptionType},
		func() {
			g.W("return func(o *%s) { o.genericEndpointMiddleware = v }\n", serverOptType)
		},
	)

	g.W("type %s func (*%s)\n", serverOptionType, serverOptType)

	g.W("type %s struct {\n", serverOptType)
	g.W("genericServerOption []%s\n", kithttpServerOption)
	g.W("genericEndpointMiddleware []%s\n", endpointMiddlewareOption)

	for i := 0; i < g.options.Interfaces().Len(); i++ {
		iface := g.options.Interfaces().At(i)

		for _, m := range iface.Methods() {
			g.W("%sServerOption []%s\n", m.NameUnExport, kithttpServerOption)
			g.W("%sEndpointMiddleware []%s\n", m.NameUnExport, endpointMiddlewareOption)
		}
	}
	g.W("}\n")

	for i := 0; i < g.options.Interfaces().Len(); i++ {
		iface := g.options.Interfaces().At(i)
		for _, m := range iface.Methods() {
			g.WriteFunc(
				fmt.Sprintf("%sServerOptions", m.NameExport),
				"",
				[]string{"opt", "..." + kithttpServerOption},
				[]string{"", serverOptionType},
				func() {
					g.W("return func(c *%s) { c.%sServerOption = opt }\n", serverOptType, m.NameUnExport)
				},
			)

			g.WriteFunc(
				fmt.Sprintf("%sServerEndpointMiddlewares", m.NameExport),
				"",
				[]string{"opt", "..." + endpointMiddlewareOption},
				[]string{"", serverOptionType},
				func() {
					g.W("return func(c *%s) { c.%sEndpointMiddleware = opt }\n", serverOptType, m.NameUnExport)
				},
			)
		}
	}

	return nil
}

func (g *httpTransport) PkgName() string {
	return ""
}

func (g *httpTransport) OutputDir() string {
	return ""
}

func (g *httpTransport) Filename() string {
	return "http_gen.go"
}

func (g *httpTransport) SetImporter(i *importer.Importer) {
	g.i = i
}

func NewHttpTransport(
	options httpTransportOptionsGateway,
) generator.Generator {
	return &httpTransport{
		options: options,
	}
}
