package gen

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	stdtypes "go/types"
	"os"
	"path/filepath"
	stdstrings "strings"

	"github.com/swipe-io/swipe/pkg/graph"

	"github.com/swipe-io/swipe/pkg/domain/model"
	"github.com/swipe-io/swipe/pkg/file"
	"github.com/swipe-io/swipe/pkg/importer"
	"github.com/swipe-io/swipe/pkg/parser"
	"github.com/swipe-io/swipe/pkg/registry"
	"github.com/swipe-io/swipe/pkg/types"
	"github.com/swipe-io/swipe/pkg/usecase/processor"
	"github.com/swipe-io/swipe/pkg/value"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/types/typeutil"
)

type importerer interface {
	SetImporter(*importer.Importer)
}

type Result struct {
	PkgPath    string
	OutputPath string
	Content    []byte
	Errs       []error
}

type Swipe struct {
	ctx         context.Context
	version     string
	wd          string
	env         []string
	patterns    []string
	commentMaps *typeutil.Map
	pkgs        []*packages.Package
	graphTypes  *graph.Graph
}

func (s *Swipe) Generate() ([]Result, []error) {
	r := registry.NewRegistry()

	errs := s.loadPackages()
	if len(errs) > 0 {
		return nil, errs
	}

	var result []Result
	files := make(map[string]*file.File)
	basePaths := map[string]struct{}{}

	for _, pkg := range s.pkgs {
		importerFactory := processor.NewImporterFactory(pkg)

		basePath, err := s.detectBasePath(pkg.GoFiles)
		if err != nil {
			return nil, []error{err}
		}

		basePaths[basePath] = struct{}{}

		for _, f := range pkg.Syntax {
			for _, decl := range f.Decls {
				switch decl := decl.(type) {
				case *ast.FuncDecl:
					call, err := s.findInjector(pkg.TypesInfo, decl)
					if err != nil {
						return nil, []error{err}
					}
					if call != nil {
						opt, err := parser.NewParser(pkg).Parse(call.Args[0])
						if err != nil {
							return nil, []error{err}
						}

						info := model.GenerateInfo{
							Pkg:        pkg,
							Pkgs:       s.pkgs,
							BasePath:   basePath,
							Version:    s.version,
							CommentMap: s.commentMaps,
							GraphTypes: s.graphTypes,
						}
						option := r.Option(opt.Name, info)
						if option == nil {
							return nil, []error{errors.New("unknown option:" + opt.Name)}
						}
						o, err := option.Parse(opt)
						if err != nil {
							return nil, []error{err}
						}
						p, err := r.Processor(opt.Name, info)
						if err != nil {
							return nil, []error{err}
						}
						if !p.SetOption(o) {
							return nil, []error{errors.New("option not suitable for processor: " + opt.Name)}
						}
						for _, g := range p.Generators() {
							if err := g.Prepare(s.ctx); err != nil {
								return nil, []error{err}
							}
							outputDir := g.OutputDir()
							if outputDir == "" {
								outputDir = basePath
							}
							filename := g.Filename()
							if filename == "" {
								filename = "swipe_gen.go"
							}

							fileKey := outputDir + filename

							i := importerFactory.Instance(fileKey)
							if is, ok := g.(importerer); ok {
								is.SetImporter(i)
							}

							if err := g.Process(s.ctx); err != nil {
								return nil, []error{err}
							}

							f, ok := files[fileKey]
							if !ok {
								f = &file.File{
									PkgName:   pkg.Name,
									PkgPath:   pkg.PkgPath,
									OutputDir: outputDir,
									Filename:  filename,
									Version:   s.version,
									Importer:  i,
								}
								files[fileKey] = f
							}

							b := g.Bytes()
							if len(b) > 0 {
								_, _ = f.Write(b)
							}
						}

						continue
					}
				case *ast.GenDecl:
					if decl.Tok == token.IMPORT {
						continue
					}
				}
			}
		}
	}

	for path := range basePaths {
		files, err := filepath.Glob(filepath.Join(path, "*_gen.*"))
		if err != nil {
			panic(err)
		}
		for _, f := range files {
			if err := os.Remove(f); err != nil {
				return nil, []error{err}
			}
		}
	}

	for _, f := range files {
		if len(f.Bytes()) > 0 {
			goSrc, err := f.Frame()
			if err != nil {
				f.Errs = append(f.Errs, err)
			}
			result = append(result, Result{
				PkgPath:    f.PkgPath,
				OutputPath: filepath.Join(f.OutputDir, f.Filename),
				Content:    goSrc,
				Errs:       f.Errs,
			})
		}
	}

	return result, nil
}

