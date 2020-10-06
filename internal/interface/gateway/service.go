package gateway

import (
	"fmt"
	"go/ast"
	"go/constant"
	stdtypes "go/types"
	stdstrings "strings"

	"github.com/swipe-io/strcase"
	"github.com/swipe-io/swipe/v2/internal/errors"
	"github.com/swipe-io/swipe/v2/internal/graph"
	"github.com/swipe-io/swipe/v2/internal/openapi"
	"github.com/swipe-io/swipe/v2/internal/strings"
	"github.com/swipe-io/swipe/v2/internal/types"
	"golang.org/x/tools/go/types/typeutil"

	"github.com/swipe-io/swipe/v2/internal/domain/model"
	"github.com/swipe-io/swipe/v2/internal/option"
	"github.com/swipe-io/swipe/v2/internal/usecase/gateway"
)

type serviceGateway struct {
	serviceID            string
	rawServiceID         string
	transport            model.TransportOption
	readme               model.ServiceReadme
	serviceType          stdtypes.Type
	serviceTypeName      *stdtypes.Named
	serviceIface         *stdtypes.Interface
	serviceMethods       []model.ServiceMethod
	graphTypes           *graph.Graph
	commentMap           *typeutil.Map
	defaultMethodOptions model.MethodHTTPTransportOption
	errors               map[uint32]*model.HTTPError
}

func (g *serviceGateway) Errors() map[uint32]*model.HTTPError {
	return g.errors
}

func (g *serviceGateway) InstrumentingEnable() bool {
	if g.defaultMethodOptions.InstrumentingEnable {
		return true
	}
	for _, transportOption := range g.transport.MethodOptions {
		if transportOption.InstrumentingEnable {
			return true
		}
	}
	return false
}

func (g *serviceGateway) LoggingEnable() bool {
	if g.defaultMethodOptions.LoggingEnable {
		return true
	}
	for _, transportOption := range g.transport.MethodOptions {
		if transportOption.LoggingEnable {
			return true
		}
	}
	return false
}

func (g *serviceGateway) ID() string {
	return g.serviceID
}

func (g *serviceGateway) RawID() string {
	return g.rawServiceID
}

func (g *serviceGateway) Transport() model.TransportOption {
	return g.transport
}

func (g *serviceGateway) Methods() []model.ServiceMethod {
	return g.serviceMethods
}

func (g *serviceGateway) Type() stdtypes.Type {
	return g.serviceType
}

func (g *serviceGateway) TypeName() *stdtypes.Named {
	return g.serviceTypeName
}

func (g *serviceGateway) Interface() *stdtypes.Interface {
	return g.serviceIface
}

func (g *serviceGateway) Readme() model.ServiceReadme {
	return g.readme
}

