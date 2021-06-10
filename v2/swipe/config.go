package swipe

import (
	"strings"

	"github.com/swipe-io/swipe/v2/internal/ast"
	"github.com/swipe-io/swipe/v2/option"

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
	Build  *option.Build
	Module *option.Module
}

type Config struct {
	WorkDir  string
	Envs     []string
	Patterns []string
	Modules  map[string]*option.Module

	Pkg          *packages.Package
	Packages     []*packages.Package
	CommentFuncs map[string][]string
}

func GetConfig(loader *ast.Loader) (*Config, error) {
	cfg := Config{
		WorkDir:      loader.WorkDir(),
		Envs:         loader.Env(),
		Patterns:     loader.Patterns(),
		Pkg:          loader.Pkg(),
		Packages:     loader.Pkgs(),
		CommentFuncs: loader.CommentFuncs(),
	}
	if err := cfg.Load(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) WalkBuilds(fn func(module *option.Module, build *option.Build) bool) {
	for _, module := range c.Modules {
		for _, build := range module.Builds {
			if !fn(module, build) {
				break
			}
		}
	}
}

func (c *Config) Load() (err error) {
	optionPkgs := map[string]string{}
	for _, plugin := range registeredPlugins {
		optionPkgs["swipe"+strings.ToLower(plugin.ID())] = plugin.ID()
	}
	c.Modules, err = option.Decode(optionPkgs, c.Pkg, c.Packages, c.CommentFuncs)
	return
}

func Options() (data map[string][]byte) {
	data = map[string][]byte{}
	for _, plugin := range registeredPlugins {
		name := "swipe" + strings.ToLower(plugin.ID())
		data[name] = append(data[name], plugin.Options()...)
	}
	return
}
