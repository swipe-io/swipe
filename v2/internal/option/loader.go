package option

import (
	"go/ast"
	stdtypes "go/types"
	"sync"

	"github.com/swipe-io/swipe/v2/internal/astloader"
	"github.com/swipe-io/swipe/v2/internal/value"
	"golang.org/x/tools/go/packages"
)

type ResultOption struct {
	Pkg    *packages.Package
	Option *Option
}

type Result struct {
	Data            *astloader.Data
	Options         []*ResultOption
	ExternalOptions []*ResultOption
}

type Loader struct {
	astLoader *astloader.Loader
	pkgName   string
}

func (l *Loader) declProcess(pkg *packages.Package, decl ast.Decl) (*ResultOption, error) {
	switch decl := decl.(type) {
	case *ast.FuncDecl:
		call, err := l.findInjector(pkg.TypesInfo, decl)
		if err != nil {
			return nil, err
		}
		if call != nil {
			opt, err := NewParser(pkg).Parse(call.Args[0])
			if err != nil {
				return nil, err
			}
			return &ResultOption{
				Pkg:    pkg,
				Option: opt,
			}, nil
		}
	}
	return nil, nil
}

func (l *Loader) loadPackages(pkgs []*packages.Package) (<-chan *ResultOption, <-chan error) {
	outCh := make(chan *ResultOption)
	errCh := make(chan error)
	go func() {
		var wg sync.WaitGroup
		for _, pkg := range pkgs {
			for _, f := range pkg.Syntax {
				for _, decl := range f.Decls {
					wg.Add(1)
					go func(pkg *packages.Package, decl ast.Decl) {
						defer wg.Done()
						result, err := l.declProcess(pkg, decl)
						if err != nil {
							errCh <- err
							return
						}
						if result != nil {
							outCh <- result
						}
					}(pkg, decl)
				}
			}
		}
		wg.Wait()
		close(errCh)
		close(outCh)
	}()
	return outCh, errCh
}

func (l *Loader) Load() (result *Result, errs []error) {
	result = &Result{}

	data, errs := l.astLoader.Process()
	if len(errs) > 0 {
		return nil, errs
	}
	result.Data = data

	optionsCh, errCh := l.loadPackages(data.Pkgs)

	go func() {
		for e := range errCh {
			errs = append(errs, e)
		}
	}()

	for option := range optionsCh {
		if data.Module.GoMod != option.Pkg.Module.GoMod {
			result.ExternalOptions = append(result.ExternalOptions, option)
			continue
		}
		result.Options = append(result.Options, option)
	}
	return
}

func (l *Loader) findInjector(info *stdtypes.Info, fn *ast.FuncDecl) (*ast.CallExpr, error) {
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
			if obj.Pkg().Name() == l.pkgName && obj.Name() != "Build" {
				continue
			}
			return call, nil
		case *ast.EmptyStmt:

			return nil, nil
		}
	}
	return nil, nil
}

func NewLoader(astLoader *astloader.Loader, pkgName string) *Loader {
	return &Loader{astLoader: astLoader, pkgName: pkgName}
}
