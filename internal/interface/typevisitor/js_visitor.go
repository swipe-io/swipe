package typevisitor

import (
	stdtypes "go/types"
	"strings"

	"github.com/fatih/structtag"

	"github.com/swipe-io/swipe/v2/internal/usecase/typevisitor"
	"github.com/swipe-io/swipe/v2/internal/writer"
)

type jsTypeVisitor struct {
	buf *writer.BaseWriter
}

func (v *jsTypeVisitor) w(format string, a ...interface{}) {
	v.buf.W(format, a...)

}

func (v *jsTypeVisitor) Out() string {
	s := v.buf.String()
	v.buf.Reset()
	return s
}

func (v *jsTypeVisitor) VisitPointer(t *stdtypes.Pointer, nested int) {
	typevisitor.ConvertType(t.Elem()).Accept(v, nested)
}

func (v *jsTypeVisitor) VisitArray(t *stdtypes.Array, nested int) {
	v.w("Array<")
	typevisitor.ConvertType(t.Elem()).Accept(v, nested)
	v.w(">")
}

func (v *jsTypeVisitor) VisitSlice(t *stdtypes.Slice, nested int) {
	v.w("Array<")
	typevisitor.ConvertType(t.Elem()).Accept(v, nested)
	v.w(">")
}

func (v *jsTypeVisitor) VisitMap(t *stdtypes.Map, nested int) {
	v.w("Object<string, ")
	typevisitor.ConvertType(t.Elem()).Accept(v, nested)
	v.w(">")
}

func (v *jsTypeVisitor) VisitNamed(t *stdtypes.Named, nested int) {
	switch stdtypes.TypeString(t.Obj().Type(), nil) {
	default:
		if _, ok := t.Obj().Type().Underlying().(*stdtypes.Struct); ok {
			v.w(t.Obj().Name())
		} else {
			typevisitor.ConvertType(t.Obj().Type().Underlying()).Accept(v, nested+1)
		}
	case "encoding/json.RawMessage":
		v.w("*")
		return
	case "github.com/pborman/uuid.UUID",
		"github.com/google/uuid.UUID",
		"time.Time":
		v.w("string")
		return
	}
}

func (v *jsTypeVisitor) writeStruct(st *stdtypes.Struct, nested int) {
	var j int
	for i := 0; i < st.NumFields(); i++ {
		f := st.Field(i)
		if f.Embedded() {
			var st *stdtypes.Struct
			if ptr, ok := f.Type().(*stdtypes.Pointer); ok {
				st = ptr.Elem().Underlying().(*stdtypes.Struct)
			} else {
				st = f.Type().Underlying().(*stdtypes.Struct)
			}
			v.writeStruct(st, nested)
			v.w(",\n")
			continue
		}
		var (
			skip bool
			name = f.Name()
		)
		if tags, err := structtag.Parse(st.Tag(i)); err == nil {
			if jsonTag, err := tags.Get("json"); err == nil {
				if jsonTag.Name == "-" {
					skip = true
				} else {
					name = jsonTag.Name
				}
			}
		}
		if skip {
			continue
		}
		if j > 0 {
			v.w("\n")
		}
		v.w("*%s", strings.Repeat(" ", nested))
		v.w(" %s:", name)
		typevisitor.ConvertType(f.Type()).Accept(v, nested+1)
		j++
	}
}

func (v *jsTypeVisitor) VisitStruct(t *stdtypes.Struct, nested int) {
	v.w("{\n")
	v.writeStruct(t, nested)
	v.w("\n* }")
}

func (v *jsTypeVisitor) VisitBasic(t *stdtypes.Basic, nested int) {
	switch t.Kind() {
	default:
		v.w("string")
	case stdtypes.Bool:
		v.w("boolean")
	case stdtypes.Float32,
		stdtypes.Float64,
		stdtypes.Int,
		stdtypes.Int8,
		stdtypes.Int16,
		stdtypes.Int32,
		stdtypes.Int64,
		stdtypes.Uint,
		stdtypes.Uint8,
		stdtypes.Uint16,
		stdtypes.Uint32,
		stdtypes.Uint64:
		v.w("number")
	}
}

func (v *jsTypeVisitor) VisitInterface(t *stdtypes.Interface, nested int) {
	v.w("object")
}

func (v *jsTypeVisitor) Visit(t stdtypes.Type) {
	typevisitor.ConvertType(t).Accept(v, 0)
}

func JSTypeVisitor(buf *writer.BaseWriter) typevisitor.TypeVisitor {
	return &jsTypeVisitor{
		buf: buf,
	}
}
