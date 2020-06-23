package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/constant"
	stdtypes "go/types"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	stdstrings "strings"

	"github.com/iancoleman/strcase"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/swipe-io/swipe/pkg/errors"
	"github.com/swipe-io/swipe/pkg/openapi"
	"github.com/swipe-io/swipe/pkg/parser"
	"github.com/swipe-io/swipe/pkg/strings"
	"github.com/swipe-io/swipe/pkg/types"
	"github.com/swipe-io/swipe/pkg/utils"
	"github.com/swipe-io/swipe/pkg/writer"
	"golang.org/x/tools/go/packages"
)

const jsonRPCClientBase = `
export class JSONRPCError extends Error {
	constructor(message, name, code, data) {
	  super(message);
	  this.name = name;
	  this.code = code;
	  this.data = data;
	}
  }  
class JSONRPCClient {
	/**
	 *
	 * @param {*} transport
	 */
	constructor(transport) {
	  this._transport = transport;
	  this._requestID = 0;
	  this._scheduleRequests = {};
	  this._commitTimerID = null;
	  this._beforeRequest = null;
	}	
	beforeRequest(fn) {
	  this._beforeRequest = fn;
	}
	__scheduleCommit() {
	  if (this._commitTimerID) {
		clearTimeout(this._commitTimerID);
	  }
	  this._commitTimerID = setTimeout(() => {
		this._commitTimerID = null;
		const scheduleRequests = { ...this._scheduleRequests };
		this._scheduleRequests = {};
		let requests = [];
		for (let key in scheduleRequests) {
		  requests.push(scheduleRequests[key].request);
		}
		this.__doRequest(requests)
		  .then((responses) => {
			for (var i = 0; i < responses.length; i++) {
			  if (responses[i].error) {
				scheduleRequests[responses[i].id].reject(convertError(responses[i].error));
				continue;
			  }
			  scheduleRequests[responses[i].id].resolve(responses[i].result);
			}
		  })
		  .catch((e) => {
			for (var key in requests) {
			  if (!scheduleRequests.hasOwnProperty(key)) {
				continue;
			  }
			  scheduleRequests[key].reject(e);
			}
		  });
	  }, 0);
	}
	makeJSONRPCRequest(id, method, params) {
	  return {
		jsonrpc: "2.0",
		id: id,
		method: method,
		params: params,
	  };
	}
	__scheduleRequest(method, params) {
	  var p = new Promise((resolve, reject) => {
		const request = this.makeJSONRPCRequest(
		  this.__requestIDGenerate(),
		  method,
		  params
		);
		this._scheduleRequests[request.id] = {
		  request,
		  resolve,
		  reject,
		};
	  });
	  this.__scheduleCommit();
	  return p;
	}
	__doRequest(request) {
	  return this._transport.doRequest(request);
	}
	__requestIDGenerate() {
	  return ++this._requestID;
	}
  }
`

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

type transportWrapResponse struct {
	enable bool
	name   string
}

type transportMethodOptions struct {
	method             transportMethod
	path               string
	pathVars           map[string]string
	headerVars         map[string]string
	queryVars          map[string]string
	wrapResponse       transportWrapResponse
	serverRequestFunc  astOptions
	serverResponseFunc astOptions
	clientRequestFunc  astOptions
	clientResponseFunc astOptions
}

type transportOpenapiMethodOption struct {
	errors []string
	tags   []string
}

type transportOpenapiDoc struct {
	enable        bool
	output        string
	servers       []openapi.Server
	info          openapi.Info
	methods       map[string]*transportOpenapiMethodOption
	defaultMethod transportOpenapiMethodOption
}

type transportClient struct {
	enable bool
}

type transportOptions struct {
	prefix string

	serverDisabled bool
	client         transportClient
	openapiDoc     transportOpenapiDoc
	fastHTTP       bool
	jsonRPC        transportJsonRPCOption
	methodOptions  map[string]transportMethodOptions
	mapCodeErrors  map[string]*errorDecodeInfo
}

type errorDecodeInfo struct {
	code      int64
	n         *stdtypes.Named
	isPointer bool
}

type TransportHTTP struct {
	ctx serviceCtx
	w   *writer.Writer
}

