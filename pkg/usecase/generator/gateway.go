package generator

import (
	"context"

	"github.com/swipe-io/swipe/pkg/domain/model"
	"github.com/swipe-io/swipe/pkg/importer"
	"github.com/swipe-io/swipe/pkg/strings"
	"github.com/swipe-io/swipe/pkg/writer"
)

type gatewayGenerator struct {
	writer.GoLangWriter
	filename string
	info     model.GenerateInfo
	o        model.GatewayOption
	i        *importer.Importer
}

func (g *gatewayGenerator) Prepare(ctx context.Context) error {
	return nil
}

func (g *gatewayGenerator) Process(ctx context.Context) error {
	ioPkg := g.i.Import("io", "io")
	epPkg := g.i.Import("endpoint", "github.com/go-kit/kit/endpoint")
	httpkitPkg := g.i.Import("endpoint", "github.com/go-kit/kit/transport/http")
	jsonrpckitPkg := g.i.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/http/jsonrpc")
	logPkg := g.i.Import("endpoint", "github.com/go-kit/kit/log")
	sdPkg := g.i.Import("sd", "github.com/go-kit/kit/sd")
	lbPkg := g.i.Import("sd", "github.com/go-kit/kit/sd/lb")
	timePkg := g.i.Import("time", "time")

	g.W("const (\nDefaultRetryMax = 99\nDefaultRetryTimeout = time.Second * 600\n)\n\n")

	g.W("type BalancerFactory func(s %s.Endpointer) %s.Balancer\n\n", sdPkg, lbPkg)

	g.W("func RetryErrorExtractor() endpoint.Middleware {\n")
	g.W("return func(next endpoint.Endpoint) endpoint.Endpoint {\n")
	g.W("return func(ctx context.Context, request interface{}) (response interface{}, err error) {\n")
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
	g.W("if _, ok := received.(%s.StatusCoder); ok {\n", httpkitPkg)
	g.W("keepTrying = false\n")
	g.W("} else if _, ok := received.(%s.ErrorCoder); ok {\n", jsonrpckitPkg)
	g.W("keepTrying = false\n")
	g.W("}\n")

	g.W("return\n")
	g.W("}\n")
	g.W("}\n\n")

	g.W("type EndpointSet struct {\n")
	for _, s := range g.o.Services {
		g.W("%s struct {\n", s.ID)
		for i := 0; i < s.Iface.NumMethods(); i++ {
			m := s.Iface.Method(i)
			g.W("%sEndpoint %s.Endpoint\n", m.Name(), epPkg)
		}
		g.W("}\n")
	}
	g.W("}\n\n")

	for _, s := range g.o.Services {
		g.W("type %sEndpointFactory interface {\n", s.ID)
		for i := 0; i < s.Iface.NumMethods(); i++ {
			m := s.Iface.Method(i)
			g.W("%sEndpointFactory(instance string) (%s.Endpoint, %s.Closer, error)\n", m.Name(), epPkg, ioPkg)
		}
		g.W("}\n\n")

		g.W("type %sOption struct {\n", s.ID)
		g.W("Instancer %s.Instancer \n", sdPkg)
		g.W("EndpointFactory %sEndpointFactory\n", s.ID)

		for i := 0; i < s.Iface.NumMethods(); i++ {
			m := s.Iface.Method(i)
			g.W("%s EndpointOption\n", m.Name())
		}
		g.W("}\n\n")
	}

	g.W("func NewGateway(")
	for i, s := range g.o.Services {
		if i > 0 {
			g.W(",")
		}
		g.W("%s %sOption", strings.LcFirst(s.ID), s.ID)
	}
	g.W(", logger %s.Logger) (ep EndpointSet) {\n", logPkg)

	g.W("{\n")
	for _, s := range g.o.Services {
		for i := 0; i < s.Iface.NumMethods(); i++ {
			m := s.Iface.Method(i)
			optName := strings.LcFirst(s.ID)
			g.W("{\n")

			g.W("if %s.%s.Balancer == nil {\n", optName, m.Name())
			g.W("%s.%s.Balancer = %s.NewRoundRobin\n", optName, m.Name(), lbPkg)
			g.W("}\n")

			g.W("if %s.%s.RetryMax == 0 {\n", optName, m.Name())
			g.W("%s.%s.RetryMax = DefaultRetryMax\n", optName, m.Name())
			g.W("}\n")

			g.W("if %s.%s.RetryTimeout == 0 {\n", optName, m.Name())
			g.W("%s.%s.RetryTimeout = DefaultRetryTimeout\n", optName, m.Name())
			g.W("}\n")

			g.W("endpointer := %[1]s.NewEndpointer(%[2]s.Instancer, %[2]s.EndpointFactory.%[3]sEndpointFactory, logger)\n", sdPkg, optName, m.Name())
			g.W(
				"ep.%[4]s.%[3]sEndpoint = %[1]s.RetryWithCallback(%[2]s.%[3]s.RetryTimeout, %[2]s.%[3]s.Balancer(endpointer), retryMax(%[2]s.%[3]s.RetryMax))\n",
				lbPkg, optName, m.Name(), s.ID,
			)
			g.W("ep.%[2]s.%[1]sEndpoint = RetryErrorExtractor()(ep.%[2]s.%[1]sEndpoint)\n", m.Name(), s.ID)
			g.W("}\n")
		}
	}
	g.W("}\n")
	g.W("return\n")
	g.W("}\n")
	return nil
}

func (g *gatewayGenerator) PkgName() string {
	return ""
}

func (g *gatewayGenerator) OutputDir() string {
	return ""
}

func (g *gatewayGenerator) Filename() string {
	return g.filename
}

func (g *gatewayGenerator) SetImporter(i *importer.Importer) {
	g.i = i
}

func NewGatewayGenerator(filename string, info model.GenerateInfo, o model.GatewayOption) Generator {
	return &gatewayGenerator{filename: filename, info: info, o: o}
}
