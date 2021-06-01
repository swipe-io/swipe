package configenv

import (
	"context"
	"fmt"
	"strconv"

	"github.com/swipe-io/strcase"
	"github.com/swipe-io/swipe/v2/option"
	"github.com/swipe-io/swipe/v2/swipe"
	"github.com/swipe-io/swipe/v2/writer"
)

type Generator struct {
	writer.GoWriter
	Struct   *option.NamedType
	FuncName string
}

func (g *Generator) Generate(ctx context.Context) []byte {
	importer := ctx.Value(swipe.ImporterKey).(swipe.Importer)

	pkgName := importer.Import(g.Struct.Pkg.Name, g.Struct.Pkg.Path)
	typeName := pkgName + g.Struct.Name.Upper()

	g.W("func %s() (c *%s, errs []error) {\n", g.FuncName, typeName)
	g.W("c = &%s{}\n", typeName)

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
		case *option.BasicType, *option.ArrayType, *option.MapType:
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

		g.W("%sParse()\n", flagPkg)

		g.W("seen := map[string]struct{}{}\n")
		g.W("%[1]sVisit(func(f *%[1]s.Flag) { seen[f.Name] = struct{}{} })\n", flagPkg)

		for _, o := range requiredFlags {
			g.W("if _, ok := seen[\"%s\"]; !ok {\n", o.opts.name)
			g.writeAppendErr(importer, o.opts)
			g.W("}")
			if !bool(o.opts.useZero) && bool(o.opts.required) {
				g.W(" else {\n")
				g.writeCheckZero(importer, o.f, o.opts)
				g.W("}\n")
			} else {
				g.W("\n")
			}
		}
	}

	g.W("return\n")
	g.W("}\n\n")

	g.W("func (cfg %s) String() string {\n", typeName)
	g.W("out := `\n")
	if len(envs) > 0 {
		fmtPkg := importer.Import("fmt", "fmt")
		for _, env := range envs {
			if env.isFlag {
				g.W("--%s ", env.name)
			} else {
				g.W("%s=", env.name)
			}
			g.W("`+%sSprintf(\"%%v\", %s)+`", fmtPkg, "cfg."+env.fieldPath)
			if env.desc != "" {
				g.W(" ; %s", env.desc)
			}
			g.Line()
		}
	}
	g.W("`\n")
	g.W("return out\n")
	g.W("}\n\n")

	return g.Bytes()
}

func (g *Generator) OutputDir() string {
	return ""
}

func (g *Generator) Filename() string {
	return "config.go"
}

func (g *Generator) writeEnv(importer swipe.Importer, f *option.VarType, opts fldOpts) {
	tmpVar := strcase.ToLowerCamel(opts.fieldPath) + "Tmp"
	pkgOS := importer.Import("os", "os")
	g.W("%s, ok := %sLookupEnv(%s)\n", tmpVar, pkgOS, strconv.Quote(opts.name))
	g.W("if ok {\n")

	g.WriteConvertType(importer, "cfg."+opts.fieldPath, tmpVar, f, nil, "errs", false, "convert "+opts.name+" error")
	g.writeCheckZero(importer, f, opts)

	g.W("}")
	if opts.required {
		g.W(" else {\n")
		g.writeAppendErr(importer, opts)
		g.W("}\n")
	} else {
		g.W("\n")
	}
}

func (g *Generator) writeFlag(i swipe.Importer, f *option.VarType, opts fldOpts) {
	if t, ok := f.Type.(*option.BasicType); ok {
		flagPkg := i.Import("flag", "flag")
		if t.IsString() {
			g.W("%[1]sStringVar(&cfg.%[2]s, \"%[3]s\", cfg.%[2]s, \"%[4]s\")\n", flagPkg, opts.fieldPath, opts.name, opts.desc)
		}
		if t.IsInt64() {
			g.W("%[1]sInt64Var(&cfg.%[2]s, \"%[3]s\", cfg.%[2]s, \"%[4]s\")\n", flagPkg, opts.fieldPath, opts.name, opts.desc)
		}
		if t.IsInt() {
			g.W("%[1]sIntVar(&cfg.%[2]s, \"%[3]s\", cfg.%[2]s, \"%[4]s\")\n", flagPkg, opts.fieldPath, opts.name, opts.desc)
		}
		if t.IsFloat64() {
			g.W("%[1]sFloat64Var(&cfg.%[2]s, \"%[3]s\", cfg.%[2]s, \"%[4]s\")\n", flagPkg, opts.fieldPath, opts.name, opts.desc)
		}
		if t.IsBool() {
			g.W("%[1]sBoolVar(&cfg.%[2]s, \"%[3]s\", cfg.%[2]s, \"%[4]s\")\n", flagPkg, opts.fieldPath, opts.name, opts.desc)
		}
	}
}

func (g *Generator) writeCheckZero(i swipe.Importer, f *option.VarType, opts fldOpts) {
	if !bool(opts.useZero) && bool(opts.required) {
		if f.Zero != "" {
			g.W("if %s == %s {\n", "cfg."+opts.fieldPath, f.Zero)
			g.writeAppendErr(i, opts)
			g.W("}\n")
		}
	}
}

func (g *Generator) writeAppendErr(i swipe.Importer, opts fldOpts) {
	errorsPkg := i.Import("errors", "errors")
	requiredMsg := strconv.Quote(fmt.Sprintf("%s %s required", opts.tagName(), opts.name))
	g.W("errs = append(errs, %sNew(%s))\n ", errorsPkg, requiredMsg)
}
