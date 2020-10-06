package typevisitor

import (
	stdtypes "go/types"

	"github.com/fatih/structtag"

	"github.com/swipe-io/swipe/v2/internal/usecase/typevisitor"
	"github.com/swipe-io/swipe/v2/internal/writer"
)

type jsTypeDefVisitor struct {
	buf *writer.BaseWriter
	jst typevisitor.TypeVisitor
}

func (v *jsTypeDefVisitor) writeStruct(st *stdtypes.Struct, nested int) {
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
			v.buf.W("\n")
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

		v.buf.W("* @property {")
		typevisitor.ConvertType(f.Type()).Accept(v, nested+1)
		v.buf.W("} ")
		v.buf.W("%s\n", name)

		j++
	}
}

func (v *jsTypeDefVisitor) Visit(t stdtypes.Type) {
	typevisitor.ConvertType(t).Accept(v, 0)
}

func (v *jsTypeDefVisitor) VisitPointer(t *stdtypes.Pointer, nested int) {
	v.jst.VisitPointer(t, nested)
}

func (v *jsTypeDefVisitor) VisitArray(t *stdtypes.Array, nested int) {
	v.jst.VisitArray(t, nested)
}

func (v *jsTypeDefVisitor) VisitSlice(t *stdtypes.Slice, nested int) {
	v.jst.VisitSlice(t, nested)
}

func (v *jsTypeDefVisitor) VisitMap(t *stdtypes.Map, nested int) {
	v.jst.VisitMap(t, nested)
}

func (v *jsTypeDefVisitor) VisitBasic(t *stdtypes.Basic, nested int) {
	v.jst.VisitBasic(t, nested)
}

func (v *jsTypeDefVisitor) VisitInterface(t *stdtypes.Interface, nested int) {
	v.jst.VisitInterface(t, nested)
}

func (v *jsTypeDefVisitor) VisitNamed(t *stdtypes.Named, nested int) {
	switch stdtypes.TypeString(t.Obj().Type(), nil) {
	case "encoding/json.RawMessage":
		v.buf.W("*")
		return
	case "github.com/pborman/uuid.UUID",
		"github.com/google/uuid.UUID",
		"time.Time":
		v.buf.W("string")
		return
	}
	if nested == 0 {
		if st, ok := t.Obj().Type().Underlying().(*stdtypes.Struct); ok {
			v.buf.W("/**\n")
			v.buf.W("* @typedef {Object} %s\n", t.Obj().Name())
			v.writeStruct(st, nested)
			v.buf.W("**/\n\n")
		}
	} else {
		if _, ok := t.Obj().Type().Underlying().(*stdtypes.Struct); ok {
			v.buf.W(t.Obj().Name())
		} else {
			typevisitor.ConvertType(t.Obj().Type().Underlying()).Accept(v, nested+1)
		}
	}
}

func (v *jsTypeDefVisitor) VisitStruct(t *stdtypes.Struct, nested int) {
	v.writeStruct(t, nested)
}

func JSTypeDefVisitor(buf *writer.BaseWriter) typevisitor.TypeVisitor {
	return &jsTypeDefVisitor{
		buf: buf,
		jst: JSTypeVisitor(buf),
	}
}
