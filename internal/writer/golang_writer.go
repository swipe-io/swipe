package writer

import (
	"fmt"
	stdtypes "go/types"
	"strconv"
	stdstrings "strings"

	"github.com/swipe-io/swipe/v2/internal/strings"
	"github.com/swipe-io/swipe/v2/internal/types"
)

type GoLangWriter struct {
	BaseWriter
}

func (w *GoLangWriter) WriteCheckErr(body func()) {
	w.W("if err != nil {\n")
	body()
	w.W("}\n")
}

func (w *GoLangWriter) WriteType(name string) {
	w.W("type %s ", name)
}

func (w *GoLangWriter) WriteFunc(name, recv string, params, results []string, body func()) {
	w.W("func")

	if recv != "" {
		w.W(" (%s)", recv)
	}

	w.W(" %s(", name)
	w.WriteSignature(params)
	w.W(") ")

	if len(results) > 0 {
		w.W("( ")
		w.WriteSignature(results)
		w.W(") ")
	}

	w.W("{\n")
	body()
	w.W("}\n\n")
}

func (w *GoLangWriter) WriteDefer(params []string, calls []string, body func()) {
	w.W("defer func(")
	w.W(stdstrings.Join(params, ","))
	w.W(") {\n")
	body()
	w.W("}(")
	w.W(stdstrings.Join(calls, ","))
	w.W(")\n")
}

func (w *GoLangWriter) WriteSignature(keyvals []string) {
	if len(keyvals) == 0 {
		return
	}
	if len(keyvals)%2 != 0 {
		panic("WriteSignature: missing Value")
	}
	for i := 0; i < len(keyvals); i += 2 {
		if i > 0 {
			w.W(", ")
		}
		name := "_"
		if keyvals[i] != "" {
			name = keyvals[i]
		}
		w.W("%s %s", name, keyvals[i+1])
	}
}

func (w *GoLangWriter) WriteTypeStruct(name string, keyvals []string) {
	w.WriteType(name)
	w.WriteStruct(keyvals, false)
	w.Line()
	w.Line()
}

func (w *GoLangWriter) WriteStruct(keyvals []string, assign bool) {
	w.W(" struct ")
	if assign {
		w.WriteStructAssign(keyvals)
	} else {
		w.WriteStructDefined(keyvals)
	}
}

