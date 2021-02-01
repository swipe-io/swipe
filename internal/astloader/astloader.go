package astloader

import (
	"context"
	"go/ast"
	"go/build"
	"go/token"
	stdtypes "go/types"
	"path/filepath"
	"strconv"
	"strings"
	stdstrings "strings"

	"golang.org/x/mod/modfile"

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

type Data struct {
	WorkDir       string
	PkgPath       string
	CommentFuncs  map[string][]string
	CommentFields map[string]map[string]string
	Pkgs          []*packages.Package
	GraphTypes    *graph.Graph
	Enums         *typeutil.Map
}

type Loader struct {
	ctx      context.Context
	wd       string
	env      []string
	patterns []string
	mod      *modfile.File
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

func (l *Loader) Process() (data *Data, errs []error) {
	var (
		err error
	)

	var pkgPath string

	if l.mod == nil {
		pkgPath = strings.Split(l.wd, filepath.Join(build.Default.GOPATH, "src")+"/")[1]
	} else {
		pkgPath = l.mod.Module.Mod.Path
	}

	data = &Data{
		WorkDir:       l.wd,
		PkgPath:       pkgPath,
		CommentFuncs:  map[string][]string{},
		CommentFields: map[string]map[string]string{},
		GraphTypes:    graph.NewGraph(),
		Enums:         new(typeutil.Map),
	}
	cfg := &packages.Config{
		Context:    l.ctx,
		Mode:       packages.NeedDeps | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedTypes | packages.NeedTypesSizes | packages.NeedImports | packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles,
		Dir:        l.wd,
		Env:        l.env,
		BuildFlags: []string{"-tags=swipe"},
	}
	escaped := make([]string, len(l.patterns))
	for i := range l.patterns {
		escaped[i] = "pattern=" + l.patterns[i]
	}
	data.Pkgs, err = packages.Load(cfg, escaped...)
	if err != nil {
		return data, []error{err}
	}

	var astNodes []nodeInfo

	for _, pkg := range data.Pkgs {
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
								data.GraphTypes.Add(&graph.Node{Object: obj})
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
							data.Enums.Set(ti, enums)
						}
					}
				case *ast.FuncDecl:
					obj := pkg.TypesInfo.ObjectOf(v.Name)
					if obj != nil {
						n := &graph.Node{Object: obj}

						data.GraphTypes.Add(n)

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
						data.GraphTypes.Iterate(func(n *graph.Node) {
							data.GraphTypes.Traverse(n, func(n *graph.Node) bool {
								if n.Object.Name() == obj.Name() && stdtypes.Identical(n.Object.Type(), obj.Type()) {
									data.GraphTypes.AddEdge(ni.node, n)
								}
								return true
							})
						})
						continue
					}
				}
			}
			if nn := data.GraphTypes.Node(obj); nn != nil {
				data.GraphTypes.AddEdge(ni.node, nn)
			}
		}
	}
	types.Inspect(data.Pkgs, func(p *packages.Package, n ast.Node) bool {
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
					data.CommentFields[obj.String()] = comments
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
						obj := p.TypesInfo.ObjectOf(name)
						data.CommentFuncs[obj.String()] = comments
					}
				}
			}
		}
		return true
	})
	for _, p := range data.Pkgs {
		for _, e := range p.Errors {
			errs = append(errs, e)
		}
	}
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
			case *ast.StructType:

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

func NewLoader(wd string, env []string, patterns []string, mod *modfile.File) *Loader {
	return &Loader{
		wd:       wd,
		env:      env,
		patterns: patterns,
		mod:      mod,
	}
}
