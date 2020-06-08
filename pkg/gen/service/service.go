package service

import (
	"fmt"
	"go/types"

	"github.com/iancoleman/strcase"

	"github.com/swipe-io/swipe/pkg/errors"
	"github.com/swipe-io/swipe/pkg/parser"
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

type serviceCtx struct {
	id            string
	typeStr       string
	iface         *types.Interface
	logging       bool
	instrumenting serviceInstrumenting
}

func (w *Service) Write(opt *parser.Option) error {
	serviceOpt := parser.MustOption(opt.Get("iface"))
	ifacePtr, ok := serviceOpt.Value.Type().(*types.Pointer)
	if !ok {
		return errors.NotePosition(serviceOpt.Position,
			fmt.Errorf("the Interface option is required must be a pointer to an interface type; found %s", w.w.TypeString(serviceOpt.Value.Type())))
	}
	iface, ok := ifacePtr.Elem().Underlying().(*types.Interface)
	if !ok {
		return errors.NotePosition(serviceOpt.Position,
			fmt.Errorf("the Interface option is required must be a pointer to an interface type; found %s", w.w.TypeString(serviceOpt.Value.Type())))
	}
	typeStr := w.w.TypeString(ifacePtr.Elem())
	id := strcase.ToCamel(typeStr)

	ctx := serviceCtx{
		id:      id,
		typeStr: typeStr,
		iface:   iface,
	}

	if err := newEndpoint(ctx, w.w).Write(); err != nil {
		return err
	}
	if _, ok := opt.Get("Logging"); ok {
		ctx.logging = true
		if err := newLogging(ctx, w.w).Write(); err != nil {
			return err
		}
	}
	if instrumentingOpt, ok := opt.Get("Instrumenting"); ok {
		ctx.instrumenting.enable = true
		if namespace, ok := instrumentingOpt.Get("Namespace"); ok {
			ctx.instrumenting.namespace = namespace.Value.String()
		}
		if subsystem, ok := instrumentingOpt.Get("Subsystem"); ok {
			ctx.instrumenting.subsystem = subsystem.Value.String()
		}
		if err := newInstrumenting(ctx, w.w).Write(); err != nil {
			return err
		}
	}

	if transportOpt, ok := opt.Get("Transport"); ok {
		if err := newTransport(ctx, w.w).Write(transportOpt); err != nil {
			return err
		}
	}

	return nil
}

func New(w *writer.Writer) *Service {
	return &Service{w: w}
}
