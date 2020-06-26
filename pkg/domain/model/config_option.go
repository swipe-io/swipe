package model

import (
	"go/ast"
	"go/types"
)

type ConfigOption struct {
	FuncName   string
	Struct     *types.Struct
	StructType types.Type
	StructExpr ast.Expr
}
