package swipe

import (
	"github.com/swipe-io/swipe/v2/internal/ast"
	"github.com/swipe-io/swipe/v2/internal/option"

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
	//CallGraph    *graph.Graph
	//findStmt     func(*stdtypes.Interface)
}

func GetConfig(loader *ast.Loader) (*Config, error) {
	cfg := Config{
		WorkDir:      loader.WorkDir(),
		Envs:         loader.Env(),
		Patterns:     loader.Patterns(),
		Pkg:          loader.Pkg(),
		Packages:     loader.Pkgs(),
		CommentFuncs: loader.CommentFuncs(),
		//CallGraph:    loader.CallGraph(),
		//findStmt:     loader.FindStmt,
	}
	if err := cfg.Load(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) Load() (err error) {
	c.Modules, err = option.Decode(c.Pkg, c.Packages, c.CommentFuncs)
	return
}
