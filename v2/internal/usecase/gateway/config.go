package gateway

import (
	"go/ast"
	"go/types"
)

type ConfigGateway interface {
	DocEnable() bool
	DocOutputDir() string
	Struct() *types.Struct
	StructType() types.Type
	StructExpr() ast.Expr
	FuncName() string
}
