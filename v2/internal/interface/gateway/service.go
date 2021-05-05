package gateway

import (
	"fmt"
	"go/ast"
	"go/constant"
	stdtypes "go/types"
	"path/filepath"
	"sort"
	stdstrings "strings"

	openapi2 "github.com/swipe-io/swipe/v2/internal/plugin/gokit/openapi"

	"github.com/swipe-io/strcase"
	"github.com/swipe-io/swipe/v2/internal/domain/model"
	"github.com/swipe-io/swipe/v2/internal/errors"
	"github.com/swipe-io/swipe/v2/internal/graph"
	"github.com/swipe-io/swipe/v2/internal/types"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/types/typeutil"
)

type ServiceGateway struct {
	pkg                      *packages.Package
	pkgPath                  string
	transportType            model.Transport
	useFast                  bool
	graphTypes               *graph.Graph
	commentFuncs             map[string][]string
	commentFields            map[string]map[string]string
	enums                    *typeutil.Map
	wd                       string
	methodOptions            map[string]model.MethodOption
	defaultMethodOptions     model.MethodOption
	clientsEnable            []string
	errors                   map[uint32]*model.HTTPError
	openapiEnable            bool
	openapiOutput            string
	openapiInfo              openapi2.Info
	openapiServers           []openapi2.Server
	openapiMethodTags        map[string][]string
	openapiDefaultMethodTags []string
	jsonRPCEnable            bool
	jsonRPCDocEnable         bool
	jsonRPCDocOutputDir      string
	jsonRPCPath              string
	readmeEnable             bool
	readmeOutput             string
	readmeTemplatePath       string
	interfaces               model.Interfaces
	hasher                   typeutil.Hasher
	appName                  string
	appID                    string
	defaultErrorEncoder      _option.Value
	externalOptions          []*_option.ResultOption
}

func (g *ServiceGateway) Enums() *typeutil.Map {
	return g.enums
}

func (g *ServiceGateway) CommentFields() map[string]map[string]string {
	return g.commentFields
}

func (g *ServiceGateway) AppID() string {
	return g.appID
}

func (g *ServiceGateway) AppName() string {
	return g.appName
}

func (g *ServiceGateway) Interfaces() model.Interfaces {
	return g.interfaces
}

func (g *ServiceGateway) UseFast() bool {
	return g.useFast
}

func (g *ServiceGateway) DefaultErrorEncoder() _option.Value {
	return g.defaultErrorEncoder
}

func (g *ServiceGateway) MethodOption(m model.ServiceMethod) model.MethodOption {
	if sign, ok := m.T.(*stdtypes.Signature); ok && sign.Recv() != nil {
		ifaceName := stdtypes.TypeString(sign.Recv().Type(), func(p *stdtypes.Package) string {
			return ""
		})
		mopt, ok := g.methodOptions[ifaceName+m.Name]
		if ok {
			return mopt
		}
	}
	return g.defaultMethodOptions
}

func (g *ServiceGateway) ClientEnable() bool {
	return len(g.clientsEnable) > 0
}

func (g *ServiceGateway) GoClientEnable() bool {
	for _, client := range g.clientsEnable {
		if client == "go" {
			return true
		}
	}
	return false
}

func (g *ServiceGateway) JSClientEnable() bool {
	for _, client := range g.clientsEnable {
		if client == "js" {
			return true
		}
	}
	return false
}

func (g *ServiceGateway) OpenapiEnable() bool {
	return g.openapiEnable
}

func (g *ServiceGateway) OpenapiOutput() string {
	return g.openapiOutput
}

func (g *ServiceGateway) OpenapiInfo() openapi2.Info {
	return g.openapiInfo
}

func (g *ServiceGateway) OpenapiServers() []openapi2.Server {
	return g.openapiServers
}

func (g *ServiceGateway) OpenapiMethodTags(name string) []string {
	return g.openapiMethodTags[name]
}

func (g *ServiceGateway) OpenapiDefaultMethodTags() []string {
	return g.openapiDefaultMethodTags
}

