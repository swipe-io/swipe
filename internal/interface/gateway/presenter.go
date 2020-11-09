package gateway

import (
	"fmt"
	stdtypes "go/types"

	"github.com/swipe-io/swipe/v2/internal/domain/model"
	"github.com/swipe-io/swipe/v2/internal/errors"
	"github.com/swipe-io/swipe/v2/internal/option"
	"github.com/swipe-io/swipe/v2/internal/usecase/gateway"
	"golang.org/x/tools/go/packages"
)

type presenterGateway struct {
	pkg     *packages.Package
	o       *option.Option
	methods []model.PresenterMethod
	name    string
}

func (g *presenterGateway) Methods() []model.PresenterMethod {
	return g.methods
}

func (g *presenterGateway) Name() string {
	return g.name
}

func (g *presenterGateway) extractNamed(t stdtypes.Type) (result *stdtypes.Named, ok bool) {
	if ptr, ok := t.(*stdtypes.Pointer); ok {
		t = ptr.Elem()
	}
	result, ok = t.(*stdtypes.Named)
	return
}

func (g *presenterGateway) load(o *option.Option) error {
	ifaceOpt, ok := o.At("PresenterInterface")
	if !ok {
		return errors.NotePosition(o.Position, fmt.Errorf("the PresenterInterface option required"))
	}
	ifacePtr, ok := ifaceOpt.Value.Type().(*stdtypes.Pointer)
	if !ok {
		return errors.NotePosition(ifaceOpt.Position,
			fmt.Errorf("the interface option must be a pointer to an interface type; found %s", stdtypes.TypeString(ifaceOpt.Value.Type(), nil)))
	}
	iface, ok := ifacePtr.Elem().Underlying().(*stdtypes.Interface)
	if !ok {
		return errors.NotePosition(ifaceOpt.Position,
			fmt.Errorf("the interface option must be a pointer to an interface type; found %s", stdtypes.TypeString(ifaceOpt.Value.Type(), nil)))
	}
	typeName := ifacePtr.Elem().(*stdtypes.Named)

	g.name = typeName.Obj().Name()

	methods := make([]model.PresenterMethod, 0, typeName.NumMethods())

	for i := 0; i < iface.NumMethods(); i++ {
		m := iface.Method(i)
		sig := m.Type().(*stdtypes.Signature)

		if sig.Params().Len() == 1 && sig.Results().Len() == 1 {

			fromNamed, ok := g.extractNamed(sig.Params().At(0).Type())
			if !ok {
				continue
			}
			toNamed, ok := g.extractNamed(sig.Results().At(0).Type())
			if !ok {
				continue
			}

			methods = append(methods, model.PresenterMethod{
				Method: m,
				From:   fromNamed,
				To:     toNamed,
			})

		}
	}

	g.methods = methods

	//stFromPtr, ok := fromOpt.Value.Type().(*stdtypes.Pointer)
	//if !ok {
	//	return errors.NotePosition(o.Position,
	//		fmt.Errorf("the PresenterFrom option must be a pointer to an struct type; found %s", stdtypes.TypeString(fromOpt.Value.Type(), nil)))
	//}
	//stToPtr, ok := toOpt.Value.Type().(*stdtypes.Pointer)
	//if !ok {
	//	return errors.NotePosition(o.Position,
	//		fmt.Errorf("the PresenterTo option must be a pointer to an struct type; found %s", stdtypes.TypeString(toOpt.Value.Type(), nil)))
	//}
	//stFrom, ok := stFromPtr.Elem().(*stdtypes.Named)
	//if !ok {
	//	return errors.NotePosition(o.Position,
	//		fmt.Errorf("the PresenterFrom option must be a named struct type; found %s", stdtypes.TypeString(toOpt.Value.Type(), nil)))
	//}
	//stTo, ok := stToPtr.Elem().(*stdtypes.Named)
	//if !ok {
	//	return errors.NotePosition(o.Position,
	//		fmt.Errorf("the PresenterTo option must be a named struct type; found %s", stdtypes.TypeString(toOpt.Value.Type(), nil)))
	//}
	//g.to = stTo
	//g.from = stFrom
	return nil
}

func NewPresenterGateway(pkg *packages.Package, o *option.Option) (gateway.PresenterGateway, error) {
	g := &presenterGateway{pkg: pkg}
	if err := g.load(o); err != nil {
		fmt.Println(err)
		return nil, err
	}
	return g, nil
}
