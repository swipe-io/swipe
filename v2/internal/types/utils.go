package types

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

func NameTypeParams(vars []*types.Var, qf types.Qualifier, filterFn FilterFn) (results []string) {
	return Params(vars, func(p *types.Var) []string {
		return []string{p.Name(), types.TypeString(p.Type(), qf)}
	}, filterFn)
}

func NameType(vars []*types.Var, qf types.Qualifier, filterFn FilterFn) (results []string) {
	return Params(vars, func(p *types.Var) []string {
		return []string{"", types.TypeString(p.Type(), qf)}
	}, filterFn)
}