func (g *ServiceGateway) TransportType() model.Transport {
	return g.transportType
}

func (g *ServiceGateway) JSONRPCEnable() bool {
	return g.jsonRPCEnable
}

func (g *ServiceGateway) JSONRPCDocEnable() bool {
	return g.jsonRPCDocEnable
}

func (g *ServiceGateway) JSONRPCDocOutput() string {
	return g.jsonRPCDocOutputDir
}

func (g *ServiceGateway) JSONRPCPath() string {
	return g.jsonRPCPath
}

func (g *ServiceGateway) ReadmeOutput() string {
	return g.readmeOutput
}

func (g *ServiceGateway) ReadmeTemplatePath() string {
	return g.readmeTemplatePath
}

func (g *ServiceGateway) Errors() map[uint32]*model.HTTPError {
	return g.errors
}

func (g *ServiceGateway) InstrumentingEnable() bool {
	if g.defaultMethodOptions.InstrumentingEnable {
		return true
	}
	for _, transportOption := range g.methodOptions {
		if transportOption.InstrumentingEnable {
			return true
		}
	}
	return false
}

func (g *ServiceGateway) LoggingEnable() bool {
	if g.defaultMethodOptions.LoggingEnable {
		return true
	}
	for _, transportOption := range g.methodOptions {
		if transportOption.LoggingEnable {
			return true
		}
	}
	return false
}

func (g *ServiceGateway) ReadmeEnable() bool {
	return g.readmeEnable
}

func (g *ServiceGateway) loadReadme(o *_option.Option) error {
	if _, ok := o.At("ReadmeEnable"); ok {
		g.readmeEnable = true
	}
	if opt, ok := o.At("ReadmeOutput"); ok {
		g.readmeOutput = opt.Value.String()
	}
	if opt, ok := o.At("ReadmeTemplatePath"); ok {
		g.readmeTemplatePath = opt.Value.String()
	}
	return nil
}