func (s *Swipe) findInjector(info *stdtypes.Info, fn *ast.FuncDecl) (*ast.CallExpr, error) {
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
			obj := value.QualifiedIdentObject(info, call.Fun)
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

func (s *Swipe) detectBasePath(paths []string) (string, error) {
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

type nodeInfo struct {
	node    *graph.Node
	objects []stdtypes.Object
}

func (s *Swipe) loadPackages() []error {
	cfg := &packages.Config{
		Context:    s.ctx,
		Mode:       packages.NeedDeps | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedTypes | packages.NeedTypesSizes | packages.NeedImports | packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles,
		Dir:        s.wd,
		Env:        s.env,
		BuildFlags: []string{"-tags=swipe"},
	}

	var (
		errs []error
		err  error
	)

	escaped := make([]string, len(s.patterns))
	for i := range s.patterns {
		escaped[i] = "pattern=" + s.patterns[i]
	}
	s.pkgs, err = packages.Load(cfg, escaped...)
	if err != nil {
		return []error{err}
	}

	astNodes := []nodeInfo{}

	for _, pkg := range s.pkgs {
		for _, syntax := range pkg.Syntax {
			for _, decl := range syntax.Decls {
				switch v := decl.(type) {
				case *ast.GenDecl:
					for _, spec := range v.Specs {
						if typeSpec, ok := spec.(*ast.TypeSpec); ok {
							obj := pkg.TypesInfo.ObjectOf(typeSpec.Name)
							s.graphTypes.Add(&graph.Node{Object: obj})
						}
					}
				case *ast.FuncDecl:
					obj := pkg.TypesInfo.ObjectOf(v.Name)

					n := &graph.Node{Object: obj}

					s.graphTypes.Add(n)

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
	for _, ni := range astNodes {
		for _, obj := range ni.objects {
			if sig, ok := obj.Type().(*stdtypes.Signature); ok {
				if sig.Recv() != nil {
					if _, ok := sig.Recv().Type().Underlying().(*stdtypes.Interface); ok {
						s.graphTypes.Iterate(func(n *graph.Node) {
							s.graphTypes.Traverse(n, func(n *graph.Node) bool {
								if n.Object.Name() == obj.Name() && stdtypes.Identical(n.Object.Type(), obj.Type()) {
									s.graphTypes.AddEdge(ni.node, n)
								}
								return true
							})
						})
						continue
					}
				}
			}
			if nn := s.graphTypes.Node(obj); nn != nil {
				s.graphTypes.AddEdge(ni.node, nn)
			}
		}
	}

	types.Inspect(s.pkgs, func(p *packages.Package, n ast.Node) bool {
		if spec, ok := n.(*ast.Field); ok {
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
					s.commentMaps.Set(t, comments)
				}
			}
		}
		return true
	})

	for _, p := range s.pkgs {
		for _, e := range p.Errors {
			errs = append(errs, e)
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

func visitBlockStmt(p *packages.Package, block *ast.BlockStmt) (values []stdtypes.TypeAndValue, objects []stdtypes.Object) {
	for _, stmt := range block.List {
		switch v := stmt.(type) {
		case *ast.SelectStmt:
			otherValues, otherObjects := visitBlockStmt(p, v.Body)
			values = append(values, otherValues...)
			objects = append(objects, otherObjects...)

		case *ast.RangeStmt:
			otherValues, otherObjects := visitBlockStmt(p, v.Body)
			values = append(values, otherValues...)
			objects = append(objects, otherObjects...)
		case *ast.ForStmt:
			otherValues, otherObjects := visitBlockStmt(p, v.Body)
			values = append(values, otherValues...)
			objects = append(objects, otherObjects...)
		case *ast.TypeSwitchStmt:
			otherValues, otherObjects := visitBlockStmt(p, v.Body)
			values = append(values, otherValues...)
			objects = append(objects, otherObjects...)
		case *ast.SwitchStmt:
			otherValues, otherObjects := visitBlockStmt(p, v.Body)
			values = append(values, otherValues...)
			objects = append(objects, otherObjects...)
		case *ast.IfStmt:
			otherValues, otherObjects := visitBlockStmt(p, v.Body)
			values = append(values, otherValues...)
			objects = append(objects, otherObjects...)
		case *ast.BlockStmt:
			otherValues, otherObjects := visitBlockStmt(p, v)
			values = append(values, otherValues...)
			objects = append(objects, otherObjects...)
		case *ast.ReturnStmt:
			for _, result := range v.Results {
				switch vv := result.(type) {
				case *ast.FuncLit:
					otherValues, otherObjects := visitBlockStmt(p, vv.Body)
					values = append(values, otherValues...)
					objects = append(objects, otherObjects...)
				case *ast.CompositeLit:
					if named, ok := p.TypesInfo.TypeOf(vv.Type).(*stdtypes.Named); ok {
						objects = append(objects, named.Obj())
					}
				case *ast.UnaryExpr, *ast.BasicLit:
					values = append(values, p.TypesInfo.Types[vv])
				case *ast.CallExpr:
					switch fv := vv.Fun.(type) {
					case *ast.SelectorExpr:
						objects = append(objects, p.TypesInfo.ObjectOf(fv.Sel))
					case *ast.Ident:
						objects = append(objects, p.TypesInfo.ObjectOf(fv))
					}
				}
			}
		}
	}
	return
}

func NewSwipe(ctx context.Context, version, wd string, env []string, patterns []string) *Swipe {
	return &Swipe{
		ctx:         ctx,
		version:     version,
		wd:          wd,
		env:         env,
		patterns:    patterns,
		commentMaps: new(typeutil.Map),
		graphTypes:  graph.NewGraph(),
	}
}
