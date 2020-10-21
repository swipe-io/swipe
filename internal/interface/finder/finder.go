package finder

import (
	"go/build"
	stdtypes "go/types"
	"path/filepath"
	stdstrings "strings"

	"github.com/swipe-io/swipe/v2/internal/domain/model"

	ig "github.com/swipe-io/swipe/v2/internal/interface/gateway"

	"github.com/swipe-io/swipe/v2/internal/option"

	"github.com/swipe-io/swipe/v2/internal/usecase/finder"
	"github.com/swipe-io/swipe/v2/internal/usecase/gateway"
)

type serviceFinder struct {
	loader *option.Loader
}

func (s *serviceFinder) Find(named *stdtypes.Named) (gateway.ServiceGateway, *model.ServiceInterface, []error) {
	pkgPathParts := stdstrings.Split(named.Obj().Pkg().Path(), "/")
	servicePath := filepath.Join(build.Default.GOPATH, "src", stdstrings.Join(pkgPathParts[:3], "/"))

	o, errs := s.loader.Load(servicePath, nil, []string{"./..."})
	if len(errs) > 0 {
		return nil, nil, errs
	}
	for _, resultOption := range o.Options {
		if resultOption.Option.Name == "Service" {
			sg, err := ig.NewServiceGateway(resultOption.Pkg, resultOption.Option, o.Data.GraphTypes, o.Data.Enums)
			if err != nil {
				return nil, nil, []error{err}
			}
			for i := 0; i < sg.Interfaces().Len(); i++ {
				iface := sg.Interfaces().At(i)
				if iface.TypeName().Obj().String() == named.Obj().String() && sg.JSONRPCEnable() {
					return sg, iface, nil
				}
			}
		}
	}
	return nil, nil, nil
}

func NewServiceFinder(loader *option.Loader) finder.ServiceFinder {
	return &serviceFinder{loader: loader}
}
