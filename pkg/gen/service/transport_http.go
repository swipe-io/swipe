package service

import (
	"fmt"
	"go/ast"
	"go/constant"
	stdtypes "go/types"
	"io/ioutil"
	"path/filepath"
	"strconv"
	stdstrings "strings"

	"github.com/iancoleman/strcase"
	"github.com/pquerna/ffjson/ffjson"
	"golang.org/x/tools/go/packages"

	"github.com/swipe-io/swipe/pkg/errors"
	"github.com/swipe-io/swipe/pkg/openapi"
	"github.com/swipe-io/swipe/pkg/parser"
	"github.com/swipe-io/swipe/pkg/strings"
	"github.com/swipe-io/swipe/pkg/types"
	"github.com/swipe-io/swipe/pkg/utils"
	"github.com/swipe-io/swipe/pkg/writer"
)

type transportJsonRPCOption struct {
	enable bool
	path   string
}

type astOptions struct {
	t    stdtypes.Type
	expr ast.Expr
}

type transportMethod struct {
	name string
	expr ast.Expr
}

type transportMethodOptions struct {
	method             transportMethod
	path               string
	pathVars           map[string]string
	headerVars         map[string]string
	queryVars          map[string]string
	serverRequestFunc  astOptions
	serverResponseFunc astOptions
	clientRequestFunc  astOptions
	clientResponseFunc astOptions
}

type transportOpenapiLicence struct {
	name string
	url  string
}

type transportOpenapiContact struct {
	name  string
	url   string
	email string
}

type transportOpenapiServer struct {
	name string
	url  string
	desc string
}

type transportOpenapiDoc struct {
	enable      bool
	version     string
	description string
	title       string
	output      string
	servers     []openapi.Server
	contact     openapi.Contact
	licence     openapi.License
}

type transportClient struct {
	enable bool
}

type transportOptions struct {
	prefix         string
	notWrapBody    bool
	serverDisabled bool
	client         transportClient
	openapiDoc     transportOpenapiDoc
	fastHTTP       bool
	jsonRPC        transportJsonRPCOption
	methodOptions  map[string]transportMethodOptions
}

type TransportHTTP struct {
	ctx serviceCtx
	w   *writer.Writer
}

func (w *TransportHTTP) Write(opt *parser.Option) error {

	_, enabledFastHTTP := opt.Get("FastEnable")

	options := &transportOptions{
		fastHTTP:      enabledFastHTTP,
		methodOptions: map[string]transportMethodOptions{},
	}

	if _, ok := opt.Get("ClientEnable"); ok {
		options.client.enable = true
	}

	if _, ok := opt.Get("ServerDisabled"); ok {
		options.serverDisabled = true
	}

	if _, ok := opt.Get("NotWrapBody"); ok {
		options.notWrapBody = true
	}

	if openapiDocOpt, ok := opt.Get("Openapi"); ok {
		options.openapiDoc.enable = true
		if v, ok := openapiDocOpt.Get("OpenapiOutput"); ok {
			options.openapiDoc.output = v.Value.String()
		}
		if v, ok := openapiDocOpt.Get("OpenapiVersion"); ok {
			options.openapiDoc.version = v.Value.String()
		}
		if v, ok := openapiDocOpt.Get("OpenapiTitle"); ok {
			options.openapiDoc.title = v.Value.String()
		}
		if v, ok := openapiDocOpt.Get("OpenapiDescription"); ok {
			options.openapiDoc.description = v.Value.String()
		}
		if v, ok := openapiDocOpt.Get("OpenapiContact"); ok {
			options.openapiDoc.contact.Name = parser.MustOption(v.Get("name")).Value.String()
			options.openapiDoc.contact.Email = parser.MustOption(v.Get("email")).Value.String()
			options.openapiDoc.contact.URL = parser.MustOption(v.Get("url")).Value.String()
		}
		if v, ok := openapiDocOpt.Get("OpenapiLicence"); ok {
			options.openapiDoc.licence.Name = parser.MustOption(v.Get("name")).Value.String()
			options.openapiDoc.licence.URL = parser.MustOption(v.Get("url")).Value.String()
		}

		if s, ok := openapiDocOpt.GetSlice("OpenapiServer"); ok {
			for _, v := range s {
				options.openapiDoc.servers = append(options.openapiDoc.servers, openapi.Server{
					Description: parser.MustOption(v.Get("description")).Value.String(),
					URL:         parser.MustOption(v.Get("url")).Value.String(),
				})
			}
		}

		if options.openapiDoc.output == "" {
			options.openapiDoc.output = "./"
		}
	}
	if jsonRpcOpt, ok := opt.Get("JSONRPC"); ok {
		options.jsonRPC.enable = true
		if path, ok := jsonRpcOpt.Get("JSONRPCPath"); ok {
			options.jsonRPC.path = path.Value.String()
		}
	}

	if methods, ok := opt.GetSlice("MethodOptions"); ok {
		for _, methodOpt := range methods {
			signOpt := parser.MustOption(methodOpt.Get("signature"))
			fnSel, ok := signOpt.Value.Expr().(*ast.SelectorExpr)
			if !ok {
				return errors.NotePosition(signOpt.Position, fmt.Errorf("the Signature value must be func selector"))
			}

			transportMethodOptions := transportMethodOptions{}

			if httpMethodOpt, ok := methodOpt.Get("Method"); ok {
				transportMethodOptions.method.name = httpMethodOpt.Value.String()
				transportMethodOptions.method.expr = httpMethodOpt.Value.Expr()
			}

			if path, ok := methodOpt.Get("Path"); ok {
				transportMethodOptions.path = path.Value.String()

				idxs, err := httpBraceIndices(transportMethodOptions.path)
				if err != nil {
					return err
				}
				if len(idxs) > 0 {
					transportMethodOptions.pathVars = make(map[string]string, len(idxs))

					var end int
					for i := 0; i < len(idxs); i += 2 {
						end = idxs[i+1]
						parts := stdstrings.SplitN(transportMethodOptions.path[idxs[i]+1:end-1], ":", 2)

						name := parts[0]
						regexp := ""

						if len(parts) == 2 {
							regexp = parts[1]
						}
						transportMethodOptions.pathVars[name] = regexp
					}
				}
			}

			if serverRequestFunc, ok := methodOpt.Get("ServerRequestFunc"); ok {
				transportMethodOptions.serverRequestFunc.t = serverRequestFunc.Value.Type()
				transportMethodOptions.serverRequestFunc.expr = serverRequestFunc.Value.Expr()
			}

			if serverResponseFunc, ok := methodOpt.Get("ServerResponseFunc"); ok {
				transportMethodOptions.serverResponseFunc.t = serverResponseFunc.Value.Type()
				transportMethodOptions.serverResponseFunc.expr = serverResponseFunc.Value.Expr()
			}

			if clientRequestFunc, ok := methodOpt.Get("ClientRequestFunc"); ok {
				transportMethodOptions.clientRequestFunc.t = clientRequestFunc.Value.Type()
				transportMethodOptions.clientRequestFunc.expr = clientRequestFunc.Value.Expr()
			}

			if clientResponseFunc, ok := methodOpt.Get("ClientResponseFunc"); ok {
				transportMethodOptions.clientResponseFunc.t = clientResponseFunc.Value.Type()
				transportMethodOptions.clientResponseFunc.expr = clientResponseFunc.Value.Expr()
			}

			if queryVars, ok := methodOpt.Get("QueryVars"); ok {
				transportMethodOptions.queryVars = map[string]string{}

				values := queryVars.Value.StringSlice()
				for i := 0; i < len(values); i += 2 {
					transportMethodOptions.queryVars[values[0]] = values[1]
				}
			}
			if headerVars, ok := methodOpt.Get("HeaderVars"); ok {
				transportMethodOptions.headerVars = map[string]string{}
				values := headerVars.Value.StringSlice()
				for i := 0; i < len(values); i += 2 {
					transportMethodOptions.headerVars[values[0]] = values[1]
				}
			}

			options.methodOptions[fnSel.Sel.Name] = transportMethodOptions
		}
	}
	options.prefix = "REST"
	if options.jsonRPC.enable {
		options.prefix = "JSONRPC"
	}

	if options.openapiDoc.enable {
		if err := w.writeOpenapiDoc(options); err != nil {
			return err
		}
	}

	errorStatusMethod := "StatusCode"
	if options.jsonRPC.enable {
		errorStatusMethod = "ErrorCode"
	}

	mapCodeErrors := map[*stdtypes.Named]string{}

	w.w.Inspect(func(p *packages.Package, n ast.Node) bool {
		if ret, ok := n.(*ast.ReturnStmt); ok {
			for _, expr := range ret.Results {
				if typeInfo, ok := p.TypesInfo.Types[expr]; ok {
					if pointer, ok := typeInfo.Type.(*stdtypes.Pointer); ok {
						if named, ok := pointer.Elem().(*stdtypes.Named); ok {
							for i := 0; i < named.NumMethods(); i++ {
								m := named.Method(i)
								if m.Name() == errorStatusMethod {
									mapCodeErrors[named] = ""
									break
								}
							}
						}
					}
				}
			}
		}
		return true
	})

	w.w.Inspect(func(p *packages.Package, n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			if fn.Name.Name == errorStatusMethod {
				if fn.Recv != nil && len(fn.Recv.List) > 0 {
					recvType := p.TypesInfo.TypeOf(fn.Recv.List[0].Type)
					recvPtr := recvType.(*stdtypes.Pointer)
					recv := recvPtr.Elem().(*stdtypes.Named)

					if _, ok := mapCodeErrors[recv]; ok {
						ast.Inspect(n, func(n ast.Node) bool {
							if ret, ok := n.(*ast.ReturnStmt); ok && len(ret.Results) == 1 {
								if v, ok := p.TypesInfo.Types[ret.Results[0]]; ok {
									if v.Value != nil && v.Value.Kind() == constant.Int {
										mapCodeErrors[recv] = v.Value.String()
									}
								}
							}
							return true
						})
					}
				}
			}
		}
		return true
	})

	fmtPkg := w.w.Import("fmt", "fmt")

	w.w.WriteFunc("ErrorDecode", "", []string{"code", "int"}, []string{"", "error"}, func() {
		w.w.Write("switch code {\n")
		w.w.Write("default:\nreturn %s.Errorf(\"error code %%d\", code)\n", fmtPkg)
		for v, code := range mapCodeErrors {
			w.w.Write("case %s:\n", code)

			pkg := w.w.Import(v.Obj().Pkg().Name(), v.Obj().Pkg().Path())
			w.w.Write("return new(%s.%s)\n", pkg, v.Obj().Name())
		}
		w.w.Write("}\n")
	})

	if options.client.enable {
		w.writeClientStruct(options)

		clientType := "client" + w.ctx.id

		w.w.Write("func NewClient%s%s(tgt string", options.prefix, w.ctx.id)

		w.w.Write(" ,opts ...%[1]sOption", clientType)

		w.w.Write(") (%s, error) {\n", w.ctx.typeStr)

		w.w.Write("c := &%s{}\n", clientType)

		w.w.Write("for _, o := range opts {\n")
		w.w.Write("o(c)\n")
		w.w.Write("}\n")

		if options.jsonRPC.enable {
			w.writeJsonRPCClient(options)
		} else {
			w.writeRestClient(options)
		}

		w.w.Write("return c, nil\n")
		w.w.Write("}\n")
	}
	if !options.serverDisabled {
		if err := w.writeHTTP(options); err != nil {
			return err
		}
	}

	return nil
}

