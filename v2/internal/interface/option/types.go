package option

import (
	"github.com/fatih/structtag"
)

type NameType struct {
	Exported    string
	NotExported string
	Var         string
}

type VarType struct {
	Name     NameType
	Embedded bool
	Exported bool
	IsField  bool
	Type     interface{}
	Comment  string
}

type StructType struct {
	Name   NameType
	Pkg    *PackageType
	Fields []*StructFieldType
}

type StructFieldType struct {
	Var  *VarType
	Tags *structtag.Tags
}

type SignType struct {
	Ctx         *VarType
	Err         *VarType
	Recv        *VarType
	Variadic    *VarType
	Params      []*VarType
	Results     []*VarType
	ResultNamed bool
}

type NamedType struct {
	Name    NameType
	Methods []*FuncType
}

type FuncType struct {
	FullName string
	Name     NameType
	Exported bool
	Sig      *SignType
	Comment  string
}

type IfaceType struct {
	Name            NameType
	Methods         []*FuncType
	Embeddeds       []interface{}
	ExplicitMethods []*FuncType
	Pkg             *PackageType
}

type ModuleType struct {
	Version string
	Path    string
}

type PackageType struct {
	Name     string
	Path     string
	Module   *ModuleType
	External bool
}

type BasicType struct {
	Name string
	Zero string
}
