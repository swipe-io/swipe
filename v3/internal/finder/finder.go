package finder

import (
	"go/ast"
	"go/types"
	stdtypes "go/types"
	"strings"

	"github.com/swipe-io/swipe/v3/option"

	"github.com/swipe-io/swipe/v3/internal/packages"

	stdpackages "golang.org/x/tools/go/packages"
)

type Interface struct {
}

type Error struct {
	PkgName string
	PkgPath string
	Name    string
	Code    int64
	ErrCode string
}

type typeInfo struct {
	obj      stdtypes.Object
	stmtList []ast.Stmt
	pkg      *stdpackages.Package
	recv     *types.Var
}

type Finder struct {
	packages           *packages.Packages
	modulePath         string
	funcDeclTypes      map[string]*typeInfo
	funcDeclIfaceTypes map[string][]*typeInfo
}

func (f *Finder) FindErrors() (result map[string]Error) {
	result = make(map[string]Error, 1024)
	_ = f.packages.TraverseObjects(func(pkg *stdpackages.Package, id *ast.Ident, obj stdtypes.Object) (err error) {
		if !strings.Contains(pkg.PkgPath, f.modulePath) {
			return
		}
		if t, ok := obj.Type().(*types.Named); ok {

			if t.Obj().Pkg() != nil && strings.Contains(t.Obj().Pkg().Path(), "swipe") {
				return
			}

			methodFunc := findMethodByNamed(t, "ErrorCode", "StatusCode")
			if methodFunc == nil {
				return
			}
			values := extractReturnValuesByFunc(pkg, methodFunc, f.funcDeclTypes)
			if len(values) != 1 {
				return
			}
			code, ok := values[0].(int64)
			if !ok {
				return
			}

			var errCode string
			if methodFunc := findMethodByNamed(t, "Code"); methodFunc != nil {
				values := extractReturnValuesByFunc(pkg, methodFunc, f.funcDeclTypes)
				if len(values) > 0 {
					errCode, _ = values[0].(string)
				}
			}
			result[t.Obj().Pkg().Path()+"/"+t.Obj().Name()] = Error{
				PkgName: t.Obj().Pkg().Name(),
				PkgPath: t.Obj().Pkg().Path(),
				Name:    t.Obj().Name(),
				Code:    code,
				ErrCode: errCode,
			}
		}
		return
	})
	return
}

func (f *Finder) FindIfaceErrors(interfaces []*option.NamedType) (result map[string]map[string][]Error) {
	result = map[string]map[string][]Error{}
	visited := map[string]struct{}{}
	errors := f.FindErrors()

	for _, named := range interfaces {
		i := named.Type.(*option.IfaceType)
		for _, m := range i.Methods {
			id := named.Pkg.Path + "/" + named.Name.Value + "." + m.Name.Value
			if fns, ok := f.funcDeclIfaceTypes[id]; ok {
				for _, info := range fns {
					visited[id] = struct{}{}
					if _, ok := result[named.Name.Value]; !ok {
						result[named.Name.Value] = map[string][]Error{}
					}
					result[named.Name.Value][info.obj.Name()] = f.findIfaceErrorsRecursive(errors, visited, info.stmtList)
				}
			}
		}
	}
	return
}

func (f *Finder) findIfaceErrorsRecursive(errors map[string]Error, visited map[string]struct{}, stmts []ast.Stmt) (results []Error) {
	for _, stmt := range stmts {
		switch t := stmt.(type) {
		case *ast.ReturnStmt:
			for _, result := range t.Results {
				call, ok := result.(*ast.CallExpr)
				if !ok {
					if unary, ok := result.(*ast.UnaryExpr); ok {
						if cpl, ok := unary.X.(*ast.CompositeLit); ok {
							if sel, ok := cpl.Type.(*ast.SelectorExpr); ok {
								if obj := f.packages.ObjectOf(sel.Sel); obj != nil {
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
					if obj := f.packages.ObjectOf(sel.Sel); obj != nil {
						if named, ok := obj.Type().(*types.Named); ok {
							if _, ok := named.Obj().Type().Underlying().(*types.Interface); ok {
								id := named.Obj().Pkg().Path() + "/" + named.Obj().Name() + "." + selFun.Sel.Name
								if _, ok := visited[id]; ok {
									break
								}
								visited[id] = struct{}{}
								if infos, ok := f.funcDeclIfaceTypes[id]; ok {
									for _, info := range infos {
										results = append(results, f.findIfaceErrorsRecursive(errors, visited, info.stmtList)...)
									}
									break
								}
							}
						}
					}
				}
			}
		case *ast.IfStmt:
			results = append(results, f.findIfaceErrorsRecursive(errors, visited, t.Body.List)...)
		case *ast.SwitchStmt:
			results = append(results, f.findIfaceErrorsRecursive(errors, visited, t.Body.List)...)
		case *ast.BlockStmt:
			results = append(results, f.findIfaceErrorsRecursive(errors, visited, t.List)...)
		case *ast.CommClause:
			results = append(results, f.findIfaceErrorsRecursive(errors, visited, t.Body)...)
		case *ast.SelectStmt:
			results = append(results, f.findIfaceErrorsRecursive(errors, visited, t.Body.List)...)
		case *ast.ForStmt:
			results = append(results, f.findIfaceErrorsRecursive(errors, visited, t.Body.List)...)
		case *ast.RangeStmt:
			results = append(results, f.findIfaceErrorsRecursive(errors, visited, t.Body.List)...)
		}
	}
	return
}

func (f *Finder) fillFuncDeclTypes() {
	_ = f.packages.TraverseDecls(func(pkg *stdpackages.Package, file *ast.File, decl ast.Decl) (err error) {
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
				f.funcDeclTypes[pkg.PkgPath+id] = &typeInfo{
					obj:      obj,
					pkg:      pkg,
					recv:     recv,
					stmtList: t.Body.List,
				}
			}
		}
		return
	})
}

func (f *Finder) fillFuncIfaceDeclTypes() {
	_ = f.packages.TraverseDecls(func(pkg *stdpackages.Package, file *ast.File, decl ast.Decl) (err error) {
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
					for _, info := range f.funcDeclTypes {
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
										f.funcDeclIfaceTypes[id] = append(f.funcDeclIfaceTypes[id], info)
									}
								}
							}
						}
						imp := stdtypes.Implements(ptr.Underlying(), iface)
						if imp {
							id := pkg.PkgPath + "/" + named.Obj().Name() + "." + info.obj.Name()
							f.funcDeclIfaceTypes[id] = append(f.funcDeclIfaceTypes[id], info)
						}
					}
				}
			}
		}
		return
	})
}

func NewFinder(packages *packages.Packages, modulePath string) *Finder {
	f := &Finder{packages: packages, modulePath: modulePath, funcDeclTypes: map[string]*typeInfo{}, funcDeclIfaceTypes: map[string][]*typeInfo{}}
	f.fillFuncDeclTypes()
	f.fillFuncIfaceDeclTypes()
	return f
}
