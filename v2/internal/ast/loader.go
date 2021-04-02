package ast

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	stdtypes "go/types"
	"strconv"
	stdstrings "strings"

	"github.com/swipe-io/swipe/v2/internal/domain/model"
	"github.com/swipe-io/swipe/v2/internal/graph"
	"github.com/swipe-io/swipe/v2/internal/types"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/types/typeutil"
)

type nodeInfo struct {
	node    *graph.Node
	objects []stdtypes.Object
}

type Loader struct {
	ctx           context.Context
	wd            string
	env           []string
	patterns      []string
	pkg           *packages.Package
	commentFuncs  map[string][]string
	commentFields map[string]map[string]string
	pkgs          []*packages.Package
	graphTypes    *graph.Graph
	enums         *typeutil.Map
}

func (l *Loader) FindPkgByID(path string) *packages.Package {
	for _, pkg := range l.pkgs {
		if pkg.PkgPath == path {
			return pkg
		}
	}
	return nil
}

func (l *Loader) CommentFields() map[string]map[string]string {
	return l.commentFields
}

func (l *Loader) CommentFuncs() map[string][]string {
	return l.commentFuncs
}

func (l *Loader) Pkg() *packages.Package {
	return l.pkg
}

func (l *Loader) Pkgs() []*packages.Package {
	return l.pkgs
}

func mode(tv stdtypes.TypeAndValue) string {
	switch {
	case tv.IsVoid():
		return "void"
	case tv.IsType():
		return "type"
	case tv.IsBuiltin():
		return "builtin"
	case tv.IsNil():
		return "nil"
	case tv.Assignable():
		if tv.Addressable() {
			return "var"
		}
		return "map"
	case tv.IsValue():
		return "value"
	default:
		return "unknown"
	}
}

func exprString(fset *token.FileSet, expr ast.Expr) string {
	var buf bytes.Buffer
	format.Node(&buf, fset, expr)
	return buf.String()
}

func (l *Loader) Interface(expr ast.Expr) {
	for _, pkg := range l.pkgs {
		if t := pkg.TypesInfo.TypeOf(expr); t != nil {

		}
	}
}

func (l *Loader) Patterns() []string {
	return l.patterns
}

func (l *Loader) Env() []string {
	return l.env
}

func (l *Loader) WorkDir() string {
	return l.wd
}

func (l *Loader) normalizeStmt(pkg *packages.Package, stmt ast.Stmt) interface{} {
	switch v := stmt.(type) {
	case *ast.SelectStmt:
		return l.normalizeBlockStmt(nil, v.Body)
	case *ast.RangeStmt:
		return l.normalizeBlockStmt(nil, v.Body)
	case *ast.ForStmt:
		return l.normalizeBlockStmt(nil, v.Body)
	case *ast.TypeSwitchStmt:
		return l.normalizeBlockStmt(nil, v.Body)
	case *ast.SwitchStmt:
		return l.normalizeBlockStmt(nil, v.Body)
	case *ast.IfStmt:
		return l.normalizeBlockStmt(nil, v.Body)
	case *ast.BlockStmt:
		return l.normalizeBlockStmt(nil, v)
	case *ast.ReturnStmt:
		for _, result := range v.Results {
			if callExpr, ok := result.(*ast.CallExpr); ok {

				fmt.Println(callExpr.Fun)
			}

			//v := pkg.TypesInfo.Types[result]
			//
			//fmt.Println(v)

		}
	}

	return nil
}

func (l *Loader) normalizeBlockStmt(pkg *packages.Package, blockStmt *ast.BlockStmt) interface{} {
	for _, stmt := range blockStmt.List {
		l.normalizeStmt(pkg, stmt)
	}
	return nil
}

