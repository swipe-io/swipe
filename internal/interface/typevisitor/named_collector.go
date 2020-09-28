package typevisitor

import (
	stdtypes "go/types"

	"github.com/swipe-io/swipe/v2/internal/usecase/typevisitor"

	"golang.org/x/tools/go/types/typeutil"
)

type namedTypeCollector struct {
	typeDefs      []*stdtypes.Named
	existsTypeDef map[uint32]struct{}
	hasher        typeutil.Hasher
}

func (v *namedTypeCollector) Visit(t stdtypes.Type) {
	typevisitor.ConvertType(t).Accept(v, 0)
}

func (v *namedTypeCollector) VisitPointer(t *stdtypes.Pointer, nested int) {
	typevisitor.ConvertType(t.Elem()).Accept(v, nested)
}

func (v *namedTypeCollector) VisitArray(t *stdtypes.Array, nested int) {
}

func (v *namedTypeCollector) VisitSlice(t *stdtypes.Slice, nested int) {
}

func (v *namedTypeCollector) VisitMap(t *stdtypes.Map, nested int) {
}

func (v *namedTypeCollector) VisitNamed(t *stdtypes.Named, nested int) {
	switch stdtypes.TypeString(t.Obj().Type(), nil) {
	case
		"encoding/json.RawMessage",
		"github.com/pborman/uuid.UUID",
		"github.com/google/uuid.UUID",
		"time.Time":
		return
	}
	if !v.exists(t) {
		v.add(t)
		v.Visit(t.Underlying())
	}
}

func (v *namedTypeCollector) VisitStruct(t *stdtypes.Struct, nested int) {
	for i := 0; i < t.NumFields(); i++ {
		p := t.Field(i)
		typevisitor.ConvertType(p.Type()).Accept(v, nested+1)
	}
}

func (v *namedTypeCollector) VisitBasic(t *stdtypes.Basic, nested int) {
}

func (v *namedTypeCollector) VisitInterface(t *stdtypes.Interface, nested int) {
}

func (v *namedTypeCollector) TypeDefs() []*stdtypes.Named {
	return v.typeDefs
}

func (v *namedTypeCollector) exists(t *stdtypes.Named) bool {
	h := v.hasher.Hash(t)
	_, ok := v.existsTypeDef[h]
	return ok
}

func (v *namedTypeCollector) add(t *stdtypes.Named) {
	h := v.hasher.Hash(t)
	v.typeDefs = append(v.typeDefs, t)
	v.existsTypeDef[h] = struct{}{}
}

func NewNamedTypeCollector() typevisitor.NamedTypeCollector {
	return &namedTypeCollector{
		existsTypeDef: map[uint32]struct{}{},
		hasher:        typeutil.MakeHasher(),
	}
}