func (g *TransportHTTP) Write(opt *parser.Option) error {

	_, enabledFastHTTP := opt.Get("FastEnable")

	options := &transportOptions{
		fastHTTP:      enabledFastHTTP,
		methodOptions: map[string]transportMethodOptions{},
		mapCodeErrors: map[string]*errorDecodeInfo{},
		openapiDoc: transportOpenapiDoc{
			methods: map[string]*transportOpenapiMethodOption{},
		},
	}

	if _, ok := opt.Get("ClientEnable"); ok {
		options.client.enable = true
	}

	if _, ok := opt.Get("ServerDisabled"); ok {
		options.serverDisabled = true
	}

	if openapiDocOpt, ok := opt.Get("Openapi"); ok {
		options.openapiDoc.enable = true
		if v, ok := openapiDocOpt.Get("OpenapiOutput"); ok {
			options.openapiDoc.output = v.Value.String()
		}
		if v, ok := openapiDocOpt.Get("OpenapiInfo"); ok {
			options.openapiDoc.info = openapi.Info{
				Title:       parser.MustOption(v.Get("title")).Value.String(),
				Description: parser.MustOption(v.Get("description")).Value.String(),
				Version:     parser.MustOption(v.Get("version")).Value.String(),
			}
		}
		if v, ok := openapiDocOpt.Get("OpenapiContact"); ok {
			options.openapiDoc.info.Contact = &openapi.Contact{
				Name:  parser.MustOption(v.Get("name")).Value.String(),
				Email: parser.MustOption(v.Get("email")).Value.String(),
				URL:   parser.MustOption(v.Get("url")).Value.String(),
			}
		}
		if v, ok := openapiDocOpt.Get("OpenapiLicence"); ok {
			options.openapiDoc.info.License = &openapi.License{
				Name: parser.MustOption(v.Get("name")).Value.String(),
				URL:  parser.MustOption(v.Get("url")).Value.String(),
			}
		}
		if s, ok := openapiDocOpt.GetSlice("OpenapiServer"); ok {
			for _, v := range s {
				options.openapiDoc.servers = append(options.openapiDoc.servers, openapi.Server{
					Description: parser.MustOption(v.Get("description")).Value.String(),
					URL:         parser.MustOption(v.Get("url")).Value.String(),
				})
			}
		}

		if openapiErrors, ok := openapiDocOpt.GetSlice("OpenapiErrors"); ok {
			for _, openapiErrorsOpt := range openapiErrors {
				var methods []string
				if methodsOpt, ok := openapiErrorsOpt.Get("methods"); ok {
					for _, expr := range methodsOpt.Value.ExprSlice() {
						fnSel, ok := expr.(*ast.SelectorExpr)
						if !ok {
							return errors.NotePosition(methodsOpt.Position, fmt.Errorf("the %s value must be func selector", methodsOpt.Name))
						}
						methods = append(methods, fnSel.Sel.Name)
						if _, ok := options.openapiDoc.methods[fnSel.Sel.Name]; !ok {
							options.openapiDoc.methods[fnSel.Sel.Name] = &transportOpenapiMethodOption{}
						}
					}
				}
				if errorsOpt, ok := openapiErrorsOpt.Get("errors"); ok {
					var errorsName []string
					for _, expr := range errorsOpt.Value.ExprSlice() {
						ptr, ok := g.w.TypeOf(expr).(*stdtypes.Pointer)
						if !ok {
							return errors.NotePosition(
								openapiErrorsOpt.Position, fmt.Errorf("the %s value must be nil pointer errors", openapiErrorsOpt.Name),
							)
						}
						named, ok := ptr.Elem().(*stdtypes.Named)
						if !ok {
							return errors.NotePosition(
								openapiErrorsOpt.Position, fmt.Errorf("the %s value must be nil pointer errors", openapiErrorsOpt.Name),
							)
						}
						errorsName = append(errorsName, named.Obj().Name())
					}
					if len(methods) > 0 {
						for _, method := range methods {
							options.openapiDoc.methods[method].errors = append(options.openapiDoc.methods[method].errors, errorsName...)
						}
					} else {
						options.openapiDoc.defaultMethod.errors = append(options.openapiDoc.defaultMethod.errors, errorsName...)
					}
				}
			}
		}

		if openapiTags, ok := openapiDocOpt.GetSlice("OpenapiTags"); ok {
			for _, openapiTagsOpt := range openapiTags {
				var methods []string
				if methodsOpt, ok := openapiTagsOpt.Get("methods"); ok {
					for _, expr := range methodsOpt.Value.ExprSlice() {
						fnSel, ok := expr.(*ast.SelectorExpr)
						if !ok {
							return errors.NotePosition(methodsOpt.Position, fmt.Errorf("the %s value must be func selector", methodsOpt.Name))
						}
						methods = append(methods, fnSel.Sel.Name)
						if _, ok := options.openapiDoc.methods[fnSel.Sel.Name]; !ok {
							options.openapiDoc.methods[fnSel.Sel.Name] = &transportOpenapiMethodOption{}
						}
					}
				}
				if tagsOpt, ok := openapiTagsOpt.Get("tags"); ok {
					if len(methods) > 0 {
						for _, method := range methods {
							options.openapiDoc.methods[method].tags = append(options.openapiDoc.methods[method].tags, tagsOpt.Value.StringSlice()...)
						}
					} else {
						options.openapiDoc.defaultMethod.tags = append(options.openapiDoc.defaultMethod.tags, tagsOpt.Value.StringSlice()...)
					}
				}
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

	if methodDefaultOpt, ok := opt.Get("MethodDefaultOptions"); ok {
		mopt, err := g.getMethodOptions(methodDefaultOpt, transportMethodOptions{})
		if err != nil {
			return err
		}
		for _, m := range g.ctx.iface.methods {
			options.methodOptions[m.name] = mopt
		}
	}

	if methods, ok := opt.GetSlice("MethodOptions"); ok {
		for _, methodOpt := range methods {
			signOpt := parser.MustOption(methodOpt.Get("signature"))
			fnSel, ok := signOpt.Value.Expr().(*ast.SelectorExpr)
			if !ok {
				return errors.NotePosition(signOpt.Position, fmt.Errorf("the Signature value must be func selector"))
			}
			baseMethodOpts := options.methodOptions[fnSel.Sel.Name]
			mopt, err := g.getMethodOptions(methodOpt, baseMethodOpts)
			if err != nil {
				return err
			}
			options.methodOptions[fnSel.Sel.Name] = mopt
		}
	}

	options.prefix = "REST"
	if options.jsonRPC.enable {
		options.prefix = "JSONRPC"
	}

	errorStatusMethod := "StatusCode"
	if options.jsonRPC.enable {
		errorStatusMethod = "ErrorCode"
	}

	g.w.Inspect(func(p *packages.Package, n ast.Node) bool {
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
							options.mapCodeErrors[named.Obj().Name()] = &errorDecodeInfo{isPointer: isPointer, n: named}
						}
					}
				}
			}
		}
		return true
	})

	g.w.Inspect(func(p *packages.Package, n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			if fn.Name.Name == errorStatusMethod {
				if fn.Recv != nil && len(fn.Recv.List) > 0 {
					recvType := p.TypesInfo.TypeOf(fn.Recv.List[0].Type)
					ptr, ok := recvType.(*stdtypes.Pointer)
					if ok {
						recvType = ptr.Elem()
					}
					if named, ok := recvType.(*stdtypes.Named); ok {
						if _, ok := options.mapCodeErrors[named.Obj().Name()]; ok {
							ast.Inspect(n, func(n ast.Node) bool {
								if ret, ok := n.(*ast.ReturnStmt); ok && len(ret.Results) == 1 {
									if v, ok := p.TypesInfo.Types[ret.Results[0]]; ok {
										if v.Value != nil && v.Value.Kind() == constant.Int {
											code, _ := constant.Int64Val(v.Value)
											options.mapCodeErrors[named.Obj().Name()].code = code
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

	if options.openapiDoc.enable {
		if err := g.writeOpenapiDoc(options); err != nil {
			return err
		}
	}

	var httpPkg string
	if options.fastHTTP {
		httpPkg = g.w.Import("fasthttp", "github.com/valyala/fasthttp")
	} else {
		httpPkg = g.w.Import("http", "net/http")
	}

	endpointPkg := g.w.Import("endpoint", "github.com/go-kit/kit/endpoint")

	g.w.Write("type httpError struct {\ncode int\n}\n")
	if options.fastHTTP {
		g.w.Write("func (e httpError) Error() string {\nreturn %s.StatusMessage(e.code)\n}\n", httpPkg)
	} else {
		g.w.Write("func (e httpError) Error() string {\nreturn %s.StatusText(e.code)\n}\n", httpPkg)
	}

	g.w.Write("func (e httpError) StatusCode() int {\nreturn e.code\n}\n")

	g.w.WriteFunc("ErrorDecode", "", []string{"code", "int"}, []string{"", "error"}, func() {
		g.w.Write("switch code {\n")
		g.w.Write("default:\nreturn httpError{code: code}\n")
		for _, i := range options.mapCodeErrors {
			g.w.Write("case %d:\n", i.code)
			pkg := g.w.Import(i.n.Obj().Pkg().Name(), i.n.Obj().Pkg().Path())
			g.w.Write("return ")
			if i.isPointer {
				g.w.Write("&")
			}
			g.w.Write("%s.%s{}\n", pkg, i.n.Obj().Name())
		}
		g.w.Write("}\n")
	})

	g.w.Write("func middlewareChain(middlewares []%[1]s.Middleware) %[1]s.Middleware {\n", endpointPkg)
	g.w.Write("return func(next %[1]s.Endpoint) %[1]s.Endpoint {\n", endpointPkg)
	g.w.Write("if len(middlewares) == 0 {\n")
	g.w.Write("return next\n")
	g.w.Write("}\n")
	g.w.Write("outer := middlewares[0]\n")
	g.w.Write("others := middlewares[1:]\n")
	g.w.Write("for i := len(others) - 1; i >= 0; i-- {\n")
	g.w.Write("next = others[i](next)\n")
	g.w.Write("}\n")
	g.w.Write("return outer(next)\n")
	g.w.Write("}\n")
	g.w.Write("}\n")

	if options.client.enable {
		g.writeClientStruct(options)

		clientType := "client" + g.ctx.id

		g.w.Write("func NewClient%s%s(tgt string", options.prefix, g.ctx.id)

		g.w.Write(" ,opts ...%[1]sOption", clientType)

		g.w.Write(") (%s, error) {\n", g.ctx.typeStr)

		g.w.Write("c := &%s{}\n", clientType)

		g.w.Write("for _, o := range opts {\n")
		g.w.Write("o(c)\n")
		g.w.Write("}\n")

		if options.jsonRPC.enable {
			g.writeJsonRPCClientGo(options)
			g.writeJsonRPCClientJS(options)
		} else {
			g.writeRestClient(options)
		}

		g.w.Write("return c, nil\n")
		g.w.Write("}\n")
	}
	if !options.serverDisabled {
		if err := g.writeHTTP(options); err != nil {
			return err
		}
	}
	return nil
}

func (g *TransportHTTP) getMethodOptions(methodOpt *parser.Option, baseMethodOpts transportMethodOptions) (transportMethodOptions, error) {
	if wrapResponseOpt, ok := methodOpt.Get("WrapResponse"); ok {
		baseMethodOpts.wrapResponse.enable = true
		baseMethodOpts.wrapResponse.name = wrapResponseOpt.Value.String()
	}

	if httpMethodOpt, ok := methodOpt.Get("Method"); ok {
		baseMethodOpts.method.name = httpMethodOpt.Value.String()
		baseMethodOpts.method.expr = httpMethodOpt.Value.Expr()
	}

	if path, ok := methodOpt.Get("Path"); ok {
		baseMethodOpts.path = path.Value.String()

		idxs, err := httpBraceIndices(baseMethodOpts.path)
		if err != nil {
			return baseMethodOpts, err
		}
		if len(idxs) > 0 {
			baseMethodOpts.pathVars = make(map[string]string, len(idxs))

			var end int
			for i := 0; i < len(idxs); i += 2 {
				end = idxs[i+1]
				parts := stdstrings.SplitN(baseMethodOpts.path[idxs[i]+1:end-1], ":", 2)

				name := parts[0]
				regexp := ""

				if len(parts) == 2 {
					regexp = parts[1]
				}
				baseMethodOpts.pathVars[name] = regexp
			}
		}
	}

	if serverRequestFunc, ok := methodOpt.Get("ServerDecodeRequestFunc"); ok {
		baseMethodOpts.serverRequestFunc.t = serverRequestFunc.Value.Type()
		baseMethodOpts.serverRequestFunc.expr = serverRequestFunc.Value.Expr()
	}

	if serverResponseFunc, ok := methodOpt.Get("ServerEncodeResponseFunc"); ok {
		baseMethodOpts.serverResponseFunc.t = serverResponseFunc.Value.Type()
		baseMethodOpts.serverResponseFunc.expr = serverResponseFunc.Value.Expr()
	}

	if clientRequestFunc, ok := methodOpt.Get("ClientEncodeRequestFunc"); ok {
		baseMethodOpts.clientRequestFunc.t = clientRequestFunc.Value.Type()
		baseMethodOpts.clientRequestFunc.expr = clientRequestFunc.Value.Expr()
	}

	if clientResponseFunc, ok := methodOpt.Get("ClientDecodeResponseFunc"); ok {
		baseMethodOpts.clientResponseFunc.t = clientResponseFunc.Value.Type()
		baseMethodOpts.clientResponseFunc.expr = clientResponseFunc.Value.Expr()
	}

	if queryVars, ok := methodOpt.Get("QueryVars"); ok {
		baseMethodOpts.queryVars = map[string]string{}

		values := queryVars.Value.StringSlice()
		for i := 0; i < len(values); i += 2 {
			baseMethodOpts.queryVars[values[i]] = values[i+1]
		}
	}
	if headerVars, ok := methodOpt.Get("HeaderVars"); ok {
		baseMethodOpts.headerVars = map[string]string{}
		values := headerVars.Value.StringSlice()
		for i := 0; i < len(values); i += 2 {
			baseMethodOpts.headerVars[values[i]] = values[i+1]
		}
	}
	return baseMethodOpts, nil
}

func (g *TransportHTTP) writeHTTP(opts *transportOptions) error {
	var (
		kithttpPkg string
	)
	endpointPkg := g.w.Import("endpoint", "github.com/go-kit/kit/endpoint")
	if opts.jsonRPC.enable {
		if opts.fastHTTP {
			kithttpPkg = g.w.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/fasthttp/jsonrpc")
		} else {
			kithttpPkg = g.w.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/http/jsonrpc")
		}
	} else {
		if opts.fastHTTP {
			kithttpPkg = g.w.Import("fasthttp", "github.com/l-vitaly/go-kit/transport/fasthttp")
		} else {
			kithttpPkg = g.w.Import("http", "github.com/go-kit/kit/transport/http")
		}
	}

	serverOptType := fmt.Sprintf("server%sOpts", g.ctx.id)
	serverOptionType := fmt.Sprintf("server%sOption", g.ctx.id)
	kithttpServerOption := fmt.Sprintf("%s.ServerOption", kithttpPkg)
	endpointMiddlewareOption := fmt.Sprintf("%s.Middleware", endpointPkg)

	g.w.Write("type %s func (*%s)\n", serverOptionType, serverOptType)

	g.w.Write("type %s struct {\n", serverOptType)
	g.w.Write("genericServerOption []%s\n", kithttpServerOption)
	g.w.Write("genericEndpointMiddleware []%s\n", endpointMiddlewareOption)

	for _, m := range g.ctx.iface.methods {
		g.w.Write("%sServerOption []%s\n", m.lcName, kithttpServerOption)
		g.w.Write("%sEndpointMiddleware []%s\n", m.lcName, endpointMiddlewareOption)
	}
	g.w.Write("}\n")

	g.w.WriteFunc(
		g.ctx.id+"GenericServerOptions",
		"",
		[]string{"v", "..." + kithttpServerOption},
		[]string{"", serverOptionType},
		func() {
			g.w.Write("return func(o *%s) { o.genericServerOption = v }\n", serverOptType)
		},
	)

	g.w.WriteFunc(
		g.ctx.id+"GenericServerEndpointMiddlewares",
		"",
		[]string{"v", "..." + endpointMiddlewareOption},
		[]string{"", serverOptionType},
		func() {
			g.w.Write("return func(o *%s) { o.genericEndpointMiddleware = v }\n", serverOptType)
		},
	)

	for _, m := range g.ctx.iface.methods {
		g.w.WriteFunc(
			g.ctx.id+m.name+"ServerOptions",
			"",
			[]string{"opt", "..." + kithttpServerOption},
			[]string{"", serverOptionType},
			func() {
				g.w.Write("return func(c *%s) { c.%sServerOption = opt }\n", serverOptType, m.lcName)
			},
		)

		g.w.WriteFunc(
			g.ctx.id+m.name+"ServerEndpointMiddlewares",
			"",
			[]string{"opt", "..." + endpointMiddlewareOption},
			[]string{"", serverOptionType},
			func() {
				g.w.Write("return func(c *%s) { c.%sEndpointMiddleware = opt }\n", serverOptType, m.lcName)
			},
		)
	}

	g.w.Write("// HTTP %s Transport\n", opts.prefix)

	if opts.jsonRPC.enable {
		g.writeJsonRPCEncodeResponse()
	} else {
		g.writeHTTPEncodeResponse(opts)
	}

	if opts.jsonRPC.enable {
		g.writeJSONRPCEndpointCodecMap(opts)
	}

	g.w.Write("func MakeHandler%s%s(s %s", opts.prefix, g.ctx.id, g.ctx.typeStr)
	if g.ctx.logging {
		logPkg := g.w.Import("log", "github.com/go-kit/kit/log")
		g.w.Write(", logger %s.Logger", logPkg)
	}
	g.w.Write(", opts ...server%sOption", g.ctx.id)
	g.w.Write(") (")
	if opts.fastHTTP {
		g.w.Write("%s.RequestHandler", g.w.Import("fasthttp", "github.com/valyala/fasthttp"))
	} else {
		g.w.Write("%s.Handler", g.w.Import("http", "net/http"))
	}

	g.w.Write(", error) {\n")

	g.w.Write("sopt := &server%sOpts{}\n", g.ctx.id)

	g.w.Write("for _, o := range opts {\n o(sopt)\n }\n")

	g.writeMiddlewares(opts)

	g.w.Write("ep := MakeEndpointSet(s)\n")

	for _, m := range g.ctx.iface.methods {
		g.w.Write("ep.%[1]sEndpoint = middlewareChain(append(sopt.genericEndpointMiddleware, sopt.%[2]sEndpointMiddleware...))(ep.%[1]sEndpoint)\n", m.name, m.lcName)
	}

	if opts.jsonRPC.enable {
		g.writeJSONRPCHandler(opts)
	} else {
		g.writeRESTHandler(opts)
	}

	g.w.Write("}\n\n")

	return nil
}

func (g *TransportHTTP) writeJsonRPCEncodeResponse() {
	ffjsonPkg := g.w.Import("ffjson", "github.com/pquerna/ffjson/ffjson")
	jsonPkg := g.w.Import("json", "encoding/json")
	contextPkg := g.w.Import("context", "context")

	g.w.Write("func encodeResponseJSONRPC%s(_ %s.Context, result interface{}) (%s.RawMessage, error) {\n", g.ctx.id, contextPkg, jsonPkg)
	g.w.Write("b, err := %s.Marshal(result)\n", ffjsonPkg)
	g.w.Write("if err != nil {\n")
	g.w.Write("return nil, err\n")
	g.w.Write("}\n")
	g.w.Write("return b, nil\n")
	g.w.Write("}\n\n")
}

func (g *TransportHTTP) writeHTTPEncodeResponse(opts *transportOptions) {
	kitEndpointPkg := g.w.Import("endpoint", "github.com/go-kit/kit/endpoint")
	jsonPkg := g.w.Import("ffjson", "github.com/pquerna/ffjson/ffjson")
	contextPkg := g.w.Import("context", "context")

	var httpPkg string

	if opts.fastHTTP {
		httpPkg = g.w.Import("fasthttp", "github.com/valyala/fasthttp")
	} else {
		httpPkg = g.w.Import("http", "net/http")
	}

	g.w.Write("type errorWrapper struct {\n")
	g.w.Write("Error string `json:\"error\"`\n")
	g.w.Write("}\n")

	g.w.Write("func encodeResponseHTTP%s(ctx %s.Context, ", g.ctx.id, contextPkg)

	if opts.fastHTTP {
		g.w.Write("w *%s.Response", httpPkg)
	} else {
		g.w.Write("w %s.ResponseWriter", httpPkg)
	}

	g.w.Write(", response interface{}) error {\n")

	if opts.fastHTTP {
		g.w.Write("h := w.Header\n")
	} else {
		g.w.Write("h := w.Header()\n")
	}

	g.w.Write("h.Set(\"Content-Type\", \"application/json; charset=utf-8\")\n")

	g.w.Write("if e, ok := response.(%s.Failer); ok && e.Failed() != nil {\n", kitEndpointPkg)

	g.w.Write("data, err := %s.Marshal(errorWrapper{Error: e.Failed().Error()})\n", jsonPkg)
	g.w.Write("if err != nil {\n")
	g.w.Write("return err\n")
	g.w.Write("}\n")

	if opts.fastHTTP {
		g.w.Write("w.SetBody(data)\n")
	} else {
		g.w.Write("w.Write(data)\n")
	}

	g.w.Write("return nil\n")
	g.w.Write("}\n")

	g.w.Write("data, err := %s.Marshal(response)\n", jsonPkg)
	g.w.Write("if err != nil {\n")
	g.w.Write("return err\n")
	g.w.Write("}\n")

	if opts.fastHTTP {
		g.w.Write("w.SetBody(data)\n")
	} else {
		g.w.Write("w.Write(data)\n")
	}

	g.w.Write("return nil\n")
	g.w.Write("}\n\n")
}

func (g *TransportHTTP) makeRestPath(opts *transportOptions, m ifaceServiceMethod) *openapi.Operation {
	mopt := opts.methodOptions[m.name]

	responseSchema := &openapi.Schema{
		Type:       "object",
		Properties: map[string]*openapi.Schema{},
	}

	requestSchema := &openapi.Schema{
		Type:       "object",
		Properties: map[string]*openapi.Schema{},
	}

	for _, p := range m.params {
		if _, ok := mopt.pathVars[p.Name()]; ok {
			continue
		}

		if _, ok := mopt.queryVars[p.Name()]; ok {
			continue
		}

		if _, ok := mopt.headerVars[p.Name()]; ok {
			continue
		}

		if types.IsContext(p.Type()) {
			continue
		}
		requestSchema.Properties[strcase.ToLowerCamel(p.Name())] = g.makeSwaggerSchema(p.Type())
	}

	if len(m.results) > 1 {
		for _, r := range m.results {
			responseSchema.Properties[strcase.ToLowerCamel(r.Name())] = g.makeSwaggerSchema(r.Type())
		}
	} else if len(m.results) == 1 {
		responseSchema = g.makeSwaggerSchema(m.results[0].Type())
	}

	if mopt.wrapResponse.enable {
		properties := openapi.Properties{}
		properties[mopt.wrapResponse.name] = responseSchema
		responseSchema = &openapi.Schema{
			Properties: properties,
		}
	}

	o := &openapi.Operation{
		Summary: m.name,
		Responses: map[string]openapi.Response{
			"200": {
				Description: "OK",
				Content: openapi.Content{
					"application/json": {
						Schema: responseSchema,
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
	}

	for name := range mopt.pathVars {
		var schema *openapi.Schema
		if fld := m.params.lookupField(name); fld != nil {
			schema = g.makeSwaggerSchema(fld.Type())
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
		if fld := m.params.lookupField(argName); fld != nil {
			schema = g.makeSwaggerSchema(fld.Type())
		}
		o.Parameters = append(o.Parameters, openapi.Parameter{
			In:     "query",
			Name:   name,
			Schema: schema,
		})
	}

	for argName, name := range mopt.headerVars {
		var schema *openapi.Schema
		if fld := m.params.lookupField(argName); fld != nil {
			schema = g.makeSwaggerSchema(fld.Type())
		}
		o.Parameters = append(o.Parameters, openapi.Parameter{
			In:     "header",
			Name:   name,
			Schema: schema,
		})
	}

	switch mopt.method.name {
	case "POST", "PUT", "PATCH":
		o.RequestBody = &openapi.RequestBody{
			Required: true,
			Content: map[string]openapi.Media{
				"application/json": {
					Schema: requestSchema,
				},
			},
		}
	}

	return o
}

func (g *TransportHTTP) makeJSONRPCPath(opts *transportOptions, m ifaceServiceMethod) *openapi.Operation {
	mopt := opts.methodOptions[m.name]

	responseSchema := &openapi.Schema{
		Type:       "object",
		Properties: map[string]*openapi.Schema{},
	}

	requestSchema := &openapi.Schema{
		Type:       "object",
		Properties: map[string]*openapi.Schema{},
	}

	if len(m.params) > 0 {
		for _, p := range m.params {
			requestSchema.Properties[strcase.ToLowerCamel(p.Name())] = g.makeSwaggerSchema(p.Type())
		}
	} else {
		requestSchema.Example = json.RawMessage("null")
	}

	if len(m.results) > 1 {
		for _, r := range m.results {
			responseSchema.Properties[strcase.ToLowerCamel(r.Name())] = g.makeSwaggerSchema(r.Type())
		}
	} else if len(m.results) == 1 {
		responseSchema = g.makeSwaggerSchema(m.results[0].Type())
	} else {
		responseSchema.Example = json.RawMessage("null")
	}

	if mopt.wrapResponse.enable {
		properties := openapi.Properties{}
		properties[mopt.wrapResponse.name] = responseSchema
		responseSchema = &openapi.Schema{
			Properties: properties,
		}
	}

	response := &openapi.Schema{
		Type: "object",
		Properties: openapi.Properties{
			"jsonrpc": &openapi.Schema{
				Type:    "string",
				Example: "2.0",
			},
			"id": &openapi.Schema{
				Type:    "string",
				Example: "c9b14c57-7503-447a-9fb9-be6f8920f31f",
			},
			"result": responseSchema,
		},
	}
	request := &openapi.Schema{
		Type: "object",
		Properties: openapi.Properties{
			"jsonrpc": &openapi.Schema{
				Type:    "string",
				Example: "2.0",
			},
			"id": &openapi.Schema{
				Type:    "string",
				Example: "c9b14c57-7503-447a-9fb9-be6f8920f31f",
			},
			"method": &openapi.Schema{
				Type: "string",
				Enum: []string{strcase.ToLowerCamel(m.name)},
			},
			"params": requestSchema,
		},
	}

	return &openapi.Operation{
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
			"x-32700": {
				Description: "Parse error. Invalid JSON was received by the server. An error occurred on the server while parsing the JSON text.",
				Content: openapi.Content{
					"application/json": {
						Schema: &openapi.Schema{
							Ref: "#/components/schemas/ParseError",
						},
					},
				},
			},
			"x-32600": {
				Description: "Invalid Request. The JSON sent is not a valid Request object.",
				Content: openapi.Content{
					"application/json": {
						Schema: &openapi.Schema{
							Ref: "#/components/schemas/InvalidRequestError",
						},
					},
				},
			},
			"x-32601": {
				Description: "Method not found. The method does not exist / is not available.",
				Content: openapi.Content{
					"application/json": {
						Schema: &openapi.Schema{
							Ref: "#/components/schemas/MethodNotFoundError",
						},
					},
				},
			},
			"x-32602": {
				Description: "Invalid params. Invalid method parameters.",
				Content: openapi.Content{
					"application/json": {
						Schema: &openapi.Schema{
							Ref: "#/components/schemas/InvalidParamsError",
						},
					},
				},
			},
			"x-32603": {
				Description: "Internal error. Internal JSON-RPC error.",
				Content: openapi.Content{
					"application/json": {
						Schema: &openapi.Schema{
							Ref: "#/components/schemas/InternalError",
						},
					},
				},
			},
		},
	}
}

func (g *TransportHTTP) writeOpenapiDoc(opts *transportOptions) error {
	swg := openapi.OpenAPI{
		OpenAPI: "3.0.0",
		Info:    opts.openapiDoc.info,
		Servers: opts.openapiDoc.servers,
		Paths:   map[string]*openapi.Path{},
		Components: openapi.Components{
			Schemas: openapi.Schemas{},
		},
	}

	if opts.jsonRPC.enable {
		swg.Components.Schemas = getOpenapiJSONRPCErrorSchemas()
	} else {
		swg.Components.Schemas["Error"] = getOpenapiRestErrorSchema()
	}

	for name, ei := range opts.mapCodeErrors {
		var s *openapi.Schema
		if opts.jsonRPC.enable {
			s = &openapi.Schema{
				Type: "object",
				Properties: openapi.Properties{
					"jsonrpc": &openapi.Schema{
						Type:    "string",
						Example: "2.0",
					},
					"id": &openapi.Schema{
						Type:    "string",
						Example: "1f1ecd1b-d729-40cd-b6f4-4011f69811fe",
					},
					"error": &openapi.Schema{
						Type: "object",
						Properties: openapi.Properties{
							"code": &openapi.Schema{
								Type:    "integer",
								Example: ei.code,
							},
							"message": &openapi.Schema{
								Type: "string",
							},
						},
					},
				},
			}
		} else {
			s = &openapi.Schema{
				Type: "object",
				Properties: openapi.Properties{
					"error": &openapi.Schema{
						Type: "string",
					},
				},
			}
		}
		swg.Components.Schemas[name] = s
	}

	for _, m := range g.ctx.iface.methods {
		mopt := opts.methodOptions[m.name]

		var (
			o       *openapi.Operation
			pathStr string
			errors  = opts.openapiDoc.defaultMethod.errors
			tags    = opts.openapiDoc.defaultMethod.tags
		)

		if openapiMethodOpt, ok := opts.openapiDoc.methods[m.name]; ok {
			errors = append(errors, openapiMethodOpt.errors...)
			tags = append(tags, openapiMethodOpt.tags...)
		}

		if opts.jsonRPC.enable {
			o = g.makeJSONRPCPath(opts, m)
			pathStr = "/" + strings.LcFirst(m.name)
			mopt.method.name = "POST"
			for _, name := range errors {
				if ei, ok := opts.mapCodeErrors[name]; ok {
					codeStr := strconv.FormatInt(ei.code, 10)
					o.Responses["x"+codeStr] = openapi.Response{
						Description: name,
						Content: openapi.Content{
							"application/json": {
								Schema: &openapi.Schema{
									Ref: "#/components/schemas/" + name,
								},
							},
						},
					}
				}
			}
		} else {
			o = g.makeRestPath(opts, m)
			pathStr = mopt.path
			for _, regexp := range mopt.pathVars {
				pathStr = stdstrings.Replace(pathStr, ":"+regexp, "", -1)
			}
		}

		o.Tags = tags

		if _, ok := swg.Paths[pathStr]; !ok {
			swg.Paths[pathStr] = &openapi.Path{}
		}

		switch mopt.method.name {
		default:
			swg.Paths[pathStr].Get = o
		case "POST":
			swg.Paths[pathStr].Post = o
		case "PUT":
			swg.Paths[pathStr].Put = o
		case "PATCH":
			swg.Paths[pathStr].Patch = o
		case "DELETE":
			swg.Paths[pathStr].Delete = o
		}
	}

	typeName := "rest"
	if opts.jsonRPC.enable {
		typeName = "jsonrpc"
	}
	output, err := filepath.Abs(filepath.Join(g.w.BasePath(), opts.openapiDoc.output))
	if err != nil {
		return err
	}
	d, _ := ffjson.Marshal(swg)
	if err := ioutil.WriteFile(filepath.Join(output, fmt.Sprintf("openapi_%s.json", typeName)), d, 0755); err != nil {
		return err
	}
	return nil
}

func getOpenapiJSONRPCErrorSchemas() openapi.Schemas {
	return openapi.Schemas{
		"ParseError": {
			Type: "object",
			Properties: openapi.Properties{
				"jsonrpc": &openapi.Schema{
					Type:    "string",
					Example: "2.0",
				},
				"id": &openapi.Schema{
					Type:    "string",
					Example: "1f1ecd1b-d729-40cd-b6f4-4011f69811fe",
				},
				"error": &openapi.Schema{
					Type: "object",
					Properties: openapi.Properties{
						"code": &openapi.Schema{
							Type:    "integer",
							Example: -32700,
						},
						"message": &openapi.Schema{
							Type:    "string",
							Example: "Parse error",
						},
					},
				},
			},
		},
		"InvalidRequestError": {
			Type: "object",
			Properties: openapi.Properties{
				"jsonrpc": &openapi.Schema{
					Type:    "string",
					Example: "2.0",
				},
				"id": &openapi.Schema{
					Type:    "string",
					Example: "1f1ecd1b-d729-40cd-b6f4-4011f69811fe",
				},
				"error": &openapi.Schema{
					Type: "object",
					Properties: openapi.Properties{
						"code": &openapi.Schema{
							Type:    "integer",
							Example: -32600,
						},
						"message": &openapi.Schema{
							Type:    "string",
							Example: "Invalid Request",
						},
					},
				},
			},
		},
		"MethodNotFoundError": {
			Type: "object",
			Properties: openapi.Properties{
				"jsonrpc": &openapi.Schema{
					Type:    "string",
					Example: "2.0",
				},
				"id": &openapi.Schema{
					Type:    "string",
					Example: "1f1ecd1b-d729-40cd-b6f4-4011f69811fe",
				},
				"error": &openapi.Schema{
					Type: "object",
					Properties: openapi.Properties{
						"code": &openapi.Schema{
							Type:    "integer",
							Example: -32601,
						},
						"message": &openapi.Schema{
							Type:    "string",
							Example: "Method not found",
						},
					},
				},
			},
		},
		"InvalidParamsError": {
			Type: "object",
			Properties: openapi.Properties{
				"jsonrpc": &openapi.Schema{
					Type:    "string",
					Example: "2.0",
				},
				"id": &openapi.Schema{
					Type:    "string",
					Example: "1f1ecd1b-d729-40cd-b6f4-4011f69811fe",
				},
				"error": &openapi.Schema{
					Type: "object",
					Properties: openapi.Properties{
						"code": &openapi.Schema{
							Type:    "integer",
							Example: -32602,
						},
						"message": &openapi.Schema{
							Type:    "string",
							Example: "Invalid params",
						},
					},
				},
			},
		},
		"InternalError": {
			Type: "object",
			Properties: openapi.Properties{
				"jsonrpc": &openapi.Schema{
					Type:    "string",
					Example: "2.0",
				},
				"id": &openapi.Schema{
					Type:    "string",
					Example: "1f1ecd1b-d729-40cd-b6f4-4011f69811fe",
				},
				"error": &openapi.Schema{
					Type: "object",
					Properties: openapi.Properties{
						"code": &openapi.Schema{
							Type:    "integer",
							Example: -32603,
						},
						"message": &openapi.Schema{
							Type:    "string",
							Example: "Internal error",
						},
					},
				},
			},
		},
	}
}

func getOpenapiRestErrorSchema() *openapi.Schema {
	return &openapi.Schema{
		Type: "object",
		Properties: openapi.Properties{
			"error": &openapi.Schema{
				Type: "string",
			},
		},
	}
}

func (g *TransportHTTP) writeJSONRPCEndpointCodecMap(opts *transportOptions) {
	var jsonrpcPkg string
	if opts.fastHTTP {
		jsonrpcPkg = g.w.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/fasthttp/jsonrpc")
	} else {
		jsonrpcPkg = g.w.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/http/jsonrpc")
	}
	stringsPkg := g.w.Import("strings", "strings")

	g.w.Write("func Make%sEndpointCodecMap(ep EndpointSet, ns ...string) %s.EndpointCodecMap {\n", g.ctx.id, jsonrpcPkg)

	g.w.Write("var namespace = %s.Join(ns, \".\")\n", stringsPkg)
	g.w.Write("if len(ns) > 0 {\n")
	g.w.Write("namespace += \".\"\n")
	g.w.Write("}\n")

	g.w.Write("return %[1]s.EndpointCodecMap{\n", jsonrpcPkg)

	for _, m := range g.ctx.iface.methods {
		g.writeJSONRPC(opts, m)
	}

	g.w.Write("}\n}\n")
}

func (g *TransportHTTP) writeJSONRPCHandler(opts *transportOptions) {
	var (
		routerPkg  string
		jsonrpcPkg string
	)

	if opts.jsonRPC.enable {
		if opts.fastHTTP {
			jsonrpcPkg = g.w.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/fasthttp/jsonrpc")
		} else {
			jsonrpcPkg = g.w.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/http/jsonrpc")
		}
	}

	if opts.fastHTTP {
		routerPkg = g.w.Import("routing", "github.com/qiangxue/fasthttp-routing")
		g.w.Write("r := %s.New()\n", routerPkg)
	} else {
		routerPkg = g.w.Import("mux", "github.com/gorilla/mux")
		g.w.Write("r := %s.NewRouter()\n", routerPkg)
	}
	g.w.Write("handler := %[1]s.NewServer(Make%sEndpointCodecMap(ep), sopt.genericServerOption...)\n", jsonrpcPkg, g.ctx.id)
	jsonRPCPath := opts.jsonRPC.path
	if opts.fastHTTP {
		r := stdstrings.NewReplacer("{", "<", "}", ">")
		jsonRPCPath = r.Replace(jsonRPCPath)

		g.w.Write("r.Post(\"%s\", func(c *routing.Context) error {\nhandler.ServeFastHTTP(c.RequestCtx)\nreturn nil\n})\n", jsonRPCPath)
	} else {
		g.w.Write("r.Methods(\"POST\").Path(\"%s\").Handler(handler)\n", jsonRPCPath)
	}
	if opts.fastHTTP {
		g.w.Write("return r.HandleRequest, nil")
	} else {
		g.w.Write("return r, nil")
	}
}

func (g *TransportHTTP) writeRESTHandler(opts *transportOptions) {
	var routerPkg string
	if opts.fastHTTP {
		routerPkg = g.w.Import("routing", "github.com/qiangxue/fasthttp-routing")
		g.w.Write("r := %s.New()\n", routerPkg)
	} else {
		routerPkg = g.w.Import("mux", "github.com/gorilla/mux")
		g.w.Write("r := %s.NewRouter()\n", routerPkg)
	}
	for _, m := range g.ctx.iface.methods {
		g.writeHTTPRest(opts, m)
	}
	if opts.fastHTTP {
		g.w.Write("return r.HandleRequest, nil")
	} else {
		g.w.Write("return r, nil")
	}
}

func (g *TransportHTTP) writeJSONRPC(opts *transportOptions, m ifaceServiceMethod) {
	mopt := opts.methodOptions[m.name]

	var (
		jsonrpcPkg string
	)
	if opts.fastHTTP {
		jsonrpcPkg = g.w.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/fasthttp/jsonrpc")
	} else {
		jsonrpcPkg = g.w.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/http/jsonrpc")
	}

	jsonPkg := g.w.Import("json", "encoding/json")
	ffjsonPkg := g.w.Import("ffjson", "github.com/pquerna/ffjson/ffjson")
	contextPkg := g.w.Import("context", "context")

	g.w.Write("namespace+\"%s\": %s.EndpointCodec{\n", m.lcName, jsonrpcPkg)
	g.w.Write(
		"Endpoint: ep.%sEndpoint,\n",
		m.name,
	)
	g.w.Write("Decode: ")

	if mopt.serverRequestFunc.expr != nil {
		g.w.WriteAST(mopt.serverRequestFunc.expr)
	} else {
		fmtPkg := g.w.Import("fmt", "fmt")

		g.w.Write("func(_ %s.Context, msg %s.RawMessage) (interface{}, error) {\n", contextPkg, jsonPkg)

		if len(m.params) > 0 {
			g.w.Write("var req %sRequest%s\n", m.lcName, g.ctx.id)
			g.w.Write("err := %s.Unmarshal(msg, &req)\n", ffjsonPkg)
			g.w.Write("if err != nil {\n")
			g.w.Write("return nil, %s.Errorf(\"couldn't unmarshal body to %sRequest%s: %%s\", err)\n", fmtPkg, m.lcName, g.ctx.id)
			g.w.Write("}\n")
			g.w.Write("return req, nil\n")

		} else {
			g.w.Write("return nil, nil\n")
		}
		g.w.Write("}")
	}

	g.w.Write(",\n")

	g.w.Write("Encode:")

	if mopt.wrapResponse.enable && len(m.results) > 0 {
		jsonPkg := g.w.Import("json", "encoding/json")
		g.w.Write("func (ctx context.Context, response interface{}) (%s.RawMessage, error) {\n", jsonPkg)
		g.w.Write("return encodeResponseJSONRPC%s(ctx, map[string]interface{}{\"%s\": response})\n", g.ctx.id, mopt.wrapResponse.name)
		g.w.Write("},\n")
	} else {
		g.w.Write("encodeResponseJSONRPC%s,\n", g.ctx.id)
	}

	g.w.Write("},\n")
}

func (g *TransportHTTP) writeHTTPRest(opts *transportOptions, m ifaceServiceMethod) {
	var (
		kithttpPkg string
		httpPkg    string
		routerPkg  string
	)
	if opts.fastHTTP {
		kithttpPkg = g.w.Import("fasthttp", "github.com/l-vitaly/go-kit/transport/fasthttp")
		httpPkg = g.w.Import("fasthttp", "github.com/valyala/fasthttp")
		routerPkg = g.w.Import("routing", "github.com/qiangxue/fasthttp-routing")
	} else {
		kithttpPkg = g.w.Import("http", "github.com/go-kit/kit/transport/http")
		httpPkg = g.w.Import("http", "net/http")
		routerPkg = g.w.Import("mux", "github.com/gorilla/mux")
	}

	contextPkg := g.w.Import("context", "context")

	mopt := opts.methodOptions[m.name]

	if opts.fastHTTP {
		g.w.Write("r.To(")

		if mopt.method.name != "" {
			g.w.WriteAST(mopt.method.expr)
		} else {
			g.w.Write(strconv.Quote("GET"))
		}

		g.w.Write(", ")

		if mopt.path != "" {
			// replace brace indices for fasthttp router
			urlPath := stdstrings.ReplaceAll(mopt.path, "{", "<")
			urlPath = stdstrings.ReplaceAll(urlPath, "}", ">")
			g.w.Write(strconv.Quote(urlPath))
		} else {
			g.w.Write(strconv.Quote("/" + m.lcName))
		}
		g.w.Write(", ")
	} else {
		g.w.Write("r.Methods(")
		if mopt.method.name != "" {
			g.w.WriteAST(mopt.method.expr)
		} else {
			g.w.Write(strconv.Quote("GET"))
		}
		g.w.Write(").")
		g.w.Write("Path(")
		if mopt.path != "" {
			g.w.Write(strconv.Quote(mopt.path))
		} else {
			g.w.Write(strconv.Quote("/" + stdstrings.ToLower(m.name)))
		}
		g.w.Write(").")

		g.w.Write("Handler(")
	}

	g.w.Write(
		"%s.NewServer(\nep.%sEndpoint,\n",
		kithttpPkg,
		m.name,
	)

	if mopt.serverRequestFunc.expr != nil {
		g.w.WriteAST(mopt.serverRequestFunc.expr)
	} else {
		g.w.Write("func(ctx %s.Context, r *%s.Request) (interface{}, error) {\n", contextPkg, httpPkg)

		if len(m.params) > 0 {
			g.w.Write("var req %sRequest%s\n", m.lcName, g.ctx.id)
			switch stdstrings.ToUpper(mopt.method.name) {
			case "POST", "PUT", "PATCH":
				fmtPkg := g.w.Import("fmt", "fmt")
				jsonPkg := g.w.Import("ffjson", "github.com/pquerna/ffjson/ffjson")
				pkgIO := g.w.Import("io", "io")

				if opts.fastHTTP {
					g.w.Write("err := %s.Unmarshal(r.Body(), &req)\n", jsonPkg)
				} else {
					ioutilPkg := g.w.Import("ioutil", "io/ioutil")

					g.w.Write("b, err := %s.ReadAll(r.Body)\n", ioutilPkg)
					g.w.WriteCheckErr(func() {
						g.w.Write("return nil, %s.Errorf(\"couldn't read body for %sRequest%s: %%s\", err)\n", fmtPkg, m.lcName, g.ctx.id)
					})
					g.w.Write("err = %s.Unmarshal(b, &req)\n", jsonPkg)
				}

				g.w.Write("if err != nil && err != %s.EOF {\n", pkgIO)
				g.w.Write("return nil, %s.Errorf(\"couldn't unmarshal body to %sRequest%s: %%s\", err)\n", fmtPkg, m.lcName, g.ctx.id)
				g.w.Write("}\n")
			}
			if len(mopt.pathVars) > 0 {
				if opts.fastHTTP {
					fmtPkg := g.w.Import("fmt", "fmt")

					g.w.Write("vars, ok := ctx.Value(%s.ContextKeyRouter).(*%s.Context)\n", kithttpPkg, routerPkg)
					g.w.Write("if !ok {\n")
					g.w.Write("return nil, %s.Errorf(\"couldn't assert %s.ContextKeyRouter to *%s.Context\")\n", fmtPkg, kithttpPkg, routerPkg)
					g.w.Write("}\n")
				} else {
					g.w.Write("vars := %s.Vars(r)\n", routerPkg)
				}
				for pathVarName := range mopt.pathVars {
					if f := m.params.lookupField(pathVarName); f != nil {
						var valueID string
						if opts.fastHTTP {
							valueID = "vars.Param(" + strconv.Quote(pathVarName) + ")"
						} else {
							valueID = "vars[" + strconv.Quote(pathVarName) + "]"
						}
						g.w.WriteConvertType("req."+strings.UcFirst(f.Name()), valueID, f, "", false, "")
					}
				}
			}
			if len(mopt.queryVars) > 0 {
				if opts.fastHTTP {
					g.w.Write("q := r.URI().QueryArgs()\n")
				} else {
					g.w.Write("q := r.URL.Query()\n")
				}
				for argName, queryName := range mopt.queryVars {
					if f := m.params.lookupField(argName); f != nil {
						var valueID string
						if opts.fastHTTP {
							valueID = "string(q.Peek(" + strconv.Quote(queryName) + "))"
						} else {
							valueID = "q.Get(" + strconv.Quote(queryName) + ")"
						}
						g.w.WriteConvertType("req."+strings.UcFirst(f.Name()), valueID, f, "", false, "")
					}
				}
			}
			for argName, headerName := range mopt.headerVars {
				if f := m.params.lookupField(argName); f != nil {
					var valueID string
					if opts.fastHTTP {
						valueID = "string(r.Header.Peek(" + strconv.Quote(headerName) + "))"
					} else {
						valueID = "r.Header.Get(" + strconv.Quote(headerName) + ")"
					}
					g.w.WriteConvertType("req."+strings.UcFirst(f.Name()), valueID, f, "", false, "")
				}
			}
			g.w.Write("return req, nil\n")
		} else {
			g.w.Write("return nil, nil\n")
		}
		g.w.Write("},\n")
	}
	if mopt.serverResponseFunc.expr != nil {
		g.w.WriteAST(mopt.serverResponseFunc.expr)
	} else {
		if opts.jsonRPC.enable {
			g.w.Write("encodeResponseJSONRPC%s", g.ctx.id)
		} else {
			if mopt.wrapResponse.enable {
				var responseWriterType string
				if opts.fastHTTP {
					responseWriterType = fmt.Sprintf("*%s.Response", httpPkg)
				} else {
					responseWriterType = fmt.Sprintf("%s.ResponseWriter", httpPkg)
				}
				g.w.Write("func (ctx context.Context, w %s, response interface{}) error {\n", responseWriterType)
				g.w.Write("return encodeResponseHTTP%s(ctx, w, map[string]interface{}{\"%s\": response})\n", g.ctx.id, mopt.wrapResponse.name)
				g.w.Write("}")
			} else {
				g.w.Write("encodeResponseHTTP%s", g.ctx.id)
			}
		}
	}

	g.w.Write(",\n")

	g.w.Write("append(sopt.genericServerOption, sopt.%sServerOption...)...,\n", m.lcName)
	g.w.Write(")")

	if opts.fastHTTP {
		g.w.Write(".RouterHandle()")
	}

	g.w.Write(")\n")
}

func (g *TransportHTTP) writeMiddlewares(opts *transportOptions) {
	if g.ctx.logging {
		g.writeLoggingMiddleware()
	}
	if g.ctx.instrumenting.enable {
		g.writeInstrumentingMiddleware()
	}
}

func (g *TransportHTTP) writeLoggingMiddleware() {
	g.w.Write("s = &loggingMiddleware%s{next: s, logger: logger}\n", g.ctx.id)
}

func (g *TransportHTTP) writeInstrumentingMiddleware() {
	stdPrometheusPkg := g.w.Import("prometheus", "github.com/prometheus/client_golang/prometheus")
	kitPrometheusPkg := g.w.Import("prometheus", "github.com/go-kit/kit/metrics/prometheus")

	g.w.Write("s = &instrumentingMiddleware%s{\nnext: s,\n", g.ctx.id)
	g.w.Write("requestCount: %s.NewCounterFrom(%s.CounterOpts{\n", kitPrometheusPkg, stdPrometheusPkg)
	g.w.Write("Namespace: %s,\n", strconv.Quote(g.ctx.instrumenting.namespace))
	g.w.Write("Subsystem: %s,\n", strconv.Quote(g.ctx.instrumenting.subsystem))
	g.w.Write("Name: %s,\n", strconv.Quote("request_count"))
	g.w.Write("Help: %s,\n", strconv.Quote("Number of requests received."))
	g.w.Write("}, []string{\"method\"}),\n")

	g.w.Write("requestLatency: %s.NewSummaryFrom(%s.SummaryOpts{\n", kitPrometheusPkg, stdPrometheusPkg)
	g.w.Write("Namespace: %s,\n", strconv.Quote(g.ctx.instrumenting.namespace))
	g.w.Write("Subsystem: %s,\n", strconv.Quote(g.ctx.instrumenting.subsystem))
	g.w.Write("Name: %s,\n", strconv.Quote("request_latency_microseconds"))
	g.w.Write("Help: %s,\n", strconv.Quote("Total duration of requests in microseconds."))
	g.w.Write("}, []string{\"method\"}),\n")
	g.w.Write("}\n")
}

func (g *TransportHTTP) writeClientStructOptions(opts *transportOptions) {
	var (
		kithttpPkg string
	)
	endpointPkg := g.w.Import("endpoint", "github.com/go-kit/kit/endpoint")
	if opts.jsonRPC.enable {
		if opts.fastHTTP {
			kithttpPkg = g.w.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/fasthttp/jsonrpc")
		} else {
			kithttpPkg = g.w.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/http/jsonrpc")
		}
	} else {
		if opts.fastHTTP {
			kithttpPkg = g.w.Import("fasthttp", "github.com/l-vitaly/go-kit/transport/fasthttp")
		} else {
			kithttpPkg = g.w.Import("http", "github.com/go-kit/kit/transport/http")
		}
	}

	clientType := "client" + g.ctx.id

	g.w.Write("type %[1]sOption func(*%[1]s)\n", clientType)

	g.w.WriteFunc(
		g.ctx.id+"GenericClientOptions",
		"",
		[]string{"opt", "..." + kithttpPkg + ".ClientOption"},
		[]string{"", clientType + "Option"},
		func() {
			g.w.Write("return func(c *%s) { c.genericClientOption = opt }\n", clientType)
		},
	)

	g.w.WriteFunc(
		g.ctx.id+"GenericClientEndpointMiddlewares",
		"",
		[]string{"opt", "..." + endpointPkg + ".Middleware"},
		[]string{"", clientType + "Option"},
		func() {
			g.w.Write("return func(c *%s) { c.genericEndpointMiddleware = opt }\n", clientType)
		},
	)

	for _, m := range g.ctx.iface.methods {
		g.w.WriteFunc(
			g.ctx.id+m.name+"ClientOptions",
			"",
			[]string{"opt", "..." + kithttpPkg + ".ClientOption"},
			[]string{"", clientType + "Option"},
			func() {
				g.w.Write("return func(c *%s) { c.%sClientOption = opt }\n", clientType, m.lcName)
			},
		)

		g.w.WriteFunc(
			g.ctx.id+m.name+"ClientEndpointMiddlewares",
			"",
			[]string{"opt", "..." + endpointPkg + ".Middleware"},
			[]string{"", clientType + "Option"},
			func() {
				g.w.Write("return func(c *%s) { c.%sEndpointMiddleware = opt }\n", clientType, m.lcName)
			},
		)
	}
}

func (g *TransportHTTP) writeClientStruct(opts *transportOptions) {
	var (
		kithttpPkg string
	)
	if opts.jsonRPC.enable {
		if opts.fastHTTP {
			kithttpPkg = g.w.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/fasthttp/jsonrpc")
		} else {
			kithttpPkg = g.w.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/http/jsonrpc")
		}
	} else {
		if opts.fastHTTP {
			kithttpPkg = g.w.Import("fasthttp", "github.com/l-vitaly/go-kit/transport/fasthttp")
		} else {
			kithttpPkg = g.w.Import("http", "github.com/go-kit/kit/transport/http")
		}
	}

	endpointPkg := g.w.Import("endpoint", "github.com/go-kit/kit/endpoint")
	contextPkg := g.w.Import("context", "context")

	clientType := "client" + g.ctx.id

	g.w.Write("type %s struct {\n", clientType)
	for _, m := range g.ctx.iface.methods {
		g.w.Write("%sEndpoint %s.Endpoint\n", m.lcName, endpointPkg)
		g.w.Write("%sClientOption []%s.ClientOption\n", m.lcName, kithttpPkg)
		g.w.Write("%sEndpointMiddleware []%s.Middleware\n", m.lcName, endpointPkg)
	}
	g.w.Write("genericClientOption []%s.ClientOption\n", kithttpPkg)
	g.w.Write("genericEndpointMiddleware []%s.Middleware\n", endpointPkg)

	g.w.Write("}\n\n")

	g.writeClientStructOptions(opts)

	for _, m := range g.ctx.iface.methods {
		var params []string

		if m.paramCtx != nil {
			params = append(params, m.paramCtx.Name(), g.w.TypeString(m.paramCtx.Type()))
		}

		params = append(params, utils.NameTypeParams(m.params, g.w.TypeString, nil)...)
		results := utils.NameType(m.results, g.w.TypeString, nil)

		if m.returnErr != nil {
			results = append(results, "", "error")
		}

		g.w.WriteFunc(m.name, "c *"+clientType, params, results, func() {
			if len(m.results) > 0 {
				g.w.Write("resp")
			} else {
				g.w.Write("_")
			}
			g.w.Write(", err := ")

			g.w.Write("c.%sEndpoint(", m.lcName)

			if m.paramCtx != nil {
				g.w.Write("%s,", m.paramCtx.Name())
			} else {
				g.w.Write("%s.Background(),", contextPkg)
			}

			if len(m.params) > 0 {
				g.w.Write("%sRequest%s", m.lcName, g.ctx.id)
				params := structKeyValue(m.params, func(p *stdtypes.Var) bool {
					return !types.IsContext(p.Type())
				})
				g.w.WriteStructAssign(params)
			} else {
				g.w.Write(" nil")
			}

			g.w.Write(")\n")

			if m.returnErr != nil {
				g.w.Write("if err != nil {\n")
				g.w.Write("return ")

				if len(m.results) > 0 {
					for i, r := range m.results {
						if i > 0 {
							g.w.Write(",")
						}
						g.w.Write(g.w.ZeroValue(r.Type()))
					}
					g.w.Write(",")
				}

				g.w.Write(" err\n")

				g.w.Write("}\n")
			}

			if len(m.results) > 0 {
				if m.resultsNamed {
					g.w.Write("response := resp.(%sResponse%s)\n", m.lcName, g.ctx.id)
				} else {
					g.w.Write("response := resp.(%s)\n", g.w.TypeString(m.results[0].Type()))
				}
			}

			g.w.Write("return ")

			if len(m.results) > 0 {
				if m.resultsNamed {
					for i, r := range m.results {
						if i > 0 {
							g.w.Write(",")
						}
						g.w.Write("response.%s", strings.UcFirst(r.Name()))
					}
				} else {
					g.w.Write("response")
				}
				g.w.Write(", ")
			}
			if m.returnErr != nil {
				g.w.Write("nil")
			}
			g.w.Write("\n")
		})
	}
}

func (g *TransportHTTP) writeRestClient(opts *transportOptions) {
	var (
		kithttpPkg string
		httpPkg    string
	)
	if opts.fastHTTP {
		kithttpPkg = g.w.Import("fasthttp", "github.com/l-vitaly/go-kit/transport/fasthttp")
	} else {
		kithttpPkg = g.w.Import("http", "github.com/go-kit/kit/transport/http")
	}
	if opts.fastHTTP {
		httpPkg = g.w.Import("fasthttp", "github.com/valyala/fasthttp")
	} else {
		httpPkg = g.w.Import("http", "net/http")
	}
	jsonPkg := g.w.Import("ffjson", "github.com/pquerna/ffjson/ffjson")
	pkgIO := g.w.Import("io", "io")
	fmtPkg := g.w.Import("fmt", "fmt")
	contextPkg := g.w.Import("context", "context")
	urlPkg := g.w.Import("url", "net/url")

	g.w.Write("u, err := %s.Parse(tgt)\n", urlPkg)

	g.w.WriteCheckErr(func() {
		g.w.Write("return nil, err")
	})

	for _, m := range g.ctx.iface.methods {
		epName := m.lcName + "Endpoint"

		mopt := opts.methodOptions[m.name]

		httpMethod := mopt.method.name
		if httpMethod == "" {
			if len(m.params) > 0 {
				httpMethod = "POST"
			} else {
				httpMethod = "GET"
			}
		}

		pathStr := mopt.path
		if pathStr == "" {
			pathStr = "/" + m.lcName
		}

		pathVars := []string{}
		for name, regexp := range mopt.pathVars {
			if p := m.params.lookupField(name); p != nil {
				if regexp != "" {
					regexp = ":" + regexp
				}
				pathStr = stdstrings.Replace(pathStr, "{"+name+regexp+"}", "%s", -1)
				pathVars = append(pathVars, g.w.GetFormatType("req."+strings.UcFirst(p.Name()), p))
			}
		}
		queryVars := []string{}
		for fName, qName := range mopt.queryVars {
			if p := m.params.lookupField(fName); p != nil {
				queryVars = append(queryVars, strconv.Quote(qName), g.w.GetFormatType("req."+strings.UcFirst(p.Name()), p))
			}
		}
		headerVars := []string{}
		for fName, hName := range mopt.headerVars {
			if p := m.params.lookupField(fName); p != nil {
				headerVars = append(headerVars, strconv.Quote(hName), g.w.GetFormatType("req."+strings.UcFirst(p.Name()), p))
			}
		}

		g.w.Write("c.%s = %s.NewClient(\n", epName, kithttpPkg)
		if mopt.method.expr != nil {
			g.w.WriteAST(mopt.method.expr)
		} else {
			g.w.Write(strconv.Quote(httpMethod))
		}
		g.w.Write(",\n")
		g.w.Write("u,\n")

		if mopt.clientRequestFunc.expr != nil {
			g.w.WriteAST(mopt.clientRequestFunc.expr)
		} else {
			g.w.Write("func(_ %s.Context, r *%s.Request, request interface{}) error {\n", contextPkg, httpPkg)

			if len(m.params) > 0 {
				g.w.Write("req, ok := request.(%sRequest%s)\n", m.lcName, g.ctx.id)
				g.w.Write("if !ok {\n")
				g.w.Write("return %s.Errorf(\"couldn't assert request as %sRequest%s, got %%T\", request)\n", fmtPkg, m.lcName, g.ctx.id)
				g.w.Write("}\n")
			}

			if opts.fastHTTP {
				g.w.Write("r.Header.SetMethod(")
			} else {
				g.w.Write("r.Method = ")
			}
			if mopt.method.expr != nil {
				g.w.WriteAST(mopt.method.expr)
			} else {
				g.w.Write(strconv.Quote(httpMethod))
			}
			if opts.fastHTTP {
				g.w.Write(")")
			}
			g.w.Write("\n")

			if opts.fastHTTP {
				g.w.Write("r.SetRequestURI(")
			} else {
				g.w.Write("r.URL.Path += ")
			}
			g.w.Write("%s.Sprintf(%s, %s)", fmtPkg, strconv.Quote(pathStr), stdstrings.Join(pathVars, ","))

			if opts.fastHTTP {
				g.w.Write(")")
			}
			g.w.Write("\n")

			if len(queryVars) > 0 {
				if opts.fastHTTP {
					g.w.Write("q := r.URI().QueryArgs()\n")
				} else {
					g.w.Write("q := r.URL.Query()\n")
				}

				for i := 0; i < len(queryVars); i += 2 {
					g.w.Write("q.Add(%s, %s)\n", queryVars[i], queryVars[i+1])
				}

				if opts.fastHTTP {
					g.w.Write("r.URI().SetQueryString(q.String())\n")
				} else {
					g.w.Write("r.URL.RawQuery = q.Encode()\n")
				}
			}

			for i := 0; i < len(headerVars); i += 2 {
				g.w.Write("r.Header.Add(%s, %s)\n", headerVars[i], headerVars[i+1])
			}

			switch stdstrings.ToUpper(httpMethod) {
			case "POST", "PUT", "PATCH":
				jsonPkg := g.w.Import("ffjson", "github.com/pquerna/ffjson/ffjson")

				g.w.Write("data, err := %s.Marshal(req)\n", jsonPkg)
				g.w.Write("if err != nil  {\n")
				g.w.Write("return %s.Errorf(\"couldn't marshal request %%T: %%s\", req, err)\n", fmtPkg)
				g.w.Write("}\n")

				if opts.fastHTTP {
					g.w.Write("r.SetBody(data)\n")
				} else {
					ioutilPkg := g.w.Import("ioutil", "io/ioutil")
					bytesPkg := g.w.Import("bytes", "bytes")

					g.w.Write("r.Body = %s.NopCloser(%s.NewBuffer(data))\n", ioutilPkg, bytesPkg)
				}
			}
			g.w.Write("return nil\n")
			g.w.Write("}")
		}
		g.w.Write(",\n")

		if mopt.clientResponseFunc.expr != nil {
			g.w.WriteAST(mopt.clientResponseFunc.expr)
		} else {
			g.w.Write("func(_ %s.Context, r *%s.Response) (interface{}, error) {\n", contextPkg, httpPkg)

			statusCode := "r.StatusCode"
			if opts.fastHTTP {
				statusCode = "r.StatusCode()"
			}

			g.w.Write("if statusCode := %s; statusCode != %s.StatusOK {\n", statusCode, httpPkg)
			g.w.Write("return nil, ErrorDecode(statusCode)\n")
			g.w.Write("}\n")

			if len(m.results) > 0 {
				var responseType string
				if m.resultsNamed {
					responseType = fmt.Sprintf("%sResponse%s", m.lcName, g.ctx.id)
				} else {
					responseType = g.w.TypeString(m.results[0].Type())
				}
				if mopt.wrapResponse.enable {
					g.w.Write("var resp struct {\nData %s `json:\"%s\"`\n}\n", responseType, mopt.wrapResponse.name)
				} else {
					g.w.Write("var resp %s\n", responseType)
				}
				if opts.fastHTTP {
					g.w.Write("err := %s.Unmarshal(r.Body(), ", jsonPkg)
				} else {
					ioutilPkg := g.w.Import("ioutil", "io/ioutil")

					g.w.Write("b, err := %s.ReadAll(r.Body)\n", ioutilPkg)
					g.w.WriteCheckErr(func() {
						g.w.Write("return nil, err\n")
					})
					g.w.Write("err = %s.Unmarshal(b, ", jsonPkg)
				}

				g.w.Write("&resp)\n")

				g.w.Write("if err != nil && err != %s.EOF {\n", pkgIO)
				g.w.Write("return nil, %s.Errorf(\"couldn't unmarshal body to %sResponse%s: %%s\", err)\n", fmtPkg, m.lcName, g.ctx.id)
				g.w.Write("}\n")

				if mopt.wrapResponse.enable {
					g.w.Write("return resp.Data, nil\n")
				} else {
					g.w.Write("return resp, nil\n")
				}
			} else {
				g.w.Write("return nil, nil\n")
			}

			g.w.Write("}")
		}

		g.w.Write(",\n")

		g.w.Write("append(c.genericClientOption, c.%sClientOption...)...,\n", m.lcName)

		g.w.Write(").Endpoint()\n")

		g.w.Write(
			"c.%[1]sEndpoint = middlewareChain(append(c.genericEndpointMiddleware, c.%[1]sEndpointMiddleware...))(c.%[1]sEndpoint)\n",
			m.lcName,
		)
	}
}

func (g *TransportHTTP) writeJsonRPCClientJS(opts *transportOptions) {
	w := new(bytes.Buffer)

	w.WriteString("// Code generated by Swipe " + g.w.SwipeVersion() + ". DO NOT EDIT.\n\n")

	w.WriteString(jsonRPCClientBase)
	fmt.Fprintf(w, "export default class extends JSONRPCClient {\n")

	for _, m := range g.ctx.iface.methods {
		mopt := opts.methodOptions[m.name]

		fmt.Fprint(w, "/**\n")
		for _, p := range m.params {
			fmt.Fprintf(w, "* @param {%s} %s\n", g.getJSDocType(p.Type()), p.Name())
		}

		if len(m.results) > 0 {
			fmt.Fprintf(w, "* @return {PromiseLike<")
			if m.resultsNamed {
				if mopt.wrapResponse.enable {
					fmt.Fprintf(w, "{%s: ", mopt.wrapResponse.name)
				} else {
					fmt.Fprint(w, "{")
				}
			}

			for i, p := range m.results {
				if i > 0 {
					fmt.Fprintf(w, ", ")
				}
				if p.Name() != "" {
					fmt.Fprintf(w, "%s: ", p.Name())
				}
				fmt.Fprint(w, g.getJSDocType(p.Type()))
			}
			if m.resultsNamed || mopt.wrapResponse.enable {
				fmt.Fprintf(w, "}")
			}
			fmt.Fprintf(w, ">}\n")
		}

		fmt.Fprint(w, "**/\n")
		fmt.Fprintf(w, "%s(", m.lcName)

		for i, p := range m.params {
			if i > 0 {
				fmt.Fprintf(w, ",")
			}
			fmt.Fprint(w, p.Name())
		}

		fmt.Fprintf(w, ") {\n")
		fmt.Fprintf(w, "return this.__scheduleRequest(\"%s\", {", m.lcName)

		for i, p := range m.params {
			if i > 0 {
				fmt.Fprintf(w, ",")
			}
			fmt.Fprintf(w, "%[1]s:%[1]s", p.Name())
		}

		fmt.Fprintf(w, "})\n")
		fmt.Fprintf(w, "}\n")
	}

	fmt.Fprintf(w, "}\n")

	for name, e := range opts.mapCodeErrors {
		fmt.Fprintf(w, "export class %[1]sError extends JSONRPCError {\nconstructor(message, data) {\nsuper(message, \"%[1]sError\", %d, data);\n}\n}\n", name, e.code)
	}
	fmt.Fprintf(w, "function convertError(e) {\n")
	fmt.Fprintf(w, "switch(e.code) {\n")
	fmt.Fprintf(w, "default:\n")
	fmt.Fprintf(w, "return new JSONRPCError(e.message, \"UnknownError\", e.code, e.data);\n")
	for name, e := range opts.mapCodeErrors {
		fmt.Fprintf(w, "case %d:\n", e.code)
		fmt.Fprintf(w, "return new %sError(e.message, e.data);\n", name)
	}
	fmt.Fprintf(w, "}\n}\n")

	_ = os.MkdirAll(g.w.BasePath()+"/jsclient", 0777)

	_ = ioutil.WriteFile(g.w.BasePath()+"/jsclient/index.js", w.Bytes(), 0755)

	err := exec.Command("prettier", "--write", g.w.BasePath()+"/jsclient/index.js").Run()
	if err != nil {
		fmt.Println(err)
	}
}

func (g *TransportHTTP) writeJsonRPCClientGo(opts *transportOptions) {
	var (
		jsonrpcPkg string
	)
	if opts.fastHTTP {
		jsonrpcPkg = g.w.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/fasthttp/jsonrpc")
	} else {
		jsonrpcPkg = g.w.Import("jsonrpc", "github.com/l-vitaly/go-kit/transport/http/jsonrpc")
	}

	urlPkg := g.w.Import("url", "net/url")
	contextPkg := g.w.Import("context", "context")
	ffjsonPkg := g.w.Import("ffjson", "github.com/pquerna/ffjson/ffjson")
	jsonPkg := g.w.Import("json", "encoding/json")
	fmtPkg := g.w.Import("fmt", "fmt")

	g.w.Write("u, err := %s.Parse(tgt)\n", urlPkg)

	g.w.WriteCheckErr(func() {
		g.w.Write("return nil, err")
	})

	for _, m := range g.ctx.iface.methods {
		mopt := opts.methodOptions[m.name]

		g.w.Write("c.%[1]sClientOption = append(\nc.%[1]sClientOption,\n", m.lcName)

		g.w.Write("%s.ClientRequestEncoder(", jsonrpcPkg)
		g.w.Write("func(_ %s.Context, obj interface{}) (%s.RawMessage, error) {\n", contextPkg, jsonPkg)

		if len(m.params) > 0 {
			g.w.Write("req, ok := obj.(%sRequest%s)\n", m.lcName, g.ctx.id)
			g.w.Write("if !ok {\n")
			g.w.Write("return nil, %s.Errorf(\"couldn't assert request as %sRequest%s, got %%T\", obj)\n", fmtPkg, m.lcName, g.ctx.id)
			g.w.Write("}\n")
			g.w.Write("b, err := %s.Marshal(req)\n", ffjsonPkg)
			g.w.Write("if err != nil {\n")
			g.w.Write("return nil, %s.Errorf(\"couldn't marshal request %%T: %%s\", obj, err)\n", fmtPkg)
			g.w.Write("}\n")
			g.w.Write("return b, nil\n")
		} else {
			g.w.Write("return nil, nil\n")
		}
		g.w.Write("}),\n")

		g.w.Write("%s.ClientResponseDecoder(", jsonrpcPkg)
		g.w.Write("func(_ %s.Context, response %s.Response) (interface{}, error) {\n", contextPkg, jsonrpcPkg)
		g.w.Write("if response.Error != nil {\n")
		g.w.Write("return nil, ErrorDecode(response.Error.Code)\n")
		g.w.Write("}\n")

		if len(m.results) > 0 {
			var responseType string
			if m.resultsNamed {
				responseType = fmt.Sprintf("%sResponse%s", m.lcName, g.ctx.id)
			} else {
				responseType = g.w.TypeString(m.results[0].Type())
			}

			if mopt.wrapResponse.enable {
				g.w.Write("var resp struct {\n Data %s `json:\"%s\"`\n}\n", responseType, mopt.wrapResponse.name)
			} else {
				g.w.Write("var resp %s\n", responseType)
			}

			g.w.Write("err := %s.Unmarshal(response.Result, &resp)\n", ffjsonPkg)
			g.w.Write("if err != nil {\n")
			g.w.Write("return nil, %s.Errorf(\"couldn't unmarshal body to %sResponse%s: %%s\", err)\n", fmtPkg, m.lcName, g.ctx.id)
			g.w.Write("}\n")

			if mopt.wrapResponse.enable {
				g.w.Write("return resp.Data, nil\n")
			} else {
				g.w.Write("return resp, nil\n")
			}
		} else {
			g.w.Write("return nil, nil\n")
		}

		g.w.Write("}),\n")

		g.w.Write(")\n")

		g.w.Write("c.%sEndpoint = %s.NewClient(\n", m.lcName, jsonrpcPkg)
		g.w.Write("u,\n")
		g.w.Write("%s,\n", strconv.Quote(m.lcName))

		g.w.Write("append(c.genericClientOption, c.%sClientOption...)...,\n", m.lcName)

		g.w.Write(").Endpoint()\n")

		g.w.Write(
			"c.%[1]sEndpoint = middlewareChain(append(c.genericEndpointMiddleware, c.%[1]sEndpointMiddleware...))(c.%[1]sEndpoint)\n",
			m.lcName,
		)
	}
}

func (g *TransportHTTP) getJSDocType(t stdtypes.Type) string {
	switch v := t.(type) {
	default:
		return "*"
	case *stdtypes.Pointer:
		return g.getJSDocType(v.Elem())
	case *stdtypes.Array:
		return fmt.Sprintf("Array.<%s>", g.getJSDocType(v.Elem()))
	case *stdtypes.Slice:
		return fmt.Sprintf("Array.<%s>", g.getJSDocType(v.Elem()))
	case *stdtypes.Map:
		return fmt.Sprintf("Object.<%s, %s>", g.getJSDocType(v.Key()), g.getJSDocType(v.Elem()))
	case *stdtypes.Named:
		switch stdtypes.TypeString(v.Obj().Type(), nil) {
		case "github.com/pborman/uuid.UUID",
			"github.com/google/uuid.UUID":
			return "string"
		case "time.Time":
			return "string"
		}
		return g.getJSDocType(v.Obj().Type().Underlying())
	case *stdtypes.Struct:
		buf := new(bytes.Buffer)
		fmt.Fprintf(buf, "{")
		for i := 0; i < v.NumFields(); i++ {
			f := v.Field(i)
			if i > 0 {
				fmt.Fprint(buf, ", ")
			}
			fmt.Fprintf(buf, "%s: %s", strcase.ToLowerCamel(f.Name()), g.getJSDocType(f.Type()))
		}
		fmt.Fprintf(buf, "}")
		return buf.String()
	case *stdtypes.Basic:
		switch v.Kind() {
		default:
			return "string"
		case stdtypes.Bool:
			return "boolean"
		case stdtypes.Float32,
			stdtypes.Float64,
			stdtypes.Int,
			stdtypes.Int8,
			stdtypes.Int16,
			stdtypes.Int32,
			stdtypes.Int64,
			stdtypes.Uint,
			stdtypes.Uint8,
			stdtypes.Uint16,
			stdtypes.Uint32,
			stdtypes.Uint64:
			return "number"
		}
	}
}

func (g *TransportHTTP) makeSwaggerSchema(t stdtypes.Type) (schema *openapi.Schema) {
	schema = &openapi.Schema{}
	switch v := t.(type) {
	case *stdtypes.Pointer:
		return g.makeSwaggerSchema(v.Elem())
	case *stdtypes.Interface:
		// TODO: not anyOf works in SwaggerUI, so the object type is used to display the field.
		schema.Type = "object"
		schema.Description = "Can be any value - string, number, boolean, array or object."
		schema.Properties = openapi.Properties{}
		schema.Example = json.RawMessage("null")
		schema.AnyOf = []openapi.Schema{
			{Type: "string", Example: "abc"},
			{Type: "integer", Example: 1},
			{Type: "number", Format: "float", Example: 1.11},
			{Type: "boolean", Example: true},
			{Type: "array"},
			{Type: "object"},
		}
	case *stdtypes.Map:
		schema.Type = "object"
		schema.Properties = openapi.Properties{}
	case *stdtypes.Slice:
		if vv, ok := v.Elem().(*stdtypes.Basic); ok && vv.Kind() == stdtypes.Byte {
			schema.Type = "string"
			schema.Format = "byte"
			schema.Example = "U3dhZ2dlciByb2Nrcw=="
		} else {
			schema.Type = "array"
			schema.Items = g.makeSwaggerSchema(v.Elem())
		}
	case *stdtypes.Basic:
		switch v.Kind() {
		case stdtypes.String:
			schema.Type = "string"
			schema.Format = "string"
			schema.Example = "abc"
		case stdtypes.Bool:
			schema.Type = "boolean"
			schema.Example = true
		case stdtypes.Int,
			stdtypes.Uint,
			stdtypes.Uint8,
			stdtypes.Uint16,
			stdtypes.Int8,
			stdtypes.Int16:
			schema.Type = "integer"
			schema.Example = 1
		case stdtypes.Uint32, stdtypes.Int32:
			schema.Type = "integer"
			schema.Format = "int32"
			schema.Example = 1
		case stdtypes.Uint64, stdtypes.Int64:
			schema.Type = "integer"
			schema.Format = "int64"
			schema.Example = 1
		case stdtypes.Float32, stdtypes.Float64:
			schema.Type = "number"
			schema.Format = "float"
			schema.Example = 1.11
		}
	case *stdtypes.Struct:
		schema.Type = "object"
		schema.Properties = openapi.Properties{}

		for i := 0; i < v.NumFields(); i++ {
			f := v.Field(i)
			schema.Properties[strcase.ToLowerCamel(f.Name())] = g.makeSwaggerSchema(f.Type())
		}
	case *stdtypes.Named:
		switch stdtypes.TypeString(v, nil) {
		case "encoding/json.RawMessage":
			schema.Type = "object"
			schema.Properties = openapi.Properties{}
			return
		case "time.Time":
			schema.Type = "string"
			schema.Format = "date-time"
			schema.Example = "1985-02-04T00:00:00.00Z"
			return
		case "github.com/pborman/uuid.UUID",
			"github.com/google/uuid.UUID":
			schema.Type = "string"
			schema.Format = "uuid"
			schema.Example = "d5c02d83-6fbc-4dd7-8416-9f85ed80de46"
			return
		}
		return g.makeSwaggerSchema(v.Obj().Type().Underlying())
	}
	return
}

func newTransportHTTP(ctx serviceCtx, w *writer.Writer) *TransportHTTP {
	return &TransportHTTP{ctx: ctx, w: w}
}