func (w *TransportHTTP) writeHTTP(opts *transportOptions) error {
	var (
		kithttpPkg string
	)
	if opts.jsonRPC.enable {
		if opts.fastHTTP {
			kithttpPkg = w.w.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/fasthttp/jsonrpc")
		} else {
			kithttpPkg = w.w.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/http/jsonrpc")
		}
	} else {
		if opts.fastHTTP {
			kithttpPkg = w.w.Import("fasthttp", "github.com/l-vitaly/go-kit/transport/fasthttp")
		} else {
			kithttpPkg = w.w.Import("http", "github.com/go-kit/kit/transport/http")
		}
	}

	serverOptType := fmt.Sprintf("server%sOpts", w.ctx.id)
	serverOptionType := fmt.Sprintf("server%sOption", w.ctx.id)
	kithttpServerOption := fmt.Sprintf("%s.ServerOption", kithttpPkg)

	w.w.Write("type %s func (*%s)\n", serverOptionType, serverOptType)

	w.w.Write("type %s struct {\n", serverOptType)
	w.w.Write("genericServerOption []%s\n", kithttpServerOption)
	for i := 0; i < w.ctx.iface.NumMethods(); i++ {
		m := w.ctx.iface.Method(i)
		lcName := strings.LcFirst(m.Name())
		w.w.Write("%sServerOption []%s\n", lcName, kithttpServerOption)
	}
	w.w.Write("}\n")

	w.w.WriteFunc(
		w.ctx.id+"GenericServerOptions",
		"",
		[]string{"v", "..." + kithttpServerOption},
		[]string{"", serverOptionType},
		func() {
			w.w.Write("return func(o *%s) { o.genericServerOption = v }\n", serverOptType)
		},
	)

	if !opts.jsonRPC.enable {
		for i := 0; i < w.ctx.iface.NumMethods(); i++ {
			m := w.ctx.iface.Method(i)
			lcName := strings.LcFirst(m.Name())

			w.w.WriteFunc(
				w.ctx.id+m.Name()+"ServerOptions",
				"",
				[]string{"opt", "..." + kithttpServerOption},
				[]string{"", serverOptionType},
				func() {
					w.w.Write("return func(c *%s) { c.%sServerOption = opt }\n", serverOptType, lcName)
				},
			)
		}
	}

	w.w.Write("// HTTP %s Transport\n", opts.prefix)

	if opts.jsonRPC.enable {
		w.writeJsonRPCEncodeResponse()
	} else {
		w.writeHTTPEncodeResponse(opts)
	}

	w.w.Write("func MakeHandler%s%s(s %s", opts.prefix, w.ctx.id, w.ctx.typeStr)
	if w.ctx.logging {
		logPkg := w.w.Import("log", "github.com/go-kit/kit/log")
		w.w.Write(", logger %s.Logger", logPkg)
	}
	w.w.Write(", opts ...server%sOption", w.ctx.id)
	w.w.Write(") (")
	if opts.fastHTTP {
		w.w.Write("%s.RequestHandler", w.w.Import("fasthttp", "github.com/valyala/fasthttp"))
	} else {
		w.w.Write("%s.Handler", w.w.Import("http", "net/http"))
	}

	w.w.Write(", error) {\n")

	w.w.Write("sopt := &server%sOpts{}\n", w.ctx.id)

	w.w.Write("for _, o := range opts {\n o(sopt)\n }\n")

	w.writeMiddlewares(opts)
	w.writeHTTPHandler(opts)

	w.w.Write("}\n\n")

	return nil
}