func (l *Loader) run() (errs []error) {
	var (
		astNodes []nodeInfo
		err      error
	)

	l.commentFuncs = map[string][]string{}
	l.commentFields = map[string]map[string]string{}
	l.graphTypes = graph.NewGraph()
	l.enums = new(typeutil.Map)

	cfg := &packages.Config{
		Context: l.ctx,
		Mode: packages.NeedDeps |
			packages.NeedSyntax |
			packages.NeedTypesInfo |
			packages.NeedTypes |
			packages.NeedTypesSizes |
			packages.NeedImports |
			packages.NeedName |
			packages.NeedModule |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles,
		Dir:        l.wd,
		Env:        l.env,
		BuildFlags: []string{"-tags=swipe"},
	}
	escaped := make([]string, len(l.patterns))
	for i := range l.patterns {
		escaped[i] = "pattern=" + l.patterns[i]
	}
	l.pkgs, err = packages.Load(cfg, escaped...)
	if err != nil {
		return []error{err}
	}
	for _, p := range l.pkgs {
		for _, e := range p.Errors {
			errs = append(errs, e)
		}
	}

	for _, pkg := range l.pkgs {
		for _, syntax := range pkg.Syntax {
			for _, decl := range syntax.Decls {
				switch t := decl.(type) {
				case *ast.FuncDecl:
					l.normalizeBlockStmt(pkg, t.Body)
				}
			}
		}
	}

	if len(errs) > 0 {
		return errs
	}
	for _, pkg := range l.pkgs {
		if l.pkg == nil && stdstrings.Contains(l.wd, pkg.Module.Dir) {
			l.pkg = pkg
		}
		for _, syntax := range pkg.Syntax {
			for _, decl := range syntax.Decls {
				switch v := decl.(type) {
				case *ast.GenDecl:
					switch v.Tok {
					case token.TYPE:
						for _, spec := range v.Specs {
							sp := spec.(*ast.TypeSpec)
							obj := pkg.TypesInfo.ObjectOf(sp.Name)
							if obj != nil {
								l.graphTypes.Add(&graph.Node{Object: obj})
							}
						}
					case token.CONST:
						var (
							iotaValue int
							iotaIncr  int
							enums     []model.Enum
						)
						if len(v.Specs) < 1 {
							continue
						}
						vs, ok := v.Specs[0].(*ast.ValueSpec)
						if !ok {
							continue
						}
						if vs.Type == nil {
							continue
						}
						ti := pkg.TypesInfo.TypeOf(vs.Type.(*ast.Ident))
						if ti != nil {
							if named, ok := ti.(*stdtypes.Named); ok && !named.Obj().Exported() {
								continue
							}
							if b, ok := ti.Underlying().(*stdtypes.Basic); ok {
								switch b.Info() {
								case stdtypes.IsUnsigned | stdtypes.IsInteger, stdtypes.IsInteger:
									for _, spec := range v.Specs {
										vs := spec.(*ast.ValueSpec)
										if len(vs.Values) == 1 {
											iotaValue, iotaIncr = types.EvalInt(vs.Values[0])
										} else {
											iotaValue += iotaIncr
										}
										enums = append(enums, model.Enum{
											Name:  vs.Names[0].Name,
											Value: strconv.Itoa(iotaValue),
										})
									}
								case stdtypes.IsString:
									for _, spec := range v.Specs {
										vs := spec.(*ast.ValueSpec)
										if len(vs.Values) == 1 {
											lit := vs.Values[0].(*ast.BasicLit)
											s, _ := strconv.Unquote(lit.Value)
											enums = append(enums, model.Enum{
												Name:  vs.Names[0].Name,
												Value: s,
											})
										}
									}
								}
							}
							l.enums.Set(ti, enums)
						}
					}
				case *ast.FuncDecl:
					obj := pkg.TypesInfo.ObjectOf(v.Name)
					if obj != nil {
						n := &graph.Node{Object: obj}

						l.graphTypes.Add(n)

						values, objects := visitBlockStmt(pkg, v.Body)

						n.AddValue(values...)

						astNodes = append(astNodes, nodeInfo{
							node:    n,
							objects: objects,
						})
					}
				}
			}
		}
	}

	for _, ni := range astNodes {
		for _, obj := range ni.objects {
			if sig, ok := obj.Type().(*stdtypes.Signature); ok {
				if sig.Recv() != nil {
					if _, ok := sig.Recv().Type().Underlying().(*stdtypes.Interface); ok {
						l.graphTypes.Iterate(func(n *graph.Node) {
							l.graphTypes.Traverse(n, func(n *graph.Node) bool {
								if n.Object.Name() == obj.Name() && stdtypes.Identical(n.Object.Type(), obj.Type()) {
									l.graphTypes.AddEdge(ni.node, n)
								}
								return true
							})
						})
						continue
					}
				}
			}
			if nn := l.graphTypes.Node(obj); nn != nil {
				l.graphTypes.AddEdge(ni.node, nn)
			}
		}
	}
	types.Inspect(l.pkgs, func(p *packages.Package, n ast.Node) bool {
		if ts, ok := n.(*ast.TypeSpec); ok {
			obj := p.TypesInfo.ObjectOf(ts.Name)
			if st, ok := ts.Type.(*ast.StructType); ok {
				comments := map[string]string{}
				for _, field := range st.Fields.List {
					if field.Comment != nil {
						if len(field.Comment.List) > 0 {
							for _, name := range field.Names {
								comments[name.Name] = stdstrings.TrimLeft(field.Comment.List[0].Text, "/")
							}
						}
					}
				}
				if len(comments) > 0 {
					l.commentFields[obj.String()] = comments
				}
			}
		} else if spec, ok := n.(*ast.Field); ok {
			t := p.TypesInfo.TypeOf(spec.Type)
			if t != nil {
				var comments []string
				if spec.Doc != nil {
					for _, comment := range spec.Doc.List {
						comments = append(comments, stdstrings.TrimLeft(comment.Text, "/"))
					}
				}
				if spec.Comment != nil {
					for _, comment := range spec.Comment.List {
						comments = append(comments, stdstrings.TrimLeft(comment.Text, "/"))
					}
				}
				if len(comments) > 0 {
					for _, name := range spec.Names {
						if obj := p.TypesInfo.ObjectOf(name); obj != nil {
							l.commentFuncs[obj.String()] = comments
						}
					}
				}
			}
		} else if spec, ok := n.(*ast.FuncDecl); ok {
			obj := p.TypesInfo.ObjectOf(spec.Name)
			if obj != nil {
				var comments []string
				if spec.Doc != nil {
					for _, comment := range spec.Doc.List {
						comments = append(comments, stdstrings.TrimLeft(comment.Text, "/"))
					}
				}
				if len(comments) > 0 {
					l.commentFuncs[obj.String()] = comments
				}
			}
		}
		return true
	})
	return
}

