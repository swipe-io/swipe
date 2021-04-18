package types

import (
	"go/ast"
	"go/token"
	"go/types"
	"strconv"

	"golang.org/x/tools/go/packages"
)

var (
	ErrorType = types.Universe.Lookup("error").Type()
	PanicType = types.Universe.Lookup("panic").Type()
	NilType   = types.Universe.Lookup("nil").Type()
)

func IsNil(t types.Type) bool {
	return types.Identical(t, NilType)
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

func HasNoEmptyValue(t types.Type) bool {
	switch u := t.Underlying().(type) {
	case *types.Array, *types.Struct:
		return true
	case *types.Basic:
		info := u.Info()
		switch {
		case info&types.IsBoolean != 0:
			return true
		}
	}
	return false
}

//func ZeroValue(t types.Type, qf types.Qualifier) string {
//	switch u := t.Underlying().(type) {
//	case *types.Array, *types.Struct:
//		return types.TypeString(t, qf) + "{}"
//	case *types.Basic:
//		info := u.Info()
//		switch {
//		case info&types.IsBoolean != 0:
//			return "false"
//		case info&(types.IsInteger|types.IsFloat|types.IsComplex) != 0:
//			return "0"
//		case info&types.IsString != 0:
//			return `""`
//		default:
//			panic("unreachable")
//		}
//	case *types.Chan, *types.Interface, *types.Map, *types.Pointer, *types.Signature, *types.Slice:
//		return "nil"
//	default:
//		panic("unreachable")
//	}
//}

func ZeroValue(t types.Type) string {
	switch u := t.Underlying().(type) {
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

func EvalBinaryExpr(expr *ast.BinaryExpr) (n, iotas int) {
	x, iotasX := EvalInt(expr.X)
	y, iotasY := EvalInt(expr.Y)
	iotas = iotas + iotasX + iotasY
	switch expr.Op {
	case token.ADD:
		n = x + y
	case token.SUB:
		n = x - y
	case token.MUL:
		n = x * y
	case token.QUO:
		n = x / y
	case token.REM:
		n = x % y
	}
	return
}

func EvalInt(expr ast.Expr) (n, iotas int) {
	switch exp := expr.(type) {
	case *ast.Ident:
		if exp.Name == "iota" {
			iotas++
		}
	case *ast.ParenExpr:
		return EvalInt(exp.X)
	case *ast.BasicLit:
		n, _ = strconv.Atoi(exp.Value)
		return
	case *ast.BinaryExpr:
		return EvalBinaryExpr(exp)
	}
	return
}
