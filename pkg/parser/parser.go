package parser

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

func MustOption(o *Option, _ bool) *Option {
	return o
}

type Value struct {
	v    interface{}
	t    types.Type
	expr ast.Expr
	pos  token.Position
}

func (v Value) Pos() token.Position {
	return v.pos
}

func (v Value) Type() types.Type {
	return v.t
}

func (v Value) Expr() ast.Expr {
	return v.expr
}

func (v Value) ExprSlice() (value []ast.Expr) {
	value, _ = v.v.([]ast.Expr)
	return
}

func (v Value) StringSlice() (value []string) {
	value, _ = v.v.([]string)
	return
}

func (v Value) String() (value string) {
	value, _ = v.v.(string)
	return
}

func (v Value) Int() (value int) {
	value, _ = v.v.(int)
	return
}

func (v Value) Bool() (value bool) {
	value, _ = v.v.(bool)
	return
}

func (v Value) Float() (value float32) {
	value, _ = v.v.(float32)
	return
}

type Option struct {
	FnObj      types.Object
	Name       string
	Value      Value
	Properties Properties
	Position   token.Position
}

type Properties map[string][]*Option

func (o Option) OneOf(key string) bool {
	if v, ok := o.Properties[key]; ok {
		return len(v) == 1
	}
	return true
}

func (o Option) Exists(key string) bool {
	_, ok := o.Properties[key]
	return ok
}

func (o Option) GetSlice(key string) ([]*Option, bool) {
	if values, ok := o.Properties[key]; ok {
		return values, true
	}
	return nil, false
}

func (o Option) Get(key string) (*Option, bool) {
	if values, ok := o.GetSlice(key); ok {
		if len(values) > 0 {
			return values[0], true
		}
	}
	return nil, false
}

type Parser struct {
	pkg *packages.Package
}

func (p Parser) Parse(expr ast.Expr) (*Option, error) {
	return p.process(expr)
}

func (p Parser) process(expr ast.Expr) (*Option, error) {
	exprPos := p.pkg.Fset.Position(expr.Pos())

	result := &Option{Position: exprPos, Properties: map[string][]*Option{}}

	expr = astutil.Unparen(expr)

	switch v := expr.(type) {
	case *ast.CallExpr:
		if e, ok := v.Fun.(*ast.ParenExpr); ok {
			return p.process(e.X)
		}

		fnObj := p.qualifiedObject(v.Fun)

		result.FnObj = fnObj
		result.Name = fnObj.Name()

		fnSig := fnObj.Type().(*types.Signature)

		for i, expr := range v.Args {
			if fnSig.Variadic() && i >= fnSig.Params().Len()-1 {
				val, err := p.process(expr)
				if err != nil {
					return nil, err
				}
				result.Properties[val.Name] = append(result.Properties[val.Name], val)
			} else {
				t := p.pkg.TypesInfo.TypeOf(expr)
				value := p.getValue(expr)
				v := Value{
					v:    value,
					t:    t,
					expr: v.Args[i],
					pos:  p.pkg.Fset.Position(v.Args[i].Pos()),
				}
				if fnSig.Params().Len() == 1 {
					result.Value = v
				} else {
					name := fnSig.Params().At(i).Name()

					result.Properties[name] = append(
						result.Properties[name],
						&Option{FnObj: fnObj, Name: name, Position: exprPos, Value: v, Properties: map[string][]*Option{}},
					)
				}
			}
		}
	}
	return result, nil
}

func (p Parser) getValue(expr ast.Expr) interface{} {
	var v interface{}
	if tv, ok := p.pkg.TypesInfo.Types[expr]; ok {
		if tv.IsValue() && tv.Value != nil {
			switch tv.Value.Kind() {
			case constant.String:
				v = constant.StringVal(tv.Value)
			case constant.Bool:
				v = constant.BoolVal(tv.Value)
			case constant.Float:
				v, _ = constant.Float64Val(tv.Value)
			case constant.Int:
				v, _ = constant.Int64Val(tv.Value)
			}
		} else {
			switch ve := expr.(type) {
			case *ast.CompositeLit:
				switch vt := ve.Type.(type) {
				case *ast.ArrayType:
					switch elt := vt.Elt.(type) {
					case *ast.Ident:
						switch elt.Name {
						default:
							v = ve.Elts
						case "string":
							v = p.makeStringSlice(ve.Elts)
						}
					}
				}
			}
		}
	}
	return v
}

func (p Parser) qualifiedObject(expr ast.Expr) types.Object {
	switch expr := expr.(type) {
	case *ast.Ident:
		return p.pkg.TypesInfo.ObjectOf(expr)
	case *ast.SelectorExpr:
		pkgName, ok := expr.X.(*ast.Ident)
		if !ok {
			return nil
		}
		if _, ok := p.pkg.TypesInfo.ObjectOf(pkgName).(*types.PkgName); !ok {
			return nil
		}
		return p.pkg.TypesInfo.ObjectOf(expr.Sel)
	default:
		return nil
	}
}

func (p *Parser) makeStringSlice(exprs []ast.Expr) (result []string) {
	for _, expr := range exprs {
		result = append(result, p.getValue(expr).(string))
	}
	return
}

func NewParser(pkg *packages.Package) *Parser {
	return &Parser{pkg: pkg}
}
