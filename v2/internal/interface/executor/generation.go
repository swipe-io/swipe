package executor

import (
	"context"

	"fmt"
	"path/filepath"

	"github.com/swipe-io/swipe/v2/internal/errors"
	"github.com/swipe-io/swipe/v2/internal/importer"
	"github.com/swipe-io/swipe/v2/internal/types"
	"github.com/swipe-io/swipe/v2/internal/usecase/executor"
	"github.com/swipe-io/swipe/v2/internal/usecase/factory"
	"github.com/swipe-io/swipe/v2/internal/usecase/frame"
	"github.com/swipe-io/swipe/v2/internal/usecase/generator"
	"github.com/swipe-io/swipe/v2/internal/usecase/processor"
	"golang.org/x/tools/go/packages"
)

type importerer interface {
	SetImporter(*importer.Importer)
}

type generationExecutor struct {
	processorFactory processor.Factory
	importerFactory  factory.ImporterFactory
	frameFactory     frame.Factory
	optionLoader     *_option.Loader
}

func (e *generationExecutor) processGenerate(pkg *packages.Package, generators []generator.Generator) (result []executor.GenerateResult) {

	for _, g := range generators {

				generated := executor.GenerateResult{}
						if err := g.Prepare(context.TODO()); err != nil {
			generated.Errs = append(generated.Errs, err)
			result = append(result, generated)
			continue
		}

		outputDir := g.OutputDir()
		if outputDir == "" {
			basePath, err := types.DetectBasePath(pkg)
			if err != nil {
				generated.Errs = append(generated.Errs, err)
				result = append(result, generated)
				continue
			}
			outputDir = basePath
		}

		generated.PkgPath = pkg.PkgPath
		generated.OutputPath = filepath.Join(outputDir, g.Filename())

		newImporter := e.importerFactory.NewImporter(generated.OutputPath, pkg)
		if g, ok := g.(importerer); ok {
			g.SetImporter(newImporter)
		}

		if err := g.Process(context.TODO()); err != nil {
			generated.Errs = append(generated.Errs, err)
			result = append(result, generated)
			continue
		}
		fr := e.frameFactory.NewFrame(generated.OutputPath, newImporter, pkg)
		content, err := fr.Frame(g.Bytes())
		if err != nil {
			generated.Content = g.Bytes()
			generated.Errs = append(generated.Errs, err)
			result = append(result, generated)
			continue
		}
		generated.Content = content

		result = append(result, generated)
	}

	return
}

func (e *generationExecutor) Execute() (results []executor.GenerateResult, errs []error) {
	opr, errs := e.optionLoader.Load()
	if len(errs) > 0 {
		return nil, errs
	}
	if len(opr.Options) == 0 {
		return nil, []error{fmt.Errorf("swipe options not found")}
	}

	for _, o := range opr.Options {
		fn, ok := e.processorFactory.Get(o.Option.Name)
		if !ok {
			errs = append(errs, errors.NotePosition(o.Option.Position,
				fmt.Errorf("unknown option name %s", o.Option.Name)))
			continue
		}
		p, err := fn(o, opr)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		processResults := e.processGenerate(o.Pkg, p.Generators(o.Pkg, opr.Data.WorkDir))

		results = append(results, processResults...)

	}
	return
}

func NewGenerationExecutor(
	processorFactory processor.Factory,
	importerFactory factory.ImporterFactory,
	frameFactory frame.Factory,
	optionLoader *_option.Loader,
) executor.GenerationExecutor {
	return &generationExecutor{processorFactory: processorFactory, importerFactory: importerFactory, frameFactory: frameFactory, optionLoader: optionLoader}

}
