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

	"github.com/swipe-io/swipe/pkg/git"

	"github.com/swipe-io/swipe/pkg/astloader"
	"github.com/swipe-io/swipe/pkg/domain/model"
	"github.com/swipe-io/swipe/pkg/file"
	"github.com/swipe-io/swipe/pkg/importer"
	"github.com/swipe-io/swipe/pkg/parser"
	"github.com/swipe-io/swipe/pkg/registry"
	"github.com/swipe-io/swipe/pkg/usecase/processor"
	"github.com/swipe-io/swipe/pkg/value"
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
	ctx     context.Context
	version string
	loader  *astloader.Loader
}

func (s *Swipe) Generate() ([]Result, []error) {
	astData, errs := s.loader.Process()
	if len(errs) > 0 {
		return nil, errs
	}

	g := git.NewGIT()
	r := registry.NewRegistry()

	gitTags, _ := g.GetTags()

	var result []Result
	files := make(map[string]*file.File)
	basePaths := map[string]struct{}{}

	for _, pkg := range astData.Pkgs {
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
							Pkgs:       astData.Pkgs,
							BasePath:   basePath,
							Version:    s.version,
							CommentMap: astData.CommentMaps,
							GraphTypes: astData.GraphTypes,
							Enums:      astData.Enums,
							GitTags:    gitTags,
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

func NewSwipe(ctx context.Context, version string, loader *astloader.Loader) *Swipe {
	return &Swipe{
		ctx:     ctx,
		version: version,
		loader:  loader,
	}
}
