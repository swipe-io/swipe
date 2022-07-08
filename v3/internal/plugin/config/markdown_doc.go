package config

import (
	"context"

	"github.com/swipe-io/swipe/v3/option"
	"github.com/swipe-io/swipe/v3/writer"
)

type MarkdownDocGenerator struct {
	w      writer.TextWriter
	Struct *option.NamedType
	Output string
}

func (g *MarkdownDocGenerator) Generate(ctx context.Context) []byte {
	g.w.W("# Config\n\n")

	var flags []fldOpts
	var envs []fldOpts

	walk(g.Struct, func(f, parent *option.VarType, opts fldOpts) {
		if opts.isFlag {
			flags = append(flags, opts)
		} else {
			envs = append(envs, opts)
		}
	})

	if len(envs) > 0 {
		g.w.W("## Environment variables\n\n")

		g.w.W("| Name | Type | Description | Required | Use Zero |\n|------|------|------|------|------|\n")

		for _, opts := range envs {
			desc := " "
			if opts.desc != "" {
				desc = opts.desc
			}
			g.w.W("|%s|<code>%s</code>|%s|%s|%s|\n", opts.name, g.getTypeSrt(opts.t), desc, opts.required, opts.useZero)
		}
	}

	if len(flags) > 0 {
		g.w.W("## Flags\n\n")
		g.w.W("| Name | Type | Description | Required | Use Zero |\n|------|------|------|------|------|\n")

		for _, opts := range flags {
			desc := " "
			if opts.desc != "" {
				desc = opts.desc
			}
			g.w.W("|%s|<code>%s</code>|%s|%s|%s|\n", opts.name, g.getTypeSrt(opts.t), desc, opts.required, opts.useZero)
		}
	}
	return g.w.Bytes()
}

func (g *MarkdownDocGenerator) getTypeSrt(t interface{}) string {
	switch t := t.(type) {
	default:
		return ""
	case *option.MapType:
		return "map[string]" + g.getTypeSrt(t.Value)
	case *option.ArrayType:
		return g.getTypeSrt(t.Value) + "[]"
	case *option.SliceType:
		return g.getTypeSrt(t.Value) + "[]"
	case *option.BasicType:
		return t.Name
	}
}

func (g *MarkdownDocGenerator) OutputPath() string {
	return g.Output
}

func (g *MarkdownDocGenerator) Filename() string {
	return "config_doc.md"
}
