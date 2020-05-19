package config

import (
	"fmt"
	"go/types"
	"strconv"

	"github.com/fatih/structtag"
	"github.com/iancoleman/strcase"

	"github.com/swipe-io/swipe/pkg/parser"
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

type Config struct {
	w *writer.Writer
}

func (c *Config) Write(opt *parser.Option) error {
	fmtPkg := c.w.Import("fmt", "fmt")

	structOpt := parser.MustOption(opt.Get("optionsStruct"))

	var (
		strt *types.Struct
	)

	if ptr, ok := structOpt.Value.Type().(*types.Pointer); ok {
		strt = ptr.Elem().Underlying().(*types.Struct)
	} else {
		strt = structOpt.Value.Type().(*types.Struct)
	}

	var funcName = "LoadConfig"
	if funcNameOpt, ok := opt.Get("FuncName"); ok {
		funcName = funcNameOpt.Value.String()
	}

	c.w.Write("func %s() (cfg %s, errs []error) {\n", funcName, c.w.TypeString(structOpt.Value.Type()))
	c.w.Write("cfg = ")
	c.w.WriteAST(structOpt.Value.Expr())
	c.w.Write("\n")

	envs := c.writeConfigStruct("cfg", "", strt, false, "errs")

	flagFound := false
	for _, env := range envs {
		if env.isFlag {
			flagFound = true
			break
		}
	}
	if flagFound {
		c.w.Write("%s.Parse()\n", c.w.Import("flag", "flag"))
	}

	c.w.Write("return\n")
	c.w.Write("}\n\n")

	c.w.Write("func (cfg %s) String() string {\n", c.w.TypeString(structOpt.Value.Type()))
	c.w.Write("out := `\n")
	for _, env := range envs {
		if env.isFlag {
			c.w.Write("--%s ", env.name)
		} else {
			c.w.Write("%s=", env.name)
		}
		c.w.Write("`+%s.Sprintf(\"%%v\", %s)+`", fmtPkg, env.assignID)
		if env.desc != "" {
			c.w.Write(" ;%s", env.desc)
		}
		c.w.WriteLn()
	}
	c.w.Write("`\n")
	c.w.Write("return out\n}\n\n")

	return nil
}

func (c *Config) writeConfigFlagBasic(name, fldName string, descr string, f *types.Var) {
	if t, ok := f.Type().(*types.Basic); ok {
		flagPkg := c.w.Import("flag", "flag")
		switch t.Kind() {
		case types.String:
			c.w.Write("%s.StringVar(&%s, \"%s\", \"\", \"%s\")\n", flagPkg, name, fldName, descr)
		case types.Int:
			c.w.Write("%s.IntVar(&%s, \"%s\", 0, \"%s\")\n", flagPkg, name, fldName, descr)
		case types.Int64:
			c.w.Write("%s.Int64Var(&%s, \"%s\", 0, \"%s\")\n", flagPkg, name, fldName, descr)
		case types.Float64:
			c.w.Write("%s.Float64Var(&%s, \"%s\", 0, \"%s\")\n", flagPkg, name, fldName, descr)
		case types.Bool:
			c.w.Write("%s.BoolVar(&%s, \"%s\", false, \"%s\")\n", flagPkg, name, fldName, descr)
		}
	}
}

func (c *Config) writeConfigBasic(name, fldName string, f *types.Var, descr string, isFlag bool) {
	if !isFlag {
		tmpVar := strcase.ToLowerCamel(name) + "Tmp"
		c.w.Write("%s, ok := %s.LookupEnv(%s)\n", tmpVar, c.w.Import("os", "os"), strconv.Quote(fldName))
		c.w.Write("if ok {\n")
		c.w.WriteConvertType(name, tmpVar, f, "errs", false)
		c.w.Write("}\n")
	} else {
		c.writeConfigFlagBasic(name, fldName, descr, f)
	}
}

func (c *Config) writeConfigStruct(name string, envParentName string, st *types.Struct, requiredAll bool, sliceErr string) (envs []env) {
	for i := 0; i < st.NumFields(); i++ {
		f := st.Field(i)

		isFlag := false
		fldName := strcase.ToScreamingSnake(normalizeCamelCase(f.Name()))
		required := requiredAll
		parentPostfix := "_"
		descr := ""

		if tags, err := structtag.Parse(st.Tag(i)); err == nil {
			if tag, err := tags.Get("desc"); err == nil {
				descr = tag.Name
			}
			if tag, err := tags.Get("env"); err == nil {
				required = tag.HasOption("required")
				if tag.Name != "" {
					fldName = tag.Name
				}
			}
			if tag, err := tags.Get("flag"); err == nil {
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
		case *types.Pointer:
			if v.Elem().String() == "net/url.URL" {
				c.writeConfigBasic(assignID, fldName, f, descr, isFlag)
			} else {
				if st, ok := v.Elem().Underlying().(*types.Struct); ok {
					if named, ok := v.Elem().(*types.Named); ok {
						c.w.Write("%s = &%s{}\n", assignID, named.Obj().Name())
					}
					envs = append(envs, c.writeConfigStruct(assignID, fldName, st, required, sliceErr)...)
					required = false // reset check empty for struct because check is generated in writeConfigStruct
				}
			}
		case *types.Struct:
			if named, ok := f.Type().(*types.Named); ok {
				c.w.Write("%s = %s{}\n", assignID, named.Obj().Name())
			}
			envs = append(envs, c.writeConfigStruct(assignID, fldName, v, required, sliceErr)...)
			required = false // reset check empty for struct because check is generated in writeConfigStruct
		case *types.Basic, *types.Slice:
			envs = append(envs, env{name: fldName, fName: f.Name(), assignID: assignID, desc: descr, required: required, isFlag: isFlag})
			c.writeConfigBasic(assignID, fldName, f, descr, isFlag)
		}

		if required {
			errorsPkg := c.w.Import("errors", "errors")

			c.w.Write("if %s == %s {\n", assignID, c.w.ZeroValue(f.Type()))

			requiredMsg := strconv.Quote(fmt.Sprintf("env %s required", fldName))

			if sliceErr != "" {
				c.w.Write("%[1]s = append(%[1]s, %[2]s.New(%[3]s))\n ", sliceErr, errorsPkg, requiredMsg)
			} else {
				c.w.Write("return nil, %s.New(%s)\n", errorsPkg, requiredMsg)
			}
			c.w.Write("}\n")
		}
	}
	return
}

func New(w *writer.Writer) *Config {
	return &Config{w: w}
}
