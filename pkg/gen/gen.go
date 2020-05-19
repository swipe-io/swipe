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
	"golang.org/x/tools/go/types/typeutil"
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

		goSrc := w.Frame()

		fmtSrc, err := format.Source(goSrc)
		if err != nil {
			result[i].Errs = append(result[i].Errs, err)
		} else {
			goSrc = fmtSrc
		}
		result[i].Content = goSrc
		result[i].OutputPath = filepath.Join(outDir, "swipe_gen.go")
	}

	return result, nil
}

func (s *Swipe) findInjector(info *types.Info, fn *ast.FuncDecl) (*ast.CallExpr, error) {
	if fn.Body == nil {
		return nil, nil
	}
	numStatements := 0
	invalid := false

	for _, stmt := range fn.Body.List {
		switch stmt := stmt.(type) {
		case *ast.ExprStmt:
			numStatements++
			if numStatements > 1 {
				invalid = true
			}
			call, ok := stmt.X.(*ast.CallExpr)
			if !ok {
				continue
			}
			obj := value.QualifiedIdentObject(info, call.Fun)
			if obj.Name() != "Build" {
				continue
			}
			if obj == nil || obj.Pkg() == nil {
				continue
			}
			return call, nil
		case *ast.EmptyStmt:
		case *ast.ReturnStmt:
			if numStatements == 0 {
				return nil, nil
			}
		}
	}
	if invalid {
		return nil, errors.New("a call to swipe make handlers indicates that this function is an injector, but injectors must consist of only the swipe.HTTP call and an optional return")
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
		Context:    s.ctx,
		Mode:       loadAllSyntax,
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

func (s *Swipe) print(lpkg *packages.Package) {
	// if app.PrintJSON {
	// 	data, _ := json.MarshalIndent(lpkg, "", "\t")
	// 	os.Stdout.Write(data)
	// 	return
	// }
	// title
	var kind string
	// TODO(matloob): If IsTest is added back print "test command" or
	// "test package" for packages with IsTest == true.
	if lpkg.Name == "main" {
		kind += "command"
	} else {
		kind += "package"
	}
	fmt.Printf("Go %s %q:\n", kind, lpkg.ID) // unique ID
	fmt.Printf("\tpackage %s\n", lpkg.Name)

	// characterize type info
	if lpkg.Types == nil {
		fmt.Printf("\thas no exported type info\n")
	} else if !lpkg.Types.Complete() {
		fmt.Printf("\thas incomplete exported type info\n")
	} else if len(lpkg.Syntax) == 0 {
		fmt.Printf("\thas complete exported type info\n")
	} else {
		fmt.Printf("\thas complete exported type info and typed ASTs\n")
	}
	if lpkg.Types != nil && lpkg.IllTyped && len(lpkg.Errors) == 0 {
		fmt.Printf("\thas an error among its dependencies\n")
	}

	// source files
	for _, src := range lpkg.GoFiles {
		fmt.Printf("\tfile %s\n", src)
	}

	// imports
	var lines []string
	for importPath, imp := range lpkg.Imports {
		var line string
		if imp.ID == importPath {
			line = fmt.Sprintf("\timport %q", importPath)
		} else {
			line = fmt.Sprintf("\timport %q => %q", importPath, imp.ID)
		}
		lines = append(lines, line)
	}
	sort.Strings(lines)
	for _, line := range lines {
		fmt.Println(line)
	}

	// errors
	for _, err := range lpkg.Errors {
		fmt.Printf("\t%s\n", err)
	}

	// package members (TypeCheck or WholeProgram mode)
	if lpkg.Types != nil {
		qual := types.RelativeTo(lpkg.Types)
		scope := lpkg.Types.Scope()
		for _, name := range scope.Names() {
			obj := scope.Lookup(name)
			// if !obj.Exported() && !app.Private {
			// continue // skip unexported names
			// }

			fmt.Printf("\t%s\n", types.ObjectString(obj, qual))
			if _, ok := obj.(*types.TypeName); ok {
				for _, meth := range typeutil.IntuitiveMethodSet(obj.Type(), nil) {
					// if !meth.Obj().Exported() && !app.Private {
					// continue // skip unexported names
					// }
					fmt.Printf("\t%s\n", types.SelectionString(meth, qual))
				}
			}
		}
	}

	fmt.Println()
}

func NewSwipe(ctx context.Context, wd string, env []string, patterns []string) *Swipe {
	return &Swipe{ctx: ctx, wd: wd, env: env, patterns: patterns}
}
