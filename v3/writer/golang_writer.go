package writer

import (
	stdstrings "strings"

	"github.com/swipe-io/swipe/v3/option"
	"github.com/swipe-io/swipe/v3/swipe"
)

type GoWriter struct {
	TextWriter
}

func (w *GoWriter) WriteCheckErr(errName string, body func()) {
	w.W("if %s != nil {\n", errName)
	body()
	w.W("}\n")
}

func (w *GoWriter) WriteType(name string) {
	w.W("type %s ", name)
}

func (w *GoWriter) WriteDefer(params []string, calls []string, body func()) {
	w.W("defer func(")
	w.W(stdstrings.Join(params, ","))
	w.W(") {\n")
	body()
	w.W("}(")
	w.W(stdstrings.Join(calls, ","))
	w.W(")\n")
}

func (w *GoWriter) WriteTypeStruct(name string, keyvals []string) {
	w.WriteType(name)
	w.WriteStruct(keyvals, false)
	w.Line()
	w.Line()
}

func (w *GoWriter) WriteStruct(keyvals []string, assign bool) {
	w.W(" struct ")
	if assign {
		w.WriteStructAssign(keyvals)
	} else {
		w.WriteStructDefined(keyvals)
	}
}

func (w *GoWriter) WriteStructDefined(keyvals []string) {
	if len(keyvals)%2 != 0 {
		panic("WriteStructDefined: missing Value")
	}
	w.W("{\n")
	for i := 0; i < len(keyvals); i += 2 {
		w.W("%s %s\n", keyvals[i], keyvals[i+1])
		continue
	}
	w.W("}")
}

func (w *GoWriter) WriteStructAssign(keyvals []string) {
	if len(keyvals)%2 != 0 {
		panic("WriteStructAssign: missing Value")
	}
	w.W("{")
	for i := 0; i < len(keyvals); i += 2 {
		if i > 0 {
			w.W(", ")
		}
		w.W("%s: %s", keyvals[i], keyvals[i+1])
	}
	w.W("}")
}

func (w *GoWriter) WriteFuncByFuncType(f *option.FuncType, importer swipe.Importer) {
	pkg := importer.Import(f.Pkg.Name, f.Pkg.Path)
	if pkg != "" {
		pkg += "."
	}
	w.W("%s%s", pkg, f.Name)
}

func (w *GoWriter) WriteFuncCallByFuncType(f *option.FuncType, params []string, importer swipe.Importer) {
	w.WriteFuncByFuncType(f, importer)
	w.W("(%s)", stdstrings.Join(params, ","))
}

func (w *GoWriter) WriteFuncCall(id, name string, params []string) {
	w.W(id + "." + name + "(")
	w.W(stdstrings.Join(params, ","))
	w.W(")\n")
}
