package service

import (
	"fmt"
	stdtypes "go/types"

	"github.com/iancoleman/strcase"

	"github.com/swipe-io/swipe/pkg/errors"
	"github.com/swipe-io/swipe/pkg/parser"
	"github.com/swipe-io/swipe/pkg/strings"
	"github.com/swipe-io/swipe/pkg/types"
	"github.com/swipe-io/swipe/pkg/writer"
)

type Service struct {
	w *writer.Writer
}

type serviceInstrumenting struct {
	enable    bool
	namespace string
	subsystem string
}

type varSlice []*stdtypes.Var

func (s varSlice) lookupField(name string) *stdtypes.Var {
	for _, p := range s {
		if p.Name() == name {
			return p
		}
	}
	return nil
}

type ifaceServiceMethod struct {
	t            *stdtypes.Func
	name         string
	lcName       string
	params       varSlice
	results      varSlice
	comments     []string
	paramCtx     *stdtypes.Var
	returnErr    *stdtypes.Var
	resultsNamed bool
}

type ifaceService struct {
	t       *stdtypes.Interface
	methods map[string]ifaceServiceMethod
}

type serviceCtx struct {
	id            string
	typeStr       string
	iface         ifaceService
	logging       bool
	instrumenting serviceInstrumenting
}

func (w *Service) Write(opt *parser.Option) error {
	serviceOpt := parser.MustOption(opt.At("iface"))

	ifacePtr, ok := serviceOpt.Value.Type().(*stdtypes.Pointer)
	if !ok {
		return errors.NotePosition(serviceOpt.Position,
			fmt.Errorf("the Interface option is required must be a pointer to an interface type; found %s", w.w.TypeString(serviceOpt.Value.Type())))
	}
	iface, ok := ifacePtr.Elem().Underlying().(*stdtypes.Interface)
	if !ok {
		return errors.NotePosition(serviceOpt.Position,
			fmt.Errorf("the Interface option is required must be a pointer to an interface type; found %s", w.w.TypeString(serviceOpt.Value.Type())))
	}

	methodsComments := w.w.FindComments(iface)

	typeStr := w.w.TypeString(ifacePtr.Elem())
	id := strcase.ToCamel(typeStr)

	ctx := serviceCtx{
		id:      id,
		typeStr: typeStr,
		iface: ifaceService{
			t:       iface,
			methods: make(map[string]ifaceServiceMethod, iface.NumMethods()),
		},
	}

	for i := 0; i < iface.NumMethods(); i++ {
		m := iface.Method(i)
		sig := m.Type().(*stdtypes.Signature)

		sm := ifaceServiceMethod{
			t:        m,
			name:     m.Name(),
			comments: methodsComments[m.Name()],
			lcName:   strings.LcFirst(m.Name()),
		}
		var (
			resultOffset, paramOffset int
		)
		if types.ContainsContext(sig.Params()) {
			sm.paramCtx = sig.Params().At(0)
			paramOffset = 1
		}
		if types.ContainsError(sig.Results()) {
			sm.returnErr = sig.Results().At(sig.Results().Len() - 1)
			resultOffset = 1
		}
		if types.IsNamed(sig.Results()) {
			sm.resultsNamed = true
		}

		if !sm.resultsNamed && sig.Results().Len()-resultOffset > 1 {
			return errors.NotePosition(serviceOpt.Position,
				fmt.Errorf("interface method with unnamed results cannot be greater than 1"))
		}

		for j := paramOffset; j < sig.Params().Len(); j++ {
			sm.params = append(sm.params, sig.Params().At(j))
		}
		for j := 0; j < sig.Results().Len()-resultOffset; j++ {
			sm.results = append(sm.results, sig.Results().At(j))
		}
		ctx.iface.methods[m.Name()] = sm
	}

	if err := newEndpoint(ctx, w.w).Write(); err != nil {
		return err
	}
	if _, ok := opt.At("Logging"); ok {
		ctx.logging = true
		if err := newLogging(ctx, w.w).Write(); err != nil {
			return err
		}
	}
	if instrumentingOpt, ok := opt.At("Instrumenting"); ok {
		ctx.instrumenting.enable = true
		if namespace, ok := instrumentingOpt.At("Namespace"); ok {
			ctx.instrumenting.namespace = namespace.Value.String()
		}
		if subsystem, ok := instrumentingOpt.At("Subsystem"); ok {
			ctx.instrumenting.subsystem = subsystem.Value.String()
		}
		if err := newInstrumenting(ctx, w.w).Write(); err != nil {
			return err
		}
	}

	if transportOpt, ok := opt.At("Transport"); ok {
		if err := newTransport(ctx, w.w).Write(transportOpt); err != nil {
			return err
		}
	}

	return nil
}

func New(w *writer.Writer) *Service {
	return &Service{w: w}
}
