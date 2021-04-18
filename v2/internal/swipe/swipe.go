package swipe

import (
	"fmt"
	"path/filepath"

	"github.com/swipe-io/swipe/v2/internal/interface/frame"

	"github.com/swipe-io/swipe/v2/internal/importer"

	"github.com/swipe-io/strcase"
)

type Importer interface {
	Import(name string, path string) string
}

type Generator interface {
	Generate(Importer) []byte
	OutputDir() string
	Filename() string
}

type GenerateResult struct {
	PkgPath    string
	OutputPath string
	Content    []byte
	Errs       []error
}

func Generate(cfg *Config) (result []GenerateResult) {
	importerMap := map[string]*importer.Importer{}
	for _, module := range cfg.Modules {
		for _, build := range module.Builds {
			// importer cache for package.
			i, ok := importerMap[build.Pkg.Path]
			if !ok {
				i = importer.NewImporter(build.Pkg)
				importerMap[build.Pkg.Path] = i
			}
			for id, config := range build.Option {
				generated := GenerateResult{
					PkgPath: build.Pkg.Path,
				}
				p, ok := registeredPlugins[id]
				if !ok {
					generated.Errs = append(generated.Errs, &warnError{Err: fmt.Errorf("plugin %q not found", id)})
					continue
				}
				if err := p.Configure(cfg, module, build, config); err != nil {
					generated.Errs = append(generated.Errs, err)
					continue
				}
				generators, errs := p.Generators()
				if len(errs) > 0 {
					for _, err := range errs {
						generated.Errs = append(generated.Errs, err)
						continue
					}
				}
				for _, g := range generators {
					outputDir := g.OutputDir()
					if outputDir == "" {
						outputDir = build.BasePath
					}
					filename := "_swipe_gen_" + strcase.ToSnake(p.ID()) + "_" + g.Filename()
					generated.OutputPath = filepath.Join(outputDir, filename)
					f := frame.NewFrame("v1.0.0", filename, i, build.Pkg)
					data, err := f.Frame(g.Generate(i))
					if err != nil {
						generated.Errs = append(generated.Errs, err)
						continue
					}
					generated.Content = data
				}
				result = append(result, generated)
			}
		}
	}
	return
}
