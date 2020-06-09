package service

import (
	"fmt"
	stdtypes "go/types"
	"strconv"
	"strings"

	"github.com/swipe-io/swipe/pkg/types"
	"github.com/swipe-io/swipe/pkg/utils"
)

func structKeyValue(vars []*stdtypes.Var, filterFn utils.FilterFn) (results []string) {
	return utils.Params(
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
	switch t := t.(type) {
	case *stdtypes.Basic:
		return name
	case *stdtypes.Slice, *stdtypes.Array, *stdtypes.Map:
		return "len(" + name + ")"
	case *stdtypes.Named:
		if t.Obj().Pkg() != nil {
			switch t.Obj().Pkg().Path() {
			case "github.com/satori/go.uuid", "github.com/google/uuid":
				return name
			}
		} else if stdtypes.Identical(t, types.ErrorType) {
			return name
		}
	}
	return ""
}

func httpBraceIndices(s string) ([]int, error) {
	var level, idx int
	var idxs []int
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '{':
			if level++; level == 1 {
				idx = i
			}
		case '}':
			if level--; level == 0 {
				idxs = append(idxs, idx, i+1)
			} else if level < 0 {
				return nil, fmt.Errorf("mux: unbalanced braces in %q", s)
			}
		}
	}
	if level != 0 {
		return nil, fmt.Errorf("mux: unbalanced braces in %q", s)
	}
	return idxs, nil
}
