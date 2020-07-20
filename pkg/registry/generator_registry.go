package registry

import (
	"github.com/swipe-io/swipe/pkg/domain/model"
	io "github.com/swipe-io/swipe/pkg/interface/option"
	uo "github.com/swipe-io/swipe/pkg/usecase/option"
	up "github.com/swipe-io/swipe/pkg/usecase/processor"
)

type Registry struct {
}

func (r *Registry) Option(name string, info model.GenerateInfo) uo.Option {
	switch name {
	case "Gateway":
		return io.NewGatewayOption(info)
	case "ConfigEnv":
		return io.NewConfigOption()
	case "Service":
		return io.NewServiceOption(info)
	}
	return nil
}

func (r *Registry) Processor(name string, info model.GenerateInfo) (up.Processor, error) {
	switch name {
	case "Gateway":
		return up.NewGatewayProcessor(info), nil
	case "ConfigEnv":
		return up.NewConfig(info), nil
	case "Service":
		return up.NewService(info), nil
	}
	return nil, nil
}

func NewRegistry() *Registry {
	return &Registry{}
}
