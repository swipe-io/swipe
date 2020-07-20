package processor

import (
	"github.com/swipe-io/swipe/pkg/domain/model"
	ug "github.com/swipe-io/swipe/pkg/usecase/generator"
)

type gatewayProcessor struct {
	option model.GatewayOption
	info   model.GenerateInfo
}

func (g *gatewayProcessor) SetOption(option interface{}) bool {
	o, ok := option.(model.GatewayOption)
	g.option = o
	return ok
}

func (g *gatewayProcessor) Generators() []ug.Generator {
	return []ug.Generator{
		ug.NewGatewayGenerator("generator_gen.go", g.info, g.option),
	}
}

func NewGatewayProcessor(info model.GenerateInfo) Processor {
	return &gatewayProcessor{info: info}
}
