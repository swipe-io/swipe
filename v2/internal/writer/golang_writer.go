package writer

import (
	"fmt"
	"strconv"
	stdstrings "strings"

	"github.com/swipe-io/strcase"
	"github.com/swipe-io/swipe/v2/internal/option"
	"github.com/swipe-io/swipe/v2/internal/swipe"
)

type GoWriter struct {
	BaseWriter
}

func (w *GoWriter) WriteCheckErr(errName string, body func()) {
	w.W("if %s != nil {\n", errName)
	body()
	w.W("}\n")
}

func (w *GoWriter) WriteType(name string) {
	w.W("type %s ", name)
}

//func (w *GoWriter) WriteFunc(name, recv string, params, results []string, body func()) {
//	w.W("func")
//
//	if recv != "" {
//		w.W(" (%s)", recv)
//	}
//
//	w.W(" %s(", name)
//	w.WriteSignature(params)
//	w.W(") ")
//
//	if len(results) > 0 {
//		w.W("( ")
//		w.WriteSignature(results)
//		w.W(") ")
//	}
//
//	w.W("{\n")
//	body()
//	w.W("}\n\n")
//}

//func (w *GoWriter) WriteVarGroup(body func()) {
//	w.W("var (\n")
//	body()
//	w.W("\n)\n")
//}

func (w *GoWriter) WriteDefer(params []string, calls []string, body func()) {
	w.W("defer func(")
	w.W(stdstrings.Join(params, ","))
	w.W(") {\n")
	body()
	w.W("}(")
	w.W(stdstrings.Join(calls, ","))
	w.W(")\n")
}

//func (w *GoWriter) WriteSignature(keyvals []string) {
//	if len(keyvals) == 0 {
//		return
//	}
//	if len(keyvals)%2 != 0 {
//		panic("WriteSignature: missing Value")
//	}
//	for i := 0; i < len(keyvals); i += 2 {
//		if i > 0 {
//			w.W(", ")
//		}
//		name := ""
//		if keyvals[i] != "" {
//			name = keyvals[i]
//		}
//		w.W("%s %s", name, keyvals[i+1])
//	}
//}

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

func (w *GoWriter) writeFormatBasicType(importer swipe.Importer, assignId, valueId string, t *option.BasicType) {
	funcName := w.getFormatFuncName(t)
	if funcName != "" {
		strconvPkg := importer.Import("strconv", "strconv")
		w.W("%s := %s.%s\n", assignId, strconvPkg, fmt.Sprintf(funcName, valueId))
	} else {
		w.W("%s := %s\n", assignId, valueId)
	}
}

