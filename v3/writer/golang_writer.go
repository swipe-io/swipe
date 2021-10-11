package writer

import (
	"fmt"
	"strconv"
	stdstrings "strings"

	"github.com/swipe-io/strcase"
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
		if t.IsPointer {
			valueId = "*" + valueId
		}
		w.W("%s := %s.%s\n", assignId, strconvPkg, fmt.Sprintf(funcName, valueId))
	} else {
		w.W("%s := %s\n", assignId, valueId)
	}
}

func (w *GoWriter) WriteConvertBasicType(importer swipe.Importer, name, assignId, valueId string, t *option.BasicType, errRet []string, errSlice string, declareVar bool, msgErrTemplate string) {
	useCheckErr := true

	tmpId := strcase.ToLowerCamel(name) + strcase.ToCamel(t.Name)

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
			w.W("%s.Errorf(\"convert error: %%v\", %s)", importer.Import("fmt", "fmt"), tmpId)
		}
		w.W("}\n")
	}
	if declareVar {
		w.W("var ")
	}
	retId := fmt.Sprintf("%s(%s)", t.Name, tmpId)
	//if t.IsString() {
	if t.IsPointer {
		ptrName := "ptr" + strcase.ToCamel(name)
		w.W("%s := %s\n", ptrName, tmpId)
		retId = "&" + ptrName
	}
	//}
	w.W("%s = %s", assignId, retId)
	w.W("\n")
}

func (w *GoWriter) WriteFormatType(importer swipe.Importer, assignId, valueId string, f *option.VarType) {
	switch t := f.Type.(type) {
	case *option.BasicType:
		w.writeFormatBasicType(importer, assignId, valueId, t)
	case *option.NamedType:
		switch t.Pkg.Path {
		case "github.com/satori/uuid", "github.com/google/uuid":
			if t.Name.Value == "UUID" {
				w.WriteFormatUUID(importer, t, assignId, valueId)
			}
		case "time":
			w.WriteFormatTime(importer, t, assignId, valueId)
		}
	}
}

func (w *GoWriter) WriteFormatTime(importer swipe.Importer, t *option.NamedType, assignId, valueId string) {
	switch t.Name.Value {
	case "Duration":
		w.W("%s := %s.String()\n", assignId, valueId)
	case "Time":
		timePkg := importer.Import("time", "time")
		w.W("%[1]s := %[3]s.Format(%[2]s.RFC3339)\n", assignId, timePkg, valueId)
	}
}

func (w *GoWriter) WriteFormatUUID(_ swipe.Importer, t *option.NamedType, assignId, valueId string) {
	w.W("%s := %s.String() \n", assignId, valueId)
}

func (w *GoWriter) WriteConvertTime(importer swipe.Importer, t *option.NamedType, valueId string) (tmpID string) {
	switch t.Name.Value {
	case "Time":
		tmpID = t.Name.Lower() + "Time"
		timePkg := importer.Import("time", "time")
		w.W("%[1]s, err := %[2]s.Parse(%[2]s.RFC3339, %[3]s)\n", tmpID, timePkg, valueId)
	case "Duration":
		tmpID = t.Name.Lower() + "Dur"
		timePkg := importer.Import("time", "time")
		w.W("%s, err := %s.ParseDuration(%s)\n", tmpID, timePkg, valueId)
	}
	return
}

func (w *GoWriter) WriteConvertUUID(importer swipe.Importer, t *option.NamedType, valueId string) (tmpID string) {
	tmpID = t.Name.Lower() + "UUID"
	uuidPkg := importer.Import(t.Pkg.Name, t.Pkg.Path)

	switch t.Pkg.Path {
	case "github.com/google/uuid":
		w.W("%s, err := %s.Parse(%s)\n", tmpID, uuidPkg, valueId)
	case "github.com/satori/uuid":
		w.W("%s, err := %s.FromString(%s)\n", tmpID, uuidPkg, valueId)
	}
	return
}

func (w *GoWriter) WriteConvertURL(importer swipe.Importer, t *option.NamedType, valueId string) (tmpID string) {
	switch t.Name.Value {
	case "URL":
		tmpID = t.Name.Lower() + "URL"
		urlPkg := importer.Import("url", "net/url")
		w.W("%s, err := %s.Parse(%s)\n", tmpID, urlPkg, valueId)
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
		w.WriteConvertBasicType(importer, f.Name.Value, assignId, valueId, t, errRet, errSlice, declareVar, msgErrTemplate)
	case *option.MapType:
		stringsPkg := importer.Import("strings", "strings")
		if k, ok := t.Key.(*option.BasicType); ok && k.IsString() {
			if v, ok := t.Value.(*option.BasicType); ok {
				tmpID = "parts" + f.Name.Lower()
				w.W("%s := %s.Split(%s, \",\")\n", tmpID, stringsPkg, valueId)
				w.W("%s = make(%s, len(%s))\n", assignId, swipe.TypeString(t, false, importer), tmpID)
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
				tmpID = "parts" + f.Name.Lower() + strcase.ToCamel(t.Name)
				w.W("%s := %s.Split(%s, \",\")\n", tmpID, stringsPkg, valueId)
				if declareVar {
					w.W("var ")
				}
				w.W("%s = make([]%s, len(%s))\n", assignId, t.Name, tmpID)
				w.W("for i, s := range %s {\n", tmpID)
				w.WriteConvertBasicType(importer, "tmp", assignId+"[i]", "s", t, errRet, errSlice, false, msgErrTemplate)
				w.W("}\n")
			} else {
				w.W("%s = %s.Split(%s, \",\")\n", assignId, stringsPkg, valueId)
			}
		}
	case *option.ArrayType:
		stringsPkg := importer.Import("strings", "strings")
		switch b := t.Value.(type) {
		case *option.BasicType:
			if b.IsNumeric() {
				tmpID = "parts" + f.Name.Lower() + strcase.ToCamel(b.Name)
				w.W("%s := %s.Split(%s, \",\")\n", tmpID, stringsPkg, valueId)
				if declareVar {
					w.W("var ")
				}
				w.W("if len(%s) > len(%s) {\n", tmpID, assignId)
				w.W("%[1]s = append(%[1]s, %[2]s.Errorf(%[3]s))\n", errSlice, importer.Import("fmt", "fmt"), strconv.Quote(msgErrTemplate+": array length must be less or equal "+strconv.FormatInt(t.Len, 10)))
				w.W("} else {\n")

				w.W("for i, s := range %s {\n", tmpID)
				w.WriteConvertBasicType(importer, "tmp", assignId+"[i]", "s", b, errRet, errSlice, false, msgErrTemplate)
				w.W("}\n")
				w.W("}\n")
			} else {
				w.W("%s = %s.Split(%s, \",\")\n", assignId, stringsPkg, valueId)
			}
		}
	case *option.NamedType:
		switch t.Pkg.Path {
		case "net/url":
			tmpID = w.WriteConvertURL(importer, t, valueId)
			if !t.IsPointer {
				tmpID = "*" + tmpID
			}
		case "github.com/satori/uuid", "github.com/google/uuid":
			if t.Name.Value == "UUID" {
				tmpID = w.WriteConvertUUID(importer, t, valueId)
				if t.IsPointer {
					tmpID = "&" + tmpID
				}
			}
		case "time":
			tmpID = w.WriteConvertTime(importer, t, valueId)
			if t.IsPointer {
				tmpID = "&" + tmpID
			}
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