func (w *TransportHTTP) writeJsonRPCEncodeResponse() {
	ffjsonPkg := w.w.Import("ffjson", "github.com/pquerna/ffjson/ffjson")
	jsonPkg := w.w.Import("json", "encoding/json")
	contextPkg := w.w.Import("context", "context")

	w.w.Write("func encodeResponseJSONRPC%s(_ %s.Context, result interface{}) (%s.RawMessage, error) {\n", w.ctx.id, contextPkg, jsonPkg)
	w.w.Write("b, err := %s.Marshal(result)\n", ffjsonPkg)
	w.w.Write("if err != nil {\n")
	w.w.Write("return nil, err\n")
	w.w.Write("}\n")
	w.w.Write("return b, nil\n")
	w.w.Write("}\n\n")
}

func (w *TransportHTTP) writeHTTPEncodeResponse(opts *transportOptions) {
	kitEndpointPkg := w.w.Import("endpoint", "github.com/go-kit/kit/endpoint")
	jsonPkg := w.w.Import("ffjson", "github.com/pquerna/ffjson/ffjson")
	contextPkg := w.w.Import("context", "context")

	var httpPkg string

	if opts.fastHTTP {
		httpPkg = w.w.Import("fasthttp", "github.com/valyala/fasthttp")
	} else {
		httpPkg = w.w.Import("http", "net/http")
	}

	w.w.Write("type errorWrapper struct {\n")
	w.w.Write("Error string `json:\"error\"`\n")
	w.w.Write("}\n")

	w.w.Write("func encodeResponseHTTP%s(ctx %s.Context, ", w.ctx.id, contextPkg)

	if opts.fastHTTP {
		w.w.Write("w *%s.Response", httpPkg)
	} else {
		w.w.Write("w %s.ResponseWriter", httpPkg)
	}

	w.w.Write(", response interface{}) error {\n")

	if opts.fastHTTP {
		w.w.Write("h := w.Header\n")
	} else {
		w.w.Write("h := w.Header()\n")
	}

	w.w.Write("h.Set(\"Content-Type\", \"application/json; charset=utf-8\")\n")

	w.w.Write("if e, ok := response.(%s.Failer); ok && e.Failed() != nil {\n", kitEndpointPkg)

	w.w.Write("data, err := %s.Marshal(errorWrapper{Error: e.Failed().Error()})\n", jsonPkg)
	w.w.Write("if err != nil {\n")
	w.w.Write("return err\n")
	w.w.Write("}\n")

	if opts.fastHTTP {
		w.w.Write("w.SetBody(data)\n")
	} else {
		w.w.Write("w.Write(data)\n")
	}

	w.w.Write("return nil\n")
	w.w.Write("}\n")

	w.w.Write("data, err := %s.Marshal(response)\n", jsonPkg)
	w.w.Write("if err != nil {\n")
	w.w.Write("return err\n")
	w.w.Write("}\n")

	if opts.fastHTTP {
		w.w.Write("w.SetBody(data)\n")
	} else {
		w.w.Write("w.Write(data)\n")
	}

	w.w.Write("return nil\n")
	w.w.Write("}\n\n")
}

func (w *TransportHTTP) makeRestPath(opts *transportOptions, m *stdtypes.Func) (string, openapi.Path) {
	msig := m.Type().(*stdtypes.Signature)
	mopt := opts.methodOptions[m.Name()]

	pathStr := mopt.path

	for _, regexp := range mopt.pathVars {
		pathStr = stdstrings.Replace(pathStr, ":"+regexp, "", -1)
	}

	o := &openapi.Operation{
		Summary:   m.Name(),
		Responses: openapi.Responses{},
	}

	for name := range mopt.pathVars {
		var schema *openapi.Schema
		if fld := types.LookupFieldSig(name, msig); fld != nil {
			schema = w.makeSwaggerSchema(fld.Type())
		}
		o.Parameters = append(o.Parameters, openapi.Parameter{
			In:       "path",
			Name:     name,
			Required: true,
			Schema:   schema,
		})
	}

	for argName, name := range mopt.queryVars {
		var schema *openapi.Schema
		if fld := types.LookupFieldSig(argName, msig); fld != nil {
			schema = w.makeSwaggerSchema(fld.Type())
		}
		o.Parameters = append(o.Parameters, openapi.Parameter{
			In:     "query",
			Name:   name,
			Schema: schema,
		})
	}

	for argName, name := range mopt.headerVars {
		var schema *openapi.Schema
		if fld := types.LookupFieldSig(argName, msig); fld != nil {
			schema = w.makeSwaggerSchema(fld.Type())
		}
		o.Parameters = append(o.Parameters, openapi.Parameter{
			In:     "header",
			Name:   name,
			Schema: schema,
		})
	}

	swgPath := openapi.Path{}
	switch mopt.method.name {
	default:
		swgPath.Get = o
	case "POST":
		swgPath.Post = o
	case "PUT":
		swgPath.Put = o
	case "PATCH":
		swgPath.Patch = o
	case "DELETE":
		swgPath.Delete = o
	}

	return pathStr, swgPath
}

func (w *TransportHTTP) makeJsonRPCPath(opts *transportOptions, m *stdtypes.Func) (string, openapi.Path) {
	msig := m.Type().(*stdtypes.Signature)
	lcName := strings.LcFirst(m.Name())

	responseParams := &openapi.Schema{
		Properties: map[string]*openapi.Schema{},
	}

	requestParams := &openapi.Schema{
		Properties: map[string]*openapi.Schema{},
	}

	paramsLen := msig.Params().Len()
	if types.HasContextInParams(msig.Params()) {
		paramsLen--
	}

	resultLen := msig.Results().Len()
	if types.HasErrorInResults(msig.Results()) {
		resultLen--
	}

	for i := 1; i < paramsLen; i++ {
		p := msig.Params().At(i)
		requestParams.Properties[strcase.ToLowerCamel(p.Name())] = w.makeSwaggerSchema(p.Type())
	}

	for i := 0; i < resultLen; i++ {
		r := msig.Results().At(i)
		responseParams.Properties[strcase.ToLowerCamel(r.Name())] = w.makeSwaggerSchema(r.Type())
	}

	response := &openapi.Schema{
		Type: "object",
		Properties: openapi.Properties{
			"jsonrpc": &openapi.Schema{
				Type: "string",
			},
			"id": &openapi.Schema{
				Type: "string",
			},
			"result": responseParams,
		},
	}
	request := &openapi.Schema{
		Type: "object",
		Properties: openapi.Properties{
			"jsonrpc": &openapi.Schema{
				Type: "string",
			},
			"id": &openapi.Schema{
				Type: "string",
			},
			"method": &openapi.Schema{
				Type: "string",
				Enum: []string{strcase.ToLowerCamel(m.Name())},
			},
			"params": requestParams,
		},
	}

	return "/" + lcName, openapi.Path{
		Post: &openapi.Operation{
			RequestBody: &openapi.RequestBody{
				Required: true,
				Content: map[string]openapi.Media{
					"application/json": {
						Schema: request,
					},
				},
			},
			Responses: map[string]openapi.Response{
				"200": {
					Description: "OK",
					Content: openapi.Content{
						"application/json": {
							Schema: response,
						},
					},
				},
				"500": {
					Description: "FAIL",
					Content: openapi.Content{
						"application/json": {
							Schema: &openapi.Schema{
								Ref: "#/components/schemas/Error",
							},
						},
					},
				},
			},
		},
	}
}

