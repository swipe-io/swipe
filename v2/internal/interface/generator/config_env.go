package generator

import (
	"context"
	"fmt"
	"go/ast"
	stdtypes "go/types"
	"strconv"
	stdstrings "strings"

	"github.com/swipe-io/swipe/v2/internal/types"

	"github.com/fatih/structtag"

	"github.com/swipe-io/strcase"
	"github.com/swipe-io/swipe/v2/internal/importer"
	"github.com/swipe-io/swipe/v2/internal/usecase/generator"
	"github.com/swipe-io/swipe/v2/internal/writer"
)

type Bool bool

func (r Bool) String() string {
	if r {
		return "yes"
	}
	return "no"
}

type fldOpts struct {
	desc      string
	name      string
	fieldPath string
	required  Bool
	useZero   Bool
	isFlag    bool
	typeStr   string
}

func (o fldOpts) tagName() string {
	if o.isFlag {
		return "flag"
	}
	return "env"
}

func getFieldOpts(f *stdtypes.Var, tag string) (result fldOpts) {
	result.typeStr = stdtypes.TypeString(f.Type(), func(p *stdtypes.Package) string {
		return p.Name()
	})
	result.name = strcase.ToScreamingSnake(f.Name())
	result.fieldPath = f.Name()
	if tags, err := structtag.Parse(tag); err == nil {
		if tag, err := tags.Get("env"); err == nil {
			for _, option := range tag.Options {
				switch option {
				case "use_zero":
					result.useZero = true
				case "required":
					result.required = true
				case "use_flag":
					result.name = strcase.ToKebab(f.Name())
					result.isFlag = true
				default:
					if stdstrings.HasPrefix(option, "desc:") {
						descParts := stdstrings.Split(option, "desc:")
						if len(descParts) == 2 {
							result.desc = descParts[1]
						}
					}
				}
			}
			if tag.Name != "" {
				result.name = tag.Name
			}
		}
	}
	return
}

func walkStructRecursive(st *stdtypes.Struct, parent *stdtypes.Var, popts fldOpts, fn func(f, parent *stdtypes.Var, opts fldOpts)) {
	for i := 0; i < st.NumFields(); i++ {
		f := st.Field(i)
		fopts := getFieldOpts(f, st.Tag(i))
		if popts.name != "" && parent != nil {
			fopts.name = popts.name + "_" + fopts.name
			fopts.fieldPath = popts.fieldPath + "." + fopts.fieldPath

		}
		switch v := f.Type().(type) {
		default:
			fn(f, parent, fopts)
		case *stdtypes.Pointer:
			if st, ok := v.Elem().Underlying().(*stdtypes.Struct); ok {
				walkStructRecursive(st, f, fopts, fn)
			}
		case *stdtypes.Named:
			switch v.String() {
			case "time.Time", "time.Duration":
				fn(f, parent, fopts)
				continue
			}
			if st, ok := v.Underlying().(*stdtypes.Struct); ok {
				walkStructRecursive(st, f, fopts, fn)
			}
		case *stdtypes.Struct:
			walkStructRecursive(v, f, fopts, fn)
		}
	}
}

func walkStruct(st *stdtypes.Struct, fn func(f, parent *stdtypes.Var, opts fldOpts)) {
	walkStructRecursive(st, nil, fldOpts{}, fn)
}

type config struct {
	writer.GoLangWriter
	i        *importer.Importer
	st       *stdtypes.Struct
	stType   stdtypes.Type
	stExpr   ast.Expr
	funcName string
}

func (g *config) Prepare(_ context.Context) error {
	return nil
}

