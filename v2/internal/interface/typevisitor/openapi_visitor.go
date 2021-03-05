package typevisitor

import (
	"encoding/json"
	stdtypes "go/types"

	"github.com/fatih/structtag"

	"github.com/swipe-io/swipe/v2/internal/openapi"

	"github.com/swipe-io/swipe/v2/internal/usecase/typevisitor"
)

type openapiVisitor struct {
	schema *openapi.Schema
}

func (v *openapiVisitor) populateSchema(st *stdtypes.Struct) {
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

func (v *openapiVisitor) Visit(t stdtypes.Type) {
	typevisitor.ConvertType(t).Accept(v, 0)
}

func (v *openapiVisitor) VisitPointer(t *stdtypes.Pointer, nested int) {
	typevisitor.ConvertType(t.Elem()).Accept(v, nested)
}

func (v *openapiVisitor) VisitArray(t *stdtypes.Array, nested int) {
	v.schema.Type = "array"
	v.schema.Items = &openapi.Schema{}
	OpenapiVisitor(v.schema.Items).Visit(t.Elem())
}

func (v *openapiVisitor) VisitSlice(t *stdtypes.Slice, nested int) {
	if vv, ok := t.Elem().(*stdtypes.Basic); ok && vv.Kind() == stdtypes.Byte {
		v.schema.Type = "string"
		v.schema.Format = "byte"
		v.schema.Example = "U3dhZ2dlciByb2Nrcw=="
	} else {
		v.schema.Type = "array"
		v.schema.Items = &openapi.Schema{}
		OpenapiVisitor(v.schema.Items).Visit(t.Elem())
	}
}

func (v *openapiVisitor) VisitMap(t *stdtypes.Map, nested int) {
	v.schema.Properties = openapi.Properties{"key": &openapi.Schema{}}
	OpenapiVisitor(v.schema.Properties["key"]).Visit(t.Elem())
}

func (v *openapiVisitor) VisitNamed(t *stdtypes.Named, nested int) {
	switch stdtypes.TypeString(t, nil) {
	default:
		v.schema.Ref = "#/components/schemas/" + t.Obj().Name()
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
}

func (v *openapiVisitor) VisitStruct(t *stdtypes.Struct, nested int) {
	v.schema.Type = "object"
	v.schema.Properties = openapi.Properties{}
	v.populateSchema(t)
}

func (v *openapiVisitor) VisitBasic(t *stdtypes.Basic, nested int) {
	switch t.Kind() {
	default:
		v.schema.Type = "string"
		v.schema.Format = "string"
		v.schema.Example = "abc"
	case stdtypes.Bool:
		v.schema.Type = "boolean"
		v.schema.Example = true
	case stdtypes.Int,
		stdtypes.Uint,
		stdtypes.Uint8,
		stdtypes.Uint16,
		stdtypes.Int8,
		stdtypes.Int16:
		v.schema.Type = "integer"
		v.schema.Example = 1
	case stdtypes.Uint32, stdtypes.Int32:
		v.schema.Type = "integer"
		v.schema.Format = "int32"
		v.schema.Example = 1
	case stdtypes.Uint64, stdtypes.Int64:
		v.schema.Type = "integer"
		v.schema.Format = "int64"
		v.schema.Example = 1
	case stdtypes.Float32, stdtypes.Float64:
		v.schema.Type = "number"
		v.schema.Format = "float"
		v.schema.Example = 1.11
	}
}

func (v *openapiVisitor) VisitInterface(t *stdtypes.Interface, nested int) {
	v.schema.Type = "object"
	v.schema.Description = "Can be any value - string, number, boolean, array or object."
	v.schema.Properties = openapi.Properties{}
	v.schema.Example = json.RawMessage("null")
	v.schema.AnyOf = []openapi.Schema{
		{Type: "string", Example: "abc"},
		{Type: "integer", Example: 1},
		{Type: "number", Format: "float", Example: 1.11},
		{Type: "boolean", Example: true},
		{Type: "array"},
		{Type: "object"},
	}
}

func OpenapiVisitor(schema *openapi.Schema) typevisitor.TypeVisitor {
	return &openapiVisitor{
		schema: schema,
	}
}
