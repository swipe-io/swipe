package generator

import (
	"context"
	"fmt"

	"github.com/swipe-io/swipe/v3/internal/plugin"

	"github.com/swipe-io/swipe/v3/internal/plugin/gokit/config"
	"github.com/swipe-io/swipe/v3/option"
	"github.com/swipe-io/swipe/v3/swipe"
	"github.com/swipe-io/swipe/v3/writer"
)

type ClientStruct struct {
	w             writer.GoWriter
	UseFast       bool
	JSONRPCEnable bool
	Interfaces    []*config.Interface
	Output        string
	Pkg           string
}

func (g *ClientStruct) Package() string {
	return g.Pkg
}

func (g *ClientStruct) Generate(ctx context.Context) []byte {
	importer := ctx.Value(swipe.ImporterKey).(swipe.Importer)

	endpointPkg := importer.Import("endpoint", "github.com/go-kit/kit/endpoint")

	if len(g.Interfaces) > 1 {
		g.w.W("type AppClient struct {\n")

		for _, iface := range g.Interfaces {
			g.w.W("%s *%s\n", UcNameWithAppPrefix(iface), ClientType(iface))
		}
		g.w.W("}\n\n")

		if g.JSONRPCEnable {
			g.w.W("func NewClientJSONRPC(tgt string")
		} else {
			g.w.W("func NewClientREST(tgt string")
		}

		g.w.W(" ,opts ...ClientOption")
		g.w.W(") (*AppClient, error) {\n")

		for _, iface := range g.Interfaces {
			name := UcNameWithAppPrefix(iface)
			lcName := LcNameWithAppPrefix(iface)

			if g.JSONRPCEnable {
				g.w.W("%s, err := NewClientJSONRPC%s(tgt, opts...)\n", lcName, name)
			} else {
				g.w.W("%s, err := NewClientREST%s(tgt, opts...)\n", lcName, name)
			}
			g.w.WriteCheckErr("err", func() {
				g.w.W("return nil, err")
			})
		}

		g.w.W("return &AppClient{\n")
		for _, iface := range g.Interfaces {
			g.w.W("%[1]s: %[2]s,\n", UcNameWithAppPrefix(iface), LcNameWithAppPrefix(iface))
		}
		g.w.W("}, nil\n")
		g.w.W("}\n\n")
	}

	if len(g.Interfaces) > 0 {
		contextPkg := importer.Import("context", "context")

		for _, iface := range g.Interfaces {
			ifaceType := iface.Named.Type.(*option.IfaceType)

			clientType := ClientType(iface)
			g.w.W("type %s struct {\n", clientType)
			for _, m := range ifaceType.Methods {
				g.w.W("%sEndpoint %s.Endpoint\n", LcNameWithAppPrefix(iface)+m.Name.Value, endpointPkg)
			}
			g.w.W("}\n\n")

			for _, m := range ifaceType.Methods {
				var (
					ctxVarName      = fmt.Sprintf("%s.TODO()", contextPkg)
					errVarName      = "err"
					assignResult    = ":"
					responseVarName = "response"
				)

				ctxVar := findContextVar(m.Sig.Params)
				errVar := findErrorVar(m.Sig.Results)

				if ctxVar != nil {
					ctxVarName = ctxVar.Name.Value
				}
				if errVar != nil {
					errVarName = errVar.Name.Value
					assignResult = ""
				}

				if plugin.LenWithoutErrors(m.Sig.Results) == 0 {
					responseVarName = "_"
				}

				g.w.W("func (c *%s) %s %s {\n", clientType, m.Name.Value, swipe.TypeString(m.Sig, false, importer))
				if responseVarName != "_" {
					g.w.W("var %s interface{}\n", responseVarName)
				}
				g.w.W("%s, %s %s= c.%sEndpoint(%s,", responseVarName, errVarName, assignResult, LcNameWithAppPrefix(iface)+m.Name.Value, ctxVarName)

				if len(m.Sig.Params) > 0 {
					g.w.W("%s{", NameRequest(m, iface))
					for _, param := range m.Sig.Params {
						if plugin.IsContext(param) {
							continue
						}
						g.w.W("%s: %s,", param.Name.Upper(), param.Name.Value)
					}
					g.w.W("}")
				} else {
					g.w.W("nil")
				}

				g.w.W(")\n")

				g.w.WriteCheckErr(errVarName, func() {
					g.w.W("return\n")
				})

				lenResults := plugin.LenWithoutErrors(m.Sig.Results)
				if lenResults > 0 {
					for _, result := range m.Sig.Results {
						if plugin.IsError(result) {
							continue
						}
						if lenResults == 1 {
							g.w.W("%s = %s.(%s)\n", result.Name.Value, responseVarName, swipe.TypeString(result.Type, false, importer))
						} else {
							g.w.W("%s = %s.(%s).%s\n", result.Name.Value, responseVarName, NameResponse(m, iface), result.Name.Upper())
						}
					}
				}
				g.w.W("return\n")
				g.w.W("}\n")
			}
		}
	}

	return g.w.Bytes()
}

func (g *ClientStruct) OutputPath() string {
	return g.Output
}

func (g *ClientStruct) Filename() string {
	return "client_struct.go"
}