func (w *TransportHTTP) writeOpenapiDoc(opts *transportOptions) error {
	swg := openapi.OpenAPI{
		OpenAPI: "3.0.0",
		Info: openapi.Info{
			Version:     opts.openapiDoc.version,
			Description: opts.openapiDoc.description,
			Title:       opts.openapiDoc.title,
			Contact:     opts.openapiDoc.contact,
			License:     opts.openapiDoc.licence,
		},
		Servers: opts.openapiDoc.servers,
		Paths:   map[string]openapi.Path{},
		Components: openapi.Components{
			Schemas: openapi.Schemas{},
		},
	}
	for i := 0; i < w.ctx.iface.NumMethods(); i++ {
		m := w.ctx.iface.Method(i)

		var (
			pathStr string
			path    openapi.Path
		)

		if opts.jsonRPC.enable {
			swg.Components.Schemas["Error"] = getOpenapiErrorSchema()
			pathStr, path = w.makeJsonRPCPath(opts, m)
		} else {
			pathStr, path = w.makeRestPath(opts, m)
		}

		swg.Paths[pathStr] = path
	}

	typeName := "rest"
	if opts.jsonRPC.enable {
		typeName = "jsonrpc"
	}
	output, err := filepath.Abs(filepath.Join(w.w.BasePath(), opts.openapiDoc.output))
	if err != nil {
		return err
	}
	d, _ := ffjson.Marshal(swg)
	if err := ioutil.WriteFile(filepath.Join(output, fmt.Sprintf("openapi_%s.json", typeName)), d, 0755); err != nil {
		return err
	}
	return nil
}

func getOpenapiErrorSchema() *openapi.Schema {
	return &openapi.Schema{
		Type: "object",
		Properties: openapi.Properties{
			"jsonrpc": &openapi.Schema{
				Type: "string",
			},
			"id": &openapi.Schema{
				Type: "string",
			},
			"error": &openapi.Schema{
				Type: "object",
				Properties: openapi.Properties{
					"code": &openapi.Schema{
						Type: "integer",
					},
					"message": &openapi.Schema{
						Type: "string",
					},
				},
			},
		},
	}
}

func (w *TransportHTTP) writeHTTPHandler(opts *transportOptions) {
	var (
		routerPkg  string
		jsonrpcPkg string
	)

	if opts.jsonRPC.enable {
		if opts.fastHTTP {
			jsonrpcPkg = w.w.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/fasthttp/jsonrpc")
		} else {
			jsonrpcPkg = w.w.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/http/jsonrpc")
		}
	}

	if opts.fastHTTP {
		routerPkg = w.w.Import("routing", "github.com/qiangxue/fasthttp-routing")
		w.w.Write("r := %s.New()\n", routerPkg)
	} else {
		routerPkg = w.w.Import("mux", "github.com/gorilla/mux")
		w.w.Write("r := %s.NewRouter()\n", routerPkg)
	}

	if opts.jsonRPC.enable {
		w.w.Write("handler := %[1]s.NewServer(%[1]s.EndpointCodecMap{\n", jsonrpcPkg)
	}

	for i := 0; i < w.ctx.iface.NumMethods(); i++ {
		m := w.ctx.iface.Method(i)
		msig := m.Type().(*stdtypes.Signature)

		if opts.jsonRPC.enable {
			w.writeHTTPJSONRPC(opts, m, msig)
		} else {
			w.writeHTTPRest(opts, m, msig)
		}
	}

	if opts.jsonRPC.enable {
		w.w.Write("}, sopt.genericServerOption...)\n")
		jsonRPCPath := opts.jsonRPC.path
		if opts.fastHTTP {
			r := stdstrings.NewReplacer("{", "<", "}", ">")
			jsonRPCPath = r.Replace(jsonRPCPath)

			w.w.Write("r.Post(\"%s\", func(c *routing.Context) error {\nhandler.ServeFastHTTP(c.RequestCtx)\nreturn nil\n})\n", jsonRPCPath)
		} else {
			w.w.Write("r.Methods(\"POST\").Path(\"%s\").Handler(handler)\n", jsonRPCPath)
		}
	}

	if opts.fastHTTP {
		w.w.Write("return r.HandleRequest, nil")
	} else {
		w.w.Write("return r, nil")
	}

}

func (w *TransportHTTP) writeHTTPJSONRPC(opts *transportOptions, m *stdtypes.Func, sig *stdtypes.Signature) {
	var (
		jsonrpcPkg string
	)

	mopt := opts.methodOptions[m.Name()]

	if opts.fastHTTP {
		jsonrpcPkg = w.w.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/fasthttp/jsonrpc")
	} else {
		jsonrpcPkg = w.w.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/http/jsonrpc")
	}
	jsonPkg := w.w.Import("json", "encoding/json")
	ffjsonPkg := w.w.Import("ffjson", "github.com/pquerna/ffjson/ffjson")
	contextPkg := w.w.Import("context", "context")

	lcName := strings.LcFirst(m.Name())
	w.w.Write("\"%s\": %s.EndpointCodec{\n", lcName, jsonrpcPkg)
	w.w.Write("Endpoint: make%sEndpoint(s),\n", m.Name())
	w.w.Write("Decode: ")

	if mopt.serverRequestFunc.expr != nil {
		w.w.WriteAST(mopt.serverRequestFunc.expr)
	} else {
		fmtPkg := w.w.Import("fmt", "fmt")

		w.w.Write("func(_ %s.Context, msg %s.RawMessage) (interface{}, error) {\n", contextPkg, jsonPkg)

		paramsLen := sig.Params().Len()
		if types.HasContextInSignature(sig) {
			paramsLen--
		}

		if paramsLen > 0 {
			w.w.Write("var req %sRequest%s\n", lcName, w.ctx.id)
			w.w.Write("err := %s.Unmarshal(msg, &req)\n", ffjsonPkg)
			w.w.Write("if err != nil {\n")
			w.w.Write("return nil, %s.Errorf(\"couldn't unmarshal body to %sRequest%s: %%s\", err)\n", fmtPkg, lcName, w.ctx.id)
			w.w.Write("}\n")
			w.w.Write("return req, nil\n")

		} else {
			w.w.Write("return nil, nil\n")
		}
		w.w.Write("}")
	}

	w.w.Write(",\n")

	if opts.jsonRPC.enable {
		w.w.Write("Encode: encodeResponseJSONRPC%s,\n", w.ctx.id)
	} else {
		w.w.Write("Encode: encodeResponseHTTP%s,\n", w.ctx.id)
	}

	w.w.Write("},\n")
}