func (g *serviceGateway) load(o *option.Option) error {
	serviceOpt := option.MustOption(o.At("iface"))
	ifacePtr, ok := serviceOpt.Value.Type().(*stdtypes.Pointer)
	if !ok {
		return errors.NotePosition(serviceOpt.Position,
			fmt.Errorf("the Iface option is required must be a pointer to an interface type; found %s", stdtypes.TypeString(serviceOpt.Value.Type(), nil)))
	}
	iface, ok := ifacePtr.Elem().Underlying().(*stdtypes.Interface)
	if !ok {
		return errors.NotePosition(serviceOpt.Position,
			fmt.Errorf("the Iface option is required must be a pointer to an interface type; found %s", stdtypes.TypeString(serviceOpt.Value.Type(), nil)))
	}

	typeName := ifacePtr.Elem().(*stdtypes.Named)
	rawID := stdstrings.Split(typeName.Obj().Pkg().Path(), "/")[2]

	g.serviceID = strcase.ToCamel(rawID)
	g.rawServiceID = rawID

	if nameOpt, ok := o.At("Name"); ok {
		if name := nameOpt.Value.String(); name != "" {
			g.serviceID = strcase.ToCamel(name)
		}
	}

	g.serviceType = ifacePtr.Elem()
	g.serviceTypeName = typeName
	g.serviceIface = iface

	errorMethodName := "StatusCode"
	if g.transport.JsonRPC.Enable {
		errorMethodName = "ErrorCode"
	}

	hasher := typeutil.MakeHasher()

	g.graphTypes.Iterate(func(n *graph.Node) {
		g.graphTypes.Traverse(n, func(n *graph.Node) bool {
			if named, ok := n.Object.Type().(*stdtypes.Named); ok {
				key := hasher.Hash(named)
				if _, ok := g.errors[key]; ok {
					return true
				}
				if e := g.findError(named, errorMethodName); e != nil {
					g.errors[key] = e
				}
			}
			return true
		})
	})

	genericErrors := map[uint32]*model.HTTPError{}

	g.graphTypes.Iterate(func(n *graph.Node) {
		g.graphTypes.Traverse(n, func(n *graph.Node) bool {
			if sig, ok := n.Object.Type().(*stdtypes.Signature); ok {
				if sig.Results().Len() == 1 {
					if stdtypes.TypeString(sig.Results().At(0).Type(), nil) == "github.com/go-kit/kit/endpoint.Middleware" {
						g.graphTypes.Traverse(n, func(n *graph.Node) bool {
							if named, ok := n.Object.Type().(*stdtypes.Named); ok {
								key := hasher.Hash(named)
								if _, ok := genericErrors[key]; ok {
									return true
								}
								if e, ok := g.errors[key]; ok {
									genericErrors[key] = e
								}
							}
							return true
						})
					}
				}
			}
			return true
		})
	})

	for i := 0; i < iface.NumMethods(); i++ {
		m := iface.Method(i)

		sig := m.Type().(*stdtypes.Signature)
		comments, _ := g.commentMap.At(m.Type()).([]string)

		sm := model.ServiceMethod{
			Type:     m,
			T:        m.Type(),
			Name:     m.Name(),
			LcName:   strings.LcFirst(m.Name()),
			Errors:   genericErrors,
			Comments: comments,
		}

		g.graphTypes.Iterate(func(n *graph.Node) {
			g.graphTypes.Traverse(n, func(n *graph.Node) bool {
				if n.Object.Name() == m.Name() && stdtypes.Identical(n.Object.Type(), m.Type()) {
					g.graphTypes.Traverse(n, func(n *graph.Node) bool {
						if named, ok := n.Object.Type().(*stdtypes.Named); ok {
							key := hasher.Hash(named)
							if _, ok := sm.Errors[key]; ok {
								return true
							}
							if e, ok := g.errors[key]; ok {
								sm.Errors[key] = e
							}
						}
						return true
					})
					for _, value := range n.Values() {
						elem := value.Type
						if ptr, ok := value.Type.(*stdtypes.Pointer); ok {
							elem = ptr.Elem()
						}
						if named, ok := elem.(*stdtypes.Named); ok {
							key := hasher.Hash(named)
							if _, ok := sm.Errors[key]; ok {
								return true
							}
							if e, ok := g.errors[key]; ok {
								sm.Errors[key] = e
							}
						}
					}
				}
				return true
			})
		})

		var (
			resultOffset, paramOffset int
		)
		if types.ContainsContext(sig.Params()) {
			sm.ParamCtx = sig.Params().At(0)
			paramOffset = 1
		}
		if types.ContainsError(sig.Results()) {
			sm.ReturnErr = sig.Results().At(sig.Results().Len() - 1)
			resultOffset = 1
		}

		if types.IsNamed(sig.Results()) && sig.Results().Len()-resultOffset > 1 {
			sm.ResultsNamed = true
		}

		if !sm.ResultsNamed && sig.Results().Len()-resultOffset > 1 {
			return errors.NotePosition(serviceOpt.Position,
				fmt.Errorf("interface method with unnamed results cannot be greater than 1"))
		}
		for j := paramOffset; j < sig.Params().Len(); j++ {
			sm.Params = append(sm.Params, sig.Params().At(j))
		}
		for j := 0; j < sig.Results().Len()-resultOffset; j++ {
			sm.Results = append(sm.Results, sig.Results().At(j))
		}
		g.serviceMethods = append(g.serviceMethods, sm)
	}

	if transportOpt, ok := o.At("Transport"); ok {
		transportOption, err := g.loadTransport(transportOpt)
		if err != nil {
			return err
		}
		g.transport = transportOption
	}
	if opt, ok := o.At("Readme"); ok {
		g.readme.Enable = true
		if readmeTemplateOpt, ok := opt.At("ReadmeTemplate"); ok {
			g.readme.TemplatePath = readmeTemplateOpt.Value.String()
		}
		g.readme.OutputDir = option.MustOption(opt.At("outputDir")).Value.String()
	}
	return nil
}

