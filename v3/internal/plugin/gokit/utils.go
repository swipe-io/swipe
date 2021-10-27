package gokit

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/types"
	stdtypes "go/types"
	"strings"

	"github.com/swipe-io/swipe/v3/option"

	"github.com/swipe-io/swipe/v3/internal/packages"
	"github.com/swipe-io/swipe/v3/internal/plugin/gokit/config"
	stdpackages "golang.org/x/tools/go/packages"
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
	pkg      *stdpackages.Package
	recv     *types.Var
}

func extractValues(pkg *stdpackages.Package, stmtList []ast.Stmt) (values []interface{}) {
	for _, stmt := range stmtList {
		if ret, ok := stmt.(*ast.ReturnStmt); ok {
			for _, result := range ret.Results {
				if v, ok := pkg.TypesInfo.Types[result]; ok {
					values = append(values, constant.Val(v.Value))
				}
			}
		}
	}
	return
}

func findErrors(modulePath string, declTypes map[string]*typeInfo, pkgs *packages.Packages) (result map[string]config.Error) {
	result = make(map[string]config.Error, 1024)
	_ = pkgs.TraverseObjects(func(pkg *stdpackages.Package, id *ast.Ident, obj stdtypes.Object) (err error) {
		if !strings.Contains(pkg.PkgPath, modulePath) {
			return
		}
		if t, ok := obj.Type().(*types.Named); ok {
			for i := 0; i < t.NumMethods(); i++ {
				m := t.Method(i)
				if m.Name() == "ErrorCode" || m.Name() == "StatusCode" {
					sig := m.Type().(*types.Signature)
					id := m.Name()
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
						values := extractValues(pkg, info.stmtList)
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
						result[t.Obj().Pkg().Path()+"/"+t.Obj().Name()] = config.Error{
							PkgName: t.Obj().Pkg().Name(),
							PkgPath: t.Obj().Pkg().Path(),
							Name:    t.Obj().Name(),
							Type:    tp,
							Code:    val,
						}
					}
				}
			}
		}
		return
	})
	return
}

func extractSelector(e ast.Expr) *ast.SelectorExpr {
	switch t := e.(type) {
	case *ast.SelectorExpr:
		return t
	}
	return nil
}

func findIfaceErrorsRecursive(pkgs *packages.Packages, funcDecl map[string]*typeInfo, ifaceTypes map[string][]*typeInfo, errors map[string]config.Error, visited map[string]struct{}, stmts []ast.Stmt) (results []config.Error) {
	for _, stmt := range stmts {
		switch t := stmt.(type) {
		case *ast.ReturnStmt:
			for _, result := range t.Results {
				call, ok := result.(*ast.CallExpr)
				if !ok {
					if unary, ok := result.(*ast.UnaryExpr); ok {
						if cpl, ok := unary.X.(*ast.CompositeLit); ok {
							if sel, ok := cpl.Type.(*ast.SelectorExpr); ok {
								if obj := pkgs.ObjectOf(sel.Sel); obj != nil {
									if e, ok := errors[obj.Pkg().Path()+"/"+obj.Name()]; ok {
										results = append(results, e)
									}
									break
								}
							}
						}
					}
					continue
				}
				selFun := extractSelector(call.Fun)
				if selFun != nil {
					sel := extractSelector(selFun.X)
					if sel == nil {
						sel = selFun
					}
					if obj := pkgs.ObjectOf(sel.Sel); obj != nil {
						if named, ok := obj.Type().(*types.Named); ok {
							if _, ok := named.Obj().Type().Underlying().(*types.Interface); ok {
								id := named.Obj().Pkg().Path() + "/" + named.Obj().Name() + "." + selFun.Sel.Name
								if _, ok := visited[id]; ok {
									break
								}
								visited[id] = struct{}{}
								if infos, ok := ifaceTypes[id]; ok {
									for _, info := range infos {
										results = append(results, findIfaceErrorsRecursive(pkgs, funcDecl, ifaceTypes, errors, visited, info.stmtList)...)
									}
									break
								}
							}
						}
					}
				}
			}
		case *ast.IfStmt:
			results = append(results, findIfaceErrorsRecursive(pkgs, funcDecl, ifaceTypes, errors, visited, t.Body.List)...)
		case *ast.SwitchStmt:
			results = append(results, findIfaceErrorsRecursive(pkgs, funcDecl, ifaceTypes, errors, visited, t.Body.List)...)
		case *ast.BlockStmt:
			results = append(results, findIfaceErrorsRecursive(pkgs, funcDecl, ifaceTypes, errors, visited, t.List)...)
		case *ast.CommClause:
			results = append(results, findIfaceErrorsRecursive(pkgs, funcDecl, ifaceTypes, errors, visited, t.Body)...)
		case *ast.SelectStmt:
			results = append(results, findIfaceErrorsRecursive(pkgs, funcDecl, ifaceTypes, errors, visited, t.Body.List)...)
		case *ast.ForStmt:
			results = append(results, findIfaceErrorsRecursive(pkgs, funcDecl, ifaceTypes, errors, visited, t.Body.List)...)
		case *ast.RangeStmt:
			results = append(results, findIfaceErrorsRecursive(pkgs, funcDecl, ifaceTypes, errors, visited, t.Body.List)...)
		}
	}
	return
}