func (w *GoLangWriter) WriteStructDefined(keyvals []string) {
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

func (w *GoLangWriter) WriteStructAssign(keyvals []string) {
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

func (w *GoLangWriter) WriteFuncCall(id, name string, params []string) {
	w.W(id + "." + name + "(")
	w.W(stdstrings.Join(params, ","))
	w.W(")\n")
}

func (w *GoLangWriter) getConvertFunc(funcName, strconvPkg, tmpId, valueId string) string {
	return fmt.Sprintf("%s, err := %s.%s", tmpId, strconvPkg, fmt.Sprintf(funcName, valueId))
}

func (w *GoLangWriter) getConvertFuncName(kind stdtypes.BasicKind) string {
	switch kind {
	case stdtypes.Int, stdtypes.Int8, stdtypes.Int16, stdtypes.Int32, stdtypes.Int64:
		return "Atoi(%s)"
	case stdtypes.Float32, stdtypes.Float64:
		return "ParseFloat(%s, " + types.GetBitSize(kind) + ")"
	case stdtypes.Uint, stdtypes.Uint8, stdtypes.Uint16, stdtypes.Uint32, stdtypes.Uint64:
		return "ParseUint(%s, 10, " + types.GetBitSize(kind) + ")"
	case stdtypes.Bool:
		return "ParseBool(%s)"
	default:
		return ""
	}
}

func (w *GoLangWriter) getFormatFuncName(kind stdtypes.BasicKind) string {
	switch kind {
	case stdtypes.Int, stdtypes.Int8, stdtypes.Int16, stdtypes.Int32, stdtypes.Int64:
		return "FormatInt(int64(%s), 10)"
	case stdtypes.Float32, stdtypes.Float64:
		return "FormatFloat(float64(%s), 'g', -1, " + types.GetBitSize(kind) + ")"
	case stdtypes.Uint, stdtypes.Uint8, stdtypes.Uint16, stdtypes.Uint32, stdtypes.Uint64:
		return "FormatUint(uint64(%s), 10)"
	case stdtypes.Bool:
		return "FormatBool(%s)"
	default:
		return ""
	}
}

func (w *GoLangWriter) GetFormatType(importFn func(string, string) string, valueId string, f *stdtypes.Var) string {
	switch t := f.Type().(type) {
	case *stdtypes.Basic:
		funcName := w.getFormatFuncName(t.Kind())
		if funcName != "" {
			return fmt.Sprintf("%s.%s", importFn("strconv", "strconv"), fmt.Sprintf(funcName, valueId))
		}
	}
	return valueId
}

func (w *GoLangWriter) writeConvertBasicType(importFn func(string, string) string, name, assignId, valueId string, t *stdtypes.Basic, sliceErr string, declareVar bool, msgErrTemplate string) {
	useCheckErr := true

	fmtPkg := importFn("fmt", "fmt")
	tmpId := stdstrings.ToLower(name) + strings.UcFirst(t.String())

	funcName := w.getConvertFuncName(t.Kind())
	if funcName != "" {
		w.W("%s\n", w.getConvertFunc(funcName, importFn("strconv", "strconv"), tmpId, valueId))
	} else {
		useCheckErr = false
		tmpId = valueId
	}
	if useCheckErr {
		if msgErrTemplate == "" {
			msgErrTemplate = "convert error"
		}

		errMsg := strconv.Quote(msgErrTemplate + ": %w")
		w.W("if err != nil {\n")
		if sliceErr == "" {
			w.W("return nil, %s.Errorf(%s, err)\n", fmtPkg, errMsg)
		} else {
			w.W("%[1]s = append(%[1]s, %s.Errorf(%s, err))\n", sliceErr, fmtPkg, errMsg)
		}
		w.W("}\n")
	}

	if declareVar {
		w.W("var ")
	}

	w.W("%s = ", assignId)
	if t.Kind() != stdtypes.String {
		w.W("%s(%s)", t.String(), tmpId)
	} else {
		w.W("%s", tmpId)
	}
	w.W("\n")
}

func (w *GoLangWriter) WriteConvertType(
	importFn func(string, string) string, assignId, valueId string, f *stdtypes.Var, sliceErr string, declareVar bool, msgErrTemplate string,
) {
	var tmpId string

	switch t := f.Type().(type) {
	case *stdtypes.Basic:
		w.writeConvertBasicType(importFn, f.Name(), assignId, valueId, t, sliceErr, declareVar, msgErrTemplate)
	case *stdtypes.Slice:
		stringsPkg := importFn("strings", "strings")
		switch t := t.Elem().(type) {
		case *stdtypes.Basic:
			switch t.Kind() {
			case stdtypes.String:
				w.W("%s = %s.Split(%s, \",\")\n", assignId, stringsPkg, valueId)
			case stdtypes.Uint,
				stdtypes.Uint8,
				stdtypes.Uint16,
				stdtypes.Uint32,
				stdtypes.Uint64,
				stdtypes.Int,
				stdtypes.Int8,
				stdtypes.Int16,
				stdtypes.Int32,
				stdtypes.Int64,
				stdtypes.Float32,
				stdtypes.Float64:
				tmpId = "parts" + stdstrings.ToLower(f.Name()) + strings.UcFirst(t.String())
				w.W("%s := %s.Split(%s, \",\")\n", tmpId, stringsPkg, valueId)
				if declareVar {
					w.W("var ")
				}
				w.W("%s = make([]%s, len(%s))\n", assignId, t.String(), tmpId)
				w.W("for i, s := range %s {\n", tmpId)
				w.writeConvertBasicType(importFn, "tmp", assignId+"[i]", "s", t, sliceErr, false, msgErrTemplate)
				w.W("}\n")
			}
		}
	case *stdtypes.Pointer:
		if t.Elem().String() == "net/url.URL" {
			urlPkg := importFn("url", "net/url")
			tmpId := stdstrings.ToLower(f.Name()) + "URL"
			w.W("%s, err := %s.Parse(%s)\n", tmpId, urlPkg, valueId)
			w.W("if err != nil {\n")
			if sliceErr == "" {
				w.W("return nil, err\n")
			} else {
				w.W("%[1]s = append(%[1]s, err)\n", sliceErr)
			}
			w.W("}\n")
			if declareVar {
				w.W("var ")
			}
			w.W("%s = %s\n", assignId, tmpId)
		}
	case *stdtypes.Named:
		if t.Obj().Pkg().Path() == "github.com/satori/go.uuid" {
			uuidPkg := importFn("", t.Obj().Pkg().Path())
			if declareVar {
				w.W("var ")
			}
			w.W("%s, err = %s.FromString(%s)\n", assignId, uuidPkg, valueId)
			w.W("if err != nil {\n")
			if sliceErr == "" {
				w.W("return nil, err\n")
			} else {
				w.W("%[1]s = append(%[1]s, err)\n", sliceErr)
			}
			w.W("}\n")
		}
	}
}
