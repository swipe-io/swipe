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
	"github.com/swipe-io/swipe/pkg/openapi"
	"github.com/swipe-io/swipe/pkg/parser"
	"github.com/swipe-io/swipe/pkg/strings"
	"github.com/swipe-io/swipe/pkg/types"
	"github.com/swipe-io/swipe/pkg/usecase/option"

	"golang.org/x/tools/go/packages"
)

type serviceOption struct {
	info model.GenerateInfo
}

func (g *serviceOption) Parse(option *parser.Option) (interface{}, error) {
	o := model.ServiceOption{
		Methods: map[string]model.ServiceMethod{},
	}

	serviceOpt := parser.MustOption(option.At("iface"))
	ifacePtr, ok := serviceOpt.Value.Type().(*stdtypes.Pointer)
	if !ok {
		return nil, errors.NotePosition(serviceOpt.Position,
			fmt.Errorf("the Interface option is required must be a pointer to an interface type; found %s", stdtypes.TypeString(serviceOpt.Value.Type(), nil)))
	}
	iface, ok := ifacePtr.Elem().Underlying().(*stdtypes.Interface)
	if !ok {
		return nil, errors.NotePosition(serviceOpt.Position,
			fmt.Errorf("the Interface option is required must be a pointer to an interface type; found %s", stdtypes.TypeString(serviceOpt.Value.Type(), nil)))
	}

	o.ID = strcase.ToCamel(stdtypes.TypeString(ifacePtr.Elem(), func(p *stdtypes.Package) string {
		return p.Name()
	}))
	o.Type = ifacePtr.Elem()
	o.Interface = iface

	for i := 0; i < iface.NumMethods(); i++ {
		m := iface.Method(i)
		sig := m.Type().(*stdtypes.Signature)
		comments, _ := g.info.CommentMap.At(m.Type()).([]string)
		sm := model.ServiceMethod{
			Type:     m,
			Name:     m.Name(),
			LcName:   strings.LcFirst(m.Name()),
			Comments: comments,
		}
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
		o.Methods[m.Name()] = sm
	}

	if _, ok := option.At("Logging"); ok {
		o.Logging = true
	}

	if instrumentingOpt, ok := option.At("Instrumenting"); ok {
		o.Instrumenting.Enable = true
		if namespace, ok := instrumentingOpt.At("Namespace"); ok {
			o.Instrumenting.Namespace = namespace.Value.String()
		}
		if subsystem, ok := instrumentingOpt.At("Subsystem"); ok {
			o.Instrumenting.Subsystem = subsystem.Value.String()
		}
	}
	if transportOpt, ok := option.At("Transport"); ok {
		transportOption, err := g.loadTransport(transportOpt)
		if err != nil {
			return nil, err
		}
		o.Transport = transportOption
	}
	return o, nil
}

func (g *serviceOption) loadTransport(opt *parser.Option) (option model.TransportOption, err error) {
	_, fastHTTP := opt.At("FastEnable")
	option = model.TransportOption{
		Protocol:      parser.MustOption(opt.At("protocol")).Value.String(),
		FastHTTP:      fastHTTP,
		MethodOptions: map[string]model.MethodHTTPTransportOption{},
		MapCodeErrors: map[string]*model.ErrorDecodeInfoHTTPTransportOption{},
		Openapi: model.OpenapiHTTPTransportOption{
			Methods: map[string]*model.OpenapiMethodOption{},
		},
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
		if openapiErrors, ok := openapiDocOpt.Slice("OpenapiErrors"); ok {
			for _, openapiErrorsOpt := range openapiErrors {
				var methods []string
				if methodsOpt, ok := openapiErrorsOpt.At("methods"); ok {
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
				if errorsOpt, ok := openapiErrorsOpt.At("errors"); ok {
					var errorsName []string
					for _, expr := range errorsOpt.Value.ExprSlice() {
						ptr, ok := g.info.Pkg.TypesInfo.TypeOf(expr).(*stdtypes.Pointer)
						if !ok {
							return option, errors.NotePosition(
								openapiErrorsOpt.Position, fmt.Errorf("the %s value must be nil pointer errors", openapiErrorsOpt.Name),
							)
						}
						named, ok := ptr.Elem().(*stdtypes.Named)
						if !ok {
							return option, errors.NotePosition(
								openapiErrorsOpt.Position, fmt.Errorf("the %s value must be nil pointer errors", openapiErrorsOpt.Name),
							)
						}
						errorsName = append(errorsName, named.Obj().Name())
					}
					if len(methods) > 0 {
						for _, method := range methods {
							option.Openapi.Methods[method].Errors = append(option.Openapi.Methods[method].Errors, errorsName...)
						}
					} else {
						option.Openapi.DefaultMethod.Errors = append(option.Openapi.DefaultMethod.Errors, errorsName...)
					}
				}
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
				return option, errors.NotePosition(signOpt.Position, fmt.Errorf("the Signature value must be func selector"))
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

	errorStatusMethod := "StatusCode"
	if option.JsonRPC.Enable {
		errorStatusMethod = "ErrorCode"
	}

	types.Inspect(g.info.Pkgs, func(p *packages.Package, n ast.Node) bool {
		if ret, ok := n.(*ast.ReturnStmt); ok {
			for _, expr := range ret.Results {
				if typeInfo, ok := p.TypesInfo.Types[expr]; ok {
					retType := typeInfo.Type
					isPointer := false

					ptr, ok := retType.(*stdtypes.Pointer)
					if ok {
						isPointer = true
						retType = ptr.Elem()
					}
					if named, ok := retType.(*stdtypes.Named); ok && named.Obj().Exported() {
						found := 0
						for i := 0; i < named.NumMethods(); i++ {
							m := named.Method(i)
							if m.Name() == errorStatusMethod || m.Name() == "Error" {
								found++
							}
						}
						if found == 2 {
							option.MapCodeErrors[named.Obj().Name()] = &model.ErrorDecodeInfoHTTPTransportOption{IsPointer: isPointer, Named: named}
						}
					}
				}
			}
		}
		return true
	})

	types.Inspect(g.info.Pkgs, func(p *packages.Package, n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			if fn.Name.Name == errorStatusMethod {
				if fn.Recv != nil && len(fn.Recv.List) > 0 {
					recvType := p.TypesInfo.TypeOf(fn.Recv.List[0].Type)
					ptr, ok := recvType.(*stdtypes.Pointer)
					if ok {
						recvType = ptr.Elem()
					}
					if named, ok := recvType.(*stdtypes.Named); ok {
						if _, ok := option.MapCodeErrors[named.Obj().Name()]; ok {
							ast.Inspect(n, func(n ast.Node) bool {
								if ret, ok := n.(*ast.ReturnStmt); ok && len(ret.Results) == 1 {
									if v, ok := p.TypesInfo.Types[ret.Results[0]]; ok {
										if v.Value != nil && v.Value.Kind() == constant.Int {
											code, _ := constant.Int64Val(v.Value)
											option.MapCodeErrors[named.Obj().Name()].Code = code
										}
									}
								}
								return true
							})
						}
					}
				}
			}
		}
		return true
	})
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
