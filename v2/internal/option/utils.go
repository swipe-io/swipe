package option

import (
	"errors"
	"fmt"
	"go/ast"
	goast "go/ast"
	"go/constant"
	"go/types"
	stdtypes "go/types"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/tools/go/packages"
)

var paramCommentRegexp = regexp.MustCompile(`(?s)@([a-zA-Z0-9_]*) (.*)`)

func parseMethodComments(comments []string) (methodComment string, paramsComment map[string]string) {
	paramsComment = make(map[string]string)
	for _, comment := range comments {
		comment = strings.TrimSpace(comment)
		if strings.HasPrefix(comment, "@") {
			matches := paramCommentRegexp.FindAllStringSubmatch(comment, -1)
			if len(matches) == 1 && len(matches[0]) == 3 {
				paramsComment[matches[0][1]] = matches[0][2]
			}
			continue
		}
		methodComment += comment
	}
	return
}

func makeStringSlice(elts []goast.Expr, info *stdtypes.Info) (result []string) {
	for _, expr := range elts {
		tv := info.Types[expr]
		result = append(result, constant.Val(tv.Value).(string))
	}
	return
}

func sigParamAt(sig *stdtypes.Signature, i int) *stdtypes.Var {
	if sig.Variadic() && i >= sig.Params().Len()-1 {
		return sig.Params().At(sig.Params().Len() - 1)
	}
	return sig.Params().At(i)
}

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

func findPkgByID(pkgs []*packages.Package, path string) *packages.Package {
	for _, pkg := range pkgs {
		if pkg.PkgPath == path {
			return pkg
		}
	}
	return nil
}

func detectBasePath(pkg *packages.Package) (string, error) {
	paths := pkg.GoFiles
	if len(paths) == 0 {
		return "", errors.New("no files to derive output directory from")
	}
	dir := filepath.Dir(paths[0])
	for _, p := range paths[1:] {
		if dir2 := filepath.Dir(p); dir2 != dir {
			return "", fmt.Errorf("found conflicting directories %q and %q", dir, dir2)
		}
	}
	return dir, nil
}

func zeroValue(t types.Type) string {
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
	case *types.Struct:
		return "{}"
	case *types.Chan, *types.Interface, *types.Map, *types.Pointer, *types.Signature, *types.Slice:
		return "nil"
	default:
		panic("unreachable")
	}
}