func (w *TransportHTTP) writeHTTPRest(opts *transportOptions, fn *stdtypes.Func, sig *stdtypes.Signature) {
	var (
		kithttpPkg string
		httpPkg    string
		routerPkg  string
	)
	if opts.fastHTTP {
		kithttpPkg = w.w.Import("fasthttp", "github.com/l-vitaly/go-kit/transport/fasthttp")
		httpPkg = w.w.Import("fasthttp", "github.com/valyala/fasthttp")
		routerPkg = w.w.Import("routing", "github.com/qiangxue/fasthttp-routing")
	} else {
		kithttpPkg = w.w.Import("http", "github.com/go-kit/kit/transport/http")
		httpPkg = w.w.Import("http", "net/http")
		routerPkg = w.w.Import("mux", "github.com/gorilla/mux")
	}

	contextPkg := w.w.Import("context", "context")

	mopt := opts.methodOptions[fn.Name()]

	lcName := strings.LcFirst(fn.Name())

	if opts.fastHTTP {
		w.w.Write("r.To(")

		if mopt.method.name != "" {
			w.w.WriteAST(mopt.method.expr)
		} else {
			w.w.Write(strconv.Quote("GET"))
		}

		w.w.Write(", ")

		if mopt.path != "" {
			// replace brace indices for fasthttp router
			urlPath := stdstrings.ReplaceAll(mopt.path, "{", "<")
			urlPath = stdstrings.ReplaceAll(urlPath, "}", ">")
			w.w.Write(strconv.Quote(urlPath))
		} else {
			w.w.Write(strconv.Quote("/" + lcName))
		}
		w.w.Write(", ")
	} else {
		w.w.Write("r.Methods(")
		if mopt.method.name != "" {
			w.w.WriteAST(mopt.method.expr)
		} else {
			w.w.Write(strconv.Quote("GET"))
		}
		w.w.Write(").")
		w.w.Write("Path(")
		if mopt.path != "" {
			w.w.Write(strconv.Quote(mopt.path))
		} else {
			w.w.Write(strconv.Quote("/" + stdstrings.ToLower(fn.Name())))
		}
		w.w.Write(").")

		w.w.Write("Handler(")
	}

	w.w.Write("%s.NewServer(\nmake%sEndpoint(s),\n", kithttpPkg, fn.Name())

	if mopt.serverRequestFunc.expr != nil {
		w.w.WriteAST(mopt.serverRequestFunc.expr)
	} else {
		w.w.Write("func(ctx %s.Context, r *%s.Request) (interface{}, error) {\n", contextPkg, httpPkg)
		paramsLen := sig.Params().Len()
		if types.HasContextInSignature(sig) {
			paramsLen--
		}
		if paramsLen > 0 {
			w.w.Write("var req %sRequest%s\n", lcName, w.ctx.id)
			switch stdstrings.ToUpper(mopt.method.name) {
			case "POST", "PUT", "PATCH":
				fmtPkg := w.w.Import("fmt", "fmt")
				jsonPkg := w.w.Import("ffjson", "github.com/pquerna/ffjson/ffjson")
				pkgIO := w.w.Import("io", "io")

				if opts.fastHTTP {
					w.w.Write("err := %s.Unmarshal(r.Body(), &req)\n", jsonPkg)
				} else {
					ioutilPkg := w.w.Import("ioutil", "io/ioutil")

					w.w.Write("b, err := %s.ReadAll(r.Body)\n", ioutilPkg)
					w.w.WriteCheckErr(func() {
						w.w.Write("return nil, %s.Errorf(\"couldn't read body for %sRequest%s: %%s\", err)\n", fmtPkg, lcName, w.ctx.id)
					})
					w.w.Write("err = %s.Unmarshal(b, &req)\n", jsonPkg)
				}

				w.w.Write("if err != nil && err != %s.EOF {\n", pkgIO)
				w.w.Write("return nil, %s.Errorf(\"couldn't unmarshal body to %sRequest%s: %%s\", err)\n", fmtPkg, lcName, w.ctx.id)
				w.w.Write("return nil, err\n")
				w.w.Write("}\n")
			}
			if len(mopt.pathVars) > 0 {
				if opts.fastHTTP {
					fmtPkg := w.w.Import("fmt", "fmt")

					w.w.Write("vars, ok := ctx.Value(%s.ContextKeyRouter).(*%s.Context)\n", kithttpPkg, routerPkg)
					w.w.Write("if !ok {\n")
					w.w.Write("return nil, %s.Errorf(\"couldn't assert %s.ContextKeyRouter to *%s.Context\")\n", fmtPkg, kithttpPkg, routerPkg)
					w.w.Write("}\n")
				} else {
					w.w.Write("vars := %s.Vars(r)\n", routerPkg)
				}
				for pathVarName := range mopt.pathVars {
					if f := types.LookupFieldSig(pathVarName, sig); f != nil {
						var valueID string
						if opts.fastHTTP {
							valueID = "vars.Param(" + strconv.Quote(pathVarName) + ")"
						} else {
							valueID = "vars[" + strconv.Quote(pathVarName) + "]"
						}
						w.w.WriteConvertType("req."+strings.UcFirst(f.Name()), valueID, f, "", false)
					}
				}
			}
			if len(mopt.queryVars) > 0 {
				if opts.fastHTTP {
					w.w.Write("q := r.URI().QueryArgs()\n")
				} else {
					w.w.Write("q := r.URL.Query()\n")
				}
				for argName, queryName := range mopt.queryVars {
					if f := types.LookupFieldSig(argName, sig); f != nil {
						var valueID string
						if opts.fastHTTP {
							valueID = "string(q.Peek(" + strconv.Quote(queryName) + "))"
						} else {
							valueID = "q.Get(" + strconv.Quote(queryName) + ")"
						}
						w.w.WriteConvertType("req."+strings.UcFirst(f.Name()), valueID, f, "", false)
					}
				}
			}
			for argName, headerName := range mopt.headerVars {
				if f := types.LookupFieldSig(argName, sig); f != nil {
					var valueID string
					if opts.fastHTTP {
						valueID = "string(r.Header.Peek(" + strconv.Quote(headerName) + "))"
					} else {
						valueID = "r.Header.Get(" + strconv.Quote(headerName) + ")"
					}
					w.w.WriteConvertType("req."+strings.UcFirst(f.Name()), valueID, f, "", false)
				}
			}
			w.w.Write("return req, nil\n")
		} else {
			w.w.Write("return nil, nil\n")
		}
		w.w.Write("},\n")
	}
	if mopt.serverResponseFunc.expr != nil {
		w.w.WriteAST(mopt.serverResponseFunc.expr)
	} else {
		if opts.jsonRPC.enable {
			w.w.Write("encodeResponseJSONRPC%s", w.ctx.id)
		} else {
			if opts.notWrapBody {
				fmtPkg := w.w.Import("fmt", "fmt")

				var responseWriterType string
				if opts.fastHTTP {
					responseWriterType = fmt.Sprintf("*%s.Response", httpPkg)
				} else {
					responseWriterType = fmt.Sprintf("%s.ResponseWriter", httpPkg)
				}

				w.w.Write("func (ctx context.Context, w %s, response interface{}) error {\n", responseWriterType)
				w.w.Write("resp, ok := response.(%sResponse%s)\n", lcName, w.ctx.id)

				w.w.Write("if !ok {\n")
				w.w.Write("return %s.Errorf(\"couldn't assert response as %sResponse%s, got %%T\", response)\n", fmtPkg, lcName, w.ctx.id)
				w.w.Write("}\n")

				w.w.Write("return encodeResponseHTTP%s(ctx, w, resp.%s)\n", w.ctx.id, strings.UcFirst(sig.Results().At(0).Name()))
				w.w.Write("}")
			} else {
				w.w.Write("encodeResponseHTTP%s", w.ctx.id)
			}
		}
	}

	w.w.Write(",\n")

	w.w.Write("append(sopt.genericServerOption, sopt.%sServerOption...)...,\n", lcName)
	w.w.Write(")")

	if opts.fastHTTP {
		w.w.Write(".RouterHandle()")
	}

	w.w.Write(")\n")
}

