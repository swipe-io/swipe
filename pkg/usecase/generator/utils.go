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
