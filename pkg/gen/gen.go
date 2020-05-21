package gen

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"path/filepath"
	"sort"

	generrors "github.com/swipe-io/swipe/pkg/errors"
	"github.com/swipe-io/swipe/pkg/parser"
	"github.com/swipe-io/swipe/pkg/value"
	"github.com/swipe-io/swipe/pkg/writer"

	"golang.org/x/tools/go/packages"
)

const loadAllSyntax = packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedImports | packages.NeedDeps | packages.NeedExportsFile | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedTypesSizes

type Generator interface {
	Write(opt *parser.Option) error
}

type Result struct {
	PkgPath    string
	OutputPath string
	Content    []byte
	Errs       []error
}

type Swipe struct {
	ctx      context.Context
	wd       string
	env      []string
	patterns []string
}

func (s *Swipe) Generate() ([]Result, []error) {
	pkgs, allPkgs, errs := s.loadPackages()
	if len(errs) > 0 {
		return nil, errs
	}

	result := make([]Result, len(pkgs))

	for i, pkg := range pkgs {
		ec := new(generrors.ErrorCollector)

		outDir, err := s.detectOutputDir(pkg.GoFiles)
		if err != nil {
			ec.Add(err)
			continue
		}

		w := writer.NewWriter(pkg, allPkgs, outDir)
		p := parser.NewParser(pkg)

		generatorOpts := make(map[string][]*parser.Option)

		for _, f := range pkg.Syntax {
			for _, decl := range f.Decls {
				switch decl := decl.(type) {
				case *ast.FuncDecl:
					call, err := s.findInjector(pkg.TypesInfo, decl)
					if err != nil {
						ec.Add(err)
						continue
					}
					if call != nil {
						opt, err := p.Parse(call.Args[0])
						if err != nil {
							ec.Add(err)
							continue
						}
						generatorOpts[opt.Name] = append(generatorOpts[opt.Name], opt)
						continue
					}
				case *ast.GenDecl:
					if decl.Tok == token.IMPORT {
						continue
					}
				}
				w.Write("// %s:\n\n", filepath.Base(pkg.Fset.File(f.Pos()).Name()))
				w.WriteAST(decl)
				w.Write("\n\n")
			}
		}

		if len(generatorOpts) > 0 {
			for name, opts := range generatorOpts {
				if f, ok := factory[name]; ok {
					gw := f(w)
					for _, opt := range opts {
						if err := gw.Write(opt); err != nil {
							ec.Add(err)
						}
					}
				}
			}

			if len(ec.Errors()) > 0 {
				result[i].Errs = ec.Errors()
			}

			goSrc := w.Frame(false)

			fmtSrc, err := format.Source(goSrc)
			if err != nil {
				result[i].Errs = append(result[i].Errs, err)
			} else {
				goSrc = fmtSrc
			}
			result[i].Content = goSrc
			result[i].OutputPath = filepath.Join(outDir, "swipe_gen.go")
		}
	}

	return result, nil
}

func (s *Swipe) findInjector(info *types.Info, fn *ast.FuncDecl) (*ast.CallExpr, error) {
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

func (s *Swipe) detectOutputDir(paths []string) (string, error) {
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

func (s *Swipe) loadPackages() (pkgs []*packages.Package, allPkgs []*packages.Package, errs []error) {
	cfg := &packages.Config{
		Context: s.ctx,
		// Mode:       loadAllSyntax,
		Mode:       packages.LoadSyntax,
		Dir:        s.wd,
		Env:        s.env,
		BuildFlags: []string{"-tags=swipe"},
	}
	escaped := make([]string, len(s.patterns))
	for i := range s.patterns {
		escaped[i] = "pattern=" + s.patterns[i]
	}
	pkgs, err := packages.Load(cfg, escaped...)
	if err != nil {
		return nil, nil, []error{err}
	}

	seen := make(map[*packages.Package]bool)

	var visit func(pkg *packages.Package)
	visit = func(pkg *packages.Package) {
		if !seen[pkg] {
			seen[pkg] = true

			var importPaths []string
			for path := range pkg.Imports {
				importPaths = append(importPaths, path)
			}
			sort.Strings(importPaths)
			for _, path := range importPaths {
				visit(pkg.Imports[path])
			}

			allPkgs = append(allPkgs, pkg)
		}
	}
	for _, pkg := range pkgs {
		visit(pkg)
	}

	for _, p := range pkgs {
		for _, e := range p.Errors {
			errs = append(errs, e)
		}
	}
	if len(errs) > 0 {
		return nil, nil, errs
	}
	return pkgs, allPkgs, nil
}

func NewSwipe(ctx context.Context, wd string, env []string, patterns []string) *Swipe {
	return &Swipe{ctx: ctx, wd: wd, env: env, patterns: patterns}
}
