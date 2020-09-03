package generator

import (
	"context"
	stdtypes "go/types"
	"path/filepath"

	"github.com/swipe-io/swipe/pkg/domain/model"
	"github.com/swipe-io/swipe/pkg/writer"
)

type configMarkdownDoc struct {
	writer.BaseWriter
	info      model.GenerateInfo
	o         model.ConfigOption
	outputDir string
	filename  string
}

func (g *configMarkdownDoc) Prepare(ctx context.Context) error {
	outputDir, err := filepath.Abs(filepath.Join(g.info.BasePath, g.o.Doc.OutputDir))
	if err != nil {
		return err
	}
	g.outputDir = outputDir
	return nil
}

func (g *configMarkdownDoc) Process(ctx context.Context) error {
	g.W("# Config\n\n")

	var flags []fldOpts
	var envs []fldOpts

	walkStruct(g.o.Struct, func(f, parent *stdtypes.Var, opts fldOpts) {
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

func (g *configMarkdownDoc) PkgName() string {
	return ""
}

func (g *configMarkdownDoc) OutputDir() string {
	return g.outputDir
}

func (g *configMarkdownDoc) Filename() string {
	return g.filename
}

func NewConfigMarkdownDoc(filename string, o model.ConfigOption, info model.GenerateInfo) Generator {
	return &configMarkdownDoc{filename: filename, o: o, info: info}
}
