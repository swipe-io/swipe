package processor

type FactoryFn func(o *_option.ResultOption, opr *_option.Result) (Processor, error)

type Factory interface {
	Register(name string, fn FactoryFn, optFn func() []byte)
	Get(name string) (FactoryFn, bool)
	GetOptGen(name string) (fn func() []byte, ok bool)
	Names() []string
}
