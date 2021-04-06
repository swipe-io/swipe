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
	switch tp := t.Obj().Type().Underlying().(type) {
	case *stdtypes.Interface:
		if nested == 0 {
			v.buf.W("/**\n")
			v.buf.W("* @typedef {Object} %s\n", t.Obj().Name())
			v.buf.W("*/\n\n")
		} else {
			v.buf.W(t.Obj().Name())
		}
	case *stdtypes.Struct:
		if nested == 0 {
			v.buf.W("/**\n")
			v.buf.W("* @typedef {Object} %s\n", t.Obj().Name())
			typevisitor.ConvertType(tp).Accept(v, nested)
			v.buf.W("*/\n\n")
		} else {
			v.buf.W(t.Obj().Name())
		}
	case *stdtypes.Map, *stdtypes.Slice, *stdtypes.Basic:
		if nested == 0 {
			v.buf.W("/**\n")
			v.buf.W("* @typedef {")
			typevisitor.ConvertType(tp).Accept(v, nested)
			v.buf.W("} %s\n", t.Obj().Name())
			v.buf.W("*/\n\n")
		} else {
			v.buf.W(t.Obj().Name())
		}
	}
}

func (v *jsTypeDefVisitor) VisitStruct(t *stdtypes.Struct, nested int) {
	var j int
	for i := 0; i < t.NumFields(); i++ {
		f := t.Field(i)
		if f.Embedded() {
			var st *stdtypes.Struct
			if ptr, ok := f.Type().(*stdtypes.Pointer); ok {
				st = ptr.Elem().Underlying().(*stdtypes.Struct)
			} else {
				st = f.Type().Underlying().(*stdtypes.Struct)
			}

			typevisitor.ConvertType(st).Accept(v, nested)

			v.buf.W("\n")
			continue
		}
		var (
			skip bool
			name = f.Name()
		)
		if tags, err := structtag.Parse(t.Tag(i)); err == nil {
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

func JSTypeDefVisitor(buf *writer.BaseWriter) typevisitor.TypeVisitor {
	return &jsTypeDefVisitor{
		buf: buf,
		jst: JSTypeVisitor(buf),
	}
}
