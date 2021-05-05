package generator

import (
	"context"

	"github.com/swipe-io/swipe/v2/internal/plugin/gokit/config"

	"github.com/swipe-io/swipe/v2/internal/option"
	"github.com/swipe-io/swipe/v2/internal/swipe"
	"github.com/swipe-io/swipe/v2/internal/writer"
)

type Endpoint struct {
	w          writer.GoWriter
	Interfaces []*config.Interface
}

func (g *Endpoint) Generate(ctx context.Context) []byte {
	importer := ctx.Value(swipe.ImporterKey).(swipe.Importer)

	g.writeEndpointMake(importer)
	g.writeReqResp(importer)
	return g.w.Bytes()
}

func (g *Endpoint) OutputDir() string {
	return ""
}

func (g *Endpoint) Filename() string {
	return "endpoint.go"
}

func (g *Endpoint) writeReqResp(importer swipe.Importer) {
	for _, iface := range g.Interfaces {
		ifaceType := iface.Named.Type.(*option.IfaceType)

		if iface.Named.Pkg.Module.External {
			continue
		}

		typeStr := iface.Named.Name.LowerCase + "Interface"
		epSetName := iface.Named.Name.UpperCase + "EndpointSet"

		g.w.W("type %s struct {\n", epSetName)
		kitEndpointPkg := importer.Import("endpoint", "github.com/go-kit/kit/endpoint")
		for _, m := range ifaceType.Methods {
			g.w.W("%sEndpoint %s.Endpoint\n", m.Name, kitEndpointPkg)
		}
		g.w.W("}\n")

		g.w.W("func Make%[1]s(svc %[2]s) %[1]s {\n", epSetName, typeStr)
		g.w.W("return %s{\n", epSetName)
		for _, m := range ifaceType.Methods {
			g.w.W("%sEndpoint: %s(svc),\n", m.Name, NameMakeEndpoint(m, iface.Named))
		}
		g.w.W("}\n")
		g.w.W("}\n")

		for _, m := range ifaceType.Methods {
			if len(m.Sig.Params) > 0 {
				g.w.W("type %s struct {\n", NameRequest(m, iface.Named))
				for _, param := range m.Sig.Params {
					if IsContext(param) {
						continue
					}
					g.w.W("%s %s `json:\"%s\"`\n", param.Name.UpperCase, importer.TypeString(param.Type), param.Name.LowerCase)
				}
				g.w.W("}\n")
			}
			if LenWithoutErrors(m.Sig.Results) > 1 {
				g.w.W("type %s struct {\n", NameResponse(m, iface.Named))
				for _, param := range m.Sig.Results {
					if IsError(param) {
						continue
					}
					g.w.W("%s %s `json:\"%s\"`\n", param.Name.UpperCase, importer.TypeString(param.Type), param.Name.LowerCase)
				}
				g.w.W("}\n")
			}
		}
	}
}

func (g *Endpoint) writeEndpointMake(importer swipe.Importer) {
	for _, iface := range g.Interfaces {
		ifaceType := iface.Named.Type.(*option.IfaceType)

		if iface.Named.Pkg.Module.External {
			continue
		}

		contextPkg := importer.Import("context", "context")
		kitEndpointPkg := importer.Import("endpoint", "github.com/go-kit/kit/endpoint")
		typeStr := iface.Named.Name.LowerCase + "Interface"

		for _, m := range ifaceType.Methods {
			g.w.W("func %s(s %s) %s.Endpoint {\n", NameMakeEndpoint(m, iface.Named), typeStr, kitEndpointPkg)
			g.w.W("return func (ctx %s.Context, request interface{}) (interface{}, error) {\n", contextPkg)

			var callParams []string
			for _, param := range m.Sig.Params {
				if IsContext(param) {
					callParams = append(callParams, "ctx")
					continue
				}
				if param.IsVariadic {
					callParams = append(callParams, "req."+param.Name.UpperCase+"...")
					continue
				}
				callParams = append(callParams, "req."+param.Name.UpperCase)
			}
			if len(m.Sig.Params) > 0 {
				g.w.W("req := request.(%s)\n", NameRequest(m, iface.Named))
			}

			if len(m.Sig.Results) > 0 {
				for i, p := range m.Sig.Results {
					if i > 0 {
						g.w.W(", ")
					}
					g.w.W(p.Name.Origin)
				}
			}

			if len(m.Sig.Results) > 0 {
				g.w.W(" := ")
			}

			g.w.WriteFuncCall("s", m.Name.Origin, callParams)

			if len(m.Sig.Results) > 0 {
				for _, result := range m.Sig.Results {
					if IsError(result) {
						g.w.WriteCheckErr(result.Name.Origin, func() {
							g.w.W("return nil, %s\n", result.Name.Origin)
						})
					}
				}
			}
			g.w.W("return ")

			resultLen := LenWithoutErrors(m.Sig.Results)
			if resultLen > 1 {
				g.w.W("%s", NameResponse(m, iface.Named))
				var resultKeyVal []string
				for _, result := range m.Sig.Results {
					if IsError(result) {
						continue
					}
					resultKeyVal = append(resultKeyVal, result.Name.UpperCase, result.Name.Origin)
				}
				g.w.WriteStructAssign(resultKeyVal)
			} else if resultLen == 1 {
				for _, result := range m.Sig.Results {
					if IsError(result) {
						continue
					}
					g.w.W("%s", result.Name.Origin)
				}
			} else {
				g.w.W("nil")
			}
			g.w.W(" ,nil\n")
			g.w.W("}\n")
			g.w.W("}\n\n")
		}
	}
}
