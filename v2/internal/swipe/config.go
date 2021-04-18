package swipe

import (
	option2 "github.com/swipe-io/swipe/v2/internal/option"
	"golang.org/x/tools/go/packages"

	"github.com/swipe-io/swipe/v2/internal/ast"
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
	Build  *option2.Build
	Module *option2.Module
}

type Config struct {
	WorkDir  string
	Envs     []string
	Patterns []string
	Modules  map[string]*option2.Module

	Pkg          *packages.Package
	Packages     []*packages.Package
	CommentFuncs map[string][]string
}

func GetConfig(loader *ast.Loader) (*Config, []error) {
	cfg := Config{
		WorkDir:      loader.WorkDir(),
		Envs:         loader.Env(),
		Patterns:     loader.Patterns(),
		Pkg:          loader.Pkg(),
		Packages:     loader.Pkgs(),
		CommentFuncs: loader.CommentFuncs(),
	}
	if err := cfg.Load(); err != nil {
		return nil, nil
	}
	return &cfg, nil
}

func (c *Config) Load() (err error) {
	c.Modules, err = option2.Decode(c.Pkg, c.Packages, c.CommentFuncs)
	return
}
