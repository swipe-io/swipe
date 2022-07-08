package generator

import (
	"context"

	"github.com/swipe-io/swipe/v3/option"

	"github.com/swipe-io/swipe/v3/internal/plugin/gokit/config"
	"github.com/swipe-io/swipe/v3/swipe"
	"github.com/swipe-io/swipe/v3/writer"
)

type GatewayGenerator struct {
	w          writer.GoWriter
	Interfaces []*config.Interface
}

func (g *GatewayGenerator) Generate(ctx context.Context) []byte {
	importer := ctx.Value(swipe.ImporterKey).(swipe.Importer)

	contextPkg := importer.Import("context", "context")
	epPkg := importer.Import("endpoint", "github.com/go-kit/kit/endpoint")
	httpKitPkg := importer.Import("endpoint", "github.com/go-kit/kit/transport/http")
	jsonRPCKitPkg := importer.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/http/jsonrpc")
	sdPkg := importer.Import("sd", "github.com/go-kit/kit/sd")
	lbPkg := importer.Import("sd", "github.com/go-kit/kit/sd/lb")
	timePkg := importer.Import("time", "time")

	g.w.W("const (\nDefaultRetryMax = 99\nDefaultRetryTimeout = time.Second * 600\n)\n\n")

	g.w.W("type BalancerFactory func(iface %s.Endpointer) %s.Balancer\n\n", sdPkg, lbPkg)

	g.w.W("func RetryErrorExtractor() %s.Middleware {\n", epPkg)
	g.w.W("return func(next %[1]s.Endpoint) %[1]s.Endpoint {\n", epPkg)
	g.w.W("return func(ctx %s.Context, request interface{}) (response interface{}, err error) {\n", contextPkg)
	g.w.W("response, err = next(ctx, request)\n")
	g.w.W("if err != nil {\n")
	g.w.W("if re, ok := err.(sd2.RetryError); ok {\n")
	g.w.W("return nil, re.Final\n")
	g.w.W("}\n}\n")
	g.w.W("return\n")
	g.w.W("}\n}\n}\n")

	g.w.W("type EndpointOption struct{\n")
	g.w.W("Balancer BalancerFactory\n")
	g.w.W("RetryMax int\n")
	g.w.W("RetryTimeout %s.Duration\n", timePkg)
	g.w.W("}\n")

	g.w.W("func retryMax(max int) %s.Callback {\n", lbPkg)
	g.w.W("return func(n int, received error) (keepTrying bool, replacement error) {\n")
	g.w.W("keepTrying = n < max\n")
	g.w.W("replacement = received\n")
	g.w.W("if _, ok := received.(%s.StatusCoder); ok {\n", httpKitPkg)
	g.w.W("keepTrying = false\n")
	g.w.W("} else if _, ok := received.(%s.ErrorCoder); ok {\n", jsonRPCKitPkg)
	g.w.W("keepTrying = false\n")
	g.w.W("}\n")

	g.w.W("return\n")
	g.w.W("}\n")
	g.w.W("}\n\n")

	for _, iface := range g.Interfaces {
		if iface.Gateway == nil {
			continue
		}
		ifaceType := iface.Named.Type.(*option.IfaceType)

		g.w.W("type %sOption struct {\n", UcNameWithAppPrefix(iface, true))
		g.w.W("Instancer %s.Instancer \n", sdPkg)
		g.w.W("Factory func(string) (%s, error) \n", NameInterface(iface))

		for _, method := range ifaceType.Methods {
			g.w.W("%s EndpointOption\n", method.Name)
		}
		g.w.W("}\n\n")
	}
	return g.w.Bytes()
}

func (g *GatewayGenerator) OutputPath() string {
	return ""
}

func (g *GatewayGenerator) Filename() string {
	return "gateway.go"
}