func (w *TransportHTTP) writeMiddlewares(opts *transportOptions) {
	if w.ctx.logging {
		w.writeLoggingMiddleware()
	}
	if w.ctx.instrumenting.enable {
		w.writeInstrumentingMiddleware()
	}
}

func (w *TransportHTTP) writeLoggingMiddleware() {
	w.w.Write("s = &loggingMiddleware%s{next: s, logger: logger}\n", w.ctx.id)
}

func (w *TransportHTTP) writeInstrumentingMiddleware() {
	stdPrometheusPkg := w.w.Import("prometheus", "github.com/prometheus/client_golang/prometheus")
	kitPrometheusPkg := w.w.Import("prometheus", "github.com/go-kit/kit/metrics/prometheus")

	w.w.Write("s = &instrumentingMiddleware%s{\nnext: s,\n", w.ctx.id)
	w.w.Write("requestCount: %s.NewCounterFrom(%s.CounterOpts{\n", kitPrometheusPkg, stdPrometheusPkg)
	w.w.Write("Namespace: %s,\n", strconv.Quote(w.ctx.instrumenting.namespace))
	w.w.Write("Subsystem: %s,\n", strconv.Quote(w.ctx.instrumenting.subsystem))
	w.w.Write("Name: %s,\n", strconv.Quote("request_count"))
	w.w.Write("Help: %s,\n", strconv.Quote("Number of requests received."))
	w.w.Write("}, []string{\"method\"}),\n")

	w.w.Write("requestLatency: %s.NewSummaryFrom(%s.SummaryOpts{\n", kitPrometheusPkg, stdPrometheusPkg)
	w.w.Write("Namespace: %s,\n", strconv.Quote(w.ctx.instrumenting.namespace))
	w.w.Write("Subsystem: %s,\n", strconv.Quote(w.ctx.instrumenting.subsystem))
	w.w.Write("Name: %s,\n", strconv.Quote("request_latency_microseconds"))
	w.w.Write("Help: %s,\n", strconv.Quote("Total duration of requests in microseconds."))
	w.w.Write("}, []string{\"method\"}),\n")
	w.w.Write("}\n")
}

func (w *TransportHTTP) writeClientStructOptions(opts *transportOptions) {
	var (
		kithttpPkg string
	)
	if opts.jsonRPC.enable {
		if opts.fastHTTP {
			kithttpPkg = w.w.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/fasthttp/jsonrpc")
		} else {
			kithttpPkg = w.w.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/http/jsonrpc")
		}
	} else {
		if opts.fastHTTP {
			kithttpPkg = w.w.Import("fasthttp", "github.com/l-vitaly/go-kit/transport/fasthttp")
		} else {
			kithttpPkg = w.w.Import("http", "github.com/go-kit/kit/transport/http")
		}
	}

	clientType := "client" + w.ctx.id

	w.w.Write("type %[1]sOption func(*%[1]s)\n", clientType)

	w.w.WriteFunc(
		w.ctx.id+"GenericClientOptions",
		"",
		[]string{"opt", "..." + kithttpPkg + ".ClientOption"},
		[]string{"", clientType + "Option"},
		func() {
			w.w.Write("return func(c *%s) { c.genericClientOption = opt }\n", clientType)
		},
	)

	for i := 0; i < w.ctx.iface.NumMethods(); i++ {
		m := w.ctx.iface.Method(i)
		lcName := strings.LcFirst(m.Name())

		w.w.WriteFunc(
			w.ctx.id+m.Name()+"ClientOptions",
			"",
			[]string{"opt", "..." + kithttpPkg + ".ClientOption"},
			[]string{"", clientType + "Option"},
			func() {
				w.w.Write("return func(c *%s) { c.%sClientOption = opt }\n", clientType, lcName)
			},
		)
	}
}

func (w *TransportHTTP) writeClientStruct(opts *transportOptions) {
	var (
		kithttpPkg string
	)
	if opts.jsonRPC.enable {
		if opts.fastHTTP {
			kithttpPkg = w.w.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/fasthttp/jsonrpc")
		} else {
			kithttpPkg = w.w.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/http/jsonrpc")
		}
	} else {
		if opts.fastHTTP {
			kithttpPkg = w.w.Import("fasthttp", "github.com/l-vitaly/go-kit/transport/fasthttp")
		} else {
			kithttpPkg = w.w.Import("http", "github.com/go-kit/kit/transport/http")
		}
	}

	endpointPkg := w.w.Import("endpoint", "github.com/go-kit/kit/endpoint")
	contextPkg := w.w.Import("context", "context")

	w.writeClientStructOptions(opts)

	clientType := "client" + w.ctx.id

	w.w.Write("type %s struct {\n", clientType)
	for i := 0; i < w.ctx.iface.NumMethods(); i++ {
		lcName := strings.LcFirst(w.ctx.iface.Method(i).Name())
		w.w.Write("%sEndpoint %s.Endpoint\n", lcName, endpointPkg)
		w.w.Write("%sClientOption []%s.ClientOption\n", lcName, kithttpPkg)
		w.w.Write("%sEndpointMiddleware []%s.Middleware\n", lcName, endpointPkg)
	}
	w.w.Write("genericClientOption []%s.ClientOption\n", kithttpPkg)
	w.w.Write("genericEndpointMiddleware []%s.Middleware\n", endpointPkg)

	w.w.Write("}\n\n")

	for i := 0; i < w.ctx.iface.NumMethods(); i++ {
		m := w.ctx.iface.Method(i)
		msig := m.Type().(*stdtypes.Signature)

		params := utils.NameTypeParams(msig.Params(), w.w.TypeString, nil)
		results := utils.NameTypeParams(msig.Results(), w.w.TypeString, nil)

		w.w.WriteFunc(m.Name(), "c *"+clientType, params, results, func() {
			hasError := types.HasErrorInResults(msig.Results())
			epResult := make([]string, 0, 2)

			resultLen := msig.Results().Len()
			if resultLen > 0 {
				epResult = append(epResult, "resp")
			}
			if hasError {
				epResult = append(epResult, "err")
			}

			if len(epResult) > 0 {
				w.w.Write("%s := ", stdstrings.Join(epResult, ","))
			}

			w.w.Write("c.%sEndpoint(", strings.LcFirst(m.Name()))

			if msig.Params().Len() > 0 {
				hasContext := types.HasContext(msig.Params().At(0).Type())
				if hasContext {
					w.w.Write("%s,", msig.Params().At(0).Name())
				} else {
					w.w.Write("%s.Background(),", contextPkg)
				}
				if hasContext && msig.Params().Len() == 1 {
					w.w.Write("nil")
				} else {
					w.w.Write("%sRequest%s", strings.LcFirst(m.Name()), w.ctx.id)
					params := structKeyValue(msig.Params(), func(p *stdtypes.Var) bool {
						if types.HasContext(p.Type()) {
							return false
						}
						return true
					})
					w.w.WriteStructAssign(params)
				}
				w.w.Write(")\n")
			}

			if hasError {
				w.w.Write("if err != nil {\n")
				w.w.Write("return ")
				for i := 0; i < msig.Results().Len(); i++ {
					r := msig.Results().At(i)
					if i > 0 {
						w.w.Write(",")
					}
					w.w.Write(r.Name())
				}
				w.w.Write("}\n")
			}

			if len(epResult) > 0 {
				w.w.Write("response := resp.(%sResponse%s)\n", strings.LcFirst(m.Name()), w.ctx.id)
				w.w.Write("return ")

				for i := 0; i < msig.Results().Len(); i++ {
					r := msig.Results().At(i)
					if i > 0 {
						w.w.Write(",")
					}
					w.w.Write("response.%s", strings.UcFirst(r.Name()))
				}
				w.w.Write("\n")
			}
		})
	}
}