func visitBlockStmts(p *packages.Package, stmts []ast.Stmt) (values []stdtypes.TypeAndValue, objects []stdtypes.Object) {
	for _, stmt := range stmts {
		otherValues, otherObjects := visitBlockStmt(p, stmt)

		values = append(values, otherValues...)
		objects = append(objects, otherObjects...)
	}
	return
}

func visitBlockStmt(p *packages.Package, stmt ast.Stmt) (values []stdtypes.TypeAndValue, objects []stdtypes.Object) {
	switch v := stmt.(type) {
	case *ast.SelectStmt:
		return visitBlockStmts(p, v.Body.List)
	case *ast.RangeStmt:
		return visitBlockStmts(p, v.Body.List)
	case *ast.ForStmt:
		return visitBlockStmts(p, v.Body.List)
	case *ast.TypeSwitchStmt:
		return visitBlockStmts(p, v.Body.List)
	case *ast.SwitchStmt:
		return visitBlockStmts(p, v.Body.List)
	case *ast.IfStmt:
		return visitBlockStmts(p, v.Body.List)
	case *ast.BlockStmt:
		return visitBlockStmts(p, v.List)
	case *ast.ReturnStmt:
		for _, result := range v.Results {
			switch vv := result.(type) {
			case *ast.FuncLit:
				otherValues, otherObjects := visitBlockStmts(p, vv.Body.List)
				values = append(values, otherValues...)
				objects = append(objects, otherObjects...)
			case *ast.CompositeLit:
				if named, ok := p.TypesInfo.TypeOf(vv.Type).(*stdtypes.Named); ok && named.Obj() != nil {
					objects = append(objects, named.Obj())
				}
			case *ast.UnaryExpr:
				if named, ok := p.TypesInfo.TypeOf(vv.X).(*stdtypes.Named); ok && named.Obj() != nil {
					objects = append(objects, named.Obj())
				}
				values = append(values, p.TypesInfo.Types[vv])
			case *ast.BasicLit:
				values = append(values, p.TypesInfo.Types[vv])
			case *ast.CallExpr:
				switch fv := vv.Fun.(type) {
				case *ast.SelectorExpr:
					if obj := p.TypesInfo.ObjectOf(fv.Sel); obj != nil {
						objects = append(objects, obj)
					}
				case *ast.Ident:
					if obj := p.TypesInfo.ObjectOf(fv); obj != nil {
						objects = append(objects, obj)
					}
				}
			}
		}
	}
	return
}

func NewLoader(wd string, env []string, patterns []string) (*Loader, []error) {
	l := &Loader{
		wd:       wd,
		env:      env,
		patterns: patterns,
	}
	errs := l.run()
	if len(errs) > 0 {
		return nil, errs
	}
	return l, nil
}
