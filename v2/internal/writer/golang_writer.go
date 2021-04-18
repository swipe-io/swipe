package writer

import (
	"fmt"
	stdtypes "go/types"
	"strconv"
	stdstrings "strings"

	"github.com/swipe-io/strcase"
	"github.com/swipe-io/swipe/v2/internal/option"
	"github.com/swipe-io/swipe/v2/internal/swipe"
)

type GoWriter struct {
	BaseWriter
}

func (w *GoWriter) WriteCheckErr(body func()) {
	w.W("if err != nil {\n")
	body()
	w.W("}\n")
}

func (w *GoWriter) WriteType(name string) {
	w.W("type %s ", name)
}

func (w *GoWriter) WriteFunc(name, recv string, params, results []string, body func()) {
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

func (w *GoWriter) WriteVarGroup(body func()) {
	w.W("var (\n")
	body()
	w.W("\n)\n")
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

func (w *GoWriter) WriteSignature(keyvals []string) {
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

func (w *GoWriter) WriteFuncCall(id, name string, params []string) {
	w.W(id + "." + name + "(")
	w.W(stdstrings.Join(params, ","))
	w.W(")\n")
}

func (w *GoWriter) getConvertFuncName(t *option.BasicType) string {
	if t.IsAnyInt() {
		return "Atoi(%s)"
	}
	if t.IsAnyUint() {
		return "ParseUint(%s, 10, " + t.BitSize() + ")"
	}
	if t.IsAnyFloat() {
		return "ParseFloat(%s, " + t.BitSize() + ")"
	}
	if t.IsBool() {
		return "ParseBool(%s)"
	}
	return ""
}

func (w *GoWriter) getFormatFuncName(t *option.BasicType) string {
	if t.IsAnyInt() {
		return "FormatInt(int64(%s), 10)"
	}
	if t.IsAnyUint() {
		return "FormatUint(uint64(%s), 10)"
	}
	if t.IsAnyFloat() {
		return "FormatFloat(float64(%s), 'g', -1, " + t.BitSize() + ")"
	}
	if t.IsBool() {
		return "FormatBool(%s)"
	}
	return ""
}

func (w *GoWriter) writeFormatBasicType(i swipe.Importer, assignId, valueId string, t *option.BasicType) {
	funcName := w.getFormatFuncName(t)
	if funcName != "" {
		strconvPkg := i.Import("strconv", "strconv")
		w.W("%s := %s.%s\n", assignId, strconvPkg, fmt.Sprintf(funcName, valueId))
	} else {
		w.W("%s := %s\n", assignId, valueId)
	}
}

func (w *GoWriter) writeConvertBasicType(i swipe.Importer, name, assignId, valueId string, t *option.BasicType, errRet []string, errSlice string, declareVar bool, msgErrTemplate string) {
	useCheckErr := true

	tmpId := stdstrings.ToLower(name) + strcase.ToCamel(t.Name)

	funcName := w.getConvertFuncName(t)
	if funcName != "" {
		strconvPkg := i.Import("strconv", "strconv")
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
			w.W("%[1]s = append(%[1]s, %[2]s.Errorf(%[3]s, err))\n", errSlice, i.Import("fmt", "fmt"), errMsg)
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
	if t.IsString() {
		w.W("%s(%s)", t.Name, tmpId)
	} else {
		w.W("%s", tmpId)
	}
	w.W("\n")
}

func (w *GoWriter) WriteFormatType(i swipe.Importer, assignId, valueId string, f *option.VarType) {
	switch t := f.Type.(type) {
	case *option.BasicType:
		w.writeFormatBasicType(i, assignId, valueId, t)
		//case *stdtypes.Named:
		//	switch t.Obj().Type().String() {
		//	case "github.com/google/uuid.UUID", "github.com/satori/uuid.UUID":
		//		w.W("%s := %s.String() \n", assignId, valueId)
		//	case "time.Duration":
		//		w.W("%s := %s.String()\n", assignId, valueId)
		//	case "time.Time":
		//		timePkg := i.Import("time", "time")
		//		w.W("%[1]s := %[3]s.Format(%[2]s.RFC3339)\n", assignId, timePkg, valueId)
		//	}
	}
}

func (w *GoWriter) WriteConvertType(
	i swipe.Importer, assignId, valueId string, f *option.VarType, errRet []string, errSlice string, declareVar bool, msgErrTemplate string,
) {
	var tmpId string

	switch t := f.Type.(type) {
	case *option.BasicType:
		w.writeConvertBasicType(i, f.Name.Origin, assignId, valueId, t, errRet, errSlice, declareVar, msgErrTemplate)
	case *option.MapType:
		stringsPkg := i.Import("strings", "strings")

		if k, ok := t.Key.(*option.BasicType); ok && k.IsString() {
			if v, ok := t.Value.(*option.BasicType); ok {
				tmpId = "parts" + f.Name.LowerCase
				w.W("%s := %s.Split(%s, \",\")\n", tmpId, stringsPkg, valueId)
				w.W("%s = make(%s, len(%s))\n", assignId, k.Name, tmpId)
				if v.IsNumeric() {
					w.W("for _, s := range %s {\n", tmpId)
					w.W("kv := %s.Split(s, \"=\")\n", stringsPkg)
					w.W("if len(kv) == 2 {\n")
					w.writeConvertBasicType(i, "tmp", assignId+"[kv[0]]", "kv[1]", v, errRet, errSlice, false, msgErrTemplate)
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
	case *option.SliceType:
		stringsPkg := i.Import("strings", "strings")
		switch t := t.Value.(type) {
		case *option.BasicType:
			if t.IsNumeric() {
				tmpId = "parts" + f.Name.LowerCase + strcase.ToCamel(t.Name)
				w.W("%s := %s.Split(%s, \",\")\n", tmpId, stringsPkg, valueId)
				if declareVar {
					w.W("var ")
				}
				w.W("%s = make([]%s, len(%s))\n", assignId, t.Name, tmpId)
				w.W("for i, s := range %s {\n", tmpId)
				w.writeConvertBasicType(i, "tmp", assignId+"[i]", "s", t, errRet, errSlice, false, msgErrTemplate)
				w.W("}\n")
			} else {
				w.W("%s = %s.Split(%s, \",\")\n", assignId, stringsPkg, valueId)
			}
		}
	case *stdtypes.Pointer:
		if t.Elem().String() == "net/url.URL" {
			urlPkg := i.Import("url", "net/url")
			tmpID := f.Name.LowerCase + "URL"
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
		tmpID := f.Name.LowerCase + "Result"

		switch t.Obj().Type().String() {
		case "github.com/satori/uuid.UUID":
			uuidPkg := i.Import(t.Obj().Pkg().Name(), t.Obj().Pkg().Path())
			w.W("%s, err := %s.FromString(%s)\n", tmpID, uuidPkg, valueId)
		case "github.com/google/uuid.UUID":
			uuidPkg := i.Import(t.Obj().Pkg().Name(), t.Obj().Pkg().Path())
			w.W("%s, err := %s.Parse(%s)\n", tmpID, uuidPkg, valueId)
		case "time.Duration":
			timePkg := i.Import("time", "time")
			w.W("%s, err := %s.ParseDuration(%s)\n", tmpID, timePkg, valueId)
		case "time.Time":
			timePkg := i.Import("time", "time")
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
