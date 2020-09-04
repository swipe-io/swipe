package generator

import (
	"context"
	"fmt"

	"github.com/gogo/protobuf/sortkeys"

	"github.com/swipe-io/swipe/pkg/domain/model"
	"github.com/swipe-io/swipe/pkg/importer"
	"github.com/swipe-io/swipe/pkg/writer"
)

type httpTransport struct {
	*writer.GoLangWriter
	filename string
	info     model.GenerateInfo
	i        *importer.Importer
	o        model.ServiceOption
}

func (g *httpTransport) Prepare(ctx context.Context) error {
	return nil
}

func (g *httpTransport) Process(ctx context.Context) error {
	var (
		kithttpPkg string
		httpPkg    string
	)
	transportOpt := g.o.Transport

	if transportOpt.JsonRPC.Enable {
		if transportOpt.FastHTTP {
			kithttpPkg = g.i.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/fasthttp/jsonrpc")
		} else {
			kithttpPkg = g.i.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/http/jsonrpc")
		}
	} else {
		if transportOpt.FastHTTP {
			kithttpPkg = g.i.Import("fasthttp", "github.com/l-vitaly/go-kit/transport/fasthttp")
		} else {
			kithttpPkg = g.i.Import("http", "github.com/go-kit/kit/transport/http")
		}
	}

	if transportOpt.FastHTTP {
		httpPkg = g.i.Import("fasthttp", "github.com/valyala/fasthttp")
	} else {
		httpPkg = g.i.Import("http", "net/http")
	}

	endpointPkg := g.i.Import("endpoint", "github.com/go-kit/kit/endpoint")

	g.W("type httpError struct {\n")
	g.W("code int\n")
	if transportOpt.JsonRPC.Enable {
		g.W("data interface{}\n")
		g.W("message string\n")
	}
	g.W("}\n")

	if transportOpt.JsonRPC.Enable {
		g.W("func (e *httpError) Error() string {\nreturn e.message\n}\n")
	} else {
		if transportOpt.FastHTTP {
			g.W("func (e *httpError) Error() string {\nreturn %s.StatusMessage(e.code)\n}\n", httpPkg)
		} else {
			g.W("func (e *httpError) Error() string {\nreturn %s.StatusText(e.code)\n}\n", httpPkg)
		}
	}

	g.W("func (e *httpError) StatusCode() int {\nreturn e.code\n}\n")

	errorDecodeParams := []string{"code", "int"}
	if transportOpt.JsonRPC.Enable {
		g.W("func (e *httpError) ErrorData() interface{} {\nreturn e.data\n}\n")
		g.W("func (e *httpError) SetErrorData(data interface{}) {\ne.data = data\n}\n")
		g.W("func (e *httpError) SetErrorMessage(message string) {\ne.message = message\n}\n")

		errorDecodeParams = append(errorDecodeParams, "message", "string", "data", "interface{}")
	}

	g.WriteFunc("ErrorDecode", "", errorDecodeParams, []string{"err", "error"}, func() {
		g.W("switch code {\n")
		g.W("default:\nerr = &httpError{code: code}\n")
		var errorKeys []uint32
		for key := range g.o.Transport.Errors {
			errorKeys = append(errorKeys, key)
		}
		sortkeys.Uint32s(errorKeys)
		for _, key := range errorKeys {
			e := g.o.Transport.Errors[key]
			g.W("case %d:\n", e.Code)
			pkg := g.i.Import(e.Named.Obj().Pkg().Name(), e.Named.Obj().Pkg().Path())
			newPrefix := ""
			if e.IsPointer {
				newPrefix = "&"
			}
			g.W("err = %s%s.%s{}\n", newPrefix, pkg, e.Named.Obj().Name())
		}
		g.W("}\n")
		if transportOpt.JsonRPC.Enable {
			g.W("if err, ok := err.(%s.ErrorData); ok {\n", kithttpPkg)
			g.W("err.SetErrorData(data)\n")
			g.W("}\n")

			g.W("if err, ok := err.(%s.ErrorMessager); ok {\n", kithttpPkg)
			g.W("err.SetErrorMessage(message)\n")
			g.W("}\n")
		}
		g.W("return")
	})

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

	serverOptType := fmt.Sprintf("server%sOpts", g.o.ID)
	serverOptionType := fmt.Sprintf("%sServerOption", g.o.ID)
	kithttpServerOption := fmt.Sprintf("%s.ServerOption", kithttpPkg)
	endpointMiddlewareOption := fmt.Sprintf("%s.Middleware", endpointPkg)

	g.W("type %s func (*%s)\n", serverOptionType, serverOptType)

	g.W("type %s struct {\n", serverOptType)
	g.W("genericServerOption []%s\n", kithttpServerOption)
	g.W("genericEndpointMiddleware []%s\n", endpointMiddlewareOption)

	for _, m := range g.o.Methods {
		g.W("%sServerOption []%s\n", m.LcName, kithttpServerOption)
		g.W("%sEndpointMiddleware []%s\n", m.LcName, endpointMiddlewareOption)
	}
	g.W("}\n")

	g.WriteFunc(
		g.o.ID+"GenericServerOptions",
		"",
		[]string{"v", "..." + kithttpServerOption},
		[]string{"", serverOptionType},
		func() {
			g.W("return func(o *%s) { o.genericServerOption = v }\n", serverOptType)
		},
	)

	g.WriteFunc(
		g.o.ID+"GenericServerEndpointMiddlewares",
		"",
		[]string{"v", "..." + endpointMiddlewareOption},
		[]string{"", serverOptionType},
		func() {
			g.W("return func(o *%s) { o.genericEndpointMiddleware = v }\n", serverOptType)
		},
	)

	for _, m := range g.o.Methods {
		g.WriteFunc(
			g.o.ID+m.Name+"ServerOptions",
			"",
			[]string{"opt", "..." + kithttpServerOption},
			[]string{"", serverOptionType},
			func() {
				g.W("return func(c *%s) { c.%sServerOption = opt }\n", serverOptType, m.LcName)
			},
		)

		g.WriteFunc(
			g.o.ID+m.Name+"ServerEndpointMiddlewares",
			"",
			[]string{"opt", "..." + endpointMiddlewareOption},
			[]string{"", serverOptionType},
			func() {
				g.W("return func(c *%s) { c.%sEndpointMiddleware = opt }\n", serverOptType, m.LcName)
			},
		)
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
	return g.filename
}

func (g *httpTransport) SetImporter(i *importer.Importer) {
	g.i = i
}

func NewHttpTransport(filename string, info model.GenerateInfo, o model.ServiceOption) Generator {
	return &httpTransport{GoLangWriter: writer.NewGoLangWriter(), filename: filename, info: info, o: o}
}
