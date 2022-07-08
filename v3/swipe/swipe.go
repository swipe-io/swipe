package swipe

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/swipe-io/strcase"
	"github.com/swipe-io/swipe/v3/internal/importer"
	"github.com/swipe-io/swipe/v3/option"
)

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
	PkgName    string
	PkgPath    string
	OutputPath string
	Imports    []string
	Content    []byte
	Errs       []error
}

func Generate(cfg *Config, prefix string) (result map[string]*GenerateResult, errs []error) {
	result = make(map[string]*GenerateResult, 512)
	importerCache := map[string]*importer.Importer{}

	for _, module := range cfg.Modules {
		if module.External {
			continue
		}
		for _, build := range module.Injects {
			for id, options := range build.Option {
				iface, ok := registeredPlugins.Load(id)
				if !ok {
					errs = append(errs, &warnError{Err: fmt.Errorf("plugin %q not found", id)})
					continue
				}

				cb := iface.(func() Plugin)
				p := cb()

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

				for _, g := range generators {
					filename := prefix + strcase.ToSnake(p.ID()) + "_" + g.Filename()

					outputPath := g.OutputPath()
					if outputPath == "" {
						outputPath = filepath.Join(build.BasePath, filename)
					} else {
						path, err := filepath.Abs(filepath.Join(cfg.WorkDir, outputPath))
						if err != nil {
							errs = append(errs, err)
							continue
						}
						outputPath = filepath.Join(path, filename)
					}

					generateResult, ok := result[outputPath]
					if !ok {
						generateResult = &GenerateResult{
							PkgPath:    build.Pkg.Path,
							OutputPath: outputPath,
						}
						result[outputPath] = generateResult
					}

					// importer cache for package.
					importerService, ok := importerCache[outputPath]
					if !ok {
						importerService = importer.NewImporter(build.Pkg)
						importerCache[outputPath] = importerService
					}

					pkgName := build.Pkg.Name
					if gp, ok := g.(GeneratorPackage); ok && gp.Package() != "" {
						pkgName = gp.Package()
					}

					generateResult.PkgName = pkgName

					ctx := context.WithValue(context.TODO(), ImporterKey, importerService)

					generateResult.Content = append(generateResult.Content, g.Generate(ctx)...)
				}
			}
		}
	}

	for _, generateResult := range result {
		importerService := importerCache[generateResult.OutputPath]
		if importerService.HasImports() {
			generateResult.Imports = importerService.SortedImports()
		}
	}

	return
}
