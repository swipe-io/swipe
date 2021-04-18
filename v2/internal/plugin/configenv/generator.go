package configenv

import (
	"fmt"
	"strconv"

	option2 "github.com/swipe-io/swipe/v2/internal/option"

	"github.com/swipe-io/strcase"

	"github.com/swipe-io/swipe/v2/internal/swipe"
	"github.com/swipe-io/swipe/v2/internal/writer"
)

type Generator struct {
	writer.GoWriter
	Struct   *option2.StructType
	FuncName string
}

func (g *Generator) Generate(i swipe.Importer) []byte {
	pkgName := i.Import(g.Struct.Pkg.Name, g.Struct.Pkg.Path)
	typeName := pkgName + g.Struct.Name.UpperCase

	g.W("func %s() (c *%s, errs []error) {\n", g.FuncName, typeName)
	g.W("c = &%s{}\n", typeName)
	g.W("}\n")

	walkStruct(g.Struct, func(f, parent *option2.VarType, opts fldOpts) {

		if parent != nil {
			fmt.Println(parent.Name)
		}
		fmt.Println(f.Name)

	})
	return g.Bytes()
}

func (g *Generator) writeEnv(i swipe.Importer, f *option2.VarType, opts fldOpts) {
	tmpVar := strcase.ToLowerCamel(opts.fieldPath) + "Tmp"
	g.W("%s, ok := %sLookupEnv(%s)\n", tmpVar, i.Import("os", "os"), strconv.Quote(opts.name))
	g.W("if ok {\n")

	g.WriteConvertType(i, "cfg."+opts.fieldPath, tmpVar, f, nil, "errs", false, "convert "+opts.name+" error")
	g.writeCheckZero(i, f, opts)

	g.W("}")
	if opts.required {
		g.W(" else {\n")
		g.writeAppendErr(i, opts)
		g.W("}\n")
	} else {
		g.W("\n")
	}
}

func (g *Generator) writeCheckZero(i swipe.Importer, f *option2.VarType, opts fldOpts) {
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

func (g *Generator) OutputDir() string {
	return ""
}

func (g *Generator) Filename() string {
	return "config.go"
}

func walkStructRecursive(st *option2.StructType, parent *option2.VarType, fPOpts fldOpts, fn func(f, parent *option2.VarType, opts fldOpts)) {
	for _, field := range st.Fields {
		fOpts := getFieldOpts(field.Var, field.Tags)
		if fPOpts.name != "" && parent != nil {
			fOpts.name = fPOpts.name + "_" + fOpts.name
			fOpts.fieldPath = fPOpts.fieldPath + "." + fOpts.fieldPath
		}

		if v, ok := field.Var.Type.(*option2.StructType); ok {
			walkStructRecursive(v, field.Var, fOpts, fn)
			continue
		}

		fn(field.Var, parent, fOpts)
	}
}

func walkStruct(st *option2.StructType, fn func(f, parent *option2.VarType, opts fldOpts)) {
	walkStructRecursive(st, nil, fldOpts{}, fn)
}
