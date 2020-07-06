package model

import (
	"go/ast"
	"go/constant"
	stdtypes "go/types"
)

type VarSlice []*stdtypes.Var

func (s VarSlice) LookupField(name string) *stdtypes.Var {
	for _, p := range s {
		if p.Name() == name {
			return p
		}
	}
	return nil
}

type ValueResult struct {
	ID    uint32
	Type  stdtypes.Type
	Value constant.Value
}

type CallResult struct {
	ID      uint32
	FnID    uint32
	IsIface bool
}

type ReturnStmt struct {
	Results []interface{}
}

type BlockStmt struct {
	Returns []*ReturnStmt
	Blocks  []*BlockStmt
}

type DeclStmt struct {
	Name  string
	Block *BlockStmt
}

type DeclType struct {
	Type     stdtypes.Type
	DeclStmt map[uint32]*DeclStmt
}

type ReturnType struct {
	Expr ast.Expr
	Type stdtypes.Type
}

type InstrumentingServiceOption struct {
	Enable    bool
	Namespace string
	Subsystem string
}

type ServiceMethod struct {
	Type         *stdtypes.Func
	Name         string
	LcName       string
	Params       VarSlice
	Results      VarSlice
	Comments     []string
	ParamCtx     *stdtypes.Var
	ReturnErr    *stdtypes.Var
	ResultsNamed bool
	Errors       []ErrorHTTPTransportOption
	T            stdtypes.Type
}

type ServiceOption struct {
	Transport     TransportOption
	Instrumenting InstrumentingServiceOption
	Logging       bool
	Methods       map[string]ServiceMethod
	Type          stdtypes.Type
	Interface     *stdtypes.Interface
	ID            string
}
