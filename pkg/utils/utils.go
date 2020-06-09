package utils

import (
	"go/types"
)

type FilterFn func(p *types.Var) bool

func Params(vars []*types.Var, fn func(p *types.Var) []string, filterFn func(p *types.Var) bool) (results []string) {
	for _, p := range vars {
		if filterFn != nil && !filterFn(p) {
			continue
		}
		results = append(results, fn(p)...)
	}
	return
}

func NameParams(vars []*types.Var, filterFn FilterFn) (results []string) {
	return Params(vars, func(p *types.Var) []string {
		return []string{p.Name()}
	}, filterFn)
}

func NameTypeParams(vars []*types.Var, typeString func(t types.Type) string, filterFn FilterFn) (results []string) {
	return Params(vars, func(p *types.Var) []string {
		return []string{p.Name(), typeString(p.Type())}
	}, filterFn)
}

func NameType(vars []*types.Var, typeString func(t types.Type) string, filterFn FilterFn) (results []string) {
	return Params(vars, func(p *types.Var) []string {
		return []string{"", typeString(p.Type())}
	}, filterFn)
}
