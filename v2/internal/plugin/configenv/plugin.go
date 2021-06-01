package configenv

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
	Name string
}

type Config struct {
	Struct *option.NamedType `mapstructure:"optionsStruct"`
	Func   *Func             `mapstructure:"ConfigEnvFuncName"`
}

type Plugin struct {
	config Config
}

func (p *Plugin) Generators() (generators []swipe.Generator, errs []error) {
	funcName := defaultFuncName
	if p.config.Func != nil {
		funcName = p.config.Func.Name
	}
	generators = append(generators, &Generator{
		Struct:   p.config.Struct,
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

func (p *Plugin) ID() string {
	return "ConfigEnv"
}
