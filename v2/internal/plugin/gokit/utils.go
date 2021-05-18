package gokit

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	stdtypes "go/types"
	"strings"

	"github.com/swipe-io/swipe/v2/internal/option"
	"github.com/swipe-io/swipe/v2/internal/plugin/gokit/config"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

func httpBraceIndices(s string) ([]int, error) {
	var level, idx int
	var idxs []int
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '{':
			if level++; level == 1 {
				idx = i
			}
		case '}':
			if level--; level == 0 {
				idxs = append(idxs, idx, i+1)
			} else if level < 0 {
				return nil, fmt.Errorf("mux: unbalanced braces in %q", s)
			}
		}
	}
	if level != 0 {
		return nil, fmt.Errorf("mux: unbalanced braces in %q", s)
	}
	return idxs, nil
}

func pathVars(path string) (map[string]string, error) {
	idxs, err := httpBraceIndices(path)
	if err != nil {
		return nil, err
	}
	pathVars := make(map[string]string, len(idxs))
	if len(idxs) > 0 {
		var end int
		for i := 0; i < len(idxs); i += 2 {
			end = idxs[i+1]
			parts := strings.SplitN(path[idxs[i]+1:end-1], ":", 2)
			name := parts[0]
			regexp := ""
			if len(parts) == 2 {
				regexp = parts[1]
			}
			pathVars[name] = regexp
		}
	}
	return pathVars, nil
}

type typeInfo struct {
	obj      stdtypes.Object
	stmtList []ast.Stmt
	pkg      *packages.Package
}

func extractValues(pkg *packages.Package, stmtList []ast.Stmt) (values []interface{}) {
	for _, stmt := range stmtList {
		if ret, ok := stmt.(*ast.ReturnStmt); ok {
			for _, result := range ret.Results {
				if l, ok := result.(*ast.BasicLit); ok {
					if v, ok := pkg.TypesInfo.Types[l]; ok {
						values = append(values, constant.Val(v.Value))
					}
				}
			}
		}
	}
	return
}

func findErrorsRecursive(funcDecl map[string]typeInfo, pkgs []*packages.Package, stmts []ast.Stmt) (result []config.Error) {
	for _, stmt := range stmts {
		ret, ok := stmt.(*ast.ReturnStmt)
		if !ok {
			continue
		}
		for _, r := range ret.Results {
			switch t := r.(type) {
			default:
				named := extractNamed(pkgs, t)
				if named != nil {
					for i := 0; i < named.NumMethods(); i++ {
						m := named.Method(i)
						if m.Name() == "ErrorCode" || m.Name() == "StatusCode" {
							info, ok := funcDecl[m.Pkg().Name()+"."+m.Name()]
							if !ok {
								continue
							}
							values := extractValues(info.pkg, info.stmtList)
							if len(values) != 1 {
								continue
							}
							val, ok := values[0].(int64)
							if !ok {
								continue
							}
							tp := config.RESTErrorType
							if m.Name() == "ErrorCode" {
								tp = config.JRPCErrorType
							}
							result = append(result, config.Error{
								PkgName:   named.Obj().Pkg().Name(),
								PkgPath:   named.Obj().Pkg().Path(),
								IsPointer: named.IsPointer,
								Name:      named.Obj().Name(),
								Type:      tp,
								Code:      val,
							})
						}
					}
				}
			case *ast.CallExpr:
				if id, ok := t.Fun.(*ast.Ident); ok && id.Obj != nil {
					if f, ok := id.Obj.Decl.(*ast.FuncDecl); ok {
						result = append(result, findErrorsRecursive(funcDecl, pkgs, f.Body.List)...)
					}
				} else if sel, ok := t.Fun.(*ast.SelectorExpr); ok {
					if xID, ok := sel.X.(*ast.Ident); ok {
						if info, ok := funcDecl[xID.Name+"."+sel.Sel.Name]; ok {
							result = append(result, findErrorsRecursive(funcDecl, pkgs, info.stmtList)...)
						}
					}
				}
			}
		}
	}
	return
}

func findErrors(funcDecl map[string]typeInfo, pkgs []*packages.Package, stmts []ast.Stmt) []config.Error {
	return findErrorsRecursive(funcDecl, pkgs, stmts)
}

type named struct {
	*stdtypes.Named
	IsPointer bool
}

func extractNamedRecursive(pkgs []*packages.Package, expr ast.Expr, isPointer bool) *named {
	expr = astutil.Unparen(expr)
	switch t := expr.(type) {
	case *ast.CompositeLit:
		for _, pkg := range pkgs {
			if v, ok := pkg.TypesInfo.Types[t.Type]; ok {
				if n, ok := v.Type.(*stdtypes.Named); ok {
					return &named{Named: n, IsPointer: isPointer}
				}
			}
		}
	case *ast.StarExpr:
		return extractNamedRecursive(pkgs, t.X, isPointer)
	case *ast.UnaryExpr:
		return extractNamedRecursive(pkgs, t.X, t.Op == token.AND)
	}
	return nil
}

func extractNamed(pkgs []*packages.Package, expr ast.Expr) *named {
	return extractNamedRecursive(pkgs, expr, false)
}

func findIfaceErrors(funcDecl map[string]typeInfo, pkgs []*packages.Package, ifaces []*config.Interface) (result map[string]map[string][]config.Error) {
	result = map[string]map[string][]config.Error{}
	for _, info := range funcDecl {
		if info.obj == nil {
			continue
		}
		sig, ok := info.obj.Type().(*stdtypes.Signature)
		if !ok || sig.Recv() == nil {
			continue
		}
		for _, iface := range ifaces {
			if ptr, ok := sig.Recv().Type().(*stdtypes.Pointer); ok {
				imp := stdtypes.Implements(ptr.Underlying(), iface.Named.Type.(*option.IfaceType).Origin)
				if imp {
					if _, ok := result[iface.Named.Name.Origin]; !ok {
						result[iface.Named.Name.Origin] = map[string][]config.Error{}
					}
					result[iface.Named.Name.Origin][info.obj.Name()] = findErrors(funcDecl, pkgs, info.stmtList)
				}
			}
		}
	}
	return
}

func makeFuncDeclTypes(pkgs []*packages.Package) (result map[string]typeInfo) {
	result = make(map[string]typeInfo, 1024)
	for _, pkg := range pkgs {
		if pkg.Name == "swipe" {
			continue
		}
		for _, syntax := range pkg.Syntax {
			for _, decl := range syntax.Decls {
				switch t := decl.(type) {
				case *ast.FuncDecl:
					obj := pkg.TypesInfo.ObjectOf(t.Name)
					if obj != nil {
						result[pkg.Name+"."+t.Name.Name] = typeInfo{
							obj:      obj,
							pkg:      pkg,
							stmtList: t.Body.List,
						}
					}
				}
			}
		}
	}
	return
}