func findIfaceErrors(funcDecl map[string]*typeInfo, ifaceTypes map[string][]*typeInfo, errors map[string]config.Error, pkgs *packages.Packages, ifaces []*config.Interface) (result map[string]map[string][]config.Error) {
	result = map[string]map[string][]config.Error{}
	visited := map[string]struct{}{}
	for _, iface := range ifaces {
		i := iface.Named.Type.(*option.IfaceType)
		for _, m := range i.Methods {
			id := iface.Named.Pkg.Path + "/" + iface.Named.Name.Value + "." + m.Name.Value
			if fns, ok := ifaceTypes[id]; ok {
				for _, info := range fns {
					visited[id] = struct{}{}
					if _, ok := result[iface.Named.Name.Value]; !ok {
						result[iface.Named.Name.Value] = map[string][]config.Error{}
					}
					result[iface.Named.Name.Value][info.obj.Name()] = findIfaceErrorsRecursive(pkgs, funcDecl, ifaceTypes, errors, visited, info.stmtList)
				}
			}
		}
	}
	return
}

func makeFuncIfaceDeclTypes(pkgs *packages.Packages, funcDecl map[string]*typeInfo) (result map[string][]*typeInfo) {
	result = make(map[string][]*typeInfo, 1024)
	_ = pkgs.TraverseDecls(func(pkg *stdpackages.Package, file *ast.File, decl ast.Decl) (err error) {
		if strings.Contains(pkg.PkgPath, "/pkg/swipe/") {
			return
		}
		switch t := decl.(type) {
		case *ast.GenDecl:
			for _, spec := range t.Specs {
				if tp, ok := spec.(*ast.TypeSpec); ok {
					obj := pkg.TypesInfo.ObjectOf(tp.Name)
					named, ok := obj.Type().(*types.Named)
					if !ok {
						continue
					}
					iface, ok := named.Obj().Type().Underlying().(*types.Interface)
					if !ok {
						continue
					}
					for _, info := range funcDecl {
						if info.recv == nil {
							continue
						}
						ptr, ok := info.recv.Type().(*stdtypes.Pointer)
						if !ok {
							continue
						}
						for i := 0; i < iface.NumEmbeddeds(); i++ {
							if embeddedNamed, ok := iface.EmbeddedType(i).(*types.Named); ok {
								if embeddedIface, ok := embeddedNamed.Obj().Type().Underlying().(*types.Interface); ok {
									imp := stdtypes.Implements(ptr.Underlying(), embeddedIface)
									if imp {
										id := pkg.PkgPath + "/" + named.Obj().Name() + "." + info.obj.Name()
										result[id] = append(result[id], info)
									}
								}
							}
						}
						imp := stdtypes.Implements(ptr.Underlying(), iface)
						if imp {
							id := pkg.PkgPath + "/" + named.Obj().Name() + "." + info.obj.Name()
							result[id] = append(result[id], info)
						}
					}
				}
			}
		}
		return
	})
	return
}

