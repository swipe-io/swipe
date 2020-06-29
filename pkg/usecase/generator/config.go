package generator

import (
	"context"
	"fmt"
	stdtypes "go/types"
	"strconv"

	"github.com/swipe-io/swipe/pkg/domain/model"

	"github.com/fatih/structtag"
	"github.com/iancoleman/strcase"

	"github.com/swipe-io/swipe/pkg/importer"
	"github.com/swipe-io/swipe/pkg/strings"
	"github.com/swipe-io/swipe/pkg/types"
	"github.com/swipe-io/swipe/pkg/writer"
)

type env struct {
	name     string
	fName    string
	assignID string
	required bool
	isFlag   bool
	desc     string
}

type config struct {
	*writer.GoLangWriter

	i *importer.Importer
	o model.ConfigOption
}

func (g *config) Process(ctx context.Context) error {
	o := g.o

	fmtPkg := g.i.Import("fmt", "fmt")

	structType := o.Struct
	structTypeStr := stdtypes.TypeString(o.StructType, g.i.QualifyPkg)

	g.W("func %s() (cfg %s, errs []error) {\n", o.FuncName, structTypeStr)
	g.W("cfg = ")
	writer.WriteAST(g, g.i, o.StructExpr)
	g.W("\n")

	envs := g.writeConfigStruct("cfg", "", structType, false, "errs")

	flagFound := false
	for _, env := range envs {
		if env.isFlag {
			flagFound = true
			break
		}
	}
	if flagFound {
		g.W("%s.Parse()\n", g.i.Import("flag", "flag"))
	}

	g.W("return\n")
	g.W("}\n\n")

	g.W("func (cfg %s) String() string {\n", structTypeStr)
	g.W("out := `\n")
	for _, env := range envs {
		if env.isFlag {
			g.W("--%s ", env.name)
		} else {
			g.W("%s=", env.name)
		}
		g.W("`+%s.Sprintf(\"%%v\", %s)+`", fmtPkg, env.assignID)
		if env.desc != "" {
			g.W(" ;%s", env.desc)
		}
		g.Line()
	}
	g.W("`\n")
	g.W("return out\n}\n\n")

	return nil
}

func (g *config) writeConfigFlagBasic(name, fldName string, desc string, f *stdtypes.Var) {
	if t, ok := f.Type().(*stdtypes.Basic); ok {
		flagPkg := g.i.Import("flag", "flag")
		switch t.Kind() {
		case stdtypes.String:
			g.W("%[1]s.StringVar(&%[2]s, \"%[3]s\", %[2]s, \"%[4]s\")\n", flagPkg, name, fldName, desc)
		case stdtypes.Int:
			g.W("%[1]s.IntVar(&%[2]s, \"%[3]s\", %[2]s, \"%[4]s\")\n", flagPkg, name, fldName, desc)
		case stdtypes.Int64:
			g.W("%[1]s.Int64Var(&%[2]s, \"%[3]s\", %[2]s, \"%[4]s\")\n", flagPkg, name, fldName, desc)
		case stdtypes.Float64:
			g.W("%[1]s.Float64Var(&%[2]s, \"%[3]s\", %[2]s, \"%[4]s\")\n", flagPkg, name, fldName, desc)
		case stdtypes.Bool:
			g.W("%[1]s.BoolVar(&%[2]s, \"%[3]s\", %[2]s, \"%[4]s\")\n", flagPkg, name, fldName, desc)
		}
	}
}

func (g *config) writeConfigBasic(name, fldName string, f *stdtypes.Var, desc string, isFlag bool) {
	if !isFlag {
		tmpVar := strcase.ToLowerCamel(name) + "Tmp"
		g.W("%s, ok := %s.LookupEnv(%s)\n", tmpVar, g.i.Import("os", "os"), strconv.Quote(fldName))
		g.W("if ok {\n")
		g.WriteConvertType(g.i.Import, name, tmpVar, f, "errs", false, "convert "+fldName+" error")
		g.W("}\n")
	} else {
		g.writeConfigFlagBasic(name, fldName, desc, f)
	}
}

func (g *config) writeConfigStruct(name string, envParentName string, st *stdtypes.Struct, requiredAll bool, sliceErr string) (envs []env) {
	for i := 0; i < st.NumFields(); i++ {
		f := st.Field(i)

		isFlag := false
		fldName := strcase.ToScreamingSnake(strings.NormalizeCamelCase(f.Name()))
		required := requiredAll
		parentPostfix := "_"
		desc := ""

		if tags, err := structtag.Parse(st.Tag(i)); err == nil {
			if tag, err := tags.Get("desc"); err == nil {
				desc = tag.Name
			}
			if tag, err := tags.Get("env"); err == nil {
				required = tag.HasOption("required")
				if tag.Name != "" {
					fldName = tag.Name
				}
			}
			if tag, err := tags.Get("flag"); err == nil {
				required = tag.HasOption("required")
				if tag.Name != "" {
					isFlag = true
					fldName = tag.Name
				}
			}
		}

		if envParentName != "" {
			fldName = envParentName + parentPostfix + fldName
		}

		assignID := name + "." + f.Name()

		switch v := f.Type().Underlying().(type) {
		case *stdtypes.Pointer:
			if v.Elem().String() == "net/url.URL" {
				g.writeConfigBasic(assignID, fldName, f, desc, isFlag)
			} else {
				if st, ok := v.Elem().Underlying().(*stdtypes.Struct); ok {
					if named, ok := v.Elem().(*stdtypes.Named); ok {
						g.W("%s = &%s{}\n", assignID, named.Obj().Name())
					}
					envs = append(envs, g.writeConfigStruct(assignID, fldName, st, required, sliceErr)...)
					required = false // reset check empty for struct because check is generated in writeConfigStruct
				}
			}
		case *stdtypes.Struct:
			if named, ok := f.Type().(*stdtypes.Named); ok {
				g.W("%s = %s{}\n", assignID, named.Obj().Name())
			}
			envs = append(envs, g.writeConfigStruct(assignID, fldName, v, required, sliceErr)...)
			required = false // reset check empty for struct because check is generated in writeConfigStruct
		case *stdtypes.Basic, *stdtypes.Slice:
			envs = append(envs, env{name: fldName, fName: f.Name(), assignID: assignID, desc: desc, required: required, isFlag: isFlag})
			g.writeConfigBasic(assignID, fldName, f, desc, isFlag)
		}

		if required {
			tagName := "env"
			if isFlag {
				tagName = "flag"
			}

			errorsPkg := g.i.Import("errors", "errors")

			g.W("if %s == %s {\n", assignID, types.ZeroValue(f.Type()))

			requiredMsg := strconv.Quote(fmt.Sprintf("%s %s required", tagName, fldName))

			if sliceErr != "" {
				g.W("%[1]s = append(%[1]s, %[2]s.New(%[3]s))\n ", sliceErr, errorsPkg, requiredMsg)
			} else {
				g.W("return nil, %s.New(%s)\n", errorsPkg, requiredMsg)
			}
			g.W("}\n")
		}
	}
	return
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

func NewConfig(o model.ConfigOption) *config {
	return &config{GoLangWriter: writer.NewGoLangWriter(), o: o}
}