func (w *TransportHTTP) writeRestClient(opts *transportOptions) {
	var (
		kithttpPkg string
		httpPkg    string
	)
	if opts.fastHTTP {
		kithttpPkg = w.w.Import("fasthttp", "github.com/l-vitaly/go-kit/transport/fasthttp")
	} else {
		kithttpPkg = w.w.Import("http", "github.com/go-kit/kit/transport/http")
	}
	if opts.fastHTTP {
		httpPkg = w.w.Import("fasthttp", "github.com/valyala/fasthttp")
	} else {
		httpPkg = w.w.Import("http", "net/http")
	}
	jsonPkg := w.w.Import("ffjson", "github.com/pquerna/ffjson/ffjson")
	pkgIO := w.w.Import("io", "io")
	fmtPkg := w.w.Import("fmt", "fmt")
	contextPkg := w.w.Import("context", "context")
	urlPkg := w.w.Import("url", "net/url")

	w.w.Write("u, err := %s.Parse(tgt)\n", urlPkg)

	w.w.WriteCheckErr(func() {
		w.w.Write("return nil, err")
	})

	for i := 0; i < w.ctx.iface.NumMethods(); i++ {
		m := w.ctx.iface.Method(i)
		msig := m.Type().(*stdtypes.Signature)
		lcName := strings.LcFirst(m.Name())

		epName := lcName + "Endpoint"

		mopts := opts.methodOptions[m.Name()]

		defaultHTTPMethod := "GET"
		w.w.Write("c.%s = %s.NewClient(\n", epName, kithttpPkg)
		if mopts.method.name != "" {
			w.w.WriteAST(mopts.method.expr)
		} else {
			w.w.Write(strconv.Quote(defaultHTTPMethod))
		}
		w.w.Write(",\n")
		w.w.Write("u,\n")

		if mopts.clientRequestFunc.expr != nil {
			w.w.WriteAST(mopts.clientRequestFunc.expr)
		} else {
			w.w.Write("func(_ %s.Context, r *%s.Request, request interface{}) error {\n", contextPkg, httpPkg)

			paramsLen := msig.Params().Len()
			if types.HasContextInParams(msig.Params()) {
				paramsLen--
			}

			if paramsLen > 0 {
				w.w.Write("req, ok := request.(%sRequest%s)\n", lcName, w.ctx.id)
				w.w.Write("if !ok {\n")
				w.w.Write("return %s.Errorf(\"couldn't assert request as %sRequest%s, got %%T\", request)\n", fmtPkg, lcName, w.ctx.id)
				w.w.Write("}\n")
			}

			pathStr := mopts.path
			pathVars := []string{}
			for name, regexp := range mopts.pathVars {
				if p := types.LookupFieldSig(name, msig); p != nil {
					if regexp != "" {
						regexp = ":" + regexp
					}
					pathStr = stdstrings.Replace(pathStr, "{"+name+regexp+"}", "%s", -1)
					pathVars = append(pathVars, w.w.GetFormatType("req."+strings.UcFirst(p.Name()), p))
				}
			}

			if opts.fastHTTP {
				w.w.Write("r.Header.SetMethod(")
			} else {
				w.w.Write("r.Method = ")
			}
			if mopts.method.name != "" {
				w.w.WriteAST(mopts.method.expr)
			} else {
				w.w.Write(strconv.Quote(defaultHTTPMethod))
			}
			if opts.fastHTTP {
				w.w.Write(")")
			}
			w.w.Write("\n")

			if opts.fastHTTP {
				w.w.Write("r.SetRequestURI(")
			} else {
				w.w.Write("r.URL.Path = ")
			}
			w.w.Write("%s.Sprintf(%s, %s)", fmtPkg, strconv.Quote(pathStr), stdstrings.Join(pathVars, ","))

			if opts.fastHTTP {
				w.w.Write(")")
			}
			w.w.Write("\n")

			if len(mopts.queryVars) > 0 {
				if opts.fastHTTP {
					w.w.Write("q := r.URI().QueryArgs()\n")
				} else {
					w.w.Write("q := r.URL.Query()\n")
				}

				for fName, qName := range mopts.queryVars {
					if p := types.LookupFieldSig(fName, msig); p != nil {
						w.w.Write("q.Add(%s, %s)\n", strconv.Quote(qName), w.w.GetFormatType("req."+strings.UcFirst(p.Name()), p))
					}
				}

				if opts.fastHTTP {
					w.w.Write("r.URI().SetQueryString(q.String())\n")
				} else {
					w.w.Write("r.URL.RawQuery = q.Encode()\n")
				}
			}

			for fName, hName := range mopts.headerVars {
				if p := types.LookupFieldSig(fName, msig); p != nil {
					w.w.Write("r.Header.Add(%s, %s)\n", strconv.Quote(hName), w.w.GetFormatType("req."+strings.UcFirst(p.Name()), p))
				}
			}

			switch stdstrings.ToUpper(mopts.method.name) {
			case "POST", "PUT", "PATCH":
				jsonPkg := w.w.Import("ffjson", "github.com/pquerna/ffjson/ffjson")

				w.w.Write("data, err := %s.Marshal(req)\n", jsonPkg)
				w.w.Write("if err != nil  {\n")
				w.w.Write("return %s.Errorf(\"couldn't marshal request %%T: %%s\", req, err)\n", fmtPkg)
				w.w.Write("}\n")

				if opts.fastHTTP {
					w.w.Write("r.SetBody(data)\n")
				} else {
					ioutilPkg := w.w.Import("ioutil", "io/ioutil")
					bytesPkg := w.w.Import("bytes", "bytes")

					w.w.Write("r.Body = %s.NopCloser(%s.NewBuffer(data))\n", ioutilPkg, bytesPkg)
				}
			}
			w.w.Write("return nil\n")
			w.w.Write("}")
		}
		w.w.Write(",\n")

		if mopts.clientResponseFunc.expr != nil {
			w.w.WriteAST(mopts.clientResponseFunc.expr)
		} else {
			w.w.Write("func(_ %s.Context, r *%s.Response) (interface{}, error) {\n", contextPkg, httpPkg)

			statusCode := "r.StatusCode"
			if opts.fastHTTP {
				statusCode = "r.StatusCode()"
			}

			w.w.Write("if statusCode := %s; statusCode != %s.StatusOK {\n", statusCode, httpPkg)
			w.w.Write("return nil, ErrorDecode(statusCode)\n")
			w.w.Write("}\n")

			w.w.Write("var resp %sResponse%s\n", lcName, w.ctx.id)

			if opts.notWrapBody {
				w.w.Write("var body %s\n", w.w.TypeString(msig.Results().At(0).Type()))
			}

			if opts.fastHTTP {
				w.w.Write("err := %s.Unmarshal(r.Body(), ", jsonPkg)
			} else {
				ioutilPkg := w.w.Import("ioutil", "io/ioutil")

				w.w.Write("b, err := %s.ReadAll(r.Body)\n", ioutilPkg)
				w.w.WriteCheckErr(func() {
					w.w.Write("return nil, err\n")
				})
				w.w.Write("err = %s.Unmarshal(b, ", jsonPkg)
			}

			if opts.notWrapBody {
				w.w.Write("&body)\n")
			} else {
				w.w.Write("&resp)\n")
			}

			w.w.Write("if err != nil && err != %s.EOF {\n", pkgIO)
			w.w.Write("return nil, %s.Errorf(\"couldn't unmarshal body to %sResponse%s: %%s\", err)\n", fmtPkg, lcName, w.ctx.id)
			w.w.Write("}\n")

			if opts.notWrapBody {
				w.w.Write("resp.%s = body\n", strings.UcFirst(msig.Results().At(0).Name()))
			}

			w.w.Write("return resp, nil\n")

			w.w.Write("}")
		}

		w.w.Write(",\n")

		w.w.Write("append(c.genericClientOption, c.%sClientOption...)...,\n", lcName)

		w.w.Write(").Endpoint()\n")

		w.w.Write("for _, e := range c.%sEndpointMiddleware {\n", lcName)
		w.w.Write("c.%[1]sEndpoint = e(c.%[1]sEndpoint)\n", lcName)
		w.w.Write("}\n")
	}
}

