package types

import (
	"fmt"
	"go/ast"
	"go/types"

	"github.com/spaolacci/murmur3"

	"golang.org/x/tools/go/packages"
)

var (
	ErrorType = types.Universe.Lookup("error").Type()
	PanicType = types.Universe.Lookup("panic").Type()
	NilType   = types.Universe.Lookup("nil").Type()
)

func Hash(name string, hash uint32) uint32 {
	h := murmur3.New32()
	_, _ = h.Write([]byte(fmt.Sprintf("%s::%d", name, hash)))
	return h.Sum32()
}

func GetBitSize(kind types.BasicKind) string {
	switch kind {
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

func IsError(t types.Type) bool {
	return types.Identical(t, ErrorType)
}

func IsPanic(t types.Type) bool {
	return types.Identical(t, PanicType)
}

func ContainsContext(t *types.Tuple) bool {
	if t.Len() > 0 {
		return IsContext(t.At(0).Type())
	}
	return false

}

func ContainsError(results *types.Tuple) bool {
	if results.Len() > 0 {
		return IsError(results.At(results.Len() - 1).Type())
	}
	return false
}

func IsContext(t types.Type) bool {
	return types.TypeString(t, nil) == "context.Context"
}

func LenWithoutErr(t *types.Tuple) int {
	len := t.Len()
	if ContainsError(t) {
		len--
	}
	return len
}

func LenWithoutContext(t *types.Tuple) int {
	len := t.Len()
	if ContainsContext(t) {
		len--
	}
	return len
}

func LookupField(name string, sig *types.Signature) *types.Var {
	for i := 0; i < sig.Params().Len(); i++ {
		p := sig.Params().At(i)
		if p.Name() == name {
			return p
		}
	}
	return nil
}

func IsNamed(t *types.Tuple) bool {
	if t.Len() > 0 {
		return t.At(0).Name() != ""
	}
	return false
}

func Inspect(pkgs []*packages.Package, f func(p *packages.Package, n ast.Node) bool) {
	for _, p := range pkgs {
		for _, syntax := range p.Syntax {
			ast.Inspect(syntax, func(n ast.Node) bool {
				return f(p, n)
			})
		}
	}
}

func ZeroValue(t types.Type) string {
	switch u := t.Underlying().(type) {
	case *types.Array, *types.Struct:
		return types.TypeString(t, func(p *types.Package) string {
			return p.Name()
		}) + "{}"
	case *types.Basic:
		info := u.Info()
		switch {
		case info&types.IsBoolean != 0:
			return "false"
		case info&(types.IsInteger|types.IsFloat|types.IsComplex) != 0:
			return "0"
		case info&types.IsString != 0:
			return `""`
		default:
			panic("unreachable")
		}
	case *types.Chan, *types.Interface, *types.Map, *types.Pointer, *types.Signature, *types.Slice:
		return "nil"
	default:
		panic("unreachable")
	}
}
