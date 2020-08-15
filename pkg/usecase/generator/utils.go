package generator

import (
	stdtypes "go/types"
	"strconv"
	"strings"

	"github.com/swipe-io/swipe/pkg/types"
)

func structKeyValue(vars []*stdtypes.Var, filterFn types.FilterFn) (results []string) {
	return types.Params(
		vars,
		func(p *stdtypes.Var) []string {
			name := p.Name()
			fieldName := strings.ToUpper(name[:1]) + name[1:]
			return []string{fieldName, name}
		},
		filterFn,
	)
}

func makeLogParams(data ...*stdtypes.Var) (result []string) {
	for _, v := range data {
		if logParam := makeLogParam(v.Name(), v.Type()); logParam != "" {
			result = append(result, strconv.Quote(v.Name()), logParam)
		}
	}
	return
}

func makeLogParam(name string, t stdtypes.Type) string {
	switch t.(type) {
	default:
		return name
	case *stdtypes.Slice, *stdtypes.Array, *stdtypes.Map, *stdtypes.Chan:
		return "len(" + name + ")"
	}
}

func isGolangNamedType(t stdtypes.Type) bool {
	t = normalizeType(t)
	switch stdtypes.TypeString(t, nil) {
	case "time.Time",
		"time.Location",
		"sql.NullBool",
		"sql.NullFloat64",
		"sql.NullInt32",
		"sql.NullInt64",
		"sql.NullString",
		"sql.NullTime":
		return true
	}
	return false
}

func normalizeType(t stdtypes.Type) stdtypes.Type {
	switch v := t.(type) {
	case *stdtypes.Pointer:
		return normalizeType(v.Elem())
	case *stdtypes.Slice:
		return normalizeType(v.Elem())
	case *stdtypes.Array:
		return normalizeType(v.Elem())
	case *stdtypes.Map:
		return normalizeType(v.Elem())
	default:
		return v
	}
}