func (w *GoWriter) WriteConvertBasicType(importer swipe.Importer, name, assignId, valueId string, t *option.BasicType, errRet []string, errSlice string, declareVar bool, msgErrTemplate string) {
	useCheckErr := true

	tmpId := stdstrings.ToLower(name) + strcase.ToCamel(t.Name)

	funcName := w.getConvertFuncName(t)
	if funcName != "" {
		strconvPkg := importer.Import("strconv", "strconv")
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
			w.W("%[1]s = append(%[1]s, %[2]s.Errorf(%[3]s, err))\n", errSlice, importer.Import("fmt", "fmt"), errMsg)
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

func (w *GoWriter) WriteFormatType(importer swipe.Importer, assignId, valueId string, f *option.VarType) {
	switch t := f.Type.(type) {
	case *option.BasicType:
		w.writeFormatBasicType(importer, assignId, valueId, t)
	case *option.NamedType:
		switch t.Pkg.Path {
		case "github.com/satori/uuid", "github.com/google/uuid":
			if t.Name.Origin == "UUID" {
				w.WriteFormatUUID(importer, t, assignId, valueId)
			}
		case "time":
			w.WriteFormatTime(importer, t, assignId, valueId)
		}
	}
}

func (w *GoWriter) WriteFormatTime(importer swipe.Importer, t *option.NamedType, assignId, valueId string) {
	switch t.Name.Origin {
	case "Duration":
		w.W("%s := %s.String()\n", assignId, valueId)
	case "Time":
		timePkg := importer.Import("time", "time")
		w.W("%[1]s := %[3]sFormat(%[2]s.RFC3339)\n", assignId, timePkg, valueId)
	}
}

func (w *GoWriter) WriteFormatUUID(_ swipe.Importer, t *option.NamedType, assignId, valueId string) {
	w.W("%s := %s.String() \n", assignId, valueId)
}

func (w *GoWriter) WriteConvertTime(importer swipe.Importer, t *option.NamedType, valueId string) (tmpID string) {
	switch t.Name.Origin {
	case "Time":
		tmpID = t.Name.LowerCase + "Time"
		timePkg := importer.Import("time", "time")
		w.W("%[1]s, err := %[2]s.Parse(%[2]s.RFC3339, %[3]s)\n", tmpID, timePkg, valueId)
	case "Duration":
		tmpID = t.Name.LowerCase + "Dur"
		timePkg := importer.Import("time", "time")
		w.W("%s, err := %s.ParseDuration(%s)\n", tmpID, timePkg, valueId)
	}
	return
}

func (w *GoWriter) WriteConvertUUID(importer swipe.Importer, t *option.NamedType, valueId string) (tmpID string) {
	tmpID = t.Name.LowerCase + "UUID"
	uuidPkg := importer.Import(t.Pkg.Name, t.Pkg.Path)

	switch t.Pkg.Path {
	case "github.com/google/uuid":
		w.W("%s, err := %sParse(%s)\n", tmpID, uuidPkg, valueId)
	case "github.com/satori/uuid":
		w.W("%s, err := %sFromString(%s)\n", tmpID, uuidPkg, valueId)
	}
	return
}

func (w *GoWriter) WriteConvertURL(importer swipe.Importer, t *option.NamedType, valueId string) (tmpID string) {
	switch t.Name.Origin {
	case "URL":
		tmpID = t.Name.LowerCase + "URL"
		urlPkg := importer.Import("url", "net/url")
		w.W("%s, err := %sParse(%s)\n", tmpID, urlPkg, valueId)
	}
	return
}

func (w *GoWriter) WriteConvertType(
	importer swipe.Importer, assignId, valueId string, f *option.VarType, errRet []string, errSlice string, declareVar bool, msgErrTemplate string,
) {
	var (
		tmpID string
	)

	switch t := f.Type.(type) {
	case *option.BasicType:
		w.WriteConvertBasicType(importer, f.Name.Origin, assignId, valueId, t, errRet, errSlice, declareVar, msgErrTemplate)
	case *option.MapType:
		stringsPkg := importer.Import("strings", "strings")

		if k, ok := t.Key.(*option.BasicType); ok && k.IsString() {
			if v, ok := t.Value.(*option.BasicType); ok {
				tmpID = "parts" + f.Name.LowerCase
				w.W("%s := %s.Split(%s, \",\")\n", tmpID, stringsPkg, valueId)
				w.W("%s = make(%s, len(%s))\n", assignId, k.Name, tmpID)
				if v.IsNumeric() {
					w.W("for _, s := range %s {\n", tmpID)
					w.W("kv := %s.Split(s, \"=\")\n", stringsPkg)
					w.W("if len(kv) == 2 {\n")
					w.WriteConvertBasicType(importer, "tmp", assignId+"[kv[0]]", "kv[1]", v, errRet, errSlice, false, msgErrTemplate)
					w.W("}\n")
					w.W("}\n")
				} else {
					w.W("for _, s := range %s {\n", tmpID)
					w.W("kv := %s.Split(s, \"=\")\n", stringsPkg)
					w.W("if len(kv) == 2 {\n")
					w.W("%s[kv[0]] = kv[1]\n", assignId)
					w.W("}\n")
					w.W("}\n")
				}
			}
		}
	case *option.SliceType:
		stringsPkg := importer.Import("strings", "strings")
		switch t := t.Value.(type) {
		case *option.BasicType:
			if t.IsNumeric() {
				tmpID = "parts" + f.Name.LowerCase + strcase.ToCamel(t.Name)
				w.W("%s := %s.Split(%s, \",\")\n", tmpID, stringsPkg, valueId)
				if declareVar {
					w.W("var ")
				}
				w.W("%s = make([]%s, len(%s))\n", assignId, t.Name, tmpID)
				w.W("for importer, s := range %s {\n", tmpID)
				w.WriteConvertBasicType(importer, "tmp", assignId+"[importer]", "s", t, errRet, errSlice, false, msgErrTemplate)
				w.W("}\n")
			} else {
				w.W("%s = %s.Split(%s, \",\")\n", assignId, stringsPkg, valueId)
			}
		}
	case *option.NamedType:
		switch t.Pkg.Path {
		case "net/url":
			tmpID = w.WriteConvertURL(importer, t, valueId)
		case "github.com/satori/uuid", "github.com/google/uuid":
			if t.Name.Origin == "UUID" {
				tmpID = w.WriteConvertUUID(importer, t, valueId)
			}
		case "time":
			tmpID = w.WriteConvertTime(importer, t, valueId)
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
