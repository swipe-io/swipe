package generator

import (
	"context"
	"encoding/json"

	"github.com/swipe-io/swipe/v3/option"

	"github.com/swipe-io/swipe/v3/internal/finder"
	"github.com/swipe-io/swipe/v3/internal/openapi"
	"github.com/swipe-io/swipe/v3/internal/plugin/echo/config"
)

type Openapi struct {
	Contact       config.OpenapiContact
	Info          config.OpenapiInfo
	MethodTags    map[string][]string
	Servers       []config.OpenapiServer
	Licence       config.OpenapiLicence
	Output        string
	Interfaces    []*config.Interface
	MethodOptions map[string]config.MethodOptions
	IfaceErrors   map[string]map[string][]finder.Error
}

func (g *Openapi) Generate(ctx context.Context) []byte {
	var interfaces []openapi.Interface
	for _, iface := range g.Interfaces {
		ifaceType := iface.Named.Type.(*option.IfaceType)
		openapiIface := openapi.Interface{
			Name:      iface.Named.Name,
			Namespace: iface.Namespace,
		}
		for _, m := range ifaceType.Methods {
			mopt := g.MethodOptions[iface.Named.Name.Value+m.Name.Value]
			tags := g.MethodTags[iface.Named.Name.Value+m.Name.Value]
			openapiIface.Methods = append(openapiIface.Methods, openapi.InterfaceMethod{
				Name:             m.Name,
				RESTMethod:       mopt.RESTMethod.Take(),
				RESTPath:         mopt.RESTPath.Take(),
				RESTQueryVars:    mopt.RESTQueryVars.Value,
				RESTPathVars:     mopt.RESTPathVars,
				Tags:             tags,
				Func:             m,
				Description:      m.Comment,
				RESTWrapResponse: mopt.RESTWrapResponse.Take(),
				RESTQueryValues:  mopt.RESTQueryValues.Value,
				RESTHeaderVars:   mopt.RESTHeaderVars.Value,
			})
		}
		interfaces = append(interfaces, openapiIface)
	}

	o := openapi.NewOpenapi(
		openapi.Info{},
		[]openapi.Server{},
		interfaces,
		map[string]map[string][]openapi.Error{},
		false,
	)
	result := o.Build()
	data, _ := json.MarshalIndent(result, "", " ")
	return data
}

func (g *Openapi) OutputPath() string {
	return g.Output
}

func (g *Openapi) Filename() string {
	return "openapi.json"
}
