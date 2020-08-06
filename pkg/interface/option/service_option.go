package option

import (
	"fmt"
	"go/ast"
	"go/constant"
	stdtypes "go/types"
	stdstrings "strings"

	"github.com/iancoleman/strcase"

	"github.com/swipe-io/swipe/pkg/domain/model"
	"github.com/swipe-io/swipe/pkg/errors"
	"github.com/swipe-io/swipe/pkg/graph"
	"github.com/swipe-io/swipe/pkg/openapi"
	"github.com/swipe-io/swipe/pkg/parser"
	"github.com/swipe-io/swipe/pkg/strings"
	"github.com/swipe-io/swipe/pkg/types"
	"github.com/swipe-io/swipe/pkg/usecase/option"

	"golang.org/x/tools/go/types/typeutil"
)

type ErrorData struct {
	Named *stdtypes.Named
	Code  int64
}

type serviceOption struct {
	info model.GenerateInfo
}

func (g *serviceOption) Parse(option *parser.Option) (interface{}, error) {
	o := model.ServiceOption{}

	serviceOpt := parser.MustOption(option.At("iface"))
	ifacePtr, ok := serviceOpt.Value.Type().(*stdtypes.Pointer)
	if !ok {
		return nil, errors.NotePosition(serviceOpt.Position,
			fmt.Errorf("the Iface option is required must be a pointer to an interface type; found %s", stdtypes.TypeString(serviceOpt.Value.Type(), nil)))
	}
	iface, ok := ifacePtr.Elem().Underlying().(*stdtypes.Interface)
	if !ok {
		return nil, errors.NotePosition(serviceOpt.Position,
			fmt.Errorf("the Iface option is required must be a pointer to an interface type; found %s", stdtypes.TypeString(serviceOpt.Value.Type(), nil)))
	}

	typeName := ifacePtr.Elem().(*stdtypes.Named)
	rawID := stdstrings.Split(typeName.Obj().Pkg().Path(), "/")[2]

	o.ID = strcase.ToCamel(rawID)
	o.RawID = rawID

	if transportOpt, ok := option.At("Transport"); ok {
		transportOption, err := g.loadTransport(transportOpt)
		if err != nil {
			return nil, err
		}
		o.Transport = transportOption
	}
	if opt, ok := option.At("Readme"); ok {
		o.Readme.Enable = true
		if readmeTemplateOpt, ok := opt.At("ReadmeTemplate"); ok {
			o.Readme.TemplatePath = readmeTemplateOpt.Value.String()
		}
		o.Readme.OutputDir = parser.MustOption(opt.At("outputDir")).Value.String()

	}
	o.Type = ifacePtr.Elem()
	o.TypeName = typeName
	o.Interface = iface

	errorMethodName := "StatusCode"
	if o.Transport.JsonRPC.Enable {
		errorMethodName = "ErrorCode"
	}

	hasher := typeutil.MakeHasher()

	g.info.GraphTypes.Iterate(func(n *graph.Node) {
		g.info.GraphTypes.Traverse(n, func(n *graph.Node) bool {
			if named, ok := n.Object.Type().(*stdtypes.Named); ok {
				key := hasher.Hash(named)
				if _, ok := o.Transport.Errors[key]; ok {
					return true
				}
				if e := g.findError(named, errorMethodName); e != nil {
					o.Transport.Errors[key] = e
				}
			}
			return true
		})
	})

	genericErrors := map[uint32]*model.ErrorHTTPTransportOption{}

	g.info.GraphTypes.Iterate(func(n *graph.Node) {
		g.info.GraphTypes.Traverse(n, func(n *graph.Node) bool {
			if sig, ok := n.Object.Type().(*stdtypes.Signature); ok {
				if sig.Results().Len() == 1 {
					if stdtypes.TypeString(sig.Results().At(0).Type(), nil) == "github.com/go-kit/kit/endpoint.Middleware" {
						g.info.GraphTypes.Traverse(n, func(n *graph.Node) bool {
							if named, ok := n.Object.Type().(*stdtypes.Named); ok {
								key := hasher.Hash(named)
								if _, ok := genericErrors[key]; ok {
									return true
								}
								if e, ok := o.Transport.Errors[key]; ok {
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
		comments, _ := g.info.CommentMap.At(m.Type()).([]string)

		sm := model.ServiceMethod{
			Type:     m,
			T:        m.Type(),
			Name:     m.Name(),
			LcName:   strings.LcFirst(m.Name()),
			Errors:   genericErrors,
			Comments: comments,
		}

		g.info.GraphTypes.Iterate(func(n *graph.Node) {
			g.info.GraphTypes.Traverse(n, func(n *graph.Node) bool {
				if n.Object.Name() == m.Name() && stdtypes.Identical(n.Object.Type(), m.Type()) {
					g.info.GraphTypes.Traverse(n, func(n *graph.Node) bool {
						if named, ok := n.Object.Type().(*stdtypes.Named); ok {
							key := hasher.Hash(named)
							if _, ok := sm.Errors[key]; ok {
								return true
							}
							if e, ok := o.Transport.Errors[key]; ok {
								sm.Errors[key] = e
							}
						}
						return true
					})
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
		if types.IsNamed(sig.Results()) {
			sm.ResultsNamed = true
		}
		if !sm.ResultsNamed && sig.Results().Len()-resultOffset > 1 {
			return nil, errors.NotePosition(serviceOpt.Position,
				fmt.Errorf("interface method with unnamed results cannot be greater than 1"))
		}
		for j := paramOffset; j < sig.Params().Len(); j++ {
			sm.Params = append(sm.Params, sig.Params().At(j))
		}
		for j := 0; j < sig.Results().Len()-resultOffset; j++ {
			sm.Results = append(sm.Results, sig.Results().At(j))
		}
		o.Methods = append(o.Methods, sm)
	}

	if _, ok := option.At("Logging"); ok {
		o.Logging = true
	}

	if instrumentingOpt, ok := option.At("Instrumenting"); ok {
		o.Instrumenting.Enable = true
		if namespace, ok := instrumentingOpt.At("namespace"); ok {
			o.Instrumenting.Namespace = namespace.Value.String()
		}
		if subsystem, ok := instrumentingOpt.At("subsystem"); ok {
			o.Instrumenting.Subsystem = subsystem.Value.String()
		}
	}

	return o, nil
}

func (g *serviceOption) findError(named *stdtypes.Named, methodName string) *model.ErrorHTTPTransportOption {
	for i := 0; i < named.NumMethods(); i++ {
		if named.Method(i).Name() != methodName {
			continue
		}
		e := g.info.GraphTypes.Node(named.Method(i))
		if e == nil {
			continue
		}
		if len(e.Values()) != 1 {
			continue
		}
		if code, ok := constant.Int64Val(e.Values()[0].Value); ok {
			return &model.ErrorHTTPTransportOption{
				Named: named,
				Code:  code,
			}
		}
	}
	return nil
}

func (g *serviceOption) loadTransport(opt *parser.Option) (option model.TransportOption, err error) {
	_, fastHTTP := opt.At("FastEnable")
	option = model.TransportOption{
		Protocol:      parser.MustOption(opt.At("protocol")).Value.String(),
		FastHTTP:      fastHTTP,
		MethodOptions: map[string]model.MethodHTTPTransportOption{},
		Errors:        map[uint32]*model.ErrorHTTPTransportOption{},
		Openapi: model.OpenapiHTTPTransportOption{
			Methods: map[string]*model.OpenapiMethodOption{},
		},
	}
	if v, ok := opt.At("MarkdownDoc"); ok {
		option.MarkdownDoc.Enable = true
		option.MarkdownDoc.OutputDir = v.Value.String()
	}
	if _, ok := opt.At("ClientEnable"); ok {
		option.Client.Enable = true
	}
	if _, ok := opt.At("ServerDisabled"); ok {
		option.ServerDisabled = true
	}
	if openapiDocOpt, ok := opt.At("Openapi"); ok {
		option.Openapi.Enable = true
		if v, ok := openapiDocOpt.At("OpenapiOutput"); ok {
			option.Openapi.Output = v.Value.String()
		}
		if v, ok := openapiDocOpt.At("OpenapiInfo"); ok {
			option.Openapi.Info = openapi.Info{
				Title:       parser.MustOption(v.At("title")).Value.String(),
				Description: parser.MustOption(v.At("description")).Value.String(),
				Version:     parser.MustOption(v.At("version")).Value.String(),
			}
		}
		if v, ok := openapiDocOpt.At("OpenapiContact"); ok {
			option.Openapi.Info.Contact = &openapi.Contact{
				Name:  parser.MustOption(v.At("name")).Value.String(),
				Email: parser.MustOption(v.At("email")).Value.String(),
				URL:   parser.MustOption(v.At("url")).Value.String(),
			}
		}
		if v, ok := openapiDocOpt.At("OpenapiLicence"); ok {
			option.Openapi.Info.License = &openapi.License{
				Name: parser.MustOption(v.At("name")).Value.String(),
				URL:  parser.MustOption(v.At("url")).Value.String(),
			}
		}
		if s, ok := openapiDocOpt.Slice("OpenapiServer"); ok {
			for _, v := range s {
				option.Openapi.Servers = append(option.Openapi.Servers, openapi.Server{
					Description: parser.MustOption(v.At("description")).Value.String(),
					URL:         parser.MustOption(v.At("url")).Value.String(),
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
							return option, errors.NotePosition(methodsOpt.Position, fmt.Errorf("the %s value must be func selector", methodsOpt.Name))
						}
						methods = append(methods, fnSel.Sel.Name)
						if _, ok := option.Openapi.Methods[fnSel.Sel.Name]; !ok {
							option.Openapi.Methods[fnSel.Sel.Name] = &model.OpenapiMethodOption{}
						}
					}
				}
				if tagsOpt, ok := openapiTagsOpt.At("tags"); ok {
					if len(methods) > 0 {
						for _, method := range methods {
							option.Openapi.Methods[method].Tags = append(option.Openapi.Methods[method].Tags, tagsOpt.Value.StringSlice()...)
						}
					} else {
						option.Openapi.DefaultMethod.Tags = append(option.Openapi.DefaultMethod.Tags, tagsOpt.Value.StringSlice()...)
					}
				}
			}
		}
		if option.Openapi.Output == "" {
			option.Openapi.Output = "./"
		}
	}
	if jsonRpcOpt, ok := opt.At("JSONRPC"); ok {
		option.JsonRPC.Enable = true
		if path, ok := jsonRpcOpt.At("JSONRPCPath"); ok {
			option.JsonRPC.Path = path.Value.String()
		}
	}
	if methodDefaultOpt, ok := opt.At("MethodDefaultOptions"); ok {
		defaultMethodOptions, err := getMethodOptions(methodDefaultOpt, model.MethodHTTPTransportOption{})
		if err != nil {
			return option, err
		}
		option.DefaultMethodOptions = defaultMethodOptions
	}

	if methods, ok := opt.Slice("MethodOptions"); ok {
		for _, methodOpt := range methods {
			signOpt := parser.MustOption(methodOpt.At("signature"))
			fnSel, ok := signOpt.Value.Expr().(*ast.SelectorExpr)
			if !ok {
				return option, errors.NotePosition(signOpt.Position, fmt.Errorf("the signature must be selector"))
			}
			baseMethodOpts := option.MethodOptions[fnSel.Sel.Name]
			mopt, err := getMethodOptions(methodOpt, baseMethodOpts)
			if err != nil {
				return option, err
			}
			option.MethodOptions[fnSel.Sel.Name] = mopt
		}
	}

	option.Prefix = "REST"
	if option.JsonRPC.Enable {
		option.Prefix = "JSONRPC"
	}

	return
}

func NewServiceOption(info model.GenerateInfo) option.Option {
	return &serviceOption{
		info: info,
	}
}

func getMethodOptions(methodOpt *parser.Option, baseMethodOpts model.MethodHTTPTransportOption) (model.MethodHTTPTransportOption, error) {
	if wrapResponseOpt, ok := methodOpt.At("WrapResponse"); ok {
		baseMethodOpts.WrapResponse.Enable = true
		baseMethodOpts.WrapResponse.Name = wrapResponseOpt.Value.String()
	}
	if httpMethodOpt, ok := methodOpt.At("Method"); ok {
		baseMethodOpts.MethodName = httpMethodOpt.Value.String()
		baseMethodOpts.Expr = httpMethodOpt.Value.Expr()
	}
	if path, ok := methodOpt.At("Path"); ok {
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
	if serverRequestFunc, ok := methodOpt.At("ServerDecodeRequestFunc"); ok {
		baseMethodOpts.ServerRequestFunc.Type = serverRequestFunc.Value.Type()
		baseMethodOpts.ServerRequestFunc.Expr = serverRequestFunc.Value.Expr()
	}
	if serverResponseFunc, ok := methodOpt.At("ServerEncodeResponseFunc"); ok {
		baseMethodOpts.ServerResponseFunc.Type = serverResponseFunc.Value.Type()
		baseMethodOpts.ServerResponseFunc.Expr = serverResponseFunc.Value.Expr()
	}
	if clientRequestFunc, ok := methodOpt.At("ClientEncodeRequestFunc"); ok {
		baseMethodOpts.ClientRequestFunc.Type = clientRequestFunc.Value.Type()
		baseMethodOpts.ClientRequestFunc.Expr = clientRequestFunc.Value.Expr()
	}
	if clientResponseFunc, ok := methodOpt.At("ClientDecodeResponseFunc"); ok {
		baseMethodOpts.ClientResponseFunc.Type = clientResponseFunc.Value.Type()
		baseMethodOpts.ClientResponseFunc.Expr = clientResponseFunc.Value.Expr()
	}
	if queryVars, ok := methodOpt.At("QueryVars"); ok {
		baseMethodOpts.QueryVars = map[string]string{}
		values := queryVars.Value.StringSlice()
		for i := 0; i < len(values); i += 2 {
			baseMethodOpts.QueryVars[values[i]] = values[i+1]
		}
	}
	if headerVars, ok := methodOpt.At("HeaderVars"); ok {
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
