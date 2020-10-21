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
	g.stExpr = o.Value.Expr()
	g.stType = o.Value.Type()

	if ptr, ok := o.Value.Type().(*stdtypes.Pointer); ok {
		g.st = ptr.Elem().Underlying().(*stdtypes.Struct)
	} else {
		g.st = o.Value.Type().(*stdtypes.Struct)
	}
	if _, ok := o.At("ConfigEnvDocEnable"); ok {
		g.docEnable = true
	}
	g.funcName = "LoadConfig"
	if opt, ok := o.At("ConfigEnvFuncName"); ok {
		g.funcName = opt.Value.String()
	}
	if opt, ok := o.At("ConfigEnvDocOutput"); ok {
		g.docOutputDir = opt.Value.String()
	}
}

func NewConfigGateway(o *option.Option) gateway.ConfigGateway {
	g := &configGateway{}
	g.load(o)
	return g
}
