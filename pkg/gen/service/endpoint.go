package service

import (
	stdtypes "go/types"
	stdstrings "strings"

	"github.com/iancoleman/strcase"

	"github.com/swipe-io/swipe/pkg/strings"
	"github.com/swipe-io/swipe/pkg/utils"
	"github.com/swipe-io/swipe/pkg/writer"
)

type Endpoint struct {
	ctx serviceCtx
	w   *writer.Writer
}

func (w *Endpoint) Write() error {
	kitEndpointPkg := w.w.Import("endpoint", "github.com/go-kit/kit/endpoint")
	contextPkg := w.w.Import("context", "context")

	for _, m := range w.ctx.iface.methods {
		resultLen := len(m.results)

		if len(m.params) > 0 {
			w.w.Write("type %sRequest%s struct {\n", m.lcName, w.ctx.id)
			for _, p := range m.params {
				w.w.Write("%s %s `json:\"%s\"`\n", strings.UcFirst(p.Name()), w.w.TypeString(p.Type()), strcase.ToLowerCamel(p.Name()))
			}
			w.w.Write("}\n")
		}

		if m.resultsNamed {
			w.w.Write("type %sResponse%s struct {\n", m.lcName, w.ctx.id)
			for _, p := range m.results {
				name := p.Name()
				w.w.Write("%s %s `json:\"%s\"`\n", strings.UcFirst(name), w.w.TypeString(p.Type()), strcase.ToLowerCamel(name))
			}
			w.w.Write("}\n")
		}

		w.w.Write("func make%sEndpoint(s %s", m.name, w.ctx.typeStr)
		w.w.Write(") %s.Endpoint {\n", kitEndpointPkg)
		w.w.Write("w := func(ctx %s.Context, request interface{}) (interface{}, error) {\n", contextPkg)

		var callParams []string

		if m.paramCtx != nil {
			callParams = append(callParams, "ctx")
		}

		callParams = append(callParams, utils.Params(m.params, func(p *stdtypes.Var) []string {
			name := p.Name()
			name = stdstrings.ToUpper(name[:1]) + name[1:]
			return []string{"req." + name}
		}, nil)...)

		if len(m.params) > 0 {
			w.w.Write("req := request.(%sRequest%s)\n", m.lcName, w.ctx.id)
		}

		if len(m.results) > 0 {
			if m.resultsNamed {
				for i, p := range m.results {
					if i > 0 {
						w.w.Write(", ")
					}
					w.w.Write(p.Name())
				}
			} else {
				w.w.Write("result")
			}
			w.w.Write(", ")
		}

		if m.returnErr != nil {
			w.w.Write("err")
		}

		w.w.Write(" := ")

		w.w.WriteFuncCall("s", m.name, callParams)

		if m.returnErr != nil {
			w.w.WriteCheckErr(func() {
				w.w.Write("return nil, err\n")
			})
		}

		w.w.Write("return ")

		if resultLen > 0 {
			if resultLen > 0 && m.resultsNamed {
				w.w.Write("%sResponse%s", m.lcName, w.ctx.id)
				w.w.WriteStructAssign(structKeyValue(m.results, nil))
			} else {
				w.w.Write("result")
			}
		} else {
			w.w.Write("nil")
		}

		w.w.Write(" ,nil\n")

		w.w.Write("}\n")
		w.w.Write("return w\n")
		w.w.Write("}\n\n")
	}

	return nil
}

func newEndpoint(ctx serviceCtx, w *writer.Writer) *Endpoint {
	return &Endpoint{ctx: ctx, w: w}
}
