package executor

import (
	"context"
	"path/filepath"
	"sync"

	"github.com/swipe-io/swipe/v2/internal/types"

	"golang.org/x/tools/go/packages"

	"github.com/swipe-io/swipe/v2/internal/usecase/generator"

	"github.com/swipe-io/swipe/v2/internal/usecase/processor"

	"github.com/swipe-io/swipe/v2/internal/importer"
	"github.com/swipe-io/swipe/v2/internal/option"
	"github.com/swipe-io/swipe/v2/internal/usecase/executor"
	"github.com/swipe-io/swipe/v2/internal/usecase/factory"
	"github.com/swipe-io/swipe/v2/internal/usecase/frame"
	"github.com/swipe-io/swipe/v2/internal/usecase/registry"
)

type importerer interface {
	SetImporter(*importer.Importer)
}

type generationExecutor struct {
	r  registry.ProcessorRegistry
	i  factory.ImporterFactory
	ff frame.Factory
	l  *option.Loader
}

func (e *generationExecutor) processGenerate(pkg *packages.Package, generators []generator.Generator) <-chan executor.GenerateResult {
	outCh := make(chan executor.GenerateResult)

	go func() {
		var wg sync.WaitGroup

		for _, g := range generators {
			wg.Add(1)

			go func(g generator.Generator) {
				defer wg.Done()

				generated := executor.GenerateResult{}

				defer func() {
					outCh <- generated
				}()

				if err := g.Prepare(context.TODO()); err != nil {
					generated.Errs = append(generated.Errs, err)
					return
				}

				outputDir := g.OutputDir()
				if outputDir == "" {
					basePath, err := types.DetectBasePath(pkg)
					if err != nil {
						generated.Errs = append(generated.Errs, err)
						return
					}
					outputDir = basePath
				}

				generated.PkgPath = pkg.PkgPath
				generated.OutputPath = filepath.Join(outputDir, g.Filename())

				newImporter := e.i.NewImporter(generated.OutputPath, pkg)
				if g, ok := g.(importerer); ok {
					g.SetImporter(newImporter)
				}

				if err := g.Process(context.TODO()); err != nil {
					generated.Errs = append(generated.Errs, err)
					return
				}
				fr := e.ff.NewFrame(generated.OutputPath, newImporter, pkg)
				content, err := fr.Frame(g.Bytes())
				if err != nil {
					generated.Content = g.Bytes()
					generated.Errs = append(generated.Errs, err)
					return
				}
				generated.Content = content
			}(g)
		}
		wg.Wait()
		close(outCh)
	}()
	return outCh
}

func (e *generationExecutor) Execute() (results []executor.GenerateResult, errs []error) {
	opr, errs := e.l.Load()
	if len(errs) > 0 {
		return nil, errs
	}
	var processors []processor.Processor
	for _, o := range opr.Options {
		p, err := e.r.NewProcessor(o, opr.ExternalOptions, opr.Data)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		processors = append(processors, p)
	}
	if len(errs) > 0 {
		return nil, errs
	}
	var wg sync.WaitGroup
	for _, p := range processors {
		wg.Add(1)
		go func(p processor.Processor) {
			defer wg.Done()
			outCh := e.processGenerate(p.Pkg(), p.Generators())
			for generateResult := range outCh {
				results = append(results, generateResult)
			}
		}(p)
	}
	wg.Wait()
	return
}

func NewGenerationExecutor(
	r registry.ProcessorRegistry,
	i factory.ImporterFactory,
	ff frame.Factory,
	l *option.Loader,
) executor.GenerationExecutor {
	return &generationExecutor{r: r, i: i, ff: ff, l: l}

}
