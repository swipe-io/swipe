package swipe

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/swipe-io/strcase"
	"github.com/swipe-io/swipe/v2/internal/frame"
	"github.com/swipe-io/swipe/v2/internal/importer"
	"github.com/swipe-io/swipe/v2/option"
)

type ContextKey string

const (
	ImporterKey ContextKey = "importer"
)

type Importer interface {
	Import(name string, path string) string
	TypeString(v interface{}) string
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
		for _, build := range module.Builds {
			for id, options := range build.Option {
				p, ok := registeredPlugins[id]
				if !ok {
					errs = append(errs, &warnError{Err: fmt.Errorf("plugin %q not found", id)})
					continue
				}
				cfgErrs := p.Configure(cfg, module, build, options.(map[string]interface{}))
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
					}
					filename := "swipe_gen_" + strcase.ToSnake(p.ID()) + "_" + g.Filename()

					// importer cache for package.
					importerService, ok := importerMap[filename]
					if !ok {
						importerService = importer.NewImporter(build.Pkg)
						importerMap[filename] = importerService
					}

					generatorResult[i].OutputPath = filepath.Join(outputDir, filename)
					f := frame.NewFrame("v1.0.0", filename, importerService, build.Pkg)

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
