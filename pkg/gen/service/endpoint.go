package service

import (
	"fmt"
	stdtypes "go/types"
	"strconv"
	stdstrings "strings"

	"github.com/iancoleman/strcase"

	"github.com/swipe-io/swipe/pkg/strings"
	"github.com/swipe-io/swipe/pkg/types"
	"github.com/swipe-io/swipe/pkg/utils"
	"github.com/swipe-io/swipe/pkg/writer"
)

type Endpoint struct {
	ctx serviceCtx
	w   *writer.Writer
}

func (w *Endpoint) Write() error {
	for i := 0; i < w.ctx.iface.NumMethods(); i++ {
		m := w.ctx.iface.Method(i)
		sig := m.Type().(*stdtypes.Signature)

		ntParams := utils.Params(sig.Params(), func(p *stdtypes.Var) []string {
			name := p.Name()
			fieldName := stdstrings.ToUpper(name[:1]) + name[1:]
			tagName := strcase.ToLowerCamel(p.Name())
			typeName := w.w.TypeString(p.Type())
			return []string{fieldName, typeName + "`json:" + strconv.Quote(tagName) + "`"}
		}, func(p *stdtypes.Var) bool {
			return !types.IsContext(p.Type())
		})

		ntResults := utils.Params(sig.Results(), func(p *stdtypes.Var) []string {
			name := p.Name()
			tagName := strcase.ToLowerCamel(name)
			typeName := w.w.TypeString(p.Type())
			fieldName := stdstrings.ToUpper(name[:1]) + name[1:]

			return []string{fieldName, typeName + "`json:" + strconv.Quote(tagName) + "`"}
		}, func(p *stdtypes.Var) bool {
			return !types.IsError(p.Type())
		})

		if len(ntParams) > 0 {
			requestName := fmt.Sprintf("%sRequest%s", strings.LcFirst(m.Name()), w.ctx.id)
			w.w.WriteTypeStruct(requestName, ntParams)
		}

		if len(ntResults) > 0 {
			responseName := fmt.Sprintf("%sResponse%s", strings.LcFirst(m.Name()), w.ctx.id)
			w.w.WriteTypeStruct(responseName, ntResults)
		}
		if err := w.writeEndpoint(m, sig); err != nil {
			return err
		}
	}

	return nil
}

func (w *Endpoint) writeEndpoint(fn *stdtypes.Func, sig *stdtypes.Signature) error {
	kitEndpointPkg := w.w.Import("endpoint", "github.com/go-kit/kit/endpoint")
	contextPkg := w.w.Import("context", "context")

	lcName := strings.LcFirst(fn.Name())

	w.w.Write("func make%sEndpoint(s %s", fn.Name(), w.ctx.typeStr)
	w.w.Write(") %s.Endpoint {\n", kitEndpointPkg)
	w.w.Write("w := func(ctx %s.Context, request interface{}) (interface{}, error) {\n", contextPkg)

	callParams := utils.Params(sig.Params(), func(p *stdtypes.Var) []string {
		if types.IsContext(p.Type()) {
			return []string{"ctx"}
		}
		name := p.Name()
		name = stdstrings.ToUpper(name[:1]) + name[1:]
		return []string{"req." + name}
	}, nil)

	if types.LenWithoutContext(sig.Params()) > 0 {
		w.w.Write("req := request.(%sRequest%s)\n", lcName, w.ctx.id)
	}

	namesResults := utils.NameParams(sig.Results(), nil)

	if len(namesResults) > 0 {
		w.w.Write(stdstrings.Join(namesResults, ","))
		w.w.Write(" := ")
	}

	w.w.WriteFuncCall("s", fn.Name(), callParams)

	if types.ContainsError(sig.Results()) {
		w.w.WriteCheckErr(func() {
			w.w.Write("return nil, err")
		})
	}

	w.w.Write("return ")

	if types.LenWithoutErr(sig.Results()) > 0 {
		w.w.Write("%sResponse%s", lcName, w.ctx.id)
		w.w.WriteStructAssign(structKeyValue(sig.Results(), func(p *stdtypes.Var) bool {
			return !types.IsError(p.Type())
		}))
	} else {
		w.w.Write("nil")
	}

	w.w.Write(" ,nil\n")

	w.w.Write("}\n")
	w.w.Write("return w\n")
	w.w.Write("}\n\n")

	return nil
}

func newEndpoint(ctx serviceCtx, w *writer.Writer) *Endpoint {
	return &Endpoint{ctx: ctx, w: w}
}
