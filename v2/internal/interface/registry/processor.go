package registry

import (
	"errors"

	"github.com/swipe-io/swipe/v2/internal/astloader"
	"github.com/swipe-io/swipe/v2/internal/git"
	ig "github.com/swipe-io/swipe/v2/internal/interface/gateway"
	"github.com/swipe-io/swipe/v2/internal/interface/processor"
	"github.com/swipe-io/swipe/v2/internal/option"
	up "github.com/swipe-io/swipe/v2/internal/usecase/processor"
	"github.com/swipe-io/swipe/v2/internal/usecase/registry"
)

type registryProcessor struct {
	l *option.Loader
}

func (r *registryProcessor) NewProcessor(o *option.ResultOption, externalOptions []*option.ResultOption, data *astloader.Data) (up.Processor, error) {
	gt := git.NewGIT()
	switch o.Option.Name {
	case "Service":
		sg, err := ig.NewServiceGateway(o.Pkg, data.PkgPath, o.Option, data.GraphTypes, data.CommentFuncs, data.CommentFields, data.Enums, data.WorkDir, externalOptions)
		if err != nil {
			return nil, err
		}
		return processor.NewService(
			sg,
			gt,
			data.WorkDir,
			o.Pkg,
		), nil
	case "ConfigEnv":
		return processor.NewConfig(
			ig.NewConfigGateway(o.Option),
			o.Pkg,
			data.WorkDir,
		), nil
	}
	return nil, errors.New("unexpected processor: " + o.Option.Name)
}

func NewRegistry(l *option.Loader) registry.ProcessorRegistry {
	return &registryProcessor{l: l}
}
