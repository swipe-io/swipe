package gateway

import (
	"fmt"
	"go/ast"
	stdtypes "go/types"

	"github.com/swipe-io/swipe/v2/internal/domain/model"
	"github.com/swipe-io/swipe/v2/internal/errors"
	"github.com/swipe-io/swipe/v2/internal/option"
	"github.com/swipe-io/swipe/v2/internal/usecase/finder"
	"github.com/swipe-io/swipe/v2/internal/usecase/gateway"

	"golang.org/x/tools/go/packages"
)

type httpGatewayGateway struct {
	pkg      *packages.Package
	services []model.GatewayServiceOption
	finder   finder.ServiceFinder
}

func (g *httpGatewayGateway) Services() []model.GatewayServiceOption {
	return g.services
}

func (g *httpGatewayGateway) parseMethodOption(iface *stdtypes.Interface, o *option.Option) (model.GatewayMethodOption, error) {
	methodOption := model.GatewayMethodOption{}
	signOpt := option.MustOption(o.At("signature"))
	fnSel, ok := signOpt.Value.Expr().(*ast.SelectorExpr)
	if !ok {
		return model.GatewayMethodOption{}, errors.NotePosition(signOpt.Position, fmt.Errorf("the signature must be selector"))
	}
	methodOption.Name = fnSel.Sel.Name
	ifaceSel := g.pkg.TypesInfo.TypeOf(fnSel.X).Underlying()
	if !stdtypes.Identical(iface, ifaceSel) {
		return model.GatewayMethodOption{}, errors.NotePosition(
			signOpt.Position,
			fmt.Errorf(
				"the method signature does not match the interface, now %s should be %s",
				ifaceSel,
				iface,
			),
		)
	}
	if opt, ok := o.At("GatewayBalancer"); ok {
		v := opt.Value.String()
		if v != "random" && v != "roundrobin" {
			return model.GatewayMethodOption{}, errors.NotePosition(opt.Position, fmt.Errorf("there can only be values: random, roundrobin"))
		}
	}
	return methodOption, nil
}

func (g *httpGatewayGateway) load(o *option.Option) error {
	serviceOpts, _ := o.Slice("GatewayService")

	for _, serviceOpt := range serviceOpts {
		ifaceOpt := option.MustOption(serviceOpt.At("iface"))
		ifacePtr, ok := ifaceOpt.Value.Type().(*stdtypes.Pointer)
		if !ok {
			return errors.NotePosition(o.Position,
				fmt.Errorf("the Iface option is required must be a pointer to an interface type; found %s", stdtypes.TypeString(o.Value.Type(), nil)))
		}

		typeName := ifacePtr.Elem().(*stdtypes.Named)

		_, iface, errs := g.finder.Find(typeName)
		if len(errs) > 0 {
			continue
		}

		so := model.GatewayServiceOption{
			Iface:         iface,
			MethodOptions: map[string]model.GatewayMethodOption{},
		}
		if methodOpt, ok := serviceOpt.At("GatewayServiceMethod"); ok {
			mo, err := g.parseMethodOption(iface.Interface(), methodOpt)
			if err != nil {
				return err
			}
			so.MethodOptions[mo.Name] = mo
		}
		g.services = append(g.services, so)
	}
	return nil
}

func NewGateway(
	pkg *packages.Package,
	o *option.Option,
	finder finder.ServiceFinder,
) (gateway.HTTPGatewayGateway, error) {
	g := &httpGatewayGateway{pkg: pkg, finder: finder}
	if err := g.load(o); err != nil {
		return nil, err
	}
	return g, nil
}