func (g *config) Process(_ context.Context) error {
	stTypeStr := stdtypes.TypeString(g.stType, g.i.QualifyPkg)

	g.W("func %s() (cfg %s, errs []error) {\n", g.funcName, stTypeStr)
	g.W("cfg = ")
	writer.WriteAST(g, g.i, g.stExpr)
	g.W("\n")

	var foundFlags bool
	var requiredFlags []struct {
		f    *stdtypes.Var
		opts fldOpts
	}
	var envs []fldOpts

	walkStruct(g.st, func(f, parent *stdtypes.Var, opts fldOpts) {
		if opts.isFlag {
			foundFlags = true
		}
		envs = append(envs, opts)

		switch v := f.Type().(type) {
		case *stdtypes.Named:
			g.writeEnv(f, opts)
		case *stdtypes.Pointer:
			if v.Elem().String() == "net/url.URL" {
				g.writeEnv(f, opts)
			}
		case *stdtypes.Basic, *stdtypes.Slice, *stdtypes.Map:
			if opts.isFlag {
				g.writeFlag(f, opts)
				if opts.required {
					requiredFlags = append(requiredFlags, struct {
						f    *stdtypes.Var
						opts fldOpts
					}{f: f, opts: opts})
				}
			} else {
				g.writeEnv(f, opts)
			}
		}
	})

	if foundFlags {
		flagPkg := g.i.Import("flag", "flag")

		g.W("%s.Parse()\n", flagPkg)

		g.W("seen := map[string]struct{}{}\n")
		g.W("%[1]s.Visit(func(f *%[1]s.Flag) { seen[f.Name] = struct{}{} })\n", flagPkg)

		for _, o := range requiredFlags {
			g.W("if _, ok := seen[\"%s\"]; !ok {\n", o.opts.name)
			g.writeAppendErr(o.opts)
			g.W("}")
			if !bool(o.opts.useZero) && bool(o.opts.required) {
				g.W(" else {\n")
				g.writeCheckZero(o.f, o.opts)
				g.W("}\n")
			} else {
				g.W("\n")
			}
		}
	}

	g.W("return\n")
	g.W("}\n\n")

	g.W("func (cfg %s) String() string {\n", stTypeStr)
	g.W("out := `\n")
	if len(envs) > 0 {
		fmtPkg := g.i.Import("fmt", "fmt")
		for _, env := range envs {
			if env.isFlag {
				g.W("--%s ", env.name)
			} else {
				g.W("%s=", env.name)
			}
			g.W("`+%s.Sprintf(\"%%v\", %s)+`", fmtPkg, "cfg."+env.fieldPath)
			if env.desc != "" {
				g.W(" ; %s", env.desc)
			}
			g.Line()
		}
	}
	g.W("`\n")
	g.W("return out\n}\n\n")

	return nil
}

func (g *config) writeAppendErr(opts fldOpts) {
	errorsPkg := g.i.Import("errors", "errors")
	requiredMsg := strconv.Quote(fmt.Sprintf("%s %s required", opts.tagName(), opts.name))
	g.W("errs = append(errs, %s.New(%s))\n ", errorsPkg, requiredMsg)
}

func (g *config) writeFlag(f *stdtypes.Var, opts fldOpts) {
	if t, ok := f.Type().(*stdtypes.Basic); ok {
		flagPkg := g.i.Import("flag", "flag")
		switch t.Kind() {
		case stdtypes.String:
			g.W("%[1]s.StringVar(&cfg.%[2]s, \"%[3]s\", cfg.%[2]s, \"%[4]s\")\n", flagPkg, opts.fieldPath, opts.name, opts.desc)
		case stdtypes.Int:
			g.W("%[1]s.IntVar(&cfg.%[2]s, \"%[3]s\", cfg.%[2]s, \"%[4]s\")\n", flagPkg, opts.fieldPath, opts.name, opts.desc)
		case stdtypes.Int64:
			g.W("%[1]s.Int64Var(&cfg.%[2]s, \"%[3]s\", cfg.%[2]s, \"%[4]s\")\n", flagPkg, opts.fieldPath, opts.name, opts.desc)
		case stdtypes.Float64:
			g.W("%[1]s.Float64Var(&cfg.%[2]s, \"%[3]s\", cfg.%[2]s, \"%[4]s\")\n", flagPkg, opts.fieldPath, opts.name, opts.desc)
		case stdtypes.Bool:
			g.W("%[1]s.BoolVar(&cfg.%[2]s, \"%[3]s\", cfg.%[2]s, \"%[4]s\")\n", flagPkg, opts.fieldPath, opts.name, opts.desc)
		}
	}
}

func (g *config) writeCheckZero(f *stdtypes.Var, opts fldOpts) {
	if !bool(opts.useZero) && bool(opts.required) {
		if !types.HasNoEmptyValue(f.Type()) {
			g.W("if %s == %s {\n", "cfg."+opts.fieldPath, types.ZeroValue(f.Type(), g.i.QualifyPkg))
			g.writeAppendErr(opts)
			g.W("}\n")
		}
	}
}

func (g *config) writeEnv(f *stdtypes.Var, opts fldOpts) {
	tmpVar := strcase.ToLowerCamel(opts.fieldPath) + "Tmp"
	g.W("%s, ok := %s.LookupEnv(%s)\n", tmpVar, g.i.Import("os", "os"), strconv.Quote(opts.name))
	g.W("if ok {\n")

	g.WriteConvertType(g.i.Import, "cfg."+opts.fieldPath, tmpVar, f, nil, "errs", false, "convert "+opts.name+" error")
	g.writeCheckZero(f, opts)

	g.W("}")
	if opts.required {
		g.W(" else {\n")
		g.writeAppendErr(opts)
		g.W("}\n")
	} else {
		g.W("\n")
	}
}

func (g *config) PkgName() string {
	return ""
}

func (g *config) OutputDir() string {
	return ""
}

func (g *config) Filename() string {
	return "config_gen.go"
}

func (g *config) SetImporter(i *importer.Importer) {
	g.i = i
}

func NewConfig(
	st *stdtypes.Struct,
	stType stdtypes.Type,
	stExpr ast.Expr,
	funcName string,
) generator.Generator {
	return &config{
		st:       st,
		stType:   stType,
		stExpr:   stExpr,
		funcName: funcName,
	}
}
