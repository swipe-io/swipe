package option

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/packages"
)

func qualifiedObject(pkg *packages.Package, expr ast.Expr) types.Object {
	switch expr := expr.(type) {
	case *ast.Ident:
		return pkg.TypesInfo.ObjectOf(expr)
	case *ast.SelectorExpr:
		pkgName, ok := expr.X.(*ast.Ident)
		if !ok {
			return nil
		}
		if _, ok := pkg.TypesInfo.ObjectOf(pkgName).(*types.PkgName); !ok {
			return nil
		}
		return pkg.TypesInfo.ObjectOf(expr.Sel)
	default:
		return nil
	}
}

func findInjector(info *types.Info, fn *ast.FuncDecl) (*ast.CallExpr, error) {
	if fn.Body == nil {
		return nil, nil
	}
	for _, stmt := range fn.Body.List {
		switch stmt := stmt.(type) {
		case *ast.ExprStmt:
			call, ok := stmt.X.(*ast.CallExpr)
			if !ok {
				continue
			}
			obj := qualifiedIdentObject(info, call.Fun)
			if obj == nil || obj.Pkg() == nil {
				continue
			}
			if obj.Name() != "Build" {
				continue
			}
			return call, nil
		case *ast.EmptyStmt:

			return nil, nil
		}
	}
	return nil, nil
}

func qualifiedIdentObject(info *types.Info, expr ast.Expr) types.Object {
	switch expr := expr.(type) {
	case *ast.Ident:
		return info.ObjectOf(expr)
	case *ast.SelectorExpr:
		pkgName, ok := expr.X.(*ast.Ident)
		if !ok {
			return nil
		}
		if _, ok := info.ObjectOf(pkgName).(*types.PkgName); !ok {
			return nil
		}
		return info.ObjectOf(expr.Sel)
	default:
		return nil
	}
}
