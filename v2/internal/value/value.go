package value

//import (
//	"go/ast"
//	"go/constant"
//	"go/types"
//)
//
//type Value struct {
//	expr ast.Expr
//	t    types.Type
//	v    interface{}
//}
//
//func (vl *Value) Type() types.Type {
//	return vl.t
//}
//
//func (vl *Value) String() (v string) {
//	v, _ = vl.v.(string)
//	return
//}
//
//func (vl *Value) Int() (v int64) {
//	v, _ = vl.v.(int64)
//	return
//}
//
//func (vl *Value) Float() (v float64) {
//	v, _ = vl.v.(float64)
//	return
//}
//
//func (vl *Value) Bool() (v bool) {
//	v, _ = vl.v.(bool)
//	return
//}
//
//func (vl *Value) HasValue() bool {
//	return vl.v != nil
//}
//
//func (vl *Value) Expr() ast.Expr {
//	return vl.expr
//}
//
//func makeValueExpr(info *types.Info, expr ast.Expr) Value {
//	tv, ok := info.Types[expr]
//	if !ok {
//		panic("unknown type")
//	}
//	value := Value{
//		expr: expr,
//		t:    tv.Type,
//	}
//
//	if tv.IsValue() && tv.Value != nil {
//		value.v = typeValueToValue(tv)
//	}
//	return value
//}
//
//func ProcessValueExpr(info *types.Info, expr ast.Expr) (Value, error) {
//	return makeValueExpr(info, expr), nil
//}
//
//func ProcessValueExprs(info *types.Info, args []ast.Expr) ([]Value, error) {
//	var values []Value
//	for _, arg := range args {
//		values = append(values, makeValueExpr(info, arg))
//	}
//	return values, nil
//}
//
//func typeValueToValue(v types.TypeAndValue) interface{} {
//	switch v.Value.Kind() {
//	case constant.String:
//		return constant.StringVal(v.Value)
//	case constant.Bool:
//		return constant.BoolVal(v.Value)
//	case constant.Float:
//		if v, ok := constant.Float64Val(v.Value); ok {
//			return v
//		}
//	case constant.Int:
//		if v, ok := constant.Int64Val(v.Value); ok {
//			return v
//		}
//	}
//	return nil
//}
//
//func QualifiedIdentObject(info *types.Info, expr ast.Expr) types.Object {
//	switch expr := expr.(type) {
//	case *ast.Ident:
//		return info.ObjectOf(expr)
//	case *ast.SelectorExpr:
//		pkgName, ok := expr.X.(*ast.Ident)
//		if !ok {
//			return nil
//		}
//		if _, ok := info.ObjectOf(pkgName).(*types.PkgName); !ok {
//			return nil
//		}
//		return info.ObjectOf(expr.Sel)
//	default:
//		return nil
//	}
//}
