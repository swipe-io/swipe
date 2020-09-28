package typevisitor

import (
	stdtypes "go/types"

	"github.com/fatih/structtag"

	"github.com/swipe-io/swipe/v2/internal/openapi"
	"github.com/swipe-io/swipe/v2/internal/usecase/typevisitor"
)

type openapiDefVisitor struct {
	schema *openapi.Schema
	ov     typevisitor.TypeVisitor
}

func (v *openapiDefVisitor) populateSchema(st *stdtypes.Struct) {
	for i := 0; i < st.NumFields(); i++ {
		f := st.Field(i)
		if !f.Embedded() {
			name := f.Name()
			if tags, err := structtag.Parse(st.Tag(i)); err == nil {
				if tag, err := tags.Get("json"); err == nil {
					name = tag.Value()
				}
			}
			if name == "-" {
				continue
			}
			v.schema.Properties[name] = &openapi.Schema{}
			OpenapiVisitor(v.schema.Properties[name]).Visit(f.Type())
		} else {
			var st *stdtypes.Struct
			if ptr, ok := f.Type().(*stdtypes.Pointer); ok {
				st = ptr.Elem().Underlying().(*stdtypes.Struct)
			} else {
				st = f.Type().Underlying().(*stdtypes.Struct)
			}
			v.populateSchema(st)
		}
	}
}

func (v *openapiDefVisitor) Visit(t stdtypes.Type) {
	typevisitor.ConvertType(t).Accept(v, 0)
}

func (v *openapiDefVisitor) VisitPointer(t *stdtypes.Pointer, nested int) {
	v.ov.VisitPointer(t, nested)
}

func (v *openapiDefVisitor) VisitArray(t *stdtypes.Array, nested int) {
	v.ov.VisitArray(t, nested)
}

func (v *openapiDefVisitor) VisitSlice(t *stdtypes.Slice, nested int) {
	v.VisitSlice(t, nested)
}

func (v *openapiDefVisitor) VisitMap(t *stdtypes.Map, nested int) {
	v.ov.VisitMap(t, nested)
}

func (v *openapiDefVisitor) VisitNamed(t *stdtypes.Named, nested int) {
	switch stdtypes.TypeString(t, nil) {
	case "encoding/json.RawMessage":
		v.schema.Type = "object"
		v.schema.Properties = openapi.Properties{}
		return
	case "time.Time":
		v.schema.Type = "string"
		v.schema.Format = "date-time"
		v.schema.Example = "1985-04-02T01:30:00.00Z"
		return
	case "github.com/pborman/uuid.UUID",
		"github.com/google/uuid.UUID":
		v.schema.Type = "string"
		v.schema.Format = "uuid"
		v.schema.Example = "d5c02d83-6fbc-4dd7-8416-9f85ed80de46"
		return
	}

	if nested == 0 {
		if st, ok := t.Obj().Type().Underlying().(*stdtypes.Struct); ok {
			v.schema.Properties = openapi.Properties{}
			v.populateSchema(st)
		}
	} else {
		v.schema.Type = "object"
		v.schema.Ref = "#/components/schemas/" + t.Obj().Name()
	}
}

func (v *openapiDefVisitor) VisitStruct(t *stdtypes.Struct, nested int) {

}

func (v *openapiDefVisitor) VisitBasic(t *stdtypes.Basic, nested int) {
	v.ov.VisitBasic(t, nested)
}

func (v *openapiDefVisitor) VisitInterface(t *stdtypes.Interface, nested int) {
	v.ov.VisitInterface(t, nested)
}

func OpenapiDefVisitor(schema *openapi.Schema) typevisitor.TypeVisitor {
	return &openapiDefVisitor{
		schema: schema,
		ov:     OpenapiVisitor(schema),
	}
}
