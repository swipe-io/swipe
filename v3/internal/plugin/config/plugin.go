package config

import (
	"github.com/mitchellh/mapstructure"
	"github.com/swipe-io/swipe/v3/option"
	"github.com/swipe-io/swipe/v3/swipe"
)

const defaultFuncName = "LoadConfig"

func init() {
	swipe.RegisterPlugin(&Plugin{})
}

type StringValue struct {
	Value string
}

type Environment struct {
	StructType *option.NamedType
	FuncName   *StringValue `swipe:"option"`
	EnableDoc  *struct{}    `swipe:"option"`
	OutputDoc  StringValue  `swipe:"option"`
}

// Config
// @swipe:"Config"
type Config struct {
	Environment Environment
}

type Plugin struct {
	config Config
}

func (p *Plugin) Generators() (generators []swipe.Generator, errs []error) {
	funcName := defaultFuncName
	if p.config.Environment.FuncName != nil {
		funcName = p.config.Environment.FuncName.Value
	}
	if p.config.Environment.EnableDoc != nil {
		generators = append(generators, &MarkdownDocGenerator{
			Struct: p.config.Environment.StructType,
			Output: p.config.Environment.OutputDoc.Value,
		})
	}
	generators = append(generators, &Generator{
		Struct:   p.config.Environment.StructType,
		FuncName: funcName,
	})
	return
}

func (p *Plugin) Configure(cfg *swipe.Config, module *option.Module, build *option.Build, options map[string]interface{}) []error {
	if err := mapstructure.Decode(options, &p.config); err != nil {
		return []error{err}
	}
	return nil
}

func (p *Plugin) Options() []byte {
	var cfg interface{} = &Config{}
	if o, ok := cfg.(interface{ Options() []byte }); ok {
		return o.Options()
	}
	return nil
}

func (p *Plugin) ID() string {
	return "Config"
}
