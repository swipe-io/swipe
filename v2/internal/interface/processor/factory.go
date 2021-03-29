package processor

import (
	"sort"

	up "github.com/swipe-io/swipe/v2/internal/usecase/processor"
)

type factory struct {
	factories map[string]up.FactoryFn
	optFnGen  map[string]func() []byte
}

func (r *factory) Register(name string, fn up.FactoryFn, optFn func() []byte) {
	r.factories[name] = fn
	r.optFnGen[name] = optFn
}

func (r *factory) Names() (names []string) {
	for k, _ := range r.factories {
		names = append(names, k)
	}
	sort.Strings(names)
	return
}

func (r *factory) GetOptGen(name string) (fn func() []byte, ok bool) {
	fn, ok = r.optFnGen[name]
	return
}

func (r *factory) Get(name string) (fn up.FactoryFn, ok bool) {
	fn, ok = r.factories[name]
	return
}

func NewFactory() up.Factory {
	return &factory{factories: map[string]up.FactoryFn{}, optFnGen: map[string]func() []byte{}}
}
