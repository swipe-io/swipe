package writer

import (
	"fmt"
	stdtypes "go/types"
	"strconv"
	stdstrings "strings"

	"github.com/swipe-io/strcase"

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

func (w *GoLangWriter) WriteVarGroup(body func()) {
	w.W("var (\n")
	body()
	w.W("\n)\n")
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
		name := ""
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

func (w *GoLangWriter) writeFormatBasicType(importFn func(string, string) string, assignId, valueId string, t *stdtypes.Basic) {
	funcName := w.getFormatFuncName(t.Kind())
	if funcName != "" {
		strconvPkg := importFn("strconv", "strconv")
		w.W("%s := %s.%s\n", assignId, strconvPkg, fmt.Sprintf(funcName, valueId))
	} else {
		w.W("%s := %s\n", assignId, valueId)
	}
}

func (w *GoLangWriter) writeConvertBasicType(importFn func(string, string) string, name, assignId, valueId string, t *stdtypes.Basic, errRet []string, errSlice string, declareVar bool, msgErrTemplate string) {
	useCheckErr := true

	tmpId := stdstrings.ToLower(name) + strcase.ToCamel(t.String())

	funcName := w.getConvertFuncName(t.Kind())
	if funcName != "" {
		strconvPkg := importFn("strconv", "strconv")
		w.W("%s, err := %s.%s\n", tmpId, strconvPkg, fmt.Sprintf(funcName, valueId))
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
		if errSlice != "" {
			w.W("%[1]s = append(%[1]s, %[2]s.Errorf(%[3]s, err))\n", errSlice, importFn("fmt", "fmt"), errMsg)
		} else {
			w.W("return ")
			if len(errRet) > 0 {
				w.W("%s, ", stdstrings.Join(errRet, ","))
			}
			w.W("err")
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

func (w *GoLangWriter) WriteFormatType(importFn func(string, string) string, assignId, valueId string, f *stdtypes.Var) {
	switch t := f.Type().(type) {
	case *stdtypes.Basic:
		w.writeFormatBasicType(importFn, assignId, valueId, t)
	case *stdtypes.Named:
		switch t.Obj().Type().String() {
		case "github.com/google/uuid.UUID", "github.com/satori/uuid.UUID":
			w.W("%s := %s.String() \n", assignId, valueId)
		case "time.Duration":
			w.W("%s := %s.String()\n", assignId, valueId)
		case "time.Time":
			timePkg := importFn("time", "time")
			w.W("%[1]s := %[3]s.Format(%[2]s.RFC3339)\n", assignId, timePkg, valueId)
		}
	}
}

func (w *GoLangWriter) WriteConvertType(
	importFn func(string, string) string, assignId, valueId string, f *stdtypes.Var, errRet []string, errSlice string, declareVar bool, msgErrTemplate string,
) {
	var tmpId string

	switch t := f.Type().(type) {
	case *stdtypes.Basic:
		w.writeConvertBasicType(importFn, f.Name(), assignId, valueId, t, errRet, errSlice, declareVar, msgErrTemplate)
	case *stdtypes.Map:
		stringsPkg := importFn("strings", "strings")
		if k, ok := t.Key().(*stdtypes.Basic); ok && k.Kind() == stdtypes.String {
			if v, ok := t.Elem().(*stdtypes.Basic); ok {
				tmpId = "parts" + stdstrings.ToLower(f.Name())
				w.W("%s := %s.Split(%s, \",\")\n", tmpId, stringsPkg, valueId)
				w.W("%s = make(%s, len(%s))\n", assignId, t.String(), tmpId)
				if isNumeric(v.Kind()) {
					w.W("for _, s := range %s {\n", tmpId)
					w.W("kv := %s.Split(s, \"=\")\n", stringsPkg)
					w.W("if len(kv) == 2 {\n")
					w.writeConvertBasicType(importFn, "tmp", assignId+"[kv[0]]", "kv[1]", v, errRet, errSlice, false, msgErrTemplate)
					w.W("}\n")
					w.W("}\n")
				} else {
					w.W("for _, s := range %s {\n", tmpId)
					w.W("kv := %s.Split(s, \"=\")\n", stringsPkg)
					w.W("if len(kv) == 2 {\n")
					w.W("%s[kv[0]] = kv[1]\n", assignId)
					w.W("}\n")
					w.W("}\n")
				}
			}
		}
	case *stdtypes.Slice:
		stringsPkg := importFn("strings", "strings")
		switch t := t.Elem().(type) {
		case *stdtypes.Basic:
			if isNumeric(t.Kind()) {
				tmpId = "parts" + stdstrings.ToLower(f.Name()) + strcase.ToCamel(t.String())
				w.W("%s := %s.Split(%s, \",\")\n", tmpId, stringsPkg, valueId)
				if declareVar {
					w.W("var ")
				}
				w.W("%s = make([]%s, len(%s))\n", assignId, t.String(), tmpId)
				w.W("for i, s := range %s {\n", tmpId)
				w.writeConvertBasicType(importFn, "tmp", assignId+"[i]", "s", t, errRet, errSlice, false, msgErrTemplate)
				w.W("}\n")
			} else {
				w.W("%s = %s.Split(%s, \",\")\n", assignId, stringsPkg, valueId)
			}
		}
	case *stdtypes.Pointer:
		if t.Elem().String() == "net/url.URL" {
			urlPkg := importFn("url", "net/url")
			tmpID := stdstrings.ToLower(f.Name()) + "URL"
			w.W("%s, err := %s.Parse(%s)\n", tmpID, urlPkg, valueId)
			w.W("if err != nil {\n")
			if errSlice != "" {
				w.W("%[1]s = append(%[1]s, err)\n", errSlice)
			} else {
				w.W("return ")
				if len(errRet) > 0 {
					w.W("%s, ", stdstrings.Join(errRet, ","))
				}
				w.W("err")
			}
			w.W("}\n")
			if declareVar {
				w.W("var ")
			}
			w.W("%s = %s\n", assignId, tmpID)
		}
	case *stdtypes.Named:
		tmpID := strcase.ToLowerCamel(f.Name()) + "Result"

		switch t.Obj().Type().String() {
		case "github.com/satori/uuid.UUID":
			uuidPkg := importFn(t.Obj().Pkg().Name(), t.Obj().Pkg().Path())
			w.W("%s, err := %s.FromString(%s)\n", tmpID, uuidPkg, valueId)
		case "github.com/google/uuid.UUID":
			uuidPkg := importFn(t.Obj().Pkg().Name(), t.Obj().Pkg().Path())
			w.W("%s, err := %s.Parse(%s)\n", tmpID, uuidPkg, valueId)
		case "time.Duration":
			timePkg := importFn("time", "time")
			w.W("%s, err := %s.ParseDuration(%s)\n", tmpID, timePkg, valueId)
		case "time.Time":
			timePkg := importFn("time", "time")
			w.W("%[1]s, err := %[2]s.Parse(%[2]s.RFC3339, %[3]s)\n", tmpID, timePkg, valueId)
		}
		w.W("if err != nil {\n")
		if errSlice != "" {
			w.W("%[1]s = append(%[1]s, err)\n", errSlice)
		} else {
			w.W("return ")
			if len(errRet) > 0 {
				w.W("%s, ", stdstrings.Join(errRet, ","))
			}
			w.W("err")
		}
		w.W("}\n")
		if declareVar {
			w.W("var ")
		}
		w.W("%s = %s\n", assignId, tmpID)
	}
}
