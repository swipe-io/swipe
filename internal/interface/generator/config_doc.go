package generator

import (
	"context"
	stdtypes "go/types"
	"path/filepath"

	"github.com/swipe-io/swipe/v2/internal/usecase/generator"
	"github.com/swipe-io/swipe/v2/internal/writer"
)

type configDoc struct {
	writer.BaseWriter
	st        *stdtypes.Struct
	workDir   string
	outputDir string
}

func (g *configDoc) Prepare(ctx context.Context) (err error) {
	g.outputDir, err = filepath.Abs(filepath.Join(g.workDir, g.outputDir))
	return
}

func (g *configDoc) Process(ctx context.Context) error {
	g.W("# Config\n\n")

	var flags []fldOpts
	var envs []fldOpts

	walkStruct(g.st, func(f, parent *stdtypes.Var, opts fldOpts) {
		if opts.isFlag {
			flags = append(flags, opts)
		} else {
			envs = append(envs, opts)
		}
	})

	g.W("## Environment variables\n\n")

	g.W("| Name | Type | Description | Required |\n|------|------|------|------|\n")

	for _, opts := range envs {
		desc := " "
		if opts.desc != "" {
			desc = opts.desc
		}
		g.W("|%s|<code>%s</code>|%s|%s|\n", opts.name, opts.typeStr, desc, opts.required)
	}

	g.W("## Flags\n\n")
	g.W("| Name | Type | Description | Required |\n|------|------|------|------|\n")

	for _, opts := range flags {
		desc := " "
		if opts.desc != "" {
			desc = opts.desc
		}
		g.W("|%s|<code>%s</code>|%s|%s|\n", opts.name, opts.typeStr, desc, opts.required)
	}

	return nil
}

func (g *configDoc) PkgName() string {
	return ""
}

func (g *configDoc) OutputDir() string {
	return g.outputDir
}

func (g *configDoc) Filename() string {
	return "config_doc_gen.md"
}

func NewConfigDoc(
	st *stdtypes.Struct,
	workDir string,
	outputDir string,
) generator.Generator {
	return &configDoc{st: st, workDir: workDir, outputDir: outputDir}
}
