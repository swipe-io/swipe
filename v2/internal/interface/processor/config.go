package processor

import (
	"github.com/swipe-io/swipe/v2/internal/interface/generator"
	uga "github.com/swipe-io/swipe/v2/internal/usecase/gateway"
	ug "github.com/swipe-io/swipe/v2/internal/usecase/generator"
	"github.com/swipe-io/swipe/v2/internal/writer"
	"golang.org/x/tools/go/packages"
)

type Config struct {
	ConfigGateway uga.ConfigGateway
}

func (p *Config) Name() string {
	return "ConfigEnv"
}

func (p *Config) Options() []byte {
	return nil
}

func (p *Config) Generators(_ *packages.Package, wd string) []ug.Generator {
	generators := []ug.Generator{
		generator.NewConfig(p.ConfigGateway.Struct(), p.ConfigGateway.StructType(), p.ConfigGateway.StructExpr(), p.ConfigGateway.FuncName()),
	}
	if p.ConfigGateway.DocEnable() {
		generators = append(generators, generator.NewConfigDoc(p.ConfigGateway.Struct(), wd, p.ConfigGateway.DocOutputDir()))
	}
	return generators
}

func ConfigOptions() []byte {
	var w writer.GoWriter
	w.W("// A ConfigEnvOption is an option env config.\ntype ConfigEnvOption string\n\n")
	w.W("// ConfigEnv option for config generation.\n//\n// To generate a configuration loader, you can use the swipe.ConfigEnv option.\n// The optionsStruct parameter is a pointer to the configuration structure.\n//\n//  The option can work with all primitives, including datetime, and an array of primitives.\n//\n//  The option supports nested structures.\n//\n//  To use the default value, just specify it as a value in the structure.\n//\n// Default func name is `LoadConfig`.\n//\n// You can use structure tags to control generation:\n//\n//  env  - name of environment var, options: `required`.\n//\n//  flag - name of flag, enable as the console flag.\n//\n//  desc - description for String function.\nfunc ConfigEnv(optionsStruct interface{}, opts ...ConfigEnvOption) Option {\n\treturn \"implementation not generated, run swipe\"\n}\n\n")
	w.W("func ConfigEnvFuncName(name string) ConfigEnvOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n")
	w.W("// ConfigEnvDocEnable enable markdown doc generate.\nfunc ConfigEnvDocEnable() ConfigEnvOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n")
	w.W("// ConfigEnvDocOutput output path markdown doc generate.\nfunc ConfigEnvDocOutput(output string) ConfigEnvOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n")
	return w.Bytes()
}