func (g *serviceGateway) findError(named *stdtypes.Named, methodName string) *model.HTTPError {
	for i := 0; i < named.NumMethods(); i++ {
		if named.Method(i).Name() != methodName {
			continue
		}

		var isPointer bool
		sig := named.Method(i).Type().(*stdtypes.Signature)
		if sig.Recv() != nil {
			if _, ok := sig.Recv().Type().(*stdtypes.Pointer); ok {
				isPointer = true
			}
		}

		e := g.graphTypes.Node(named.Method(i))

		if e == nil {
			continue
		}
		if len(e.Values()) != 1 {
			continue
		}
		if code, ok := constant.Int64Val(e.Values()[0].Value); ok {
			return &model.HTTPError{
				Named:     named,
				Code:      code,
				IsPointer: isPointer,
			}
		}
	}
	return nil
}

func (g *serviceGateway) loadTransport(o *option.Option) (transportOption model.TransportOption, err error) {
	_, fastHTTP := o.At("FastEnable")
	transportOption = model.TransportOption{
		Protocol:      option.MustOption(o.At("protocol")).Value.String(),
		FastHTTP:      fastHTTP,
		MethodOptions: map[string]model.MethodHTTPTransportOption{},
		Openapi: model.OpenapiHTTPTransportOption{
			Methods: map[string]*model.OpenapiMethodOption{},
		},
	}
	if v, ok := o.At("MarkdownDoc"); ok {
		transportOption.MarkdownDoc.Enable = true
		transportOption.MarkdownDoc.OutputDir = v.Value.String()
	}
	if _, ok := o.At("ClientEnable"); ok {
		transportOption.Client.Enable = true
	}
	if _, ok := o.At("ServerDisabled"); ok {
		transportOption.ServerDisabled = true
	}
	if openapiDocOpt, ok := o.At("Openapi"); ok {
		transportOption.Openapi.Enable = true
		if v, ok := openapiDocOpt.At("OpenapiOutput"); ok {
			transportOption.Openapi.Output = v.Value.String()
		}
		if v, ok := openapiDocOpt.At("OpenapiInfo"); ok {
			transportOption.Openapi.Info = openapi.Info{
				Title:       option.MustOption(v.At("title")).Value.String(),
				Description: option.MustOption(v.At("description")).Value.String(),
				Version:     option.MustOption(v.At("version")).Value.String(),
			}
		}
		if v, ok := openapiDocOpt.At("OpenapiContact"); ok {
			transportOption.Openapi.Info.Contact = &openapi.Contact{
				Name:  option.MustOption(v.At("name")).Value.String(),
				Email: option.MustOption(v.At("email")).Value.String(),
				URL:   option.MustOption(v.At("url")).Value.String(),
			}
		}
		if v, ok := openapiDocOpt.At("OpenapiLicence"); ok {
			transportOption.Openapi.Info.License = &openapi.License{
				Name: option.MustOption(v.At("name")).Value.String(),
				URL:  option.MustOption(v.At("url")).Value.String(),
			}
		}
		if s, ok := openapiDocOpt.Slice("OpenapiServer"); ok {
			for _, v := range s {
				transportOption.Openapi.Servers = append(transportOption.Openapi.Servers, openapi.Server{
					Description: option.MustOption(v.At("description")).Value.String(),
					URL:         option.MustOption(v.At("url")).Value.String(),
				})
			}
		}
		if openapiTags, ok := openapiDocOpt.Slice("OpenapiTags"); ok {
			for _, openapiTagsOpt := range openapiTags {
				var methods []string
				if methodsOpt, ok := openapiTagsOpt.At("methods"); ok {
					for _, expr := range methodsOpt.Value.ExprSlice() {
						fnSel, ok := expr.(*ast.SelectorExpr)
						if !ok {
							return transportOption, errors.NotePosition(methodsOpt.Position, fmt.Errorf("the %s value must be func selector", methodsOpt.Name))
						}
						methods = append(methods, fnSel.Sel.Name)
						if _, ok := transportOption.Openapi.Methods[fnSel.Sel.Name]; !ok {
							transportOption.Openapi.Methods[fnSel.Sel.Name] = &model.OpenapiMethodOption{}
						}
					}
				}
				if tagsOpt, ok := openapiTagsOpt.At("tags"); ok {
					if len(methods) > 0 {
						for _, method := range methods {
							transportOption.Openapi.Methods[method].Tags = append(transportOption.Openapi.Methods[method].Tags, tagsOpt.Value.StringSlice()...)
						}
					} else {
						transportOption.Openapi.DefaultMethod.Tags = append(transportOption.Openapi.DefaultMethod.Tags, tagsOpt.Value.StringSlice()...)
					}
				}
			}
		}
		if transportOption.Openapi.Output == "" {
			transportOption.Openapi.Output = "./"
		}
	}
	if jsonRpcOpt, ok := o.At("JSONRPC"); ok {
		transportOption.JsonRPC.Enable = true
		if path, ok := jsonRpcOpt.At("JSONRPCPath"); ok {
			transportOption.JsonRPC.Path = path.Value.String()
		}
	}
	if methodDefaultOpt, ok := o.At("MethodDefaultOptions"); ok {
		defaultMethodOptions, err := getMethodOptions(methodDefaultOpt, model.MethodHTTPTransportOption{})
		if err != nil {
			return transportOption, err
		}

		for _, method := range g.serviceMethods {
			transportOption.MethodOptions[method.Name] = defaultMethodOptions
		}

		g.defaultMethodOptions = defaultMethodOptions
	}

	if methods, ok := o.Slice("MethodOptions"); ok {
		for _, methodOpt := range methods {
			signOpt := option.MustOption(methodOpt.At("signature"))
			fnSel, ok := signOpt.Value.Expr().(*ast.SelectorExpr)
			if !ok {
				return transportOption, errors.NotePosition(signOpt.Position, fmt.Errorf("the signature must be selector"))
			}

			baseMethodOpts := transportOption.MethodOptions[fnSel.Sel.Name]
			mopt, err := getMethodOptions(methodOpt, baseMethodOpts)
			if err != nil {
				return transportOption, err
			}
			transportOption.MethodOptions[fnSel.Sel.Name] = mopt
		}
	}

	transportOption.Prefix = "REST"
	if transportOption.JsonRPC.Enable {
		transportOption.Prefix = "JSONRPC"
	}

	return
}

