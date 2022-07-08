package convert

import (
	"fmt"
	"io"
	"strconv"

	"github.com/swipe-io/swipe/v3/option"
	"github.com/swipe-io/swipe/v3/swipe"
	"github.com/swipe-io/swipe/v3/writer"
)

type Builder struct {
	importer    swipe.Importer
	declareErr  bool
	fieldName   option.String
	fieldType   interface{}
	assignOp    string
	assignVar   string
	valueVar    string
	errorReturn func() string
	w           writer.GoWriter
}

func (b *Builder) SetDeclareErr(declareErr bool) *Builder {
	b.declareErr = declareErr
	return b
}

func (b *Builder) SetAssignOp(assignOp string) {
	b.assignOp = assignOp
}

func (b *Builder) SetFieldType(fieldType interface{}) *Builder {
	b.fieldType = fieldType
	return b
}

func (b *Builder) SetFieldName(fieldName option.String) *Builder {
	b.fieldName = fieldName
	return b
}

func (b *Builder) SetErrorReturn(errorReturn func() string) *Builder {
	b.errorReturn = errorReturn
	return b
}

func (b *Builder) SetAssignVar(assignVar string) *Builder {
	b.assignVar = assignVar
	return b
}

func (b *Builder) SetValueVar(valueVar string) *Builder {
	b.valueVar = valueVar
	return b
}

func (b *Builder) Write(w io.Writer) {
	if b.fieldType == nil {
		panic("field type must not be nil")
	}
	if b.assignOp == "" {
		b.assignOp = "="
	}
	switch t := b.fieldType.(type) {
	case *option.BasicType:
		b.writeBasicType(t)
	case *option.MapType:
		b.writeMapType(t)
	case *option.SliceType:
		b.writeSliceType(t)
	case *option.ArrayType:
		b.writeArrayType(t)
	case *option.NamedType:
		b.writeNameType(t)
	}
	_, _ = fmt.Fprint(w, b.w.String())
}

