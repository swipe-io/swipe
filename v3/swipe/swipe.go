package swipe

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/swipe-io/swipe/v3/frame"

	"github.com/swipe-io/strcase"
	"github.com/swipe-io/swipe/v3/internal/importer"
	"github.com/swipe-io/swipe/v3/option"
)

const Version = "v3.0.19"

type ContextKey string

const (
	ImporterKey ContextKey = "importer"
)

type Importer interface {
	Import(name string, path string) string
}

type AstFinder interface {
	FindImplIface(ifaceType option.IfaceType)
}

type GenerateResult struct {
	PkgPath    string
	OutputPath string
	Content    []byte
	Errs       []error
}

func Generate(cfg *Config) (result []GenerateResult, errs []error) {
	result = make([]GenerateResult, 0, 100)
	importerMap := map[string]*importer.Importer{}
	for _, module := range cfg.Modules {
		if module.External {
			continue
		}
		for _, build := range module.Injects {
			for id, options := range build.Option {
				p, ok := registeredPlugins[id]
				if !ok {
					errs = append(errs, &warnError{Err: fmt.Errorf("plugin %q not found", id)})
					continue
				}
				cfgErrs := p.Configure(cfg, module, options.(map[string]interface{}))
				if len(cfgErrs) > 0 {
					errs = append(errs, cfgErrs...)
					continue
				}
				generators, genErrs := p.Generators()
				if len(genErrs) > 0 {
					errs = append(errs, genErrs...)
					continue
				}
				generatorResult := make([]GenerateResult, len(generators))
				for i, g := range generators {
					generatorResult[i].PkgPath = build.Pkg.Path

					outputDir := g.OutputDir()
					if outputDir == "" {
						outputDir = build.BasePath
					} else {
						path, err := filepath.Abs(filepath.Join(cfg.WorkDir, outputDir))
						if err != nil {
							generatorResult[i].Errs = append(generatorResult[i].Errs, err)
							continue
						}
						outputDir = path
					}
					filename := "swipe_gen_" + strcase.ToSnake(p.ID()) + "_" + g.Filename()

					importerKey := build.Pkg.Path + filename

					// importer cache for package.
					importerService, ok := importerMap[importerKey]
					if !ok {
						importerService = importer.NewImporter(build.Pkg)
						importerMap[filename] = importerService
					}

					generatorResult[i].OutputPath = filepath.Join(outputDir, filename)

					pkgName := build.Pkg.Name
					if gp, ok := g.(GeneratorPackage); ok && gp.Package() != "" {
						pkgName = gp.Package()
					}

					f := frame.NewFrame(Version, filename, importerService, pkgName)

					ctx := context.WithValue(context.TODO(), ImporterKey, importerService)

					data, err := f.Frame(g.Generate(ctx))
					if err != nil {
						generatorResult[i].Errs = append(generatorResult[i].Errs, err)
						continue
					}
					generatorResult[i].Content = data
				}
				result = append(result, generatorResult...)
			}
		}
	}
	return
}
