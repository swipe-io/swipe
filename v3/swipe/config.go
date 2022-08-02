package swipe

import (
	"strings"

	packages2 "github.com/swipe-io/swipe/v3/internal/packages"

	"github.com/swipe-io/swipe/v3/internal/ast"
	"github.com/swipe-io/swipe/v3/option"

	"golang.org/x/tools/go/packages"
)

type warnError struct {
	Err error
}

func (e *warnError) Warn() error {
	return e.Err
}

func (e *warnError) Error() string {
	return e.Err.Error()
}

type PluginConfig struct {
	Plugin Plugin
	Build  *option.Inject
	Module *option.Module
}

type Config struct {
	WorkDir       string
	Envs          []string
	Patterns      []string
	Modules       map[string]*option.Module
	Module        *packages.Module
	Packages      *packages2.Packages
	CommentFuncs  map[string][]string
	CommentFields *ast.CommentFields
}

func GetConfig(loader *ast.Loader) (*Config, error) {
	cfg := Config{
		WorkDir:       loader.WorkDir(),
		Envs:          loader.Env(),
		Patterns:      loader.Patterns(),
		Module:        loader.Module(),
		Packages:      packages2.NewPackages(loader.Pkgs()),
		CommentFuncs:  loader.CommentFuncs(),
		CommentFields: loader.CommentFields(),
	}
	if err := cfg.Load(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) WalkBuilds(fn func(module *option.Module, build *option.Inject) bool) {
	for _, module := range c.Modules {
		for _, build := range module.Injects {
			if !fn(module, build) {
				break
			}
		}
	}
}

func (c *Config) Load() (err error) {
	optionPackages := map[string]string{}
	registeredPlugins.Range(func(key, value any) bool {
		pluginID := key.(string)
		optionPackages[strings.ToLower(pluginID)] = pluginID
		return true
	})
	c.Modules, err = option.Decode(optionPackages, c.Module, c.Packages, c.CommentFuncs, c.CommentFields)
	return
}

func Options() (data map[string][]byte) {
	data = map[string][]byte{}
	registeredPlugins.Range(func(key, value any) bool {
		pluginID := key.(string)
		f := value.(func() Plugin)
		p := f()
		name := strings.ToLower(pluginID)
		data[name] = append(data[name], p.Options()...)
		return true
	})
	return
}
