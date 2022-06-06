package generator

import (
	"context"

	"github.com/swipe-io/swipe/v3/swipe"
	"github.com/swipe-io/swipe/v3/writer"
)

type MiddlewareChain struct {
	w      writer.GoWriter
	Output string
	Pkg    string
}

func (g *MiddlewareChain) Generate(ctx context.Context) []byte {
	importer := ctx.Value(swipe.ImporterKey).(swipe.Importer)

	endpointPkg := importer.Import("endpoint", "github.com/go-kit/kit/endpoint")

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

	return g.w.Bytes()
}

func (g *MiddlewareChain) Package() string {
	return g.Pkg
}

func (g *MiddlewareChain) OutputDir() string {
	return g.Output
}

func (g *MiddlewareChain) Filename() string {
	return "middleware_chain.go"
}
