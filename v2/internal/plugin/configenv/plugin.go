package configenv

import (
	"github.com/mitchellh/mapstructure"
	option2 "github.com/swipe-io/swipe/v2/internal/option"
	"github.com/swipe-io/swipe/v2/internal/swipe"
)

const defaultFuncName = "LoadConfig"

func init() {
	swipe.RegisterPlugin(&Plugin{})
}

type Func struct {
	Name string
}

type Config struct {
	Struct *option2.StructType `mapstructure:"optionsStruct"`
	Func   *Func               `mapstructure:"ConfigEnvFuncName"`
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

func (p *Plugin) Configure(cfg *swipe.Config, module *option2.Module, build *option2.Build, config interface{}) error {
	return mapstructure.Decode(config, &p.config)
}

func (p *Plugin) ID() string {
	return "ConfigEnv"
}