func getMethodOptions(o *option.Option, baseMethodOpts model.MethodHTTPTransportOption) (model.MethodHTTPTransportOption, error) {
	if loggingOpt, ok := o.At("Logging"); ok {
		baseMethodOpts.LoggingEnable = loggingOpt.Value.Bool()
	}
	if loggingParamsOpt, ok := o.At("LoggingParams"); ok {
		baseMethodOpts.LoggingIncludeParams = map[string]struct{}{}
		baseMethodOpts.LoggingExcludeParams = map[string]struct{}{}

		includes := option.MustOption(loggingParamsOpt.At("includes")).Value.StringSlice()
		excludes := option.MustOption(loggingParamsOpt.At("excludes")).Value.StringSlice()
		for _, field := range includes {
			baseMethodOpts.LoggingIncludeParams[field] = struct{}{}
		}
		for _, field := range excludes {
			baseMethodOpts.LoggingExcludeParams[field] = struct{}{}
		}
	}
	if instrumentingOpt, ok := o.At("Instrumenting"); ok {
		baseMethodOpts.InstrumentingEnable = instrumentingOpt.Value.Bool()
	}
	if wrapResponseOpt, ok := o.At("WrapResponse"); ok {
		baseMethodOpts.WrapResponse.Enable = true
		baseMethodOpts.WrapResponse.Name = wrapResponseOpt.Value.String()
	}
	if httpMethodOpt, ok := o.At("Method"); ok {
		baseMethodOpts.MethodName = httpMethodOpt.Value.String()
		baseMethodOpts.Expr = httpMethodOpt.Value.Expr()
	}
	if path, ok := o.At("Path"); ok {
		baseMethodOpts.Path = path.Value.String()

		idxs, err := httpBraceIndices(baseMethodOpts.Path)
		if err != nil {
			return baseMethodOpts, err
		}
		if len(idxs) > 0 {
			baseMethodOpts.PathVars = make(map[string]string, len(idxs))

			var end int
			for i := 0; i < len(idxs); i += 2 {
				end = idxs[i+1]
				parts := stdstrings.SplitN(baseMethodOpts.Path[idxs[i]+1:end-1], ":", 2)

				name := parts[0]
				regexp := ""

				if len(parts) == 2 {
					regexp = parts[1]
				}
				baseMethodOpts.PathVars[name] = regexp
			}
		}
	}
	if serverRequestFunc, ok := o.At("ServerDecodeRequestFunc"); ok {
		baseMethodOpts.ServerRequestFunc.Type = serverRequestFunc.Value.Type()
		baseMethodOpts.ServerRequestFunc.Expr = serverRequestFunc.Value.Expr()
	}
	if serverResponseFunc, ok := o.At("ServerEncodeResponseFunc"); ok {
		baseMethodOpts.ServerResponseFunc.Type = serverResponseFunc.Value.Type()
		baseMethodOpts.ServerResponseFunc.Expr = serverResponseFunc.Value.Expr()
	}
	if clientRequestFunc, ok := o.At("ClientEncodeRequestFunc"); ok {
		baseMethodOpts.ClientRequestFunc.Type = clientRequestFunc.Value.Type()
		baseMethodOpts.ClientRequestFunc.Expr = clientRequestFunc.Value.Expr()
	}
	if clientResponseFunc, ok := o.At("ClientDecodeResponseFunc"); ok {
		baseMethodOpts.ClientResponseFunc.Type = clientResponseFunc.Value.Type()
		baseMethodOpts.ClientResponseFunc.Expr = clientResponseFunc.Value.Expr()
	}
	if queryVars, ok := o.At("QueryVars"); ok {
		baseMethodOpts.QueryVars = map[string]string{}
		values := queryVars.Value.StringSlice()
		for i := 0; i < len(values); i += 2 {
			baseMethodOpts.QueryVars[values[i]] = values[i+1]
		}
	}
	if headerVars, ok := o.At("HeaderVars"); ok {
		baseMethodOpts.HeaderVars = map[string]string{}
		values := headerVars.Value.StringSlice()
		for i := 0; i < len(values); i += 2 {
			baseMethodOpts.HeaderVars[values[i]] = values[i+1]
		}
	}
	return baseMethodOpts, nil
}

func httpBraceIndices(s string) ([]int, error) {
	var level, idx int
	var idxs []int
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '{':
			if level++; level == 1 {
				idx = i
			}
		case '}':
			if level--; level == 0 {
				idxs = append(idxs, idx, i+1)
			} else if level < 0 {
				return nil, fmt.Errorf("mux: unbalanced braces in %q", s)
			}
		}
	}
	if level != 0 {
		return nil, fmt.Errorf("mux: unbalanced braces in %q", s)
	}
	return idxs, nil
}

func NewServiceGateway(
	o *option.Option,
	graphTypes *graph.Graph,
	commentMap *typeutil.Map,
) (gateway.ServiceGateway, error) {
	g := &serviceGateway{
		graphTypes: graphTypes,
		commentMap: commentMap,
		errors:     map[uint32]*model.HTTPError{},
	}
	if err := g.load(o); err != nil {
		return nil, err
	}
	return g, nil
}
