package generator

import (
	"context"

	"github.com/swipe-io/swipe/v2/internal/domain/model"
	"github.com/swipe-io/swipe/v2/internal/importer"
	"github.com/swipe-io/swipe/v2/internal/strings"
	"github.com/swipe-io/swipe/v2/internal/usecase/generator"
	"github.com/swipe-io/swipe/v2/internal/writer"
)

type httpGatewayGenerator struct {
	writer.GoLangWriter
	i        *importer.Importer
	services []model.GatewayServiceOption
}

func (g *httpGatewayGenerator) Prepare(ctx context.Context) error {
	return nil
}

func (g *httpGatewayGenerator) Process(ctx context.Context) error {
	ioPkg := g.i.Import("io", "io")
	contextPkg := g.i.Import("context", "context")
	epPkg := g.i.Import("endpoint", "github.com/go-kit/kit/endpoint")
	httpKitPkg := g.i.Import("endpoint", "github.com/go-kit/kit/transport/http")
	jsonRPCKitPkg := g.i.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/http/jsonrpc")
	logPkg := g.i.Import("endpoint", "github.com/go-kit/kit/log")
	sdPkg := g.i.Import("sd", "github.com/go-kit/kit/sd")
	lbPkg := g.i.Import("sd", "github.com/go-kit/kit/sd/lb")
	timePkg := g.i.Import("time", "time")

	g.W("const (\nDefaultRetryMax = 99\nDefaultRetryTimeout = time.Second * 600\n)\n\n")

	g.W("type BalancerFactory func(s %s.Endpointer) %s.Balancer\n\n", sdPkg, lbPkg)

	g.W("func RetryErrorExtractor() %s.Middleware {\n", epPkg)
	g.W("return func(next %[1]s.Endpoint) %[1]s.Endpoint {\n", epPkg)
	g.W("return func(ctx %s.Context, request interface{}) (response interface{}, err error) {\n", contextPkg)
	g.W("response, err = next(ctx, request)\n")
	g.W("if err != nil {\n")
	g.W("if re, ok := err.(sd2.RetryError); ok {\n")
	g.W("return nil, re.Final\n")
	g.W("}\n}\n")
	g.W("return\n")
	g.W("}\n}\n}\n")

	g.W("type EndpointOption struct{\n")
	g.W("Balancer BalancerFactory\n")
	g.W("RetryMax int\n")
	g.W("RetryTimeout %s.Duration\n", timePkg)
	g.W("}\n")

	g.W("func retryMax(max int) %s.Callback {\n", lbPkg)
	g.W("return func(n int, received error) (keepTrying bool, replacement error) {\n")
	g.W("keepTrying = n < max\n")
	g.W("replacement = received\n")
	g.W("if _, ok := received.(%s.StatusCoder); ok {\n", httpKitPkg)
	g.W("keepTrying = false\n")
	g.W("} else if _, ok := received.(%s.ErrorCoder); ok {\n", jsonRPCKitPkg)
	g.W("keepTrying = false\n")
	g.W("}\n")

	g.W("return\n")
	g.W("}\n")
	g.W("}\n\n")

	g.W("type EndpointSet struct {\n")
	for _, s := range g.services {
		g.W("%s struct {\n", s.Iface.Name())
		for _, method := range s.Iface.Methods() {
			g.W("%sEndpoint %s.Endpoint\n", method.Name, epPkg)
		}
		g.W("}\n")
	}
	g.W("}\n\n")

	for _, s := range g.services {
		g.W("type %sEndpointFactory interface {\n", s.Iface.Name())
		for _, method := range s.Iface.Methods() {
			g.W("%sEndpointFactory(instance string) (%s.Endpoint, %s.Closer, error)\n", method.Name, epPkg, ioPkg)
		}
		g.W("}\n\n")

		g.W("type %sOption struct {\n", s.Iface.Name())
		g.W("Instancer %s.Instancer \n", sdPkg)
		g.W("EndpointFactory %sEndpointFactory\n", s.Iface.Name())

		for _, method := range s.Iface.Methods() {
			g.W("%s EndpointOption\n", method.Name)
		}
		g.W("}\n\n")
	}

	g.W("func NewGateway(")
	for i, s := range g.services {
		if i > 0 {
			g.W(",")
		}
		g.W("%s %sOption", strings.LcFirst(s.Iface.Name()), s.Iface.Name())
	}
	g.W(", logger %s.Logger) (ep EndpointSet) {\n", logPkg)

	g.W("{\n")
	for _, s := range g.services {
		for _, method := range s.Iface.Methods() {
			optName := strings.LcFirst(s.Iface.Name())
			g.W("{\n")

			g.W("if %s.%s.Balancer == nil {\n", optName, method.Name)
			g.W("%s.%s.Balancer = %s.NewRoundRobin\n", optName, method.Name, lbPkg)
			g.W("}\n")

			g.W("if %s.%s.RetryMax == 0 {\n", optName, method.Name)
			g.W("%s.%s.RetryMax = DefaultRetryMax\n", optName, method.Name)
			g.W("}\n")

			g.W("if %s.%s.RetryTimeout == 0 {\n", optName, method.Name)
			g.W("%s.%s.RetryTimeout = DefaultRetryTimeout\n", optName, method.Name)
			g.W("}\n")

			g.W("endpointer := %[1]s.NewEndpointer(%[2]s.Instancer, %[2]s.EndpointFactory.%[3]sEndpointFactory, logger)\n", sdPkg, optName, method.Name)
			g.W(
				"ep.%[4]s.%[3]sEndpoint = %[1]s.RetryWithCallback(%[2]s.%[3]s.RetryTimeout, %[2]s.%[3]s.Balancer(endpointer), retryMax(%[2]s.%[3]s.RetryMax))\n",
				lbPkg, optName, method.Name, s.Iface.Name(),
			)
			g.W("ep.%[2]s.%[1]sEndpoint = RetryErrorExtractor()(ep.%[2]s.%[1]sEndpoint)\n", method.Name, s.Iface.Name())
			g.W("}\n")
		}
	}
	g.W("}\n")
	g.W("return\n")
	g.W("}\n")
	return nil
}

func (g *httpGatewayGenerator) PkgName() string {
	return ""
}

func (g *httpGatewayGenerator) OutputDir() string {
	return ""
}

func (g *httpGatewayGenerator) Filename() string {
	return "gateway_gen.go"
}

func (g *httpGatewayGenerator) SetImporter(i *importer.Importer) {
	g.i = i
}

func NewGatewayGenerator(
	services []model.GatewayServiceOption,
) generator.Generator {
	return &httpGatewayGenerator{
		services: services,
	}
}