func (g *ServiceGateway) loadService(o *_option.Option, genericErrors map[uint32]*model.HTTPError, ifaceLen int) (*model.ServiceInterface, error) {
	ifaceOpt := _option.MustOption(o.At("iface"))
	nsOpt := _option.MustOption(o.At("ns"))

	ifacePtr, ok := ifaceOpt.Value.Type().(*stdtypes.Pointer)
	if !ok {
		return nil, errors.NotePosition(o.Position,
			fmt.Errorf("the iface option is required must be a pointer to an interface type; found %s", stdtypes.TypeString(o.Value.Type(), nil)))
	}

	ifaceType, ok := ifacePtr.Elem().Underlying().(*stdtypes.Interface)
	if !ok {
		return nil, errors.NotePosition(o.Position,
			fmt.Errorf("the iface option is required must be a pointer to an interface type; found %s", stdtypes.TypeString(o.Value.Type(), nil)))
	}

	var graphTypes = g.graphTypes
	var externalPkgFound bool

	ifaceNamed := ifacePtr.Elem().(*stdtypes.Named)

	appPkg := g.pkg

	for _, extOpt := range g.externalOptions {
		if ifaces, ok := extOpt.Option.Slice("Interface"); ok {
			for _, o := range ifaces {
				ifaceExtOpt := _option.MustOption(o.At("iface"))
				if ifaceExtPtr, ok := ifaceExtOpt.Value.Type().(*stdtypes.Pointer); ok {
					ifaceExtType := ifaceExtPtr.Elem().Underlying().(*stdtypes.Interface)
					if ifaceExtType.NumEmbeddeds() > 0 {
						for i := 0; i < ifaceExtType.NumEmbeddeds(); i++ {
							if ifaceExtType.EmbeddedType(i).String() == ifacePtr.Elem().String() {
								appPkg = extOpt.Pkg
								externalPkgFound = true
							}
						}
					}
					if ifaceExtPtr.Elem().String() == ifacePtr.Elem().String() {
						appPkg = extOpt.Pkg
						externalPkgFound = true
						break
					}
				}
			}
		}
	}

	basePkgService := stdstrings.Join(stdstrings.Split(ifaceNamed.Obj().Pkg().Path(), "/")[:3], "/")
	basePkgInternal := stdstrings.Join(stdstrings.Split(g.pkg.PkgPath, "/")[:3], "/")
	external := basePkgService != basePkgInternal

	var appName string
	id := stdstrings.Split(appPkg.PkgPath, "/")[:3][2]
	appName = strcase.ToCamel(id)

	if external && !externalPkgFound {
		return nil, errors.NotePosition(o.Position,
			fmt.Errorf("you need to add an external service package for %s", stdtypes.TypeString(o.Value.Type(), nil)))
	}

	ifaceUcName := ifaceNamed.Obj().Name()
	ifaceLcName := strcase.ToLowerCamel(ifaceUcName)

	ns := nsOpt.Value.String()

	var serviceMethods []model.ServiceMethod

	for i := 0; i < ifaceType.NumMethods(); i++ {
		m := ifaceType.Method(i)

		methodErrors := map[uint32]*model.HTTPError{}
		for key, httpError := range genericErrors {
			methodErrors[key] = httpError
		}

		sig := m.Type().(*stdtypes.Signature)

		comments, _ := g.commentFuncs[m.String()]

		ifaceNameUc := ifaceUcName + m.Name()
		ifaceNameLc := ifaceLcName + m.Name()

		nameRequest := m.Name() + "Request"
		nameResponse := m.Name() + "Response"

		if ifaceLen > 1 {
			nameRequest = ifaceNameUc + "Request"
			nameResponse = ifaceNameUc + "Response"
		}

		sm := model.ServiceMethod{
			Type:         m,
			T:            m.Type(),
			Name:         m.Name(),
			LcName:       strcase.ToLowerCamel(m.Name()),
			IfaceUcName:  ifaceNameUc,
			IfaceLcName:  ifaceNameLc,
			NameRequest:  nameRequest,
			NameResponse: nameResponse,
			Comments:     comments,
			Variadic:     sig.Variadic(),
		}

		if g.MethodOption(sm).Exclude {
			continue
		}

		graphTypes.Iterate(func(n *graph.Node) {
			graphTypes.Traverse(n, func(n *graph.Node) bool {
				if n.Object.Name() == m.Name() && stdtypes.Identical(n.Object.Type(), m.Type()) {
					graphTypes.Traverse(n, func(n *graph.Node) bool {
						if named, ok := n.Object.Type().(*stdtypes.Named); ok {
							key := g.hasher.Hash(named)
							if _, ok := methodErrors[key]; ok {
								return true
							}
							if e, ok := g.errors[key]; ok {
								methodErrors[key] = e
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
							key := g.hasher.Hash(named)
							if _, ok := methodErrors[key]; ok {
								return true
							}
							if e, ok := g.errors[key]; ok {
								methodErrors[key] = e
							}
						}
					}
				}
				return true
			})
		})

		var (
			resultOffset, variadicOffset, paramOffset int
		)
		if types.ContainsContext(sig.Params()) {
			sm.ParamCtx = sig.Params().At(0)
			paramOffset = 1
		}

		if sm.Variadic {
			sm.ParamVariadic = sig.Params().At(sig.Params().Len() - 1)
			variadicOffset = 1
		}

		if types.ContainsError(sig.Results()) {
			sm.ReturnErr = sig.Results().At(sig.Results().Len() - 1)
			resultOffset = 1
		}

		if types.IsNamed(sig.Results()) && sig.Results().Len()-resultOffset > 1 {
			sm.ResultsNamed = true
		}

		if !sm.ResultsNamed && sig.Results().Len()-resultOffset > 1 {
			return nil, errors.NotePosition(o.Position,
				fmt.Errorf("interface method with unnamed results cannot be greater than 1"))
		}
		for j := paramOffset; j < sig.Params().Len()-variadicOffset; j++ {
			sm.Params = append(sm.Params, sig.Params().At(j))
		}
		for j := 0; j < sig.Results().Len()-resultOffset; j++ {
			sm.Results = append(sm.Results, sig.Results().At(j))
		}

		for _, httpError := range methodErrors {
			sm.Errors = append(sm.Errors, httpError)
		}

		sort.Sort(sm.Errors)

		serviceMethods = append(serviceMethods, sm)
	}
	return model.NewServiceInterface(
		ifaceUcName,
		ifaceLcName,
		ifacePtr.Elem(),
		ifaceNamed,
		ifaceType,
		serviceMethods,
		external,
		appPkg,
		appName,
		ns,
	), nil
}

func (g *ServiceGateway) load(o *_option.Option) error {

	parts := stdstrings.Split(g.pkgPath, string(filepath.Separator))

	g.appName = parts[len(parts)-1]
	if nameOpt, ok := o.At("AppName"); ok {
		if name := nameOpt.Value.String(); name != "" {
			g.appName = strcase.ToCamel(name)
		}
	}
	g.appID = strcase.ToCamel(g.appName)
	if err := g.loadJSONRPC(o); err != nil {
		return err
	}
	if err := g.loadOpenapi(o); err != nil {
		return err
	}
	if err := g.loadMethodOptions(o); err != nil {
		return err
	}
	if err := g.loadReadme(o); err != nil {
		return err
	}

	errorMethodName := "StatusCode"
	if g.jsonRPCEnable {
		errorMethodName = "ErrorCode"
	}

	g.graphTypes.Iterate(func(n *graph.Node) {
		g.graphTypes.Traverse(n, func(n *graph.Node) bool {
			if named, ok := n.Object.Type().(*stdtypes.Named); ok {
				key := g.hasher.Hash(named)
				if _, ok := g.errors[key]; ok {
					return true
				}
				if e := g.findError(named, errorMethodName); e != nil {
					e.ID = key
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
								key := g.hasher.Hash(named)
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

	if ifaces, ok := o.Slice("Interface"); ok {
		for _, iface := range ifaces {
			svc, err := g.loadService(iface, genericErrors, len(ifaces))
			if err != nil {
				return err
			}
			if len(svc.Methods()) > 0 {
				g.interfaces = append(g.interfaces, svc)
			}
		}
	}

	if o, ok := o.At("DefaultErrorEncoder"); ok {
		g.defaultErrorEncoder = o.Value
	}
	if _, ok := o.At("HTTPServer"); ok {
		g.transportType = model.HTTPTransport
	}
	if _, ok := o.At("HTTPFast"); ok {
		g.useFast = true
	}
	if o, ok := o.At("ClientsEnable"); ok {
		g.clientsEnable = o.Value.StringSlice()
	}

	return nil
}

func (g *ServiceGateway) findError(named *stdtypes.Named, methodName string) *model.HTTPError {
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

func (g *ServiceGateway) loadOpenapi(o *_option.Option) (err error) {
	if _, ok := o.At("OpenapiEnable"); ok {
		g.openapiEnable = true
	}
	if v, ok := o.At("Output"); ok {
		g.openapiOutput = v.Value.String()
	}
	if v, ok := o.At("Info"); ok {
		g.openapiInfo = openapi2.Info{
			Title:       _option.MustOption(v.At("title")).Value.String(),
			Description: _option.MustOption(v.At("description")).Value.String(),
			Version:     _option.MustOption(v.At("version")).Value.String(),
		}
	}
	if v, ok := o.At("OpenapiContact"); ok {
		g.openapiInfo.Contact = &openapi2.Contact{
			Name:  _option.MustOption(v.At("name")).Value.String(),
			Email: _option.MustOption(v.At("email")).Value.String(),
			URL:   _option.MustOption(v.At("url")).Value.String(),
		}
	}
	if v, ok := o.At("OpenapiLicence"); ok {
		g.openapiInfo.License = &openapi2.License{
			Name: _option.MustOption(v.At("name")).Value.String(),
			URL:  _option.MustOption(v.At("url")).Value.String(),
		}
	}
	if s, ok := o.Slice("OpenapiServer"); ok {
		for _, v := range s {
			g.openapiServers = append(g.openapiServers, openapi2.Server{
				Description: _option.MustOption(v.At("description")).Value.String(),
				URL:         _option.MustOption(v.At("url")).Value.String(),
			})
		}
	}
	if openapiTags, ok := o.Slice("Tags"); ok {
		for _, openapiTagsOpt := range openapiTags {
			var methods []string
			if methodsOpt, ok := openapiTagsOpt.At("methods"); ok {
				for _, expr := range methodsOpt.Value.ExprSlice() {
					fnSel, ok := expr.(*ast.SelectorExpr)
					if !ok {
						return errors.NotePosition(methodsOpt.Position, fmt.Errorf("the %s value must be func selector", methodsOpt.Name))
					}
					methods = append(methods, fnSel.Sel.Name)
					if _, ok := g.openapiMethodTags[fnSel.Sel.Name]; !ok {
						g.openapiMethodTags[fnSel.Sel.Name] = []string{}
					}
				}
			}
			if tagsOpt, ok := openapiTagsOpt.At("tags"); ok {
				if len(methods) > 0 {
					for _, method := range methods {
						g.openapiMethodTags[method] = append(g.openapiMethodTags[method], tagsOpt.Value.StringSlice()...)
					}
				} else {
					g.openapiDefaultMethodTags = append(g.openapiDefaultMethodTags, tagsOpt.Value.StringSlice()...)
				}
			}
		}
	}
	if g.openapiOutput == "" {
		g.openapiOutput = "./"
	}
	return nil
}

func (g *ServiceGateway) loadMethodOptions(o *_option.Option) (err error) {
	if methodDefaultOpt, ok := o.At("MethodDefaultOptions"); ok {
		g.defaultMethodOptions, err = getMethodOptions(methodDefaultOpt, model.MethodOption{})
		if err != nil {
			return err
		}
	}
	if methods, ok := o.Slice("MethodOptions"); ok {
		for _, methodOpt := range methods {
			signOpt := _option.MustOption(methodOpt.At("signature"))
			fnSel, ok := signOpt.Value.Expr().(*ast.SelectorExpr)
			if !ok {
				return errors.NotePosition(signOpt.Position, fmt.Errorf("the signature must be selector"))
			}
			mopt, err := getMethodOptions(methodOpt, g.defaultMethodOptions)
			if err != nil {
				return err
			}
			obj := g.pkg.TypesInfo.ObjectOf(fnSel.Sel)
			if obj != nil {
				if sign, ok := obj.Type().(*stdtypes.Signature); ok && sign.Recv() != nil {
					ifaceName := stdtypes.TypeString(sign.Recv().Type(), func(p *stdtypes.Package) string {
						return ""
					})
					g.methodOptions[ifaceName+obj.Name()] = mopt
				}
			}
		}
	}
	return
}

func (g *ServiceGateway) loadJSONRPC(o *_option.Option) (err error) {
	if _, ok := o.At("JSONRPCEnable"); ok {
		g.jsonRPCEnable = true
	}
	if _, ok := o.At("JSONRPCDocEnable"); ok {
		g.jsonRPCDocEnable = true
	}
	if opt, ok := o.At("JSONRPCDocOutput"); ok {
		g.jsonRPCDocOutputDir = opt.Value.String()
	}
	if opt, ok := o.At("JSONRPCPath"); ok {
		g.jsonRPCPath = opt.Value.String()
	}
	return
}

func getMethodOptions(o *_option.Option, baseMethodOpts model.MethodOption) (model.MethodOption, error) {
	if opt, ok := o.At("Exclude"); ok {
		baseMethodOpts.Exclude = opt.Value.Bool()
	}
	if opt, ok := o.At("Logging"); ok {
		baseMethodOpts.LoggingEnable = opt.Value.Bool()
	}
	if opt, ok := o.At("LoggingParams"); ok {
		baseMethodOpts.LoggingIncludeParams = map[string]struct{}{}
		baseMethodOpts.LoggingExcludeParams = map[string]struct{}{}

		includes := _option.MustOption(opt.At("includes")).Value.StringSlice()
		excludes := _option.MustOption(opt.At("excludes")).Value.StringSlice()
		for _, field := range includes {
			baseMethodOpts.LoggingIncludeParams[field] = struct{}{}
		}
		for _, field := range excludes {
			baseMethodOpts.LoggingExcludeParams[field] = struct{}{}
		}
	}

	baseMethodOpts.LoggingContext = map[string]ast.Expr{}

	if opts, ok := o.Slice("LoggingContext"); ok {
		for _, opt := range opts {
			key := _option.MustOption(opt.At("key")).Value.Expr()
			name := _option.MustOption(opt.At("name")).Value.String()
			baseMethodOpts.LoggingContext[name] = key
		}
	}
	if opt, ok := o.At("Instrumenting"); ok {
		baseMethodOpts.InstrumentingEnable = opt.Value.Bool()
	}
	if opt, ok := o.At("RESTWrapResponse"); ok {
		baseMethodOpts.WrapResponse.Enable = true
		baseMethodOpts.WrapResponse.Name = opt.Value.String()
	}
	if opt, ok := o.At("RESTMethod"); ok {
		baseMethodOpts.MethodName = opt.Value.String()
		baseMethodOpts.Expr = opt.Value.Expr()
	}
	if path, ok := o.At("RESTPath"); ok {
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
	if opt, ok := o.At("ServerDecodeRequestFunc"); ok {
		baseMethodOpts.ServerRequestFunc.Type = opt.Value.Type()
		baseMethodOpts.ServerRequestFunc.Expr = opt.Value.Expr()
	}
	if opt, ok := o.At("ServerEncodeResponseFunc"); ok {
		baseMethodOpts.ServerResponseFunc.Type = opt.Value.Type()
		baseMethodOpts.ServerResponseFunc.Expr = opt.Value.Expr()
	}
	if opt, ok := o.At("ClientEncodeRequestFunc"); ok {
		baseMethodOpts.ClientRequestFunc.Type = opt.Value.Type()
		baseMethodOpts.ClientRequestFunc.Expr = opt.Value.Expr()
	}
	if opt, ok := o.At("ClientDecodeResponseFunc"); ok {
		baseMethodOpts.ClientResponseFunc.Type = opt.Value.Type()
		baseMethodOpts.ClientResponseFunc.Expr = opt.Value.Expr()
	}
	if opt, ok := o.At("RESTQueryVars"); ok {
		baseMethodOpts.QueryVars = map[string]string{}
		values := opt.Value.StringSlice()
		for i := 0; i < len(values); i += 2 {
			baseMethodOpts.QueryVars[values[i]] = values[i+1]
		}
	}
	if opt, ok := o.At("RESTHeaderVars"); ok {
		baseMethodOpts.HeaderVars = map[string]string{}
		values := opt.Value.StringSlice()
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

func NewServiceGateway(pkg *packages.Package, pkgPath string, o *_option.Option, graphTypes *graph.Graph, commentFuncs map[string][]string, commentFields map[string]map[string]string, enums *typeutil.Map, wd string, externalOptions []*_option.ResultOption) (*ServiceGateway, error) {
	g := &ServiceGateway{
		pkg:               pkg,
		pkgPath:           pkgPath,
		graphTypes:        graphTypes,
		commentFuncs:      commentFuncs,
		commentFields:     commentFields,
		enums:             enums,
		wd:                wd,
		methodOptions:     map[string]model.MethodOption{},
		openapiMethodTags: map[string][]string{},
		errors:            map[uint32]*model.HTTPError{},
		hasher:            typeutil.MakeHasher(),
		externalOptions:   externalOptions,
	}
	if err := g.load(o); err != nil {
		return nil, err
	}
	return g, nil
}
