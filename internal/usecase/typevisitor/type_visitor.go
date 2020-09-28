package typevisitor

import (
	stdtypes "go/types"
)

type Type interface {
	Accept(visitor TypeVisitor, nested int)
}

type pointerWrapper struct {
	*stdtypes.Pointer
}

func (p *pointerWrapper) Accept(visitor TypeVisitor, nested int) {
	visitor.VisitPointer(p.Pointer, nested)
}

type arrayWrapper struct {
	*stdtypes.Array
}

func (a *arrayWrapper) Accept(visitor TypeVisitor, nested int) {
	visitor.VisitArray(a.Array, nested)
}

type sliceWrapper struct {
	*stdtypes.Slice
}

func (s *sliceWrapper) Accept(visitor TypeVisitor, nested int) {
	visitor.VisitSlice(s.Slice, nested)
}

type mapWrapper struct {
	*stdtypes.Map
}

func (m *mapWrapper) Accept(visitor TypeVisitor, nested int) {
	visitor.VisitMap(m.Map, nested)
}

type namedWrapper struct {
	*stdtypes.Named
}

func (n *namedWrapper) Accept(visitor TypeVisitor, nested int) {
	visitor.VisitNamed(n.Named, nested)
}

type structWrapper struct {
	*stdtypes.Struct
}

func (st *structWrapper) Accept(visitor TypeVisitor, nested int) {
	visitor.VisitStruct(st.Struct, nested)
}

type basicWrapper struct {
	*stdtypes.Basic
}

func (b *basicWrapper) Accept(visitor TypeVisitor, nested int) {
	visitor.VisitBasic(b.Basic, nested)
}

type interfaceWrapper struct {
	*stdtypes.Interface
}

func (i *interfaceWrapper) Accept(visitor TypeVisitor, nested int) {
	visitor.VisitInterface(i.Interface, nested)
}

func ConvertType(t stdtypes.Type) Type {
	switch v := t.(type) {
	case *stdtypes.Pointer:
		return &pointerWrapper{
			Pointer: v,
		}
	case *stdtypes.Array:
		return &arrayWrapper{
			Array: v,
		}
	case *stdtypes.Slice:
		return &sliceWrapper{
			Slice: v,
		}
	case *stdtypes.Map:
		return &mapWrapper{
			Map: v,
		}
	case *stdtypes.Named:
		return &namedWrapper{
			Named: v,
		}
	case *stdtypes.Struct:
		return &structWrapper{
			Struct: v,
		}
	case *stdtypes.Basic:
		return &basicWrapper{
			Basic: v,
		}
	case *stdtypes.Interface:
		return &interfaceWrapper{
			Interface: v,
		}
	}
	panic("unexpected type:" + stdtypes.TypeString(t, nil))
}

type TypeVisitor interface {
	Visit(t stdtypes.Type)
	VisitPointer(t *stdtypes.Pointer, nested int)
	VisitArray(t *stdtypes.Array, nested int)
	VisitSlice(t *stdtypes.Slice, nested int)
	VisitMap(t *stdtypes.Map, nested int)
	VisitNamed(t *stdtypes.Named, nested int)
	VisitStruct(t *stdtypes.Struct, nested int)
	VisitBasic(t *stdtypes.Basic, nested int)
	VisitInterface(t *stdtypes.Interface, nested int)
}

type NamedTypeCollector interface {
	TypeVisitor
	TypeDefs() []*stdtypes.Named
}
