package registry

import (
	"errors"

	"github.com/swipe-io/swipe/v2/internal/usecase/finder"

	"github.com/swipe-io/swipe/v2/internal/astloader"
	"github.com/swipe-io/swipe/v2/internal/git"
	ig "github.com/swipe-io/swipe/v2/internal/interface/gateway"
	"github.com/swipe-io/swipe/v2/internal/interface/processor"
	"github.com/swipe-io/swipe/v2/internal/option"
	up "github.com/swipe-io/swipe/v2/internal/usecase/processor"
	"github.com/swipe-io/swipe/v2/internal/usecase/registry"
)

type registryProcessor struct {
	finder finder.ServiceFinder
}

func (r *registryProcessor) NewProcessor(o *option.ResultOption, data *astloader.Data) (up.Processor, error) {
	switch o.Option.Name {
	case "Gateway":
		hg, err := ig.NewGateway(o.Pkg, o.Option, r.finder)
		if err != nil {
			return nil, err
		}
		return processor.NewGatewayProcessor(hg, o.Pkg), nil
	case "Service":
		gt := git.NewGIT()
		sg, err := ig.NewServiceGateway(o.Pkg, o.Option, data.GraphTypes, data.CommentFuncs)
		if err != nil {
			return nil, err
		}
		return processor.NewService(
			sg,
			gt,
			data.CommentFields,
			data.Enums,
			data.WorkDir,
			o.Pkg,
		), nil
	case "ConfigEnv":
		return processor.NewConfig(
			ig.NewConfigGateway(o.Option),
			o.Pkg,
			data.WorkDir,
		), nil
	case "Presenter":
		pg, err := ig.NewPresenterGateway(o.Pkg, o.Option)
		if err != nil {
			return nil, err
		}
		return processor.NewPresenterGatewayProcessor(pg, o.Pkg), nil
	}
	return nil, errors.New("unexpected processor: " + o.Option.Name)
}

func NewRegistry(finder finder.ServiceFinder) registry.ProcessorRegistry {
	return &registryProcessor{finder: finder}
}
