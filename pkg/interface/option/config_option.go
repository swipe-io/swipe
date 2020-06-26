package option

import (
	"go/types"

	"github.com/swipe-io/swipe/pkg/domain/model"
	"github.com/swipe-io/swipe/pkg/parser"
	"github.com/swipe-io/swipe/pkg/usecase/option"
)

type configOption struct {
}

func (g *configOption) Parse(option *parser.Option) (interface{}, error) {
	o := model.ConfigOption{}

	structOpt := parser.MustOption(option.At("optionsStruct"))

	o.StructExpr = structOpt.Value.Expr()
	o.StructType = structOpt.Value.Type()

	if ptr, ok := structOpt.Value.Type().(*types.Pointer); ok {
		o.Struct = ptr.Elem().Underlying().(*types.Struct)
	} else {
		o.Struct = structOpt.Value.Type().(*types.Struct)
	}
	o.FuncName = "LoadConfig"
	if funcNameOpt, ok := option.At("FuncName"); ok {
		o.FuncName = funcNameOpt.Value.String()
	}
	return o, nil
}

func NewConfigOption() option.Option {
	return &configOption{}
}