func makeFuncDeclTypes(pkgs *packages.Packages) (result map[string]*typeInfo) {
	result = make(map[string]*typeInfo, 1024)
	_ = pkgs.TraverseDecls(func(pkg *stdpackages.Package, file *ast.File, decl ast.Decl) (err error) {
		if strings.Contains(pkg.PkgPath, "/pkg/swipe/") {
			return
		}
		switch t := decl.(type) {
		case *ast.FuncDecl:
			obj := pkg.TypesInfo.ObjectOf(t.Name)
			if obj != nil {
				sig := pkg.TypesInfo.TypeOf(t.Name).(*types.Signature)
				id := t.Name.Name
				recv := sig.Recv()
				if recv != nil {
					recvType := recv.Type()
					if ptr, ok := recvType.(*types.Pointer); ok {
						recvType = ptr.Elem()
					}
					recvNamed := recvType.(*types.Named)
					id = "/" + recvNamed.Obj().Name() + "." + id
				} else {
					id = "." + id
				}
				result[pkg.PkgPath+id] = &typeInfo{
					obj:      obj,
					pkg:      pkg,
					recv:     recv,
					stmtList: t.Body.List,
				}
			}
		}
		return
	})
	return
}

func fillMethodDefaultOptions(method, methodDefault config.MethodOptions) config.MethodOptions {
	if !method.RESTMethod.IsValid() {
		method.RESTMethod = methodDefault.RESTMethod
	}
	if !method.RESTMultipartMaxMemory.IsValid() {
		method.RESTMultipartMaxMemory = methodDefault.RESTMultipartMaxMemory
	}
	if !method.RESTBodyType.IsValid() {
		method.RESTBodyType = methodDefault.RESTBodyType
	}
	if method.RESTHeaderVars.Value == nil {
		method.RESTHeaderVars.Value = methodDefault.RESTHeaderVars.Value
	}
	if !method.RESTPath.IsValid() {
		method.RESTPath = methodDefault.RESTPath
	}
	if method.RESTQueryValues.Value == nil {
		method.RESTQueryValues.Value = methodDefault.RESTQueryValues.Value
	}
	if method.RESTQueryVars.Value == nil {
		method.RESTQueryVars.Value = methodDefault.RESTQueryVars.Value
	}
	if !method.RESTWrapResponse.IsValid() {
		method.RESTWrapResponse.Value = methodDefault.RESTWrapResponse.Value
	}
	if method.ClientEncodeRequest.Value == nil {
		method.ClientEncodeRequest.Value = methodDefault.ClientEncodeRequest.Value
	}
	if method.ClientDecodeResponse.Value == nil {
		method.ClientDecodeResponse.Value = methodDefault.ClientDecodeResponse.Value
	}
	if method.ServerDecodeRequest.Value == nil {
		method.ServerDecodeRequest.Value = methodDefault.ServerDecodeRequest.Value
	}
	if method.ServerEncodeResponse.Value == nil {
		method.ServerEncodeResponse.Value = methodDefault.ServerEncodeResponse.Value
	}
	if !method.Instrumenting.IsValid() {
		method.Instrumenting = methodDefault.Instrumenting
	}
	if !method.Logging.IsValid() {
		method.Logging = methodDefault.Logging
	}
	if method.LoggingContext == nil {
		method.LoggingContext = methodDefault.LoggingContext
	}
	if method.LoggingParams.Excludes == nil {
		method.LoggingParams.Excludes = methodDefault.LoggingParams.Excludes
	}
	if method.LoggingParams.Includes == nil {
		method.LoggingParams.Includes = methodDefault.LoggingParams.Includes
	}
	return method
}
