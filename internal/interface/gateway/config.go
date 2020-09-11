package gateway

import (
	"go/ast"
	stdtypes "go/types"

	"github.com/swipe-io/swipe/v2/internal/option"
	"github.com/swipe-io/swipe/v2/internal/usecase/gateway"
)

type configGateway struct {
	stType       stdtypes.Type
	stExpr       ast.Expr
	st           *stdtypes.Struct
	funcName     string
	docEnable    bool
	docOutputDir string
}

func (g *configGateway) DocOutputDir() string {
	return g.docOutputDir
}

func (g *configGateway) DocEnable() bool {
	return g.docEnable
}

func (g *configGateway) Struct() *stdtypes.Struct {
	return g.st
}

func (g *configGateway) StructType() stdtypes.Type {
	return g.stType
}

func (g *configGateway) StructExpr() ast.Expr {
	return g.stExpr
}

func (g *configGateway) FuncName() string {
	return g.funcName
}

func (g *configGateway) load(o *option.Option) {
	structOpt := option.MustOption(o.At("optionsStruct"))

	g.stExpr = structOpt.Value.Expr()
	g.stType = structOpt.Value.Type()

	if ptr, ok := structOpt.Value.Type().(*stdtypes.Pointer); ok {
		g.st = ptr.Elem().Underlying().(*stdtypes.Struct)
	} else {
		g.st = structOpt.Value.Type().(*stdtypes.Struct)
	}
	g.funcName = "LoadConfig"
	if funcNameOpt, ok := o.At("FuncName"); ok {
		g.funcName = funcNameOpt.Value.String()
	}
	if markdownDocOpt, ok := o.At("ConfigMarkdownDoc"); ok {
		g.docEnable = true
		g.docOutputDir = markdownDocOpt.Value.String()
	}
}

func NewConfigGateway(o *option.Option) gateway.ConfigGateway {
	g := &configGateway{}
	g.load(o)
	return g
}
