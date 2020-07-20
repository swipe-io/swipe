package option

import (
	"fmt"
	"go/ast"
	stdtypes "go/types"
	stdstrings "strings"

	"github.com/iancoleman/strcase"

	"github.com/swipe-io/swipe/pkg/domain/model"
	"github.com/swipe-io/swipe/pkg/errors"
	"github.com/swipe-io/swipe/pkg/parser"
	"github.com/swipe-io/swipe/pkg/usecase/option"
)

type gatewayOption struct {
	info model.GenerateInfo
}

func (g *gatewayOption) parseMethodOption(iface *stdtypes.Interface, option *parser.Option) (model.GatewayMethodOption, error) {
	o := model.GatewayMethodOption{}
	signOpt := parser.MustOption(option.At("signature"))
	fnSel, ok := signOpt.Value.Expr().(*ast.SelectorExpr)
	if !ok {
		return model.GatewayMethodOption{}, errors.NotePosition(signOpt.Position, fmt.Errorf("the signature must be selector"))
	}
	o.Name = fnSel.Sel.Name
	ifaceSel := g.info.Pkg.TypesInfo.TypeOf(fnSel.X).Underlying()
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
	if opt, ok := option.At("GatewayBalancer"); ok {
		v := opt.Value.String()
		if v != "random" && v != "roundrobin" {
			return model.GatewayMethodOption{}, errors.NotePosition(opt.Position, fmt.Errorf("there can only be values: random, roundrobin"))
		}
	}
	return o, nil
}

func (g *gatewayOption) Parse(option *parser.Option) (interface{}, error) {
	o := model.GatewayOption{}

	serviceOpts, _ := option.Slice("GatewayService")

	for _, serviceOpt := range serviceOpts {
		so := model.GatewayServiceOption{
			MethodOptions: map[string]model.GatewayMethodOption{},
		}
		ifaceOpt := parser.MustOption(serviceOpt.At("iface"))
		ifacePtr, ok := ifaceOpt.Value.Type().(*stdtypes.Pointer)
		if !ok {
			return nil, errors.NotePosition(option.Position,
				fmt.Errorf("the Iface option is required must be a pointer to an interface type; found %s", stdtypes.TypeString(option.Value.Type(), nil)))
		}
		iface, ok := ifacePtr.Elem().Underlying().(*stdtypes.Interface)
		if !ok {
			return nil, errors.NotePosition(option.Position,
				fmt.Errorf("the Iface option is required must be a pointer to an interface type; found %s", stdtypes.TypeString(option.Value.Type(), nil)))
		}

		typeName := ifacePtr.Elem().(*stdtypes.Named)
		rawID := stdstrings.Split(typeName.Obj().Pkg().Path(), "/")[2]

		so.ID = strcase.ToCamel(rawID)
		so.RawID = rawID
		so.Type = ifacePtr.Elem()
		so.TypeName = typeName
		so.Iface = iface

		if methodOpt, ok := serviceOpt.At("GatewayServiceMethod"); ok {
			mo, err := g.parseMethodOption(iface, methodOpt)
			if err != nil {
				return nil, err
			}
			so.MethodOptions[mo.Name] = mo
		}
		o.Services = append(o.Services, so)
	}
	return o, nil
}

func NewGatewayOption(info model.GenerateInfo) option.Option {
	return &gatewayOption{info: info}
}
