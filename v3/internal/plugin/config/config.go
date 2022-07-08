package config

import (
	"context"
	"fmt"
	"strconv"

	"github.com/swipe-io/swipe/v3/internal/convert"

	"github.com/swipe-io/strcase"
	"github.com/swipe-io/swipe/v3/option"
	"github.com/swipe-io/swipe/v3/swipe"
	"github.com/swipe-io/swipe/v3/writer"
)

type Generator struct {
	w        writer.GoWriter
	Struct   *option.NamedType
	FuncName string
}

func (g *Generator) Generate(ctx context.Context) []byte {
	importer := ctx.Value(swipe.ImporterKey).(swipe.Importer)

	typeName := swipe.TypeString(g.Struct, false, importer)

	g.w.W("func %s() (cfg *%s, errs []error) {\n", g.FuncName, typeName)
	g.w.W("cfg = &%s{}\n", typeName)

	var (
		foundFlags    bool
		envs          []fldOpts
		requiredFlags []struct {
			f    *option.VarType
			opts fldOpts
		}
	)

	walk(g.Struct, func(f, parent *option.VarType, opts fldOpts) {
		if opts.isFlag {
			foundFlags = true
		}
		envs = append(envs, opts)
		switch f.Type.(type) {
		case *option.NamedType:
			g.writeEnv(importer, f, opts)
		case *option.BasicType, *option.SliceType, *option.ArrayType, *option.MapType:
			if opts.isFlag {
				g.writeFlag(importer, f, opts)
				if opts.required {
					requiredFlags = append(requiredFlags, struct {
						f    *option.VarType
						opts fldOpts
					}{f: f, opts: opts})
				}
			} else {
				g.writeEnv(importer, f, opts)
			}
		}
	})

	if foundFlags {
		flagPkg := importer.Import("flag", "flag")

		g.w.W("%s.Parse()\n", flagPkg)

		g.w.W("seen := map[string]struct{}{}\n")
		g.w.W("%[1]s.Visit(func(f *%[1]s.Flag) { seen[f.Name] = struct{}{} })\n", flagPkg)

		for _, o := range requiredFlags {
			g.w.W("if _, ok := seen[\"%s\"]; !ok {\n", o.opts.name)
			g.writeAppendErr(importer, o.opts)
			g.w.W("}")
			if !bool(o.opts.useZero) && bool(o.opts.required) {
				g.w.W(" else {\n")
				g.writeCheckZero(importer, o.f, o.opts)
				g.w.W("}\n")
			} else {
				g.w.W("\n")
			}
		}
	}

	g.w.W("return\n")
	g.w.W("}\n\n")

	g.w.W("func (cfg *%s) String() string {\n", typeName)
	g.w.W("out := `\n")
	if len(envs) > 0 {
		fmtPkg := importer.Import("fmt", "fmt")
		for _, env := range envs {
			if env.isFlag {
				g.w.W("--%s ", env.name)
			} else {
				g.w.W("%s=", env.name)
			}
			g.w.W("`+%s.Sprintf(\"%%v\", %s)+`", fmtPkg, "cfg."+env.fieldPath)
			if env.desc != "" {
				g.w.W(" ; %s", env.desc)
			}
			g.w.Line()
		}
	}
	g.w.W("`\n")
	g.w.W("return out\n")
	g.w.W("}\n\n")

	return g.w.Bytes()
}

func (g *Generator) OutputPath() string {
	return ""
}

func (g *Generator) Filename() string {
	return "loader.go"
}

func (g *Generator) writeEnv(importer swipe.Importer, f *option.VarType, opts fldOpts) {
	tmpVar := strcase.ToLowerCamel(opts.fieldPath) + "Tmp"
	pkgOS := importer.Import("os", "os")

	g.w.W("%s, ok := %s.LookupEnv(%s)\n", tmpVar, pkgOS, strconv.Quote(opts.name))
	g.w.W("if ok {\n")

	convert.NewBuilder(importer).
		SetDeclareErr(true).
		SetAssignVar("cfg." + opts.fieldPath).
		SetValueVar(tmpVar).
		SetFieldName(f.Name).
		SetFieldType(f.Type).
		SetErrorReturn(func() string {
			return fmt.Sprintf("errs = append(errs, %s.New(%s))", importer.Import("errors", "errors"), strconv.Quote("convert "+opts.name+" error"))
		}).
		Write(&g.w)

	g.writeCheckZero(importer, f, opts)

	g.w.W("}")
	if opts.required {
		g.w.W(" else {\n")
		g.writeAppendErr(importer, opts)
		g.w.W("}\n")
	} else {
		g.w.W("\n")
	}
}

func (g *Generator) writeFlag(i swipe.Importer, f *option.VarType, opts fldOpts) {
	if t, ok := f.Type.(*option.BasicType); ok {
		flagPkg := i.Import("flag", "flag")
		if t.IsString() {
			g.w.W("%[1]s.StringVar(&cfg.%[2]s, \"%[3]s\", cfg.%[2]s, \"%[4]s\")\n", flagPkg, opts.fieldPath, opts.name, opts.desc)
		}
		if t.IsInt64() {
			g.w.W("%[1]s.Int64Var(&cfg.%[2]s, \"%[3]s\", cfg.%[2]s, \"%[4]s\")\n", flagPkg, opts.fieldPath, opts.name, opts.desc)
		}
		if t.IsInt() {
			g.w.W("%[1]s.IntVar(&cfg.%[2]s, \"%[3]s\", cfg.%[2]s, \"%[4]s\")\n", flagPkg, opts.fieldPath, opts.name, opts.desc)
		}
		if t.IsFloat64() {
			g.w.W("%[1]s.Float64Var(&cfg.%[2]s, \"%[3]s\", cfg.%[2]s, \"%[4]s\")\n", flagPkg, opts.fieldPath, opts.name, opts.desc)
		}
		if t.IsBool() {
			g.w.W("%[1]s.BoolVar(&cfg.%[2]s, \"%[3]s\", cfg.%[2]s, \"%[4]s\")\n", flagPkg, opts.fieldPath, opts.name, opts.desc)
		}
	}
}

func (g *Generator) writeCheckZero(i swipe.Importer, f *option.VarType, opts fldOpts) {
	if !bool(opts.useZero) && bool(opts.required) {
		if f.Zero != "" {
			g.w.W("if %s == %s {\n", "cfg."+opts.fieldPath, f.Zero)
			g.writeAppendErr(i, opts)
			g.w.W("}\n")
		}
	}
}

func (g *Generator) writeAppendErr(i swipe.Importer, opts fldOpts) {
	errorsPkg := i.Import("errors", "errors")
	requiredMsg := strconv.Quote(fmt.Sprintf("%s %s required", opts.tagName(), opts.name))
	g.w.W("errs = append(errs, %s.New(%s))\n ", errorsPkg, requiredMsg)
}
