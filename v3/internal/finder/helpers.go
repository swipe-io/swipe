package finder

import (
	"go/ast"
	"go/constant"
	"go/types"
	stdtypes "go/types"

	stdpackages "golang.org/x/tools/go/packages"
)

func findMethodByNamed(t *stdtypes.Named, name ...string) *stdtypes.Func {
	for i := 0; i < t.NumMethods(); i++ {
		m := t.Method(i)
		for _, s := range name {
			if m.Name() == s {
				return m
			}
		}
	}
	return nil
}

func extractReturnValuesByFunc(pkg *stdpackages.Package, f *stdtypes.Func, declTypes map[string]*typeInfo) []interface{} {
	sig := f.Type().(*types.Signature)
	id := f.Name()
	if recv := sig.Recv(); recv != nil {
		recvType := recv.Type()
		if ptr, ok := recvType.(*types.Pointer); ok {
			recvType = ptr.Elem()
		}
		recvNamed := recvType.(*types.Named)
		id = "/" + recvNamed.Obj().Name() + "." + id
	} else {
		id = "." + id
	}
	if info, ok := declTypes[pkg.PkgPath+id]; ok {
		return extractValues(pkg, info.stmtList)
	}
	return nil
}

func extractValues(pkg *stdpackages.Package, stmtList []ast.Stmt) (values []interface{}) {
	for _, stmt := range stmtList {
		if ret, ok := stmt.(*ast.ReturnStmt); ok {
			for _, result := range ret.Results {
				if v, ok := pkg.TypesInfo.Types[result]; ok {
					tv := constant.Val(v.Value)
					if tv != nil {
						values = append(values, tv)
					} else {
						values = append(values, v.Type)
					}
				}
			}
		}
	}
	return
}

func extractSelector(e ast.Expr) *ast.SelectorExpr {
	switch t := e.(type) {
	case *ast.SelectorExpr:
		return t
	}
	return nil
}
