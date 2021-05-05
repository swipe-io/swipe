package value

//import (
//	"go/ast"
//	"go/constant"
//	"go/types"
//)
//
//type ValueType struct {
//	expr ast.Expr
//	t    types.Type
//	v    interface{}
//}
//
//func (vl *ValueType) Type() types.Type {
//	return vl.t
//}
//
//func (vl *ValueType) String() (v string) {
//	v, _ = vl.v.(string)
//	return
//}
//
//func (vl *ValueType) Int() (v int64) {
//	v, _ = vl.v.(int64)
//	return
//}
//
//func (vl *ValueType) Float() (v float64) {
//	v, _ = vl.v.(float64)
//	return
//}
//
//func (vl *ValueType) Bool() (v bool) {
//	v, _ = vl.v.(bool)
//	return
//}
//
//func (vl *ValueType) HasValue() bool {
//	return vl.v != nil
//}
//
//func (vl *ValueType) Expr() ast.Expr {
//	return vl.expr
//}
//
//func makeValueExpr(info *types.Info, expr ast.Expr) ValueType {
//	tv, ok := info.Types[expr]
//	if !ok {
//		panic("unknown type")
//	}
//	value := ValueType{
//		expr: expr,
//		t:    tv.Type,
//	}
//
//	if tv.IsValue() && tv.ValueType != nil {
//		value.v = typeValueToValue(tv)
//	}
//	return value
//}
//
//func ProcessValueExpr(info *types.Info, expr ast.Expr) (ValueType, error) {
//	return makeValueExpr(info, expr), nil
//}
//
//func ProcessValueExprs(info *types.Info, args []ast.Expr) ([]ValueType, error) {
//	var values []ValueType
//	for _, arg := range args {
//		values = append(values, makeValueExpr(info, arg))
//	}
//	return values, nil
//}
//
//func typeValueToValue(v types.TypeAndValue) interface{} {
//	switch v.ValueType.Kind() {
//	case constant.String:
//		return constant.StringVal(v.ValueType)
//	case constant.Bool:
//		return constant.BoolVal(v.ValueType)
//	case constant.Float:
//		if v, ok := constant.Float64Val(v.ValueType); ok {
//			return v
//		}
//	case constant.Int:
//		if v, ok := constant.Int64Val(v.ValueType); ok {
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
