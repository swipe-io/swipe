package service

import (
	"fmt"
	stdtypes "go/types"
	"strconv"
	"strings"

	"github.com/swipe-io/swipe/pkg/types"
	"github.com/swipe-io/swipe/pkg/utils"
)

func structKeyValue(tuple *stdtypes.Tuple, filterFn utils.FilterFn) (results []string) {
	return utils.Params(
		tuple,
		func(p *stdtypes.Var) []string {
			name := p.Name()
			fieldName := strings.ToUpper(name[:1]) + name[1:]
			return []string{fieldName, name}
		},
		filterFn,
	)
}

func logParams(data ...*stdtypes.Tuple) (result []string) {
	for _, tuple := range data {
		for i := 0; i < tuple.Len(); i++ {
			v := tuple.At(i)
			if logParam := logParam(v); logParam != "" {
				result = append(result, strconv.Quote(v.Name()), logParam)
			}
		}
	}
	return
}

func logParam(p *stdtypes.Var) string {
	switch t := p.Type().(type) {
	case *stdtypes.Basic:
		return p.Name()
	case *stdtypes.Slice, *stdtypes.Array, *stdtypes.Map:
		return "len(" + p.Name() + ")"
	case *stdtypes.Named:
		if t.Obj().Pkg() != nil {
			switch t.Obj().Pkg().Path() {
			case "github.com/satori/go.uuid":
				return p.Name()
			}
		} else if stdtypes.Identical(t, types.ErrorType) {
			return p.Name()
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