func (w *TransportHTTP) writeJsonRPCClient(opts *transportOptions) {
	var (
		jsonrpcPkg string
	)
	if opts.fastHTTP {
		jsonrpcPkg = w.w.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/fasthttp/jsonrpc")
	} else {
		jsonrpcPkg = w.w.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/http/jsonrpc")
	}

	urlPkg := w.w.Import("url", "net/url")
	contextPkg := w.w.Import("context", "context")
	ffjsonPkg := w.w.Import("ffjson", "github.com/pquerna/ffjson/ffjson")
	jsonPkg := w.w.Import("json", "encoding/json")
	fmtPkg := w.w.Import("fmt", "fmt")

	w.w.Write("u, err := %s.Parse(tgt)\n", urlPkg)

	w.w.WriteCheckErr(func() {
		w.w.Write("return nil, err")
	})

	for i := 0; i < w.ctx.iface.NumMethods(); i++ {
		m := w.ctx.iface.Method(i)
		sig := m.Type().(*stdtypes.Signature)
		lcName := strings.LcFirst(m.Name())

		w.w.Write("c.%[1]sClientOption = append(\nc.%[1]sClientOption,\n", lcName)

		w.w.Write("%s.ClientRequestEncoder(", jsonrpcPkg)
		w.w.Write("func(_ %s.Context, obj interface{}) (%s.RawMessage, error) {\n", contextPkg, jsonPkg)

		pramsLen := sig.Params().Len()
		if types.HasContextInSignature(sig) {
			pramsLen--
		}

		if pramsLen > 0 {
			w.w.Write("req, ok := obj.(%sRequest%s)\n", lcName, w.ctx.id)
			w.w.Write("if !ok {\n")
			w.w.Write("return nil, %s.Errorf(\"couldn't assert request as %sRequest%s, got %%T\", obj)\n", fmtPkg, lcName, w.ctx.id)
			w.w.Write("}\n")
			w.w.Write("b, err := %s.Marshal(req)\n", ffjsonPkg)
			w.w.Write("if err != nil {\n")
			w.w.Write("return nil, %s.Errorf(\"couldn't marshal request %%T: %%s\", obj, err)\n", fmtPkg)
			w.w.Write("}\n")
			w.w.Write("return b, nil\n")
		} else {
			w.w.Write("return nil, nil\n")
		}
		w.w.Write("}),\n")

		w.w.Write("%s.ClientResponseDecoder(", jsonrpcPkg)
		w.w.Write("func(_ %s.Context, response %s.Response) (interface{}, error) {\n", contextPkg, jsonrpcPkg)
		w.w.Write("if response.Error != nil {\n")
		w.w.Write("return nil, ErrorDecode(response.Error.Code)\n")
		w.w.Write("}\n")
		w.w.Write("var res %sResponse%s\n", lcName, w.ctx.id)
		w.w.Write("err := %s.Unmarshal(response.Result, &res)\n", ffjsonPkg)
		w.w.Write("if err != nil {\n")
		w.w.Write("return nil, %s.Errorf(\"couldn't unmarshal body to %sResponse%s: %%s\", err)\n", fmtPkg, lcName, w.ctx.id)
		w.w.Write("}\n")
		w.w.Write("return res, nil\n")

		w.w.Write("}),\n")

		w.w.Write(")\n")

		w.w.Write("c.%sEndpoint = %s.NewClient(\n", lcName, jsonrpcPkg)
		w.w.Write("u,\n")
		w.w.Write("%s,\n", strconv.Quote(lcName))

		w.w.Write("append(c.genericClientOption, c.%sClientOption...)...,\n", lcName)

		w.w.Write(").Endpoint()\n")

		w.w.Write("for _, e := range c.%sEndpointMiddleware {\n", lcName)
		w.w.Write("c.%[1]sEndpoint = e(c.%[1]sEndpoint)\n", lcName)
		w.w.Write("}\n")
	}
}

func (w *TransportHTTP) makeSwaggerSchema(t stdtypes.Type) (schema *openapi.Schema) {
	schema = &openapi.Schema{}
	switch v := t.(type) {
	case *stdtypes.Slice:
		if vv, ok := v.Elem().(*stdtypes.Basic); ok && vv.Kind() == stdtypes.Byte {
			schema.Type = "string"
			schema.Format = "byte"
		} else {
			schema.Type = "array"
			schema.Items = w.makeSwaggerSchema(v.Elem())
		}
	case *stdtypes.Basic:
		switch v.Kind() {
		case stdtypes.String:
			schema.Type = "string"
			schema.Format = "string"
			schema.Example = "abc"
		case stdtypes.Bool:
			schema.Type = "boolean"
			schema.Example = "true"
		case stdtypes.Int8, stdtypes.Int16:
			schema.Type = "integer"
			schema.Example = "1"
		case stdtypes.Int32:
			schema.Type = "integer"
			schema.Format = "int32"
			schema.Example = "1"
		case stdtypes.Int, stdtypes.Int64:
			schema.Type = "integer"
			schema.Format = "int64"
			schema.Example = "1"
		case stdtypes.Float32, stdtypes.Float64:
			schema.Type = "number"
			schema.Format = "float"
			schema.Example = "1.1"
		}
	case *stdtypes.Struct:
		schema.Type = "object"
		schema.Properties = map[string]*openapi.Schema{}

		for i := 0; i < v.NumFields(); i++ {
			f := v.Field(i)
			schema.Properties[strcase.ToLowerCamel(f.Name())] = w.makeSwaggerSchema(f.Type())
		}
	case *stdtypes.Named:
		switch stdtypes.TypeString(v, nil) {
		case "time.Time":
			schema.Type = "string"
			schema.Format = "date-time"
			return
		case "github.com/pborman/uuid.UUID":
			schema.Type = "string"
			schema.Format = "uuid"
			return
		}
		return w.makeSwaggerSchema(v.Obj().Type().Underlying())
	}
	return
}

func newTransportHTTP(ctx serviceCtx, w *writer.Writer) *TransportHTTP {
	return &TransportHTTP{ctx: ctx, w: w}
}
