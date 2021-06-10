package config

import (
	"github.com/mitchellh/mapstructure"

	"github.com/swipe-io/swipe/v2/option"
	"github.com/swipe-io/swipe/v2/swipe"
)

const defaultFuncName = "LoadConfig"

func init() {
	swipe.RegisterPlugin(&Plugin{})
}

type Func struct {
	Value string
}

type Environment struct {
	StructType *option.NamedType
	FuncName   *Func `swipe:"option"`
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