func (b *Builder) convertFuncName(t *option.BasicType) string {
	if t.IsAnyInt() {
		return "ParseInt(%s, 10, " + t.BitSize() + ")"
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
	panic(fmt.Sprintf("convert unknown basic type: %v", t))
}

func (b *Builder) writeBasicType(t *option.BasicType) {
	if t.IsString() {
		if t.IsPointer {
			tmpVar := b.fieldName.Lower() + "Str"
			b.w.W("%s := %s\n", tmpVar, b.valueVar)
			b.valueVar = "&" + tmpVar
		}
		b.w.W("%s%s%s\n", b.assignVar, b.assignOp, b.valueVar)
		return
	}
	strconvPkg := b.importer.Import("strconv", "strconv")
	funcName := b.convertFuncName(t)

	var tmpVar string
	if t.Name != "int64" || t.IsPointer {
		tmpVar = b.fieldName.Lower() + "Tmp"
	}

	if b.declareErr && tmpVar == "" {
		b.w.W("var err error\n")
	}

	if tmpVar != "" {
		b.w.W("%s, err := ", tmpVar)
	} else {
		b.w.W("%s, err %s ", b.assignVar, b.assignOp)
	}

	b.w.W("%s.%s\n", strconvPkg, fmt.Sprintf(funcName, b.valueVar))
	b.w.W("if err != nil {\n %s\n}\n", b.errorReturn())

	if tmpVar != "" {
		if t.Name != "int64" {
			tmpVar = t.Name + "(" + tmpVar + ")"
		}
		if t.IsPointer {
			newTmpVar := b.fieldName.Lower()
			b.w.W("%s := %s\n", newTmpVar, tmpVar)
			tmpVar = "&" + newTmpVar
		}
		b.w.W("%s %s %s\n", b.assignVar, b.assignOp, tmpVar)
	}
}

func (b *Builder) writeMapType(t *option.MapType) {
	stringsPkg := b.importer.Import("strings", "strings")
	if k, ok := t.Key.(*option.BasicType); ok && k.IsString() {
		if v, ok := t.Value.(*option.BasicType); ok {
			tmpID := b.fieldName.Lower() + "Parts"

			b.w.W("%s := %s.Split(%s, \",\")\n", tmpID, stringsPkg, b.valueVar)
			b.w.W("%s = make(%s, len(%s))\n", b.assignVar, swipe.TypeString(t, false, b.importer), tmpID)
			if v.IsNumeric() {
				b.w.W("for _, s := range %s {\n", tmpID)
				b.w.W("kv := %s.Split(s, \"=\")\n", stringsPkg)
				b.w.W("if len(kv) == 2 {\n")

				NewBuilder(b.importer).
					SetDeclareErr(b.declareErr).
					SetAssignVar(b.assignVar + "[kv[0]]").
					SetValueVar("kv[1]").
					SetFieldName(b.fieldName).
					SetFieldType(v).
					SetErrorReturn(b.errorReturn).
					Write(&b.w)

				b.w.W("}\n")
				b.w.W("}\n")
			} else {
				b.w.W("for _, s := range %s {\n", tmpID)
				b.w.W("kv := %s.Split(s, \"=\")\n", stringsPkg)
				b.w.W("if len(kv) == 2 {\n")
				b.w.W("%s[kv[0]] = kv[1]\n", b.assignVar)
				b.w.W("}\n")
				b.w.W("}\n")
			}
		}
	}
}

func (b *Builder) writeSliceType(t *option.SliceType) {
	stringsPkg := b.importer.Import("strings", "strings")
	switch v := t.Value.(type) {
	case *option.BasicType:
		if v.IsNumeric() {
			tmpID := b.fieldName.Lower() + "Parts"
			b.w.W("%s := %s.Split(%s, \",\")\n", tmpID, stringsPkg, b.valueVar)

			ptrPrefix := ""
			if v.IsPointer {
				ptrPrefix = "*"
			}

			b.w.W("%s = make([]%s%s, len(%s))\n", b.assignVar, ptrPrefix, v.Name, tmpID)
			b.w.W("for i, s := range %s {\n", tmpID)

			NewBuilder(b.importer).
				SetDeclareErr(b.declareErr).
				SetAssignVar(b.assignVar + "[i]").
				SetValueVar("s").
				SetFieldName(b.fieldName).
				SetFieldType(v).
				SetErrorReturn(b.errorReturn).
				Write(&b.w)

			b.w.W("}\n")
		} else {
			b.w.W("%s = %s.Split(%s, \",\")\n", b.assignVar, stringsPkg, b.valueVar)
		}
	}
}

func (b *Builder) writeArrayType(t *option.ArrayType) {
	stringsPkg := b.importer.Import("strings", "strings")
	switch v := t.Value.(type) {
	case *option.BasicType:
		if v.IsNumeric() {
			tmpID := b.fieldName.Lower() + "Parts"
			b.w.W("%s := %s.Split(%s, \",\")\n", tmpID, stringsPkg, b.valueVar)
			b.w.W("if len(%s) > len(%s) {\n", tmpID, b.assignVar)
			b.w.W("panic(\"%s\")\n", "array length must be less or equal"+strconv.FormatInt(t.Len, 10))
			b.w.W("} else {\n")

			b.w.W("for i, s := range %s {\n", tmpID)

			NewBuilder(b.importer).
				SetDeclareErr(b.declareErr).
				SetAssignVar(b.assignVar + "[i]").
				SetValueVar("s").
				SetFieldType(v).
				SetErrorReturn(b.errorReturn).
				Write(&b.w)

			b.w.W("}\n")
			b.w.W("}\n")
		} else {
			b.w.W("%s = %s.Split(%s, \",\")\n", b.assignVar, stringsPkg, b.valueVar)
		}
	}
}

func (b *Builder) writeNameType(t *option.NamedType) {
	if b.declareErr {
		b.w.W("var err error\n")
	}
	switch t.Pkg.Path {
	case "net/url":
		switch t.Name.Value {
		case "URL":
			urlPkg := b.importer.Import("url", "net/url")
			b.w.W("%s, err %s %s.Parse(%s)\n", b.assignVar, b.assignOp, urlPkg, b.valueVar)
		}
	case "github.com/satori/uuid", "github.com/google/uuid":
		if t.Name.Value == "UUID" {
			uuidPkg := b.importer.Import(t.Pkg.Name, t.Pkg.Path)
			switch t.Pkg.Path {
			case "github.com/google/uuid":
				b.w.W("%s, err %s %s.Parse(%s)\n", b.assignVar, b.assignOp, uuidPkg, b.valueVar)
			case "github.com/satori/uuid":
				b.w.W("%s, err %s %s.FromString(%s)\n", b.assignVar, b.assignOp, uuidPkg, b.valueVar)
			}
		}
	case "time":
		switch t.Name.Value {
		case "Time":
			timePkg := b.importer.Import("time", "time")
			b.w.W("%[1]s, err %[4]s %[2]s.Parse(%[2]s.RFC3339, %[3]s)\n", b.assignVar, timePkg, b.valueVar, b.assignOp)
		case "Duration":
			timePkg := b.importer.Import("time", "time")
			b.w.W("%s, err %s %s.ParseDuration(%s)\n", b.assignVar, b.assignOp, timePkg, b.valueVar)
		}
	}
	b.w.W("if err != nil {\n %s\n}\n", b.errorReturn())
}

func NewBuilder(importer swipe.Importer) *Builder {
	return &Builder{importer: importer}
}
