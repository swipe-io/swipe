package processor

import "github.com/swipe-io/swipe/v2/internal/option"

type FactoryFn func(o *option.ResultOption, opr *option.Result) (Processor, error)

type Factory interface {
	Register(name string, fn FactoryFn, optFn func() []byte)
	Get(name string) (FactoryFn, bool)
	GetOptGen(name string) (fn func() []byte, ok bool)
	Names() []string
}
