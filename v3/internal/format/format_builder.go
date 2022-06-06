package format

import (
	"fmt"
	"io"

	"github.com/swipe-io/swipe/v3/option"
	"github.com/swipe-io/swipe/v3/swipe"
	"github.com/swipe-io/swipe/v3/writer"
)

type Builder struct {
	importer  swipe.Importer
	fieldType interface{}
	assignVar string
	valueVar  string
	w         writer.GoWriter
}

func (b *Builder) SetFieldType(fieldType interface{}) *Builder {
	b.fieldType = fieldType
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
	switch t := b.fieldType.(type) {
	case *option.BasicType:
		b.writeBasicType(t)
	case *option.SliceType:
		b.writeSliceType(t)
	case *option.NamedType:
		b.writeNameType(t)
	}
	_, _ = fmt.Fprint(w, b.w.String())
}

func (b *Builder) formatFuncName(t *option.BasicType) string {
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

func (b *Builder) writeBasicType(t *option.BasicType) {
	funcName := b.formatFuncName(t)
	if funcName != "" {
		valueVar := b.valueVar
		strconvPkg := b.importer.Import("strconv", "strconv")
		if t.IsPointer {
			valueVar = "*" + valueVar
		}
		b.w.W("%s := %s.%s\n", b.assignVar, strconvPkg, fmt.Sprintf(funcName, valueVar))
		return
	}
	b.w.W("%s := %s\n", b.assignVar, b.valueVar)
}

func (b *Builder) writeSliceType(t *option.SliceType) {
	if bt, ok := t.Value.(*option.BasicType); ok {
		b.w.W("var %s string\n", b.assignVar)
		b.w.W("for i, s := range %s {\n", b.valueVar)
		b.w.W("if i > 0 {\n %s += \",\"\n}\n", b.assignVar)
		NewBuilder(b.importer).
			SetAssignVar("v").
			SetValueVar("s").
			SetFieldType(bt).
			Write(&b.w)
		b.w.W("%s += v\n", b.assignVar)
		b.w.W("}\n")
	}
}

func (b *Builder) writeNameType(t *option.NamedType) {
	switch t.Pkg.Path {
	case "github.com/satori/uuid", "github.com/google/uuid":
		if t.Name.Value == "UUID" {
			b.w.W("%s := %s.String() \n", b.assignVar, b.valueVar)
		}
	case "time":
		switch t.Name.Value {
		case "Duration":
			b.w.W("%s := %s.String()\n", b.assignVar, b.valueVar)
		case "Time":
			timePkg := b.importer.Import("time", "time")
			b.w.W("%[1]s := %[3]s.Format(%[2]s.RFC3339)\n", b.assignVar, timePkg, b.valueVar)
		}
	}
}

func NewBuilder(importer swipe.Importer) *Builder {
	return &Builder{importer: importer}
}
