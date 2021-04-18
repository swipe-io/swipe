package option

import (
	"go/types"
	stdtypes "go/types"

	"github.com/fatih/structtag"
)

type NameType struct {
	Origin    string
	UpperCase string
	LowerCase string
	Var       string
}

type VarType struct {
	Name     NameType
	Embedded bool
	Exported bool
	IsField  bool
	Type     interface{}
	Comment  string
	Zero     string
}

type StructType struct {
	Name      NameType
	Pkg       *PackageType
	Fields    []*StructFieldType
	IsPointer bool
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
	Version  string
	Path     string
	External bool
}

type PackageType struct {
	Name   string
	Path   string
	Module *ModuleType
}

type BasicType struct {
	Name      string
	IsPointer bool
	kind      stdtypes.BasicKind
}

func (t BasicType) BitSize() string {
	switch t.kind {
	case types.Int8, types.Uint8:
		return "8"
	case types.Int16, types.Uint16:
		return "16"
	case types.Int32, types.Float32, types.Uint32:
		return "32"
	default: // for types.Int, types.Uint, types.Float64, types.Uint64, types.Int64 and other.
		return "64"
	}
}

func (t BasicType) IsString() bool {
	return t.kind == stdtypes.String
}

func (t BasicType) IsNumeric() bool {
	switch t.kind {
	default:
		return false
	case stdtypes.Uint,
		stdtypes.Uint8,
		stdtypes.Uint16,
		stdtypes.Uint32,
		stdtypes.Uint64,
		stdtypes.Int,
		stdtypes.Int8,
		stdtypes.Int16,
		stdtypes.Int32,
		stdtypes.Int64,
		stdtypes.Float32,
		stdtypes.Float64:
		return true
	}
}

func (t BasicType) IsAnyInt() bool {
	switch t.kind {
	case stdtypes.Int, stdtypes.Int8, stdtypes.Int16, stdtypes.Int32, stdtypes.Int64:
		return true
	}
	return false
}

func (t BasicType) IsInt() bool {
	return t.kind == stdtypes.Int
}

func (t BasicType) IsInt8() bool {
	return t.kind == stdtypes.Int8
}

func (t BasicType) IsInt16() bool {
	return t.kind == stdtypes.Int16
}

func (t BasicType) IsInt32() bool {
	return t.kind == stdtypes.Int32
}

func (t BasicType) IsInt64() bool {
	return t.kind == stdtypes.Int64
}

func (t BasicType) IsAnyUint() bool {
	switch t.kind {
	case stdtypes.Uint, stdtypes.Uint8, stdtypes.Uint16, stdtypes.Uint32, stdtypes.Uint64:
		return true
	}
	return false
}

func (t BasicType) IsUint() bool {
	return t.kind == stdtypes.Uint
}

func (t BasicType) IsUint8() bool {
	return t.kind == stdtypes.Uint8
}

func (t BasicType) IsUint16() bool {
	return t.kind == stdtypes.Uint16
}

func (t BasicType) IsUint32() bool {
	return t.kind == stdtypes.Uint32
}

func (t BasicType) IsUint64() bool {
	return t.kind == stdtypes.Uint64
}

func (t BasicType) IsAnyFloat() bool {
	switch t.kind {
	case stdtypes.Float32, stdtypes.Float64:
		return true
	}
	return false
}

func (t BasicType) IsFloat32() bool {
	return t.kind == stdtypes.Float32
}

func (t BasicType) IsFloat64() bool {
	return t.kind == stdtypes.Float64
}

func (t BasicType) IsBool() bool {
	return t.kind == stdtypes.Bool
}

type SelectorType struct {
	Sel interface{}
	X   interface{}
}

type PositionType struct {
	Column   int
	Filename string
	Line     int
	Offset   int
	IsValid  bool
}

type MapType struct {
	Key       interface{}
	Value     interface{}
	IsPointer bool
}

type SliceType struct {
	Value     interface{}
	IsPointer bool
}

type ArrayType struct {
	Value     interface{}
	Len       int64
	IsPointer bool
}
