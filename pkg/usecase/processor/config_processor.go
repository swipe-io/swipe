package processor

import (
	"github.com/swipe-io/swipe/pkg/domain/model"
	ug "github.com/swipe-io/swipe/pkg/usecase/generator"
)

type config struct {
	info   model.GenerateInfo
	option model.ConfigOption
}

func (p *config) SetOption(option interface{}) bool {
	o, ok := option.(model.ConfigOption)
	p.option = o
	return ok
}

func (p *config) Generators() []ug.Generator {
	generators := []ug.Generator{
		ug.NewConfig(p.option),
	}
	if p.option.Doc.Enable {
		generators = append(generators, ug.NewConfigMarkdownDoc("config_doc_gen.md", p.option, p.info))
	}
	return generators
}

func NewConfig(info model.GenerateInfo) Processor {
	return &config{info: info}
}
