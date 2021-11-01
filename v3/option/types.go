package option

import (
	"go/types"
	stdtypes "go/types"

	"github.com/fatih/structtag"

	"github.com/swipe-io/strcase"
)

type String struct {
	Value      string
	upperValue string
	lowerValue string
}

func (n String) Upper() string {
	if n.upperValue == "" {
		n.upperValue = strcase.ToCamel(n.Value)
	}
	return n.upperValue
}

func (n String) Lower() string {
	if n.lowerValue == "" {
		n.lowerValue = strcase.ToLowerCamel(n.Value)
	}
	return n.lowerValue
}

func (n String) String() string {
	return n.Value
}

type VarsType []*VarType

type VarType struct {
	Name       String
	Embedded   bool
	Exported   bool
	IsField    bool
	IsVariadic bool
	IsContext  bool
	Type       interface{}
	Comment    string
	Zero       string

	originType stdtypes.Type
}

type StructType struct {
	Fields    []*StructFieldType
	IsPointer bool

	originType stdtypes.Type
}

type StructFieldType struct {
	Var  *VarType
	Tags *structtag.Tags
}

type SignType struct {
	Params     VarsType
	Results    VarsType
	IsVariadic bool
	IsNamed    bool
	Recv       interface{}
}

type FuncType struct {
	Pkg      *PackageType
	FullName string
	Name     String
	Exported bool
	Sig      *SignType
	Comment  string
}

func (f *FuncType) ID() string {
	return f.Pkg.Path + "." + f.Name.Value
}

type IfaceType struct {
	Origin          *stdtypes.Interface
	Methods         []*FuncType
	Embeddeds       []interface{}
	ExplicitMethods []*FuncType
}

type ModuleType struct {
	ID       string
	Version  string
	Path     string
	Dir      string
	External bool
}

type PackageType struct {
	Name   string
	Path   string
	Module *ModuleType
	Types  *stdtypes.Package
}

type NamedType struct {
	Obj       stdtypes.Object
	Name      String
	Type      interface{}
	Pkg       *PackageType
	IsPointer bool
	Methods   []*FuncType
}

func (n *NamedType) ID() string {
	return n.Pkg.Path + "." + n.Name.Value
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

func (t BasicType) IsByte() bool {
	return t.kind == stdtypes.Byte
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
