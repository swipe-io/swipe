package model

import (
	"go/ast"
	"go/types"
)

type ConfigDocOption struct {
	Enable    bool
	OutputDir string
}

type ConfigOption struct {
	FuncName   string
	Struct     *types.Struct
	StructType types.Type
	StructExpr ast.Expr
	Doc        ConfigDocOption
}

type Env struct {
	Name string
}
