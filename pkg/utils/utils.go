package utils

import (
	"go/types"
)

type FilterFn func(p *types.Var) bool

func Params(tuple *types.Tuple, fn func(p *types.Var) []string, filterFn func(p *types.Var) bool) (results []string) {
	for i := 0; i < tuple.Len(); i++ {
		p := tuple.At(i)
		if filterFn != nil && !filterFn(p) {
			continue
		}
		results = append(results, fn(p)...)
	}
	return
}

func NameParams(tuple *types.Tuple, filterFn FilterFn) (results []string) {
	return Params(
		tuple,
		func(p *types.Var) []string {
			return []string{p.Name()}
		},
		filterFn,
	)
}

func NameTypeParams(tuple *types.Tuple, typeString func(t types.Type) string, filterFn FilterFn) (results []string) {
	return Params(tuple, func(p *types.Var) []string {
		return []string{p.Name(), typeString(p.Type())}
	}, filterFn)
}
