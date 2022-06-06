package generator

import (
	"context"
	"fmt"

	"github.com/swipe-io/swipe/v3/internal/plugin/gokit/config"
	"github.com/swipe-io/swipe/v3/option"
	"github.com/swipe-io/swipe/v3/swipe"
	"github.com/swipe-io/swipe/v3/writer"
)

type ClientHelpers struct {
	w             writer.GoWriter
	Interfaces    []*config.Interface
	JSONRPCEnable bool
	UseFast       bool
	IfaceErrors   map[string]map[string][]config.Error
	Output        string
	Pkg           string
}

func (g *ClientHelpers) Generate(ctx context.Context) []byte {
	importer := ctx.Value(swipe.ImporterKey).(swipe.Importer)

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
	endpointPkg := importer.Import("endpoint", "github.com/go-kit/kit/endpoint")

	kitHTTPClientOption := fmt.Sprintf("%s.ClientOption", kitHTTPPkg)
	endpointMiddlewareOption := fmt.Sprintf("%s.Middleware", endpointPkg)
	clientOptionType := "ClientOption"

	g.w.W("type Option func (*opts)\n\n")

	g.w.W("func ClientOptions(opt ...%s) Option {\n", kitHTTPClientOption)
	g.w.W("return func(c *opts) { c.clientOption = opt }\n")
	g.w.W("}\n")

	g.w.W("func MiddlewareOption(opt ...%s) Option {\n", endpointMiddlewareOption)
	g.w.W("return func(c *opts) { c.endpointMiddleware = opt }\n")
	g.w.W("}\n")

	//if !g.JSONRPCEnable {
	//	g.w.W("func DecodeRequestFuncOption(opt %s.CreateRequestFunc) Option {\n", kitHTTPPkg)
	//	g.w.W("return func(c *opts) { c.createReqFunc = opt }\n")
	//	g.w.W("}\n")
	//
	//	g.w.W("func EncodeResponseFuncOption(opt %s.DecodeResponseFunc) Option {\n", kitHTTPPkg)
	//	g.w.W("return func(c *opts) { c.decRespFunc = opt }\n")
	//	g.w.W("}\n")
	//}

	g.w.W("type opts struct {\n")
	g.w.W("clientOption []%s\n", kitHTTPClientOption)
	g.w.W("endpointMiddleware []%s\n", endpointMiddlewareOption)
	//if !g.JSONRPCEnable {
	//	g.w.W("createReqFunc %s.CreateRequestFunc\n", kitHTTPPkg)
	//	g.w.W("decRespFunc %s.DecodeResponseFunc\n", kitHTTPPkg)
	//}
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
				errorsDub := map[int64]struct{}{}
				for _, e := range methodErrors {
					if _, ok := errorsDub[e.Code]; ok {
						continue
					}
					errorsDub[e.Code] = struct{}{}

					g.w.W("case %d:\n", e.Code)
					pkgName := importer.Import(e.PkgName, e.PkgPath)
					if pkgName != "" {
						pkgName += "."
					}

					g.w.W("err = &%s%s{}\n", pkgName, e.Name)
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

	return g.w.Bytes()
}

func (g *ClientHelpers) Package() string {
	return g.Pkg
}

func (g *ClientHelpers) OutputDir() string {
	return g.Output
}

func (g *ClientHelpers) Filename() string {
	return "client_helpers.go"
}
