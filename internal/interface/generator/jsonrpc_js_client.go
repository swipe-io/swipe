package generator

import (
	"context"
	stdtypes "go/types"
	"strconv"

	"github.com/gogo/protobuf/sortkeys"
	"github.com/swipe-io/swipe/v2/internal/domain/model"
	"github.com/swipe-io/swipe/v2/internal/interface/typevisitor"
	"github.com/swipe-io/swipe/v2/internal/usecase/generator"
	"github.com/swipe-io/swipe/v2/internal/writer"

	"golang.org/x/tools/go/types/typeutil"
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
			for (let i = 0; i < responses.length; i++) {
			  if (responses[i].error) {
				scheduleRequests[responses[i].id].reject(convertError(responses[i].error));
				continue;
			  }
			  scheduleRequests[responses[i].id].resolve(responses[i].result);
			}
		  })
         .catch((e) => {
           for (let key in requests) {
             if (!requests.hasOwnProperty(key)) {
               continue;
             }
             if (scheduleRequests.hasOwnProperty(requests[key].id)) {
               scheduleRequests[requests[key].id].reject(e)
             }
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
	  const p = new Promise((resolve, reject) => {
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

type jsonRPCJSClient struct {
	writer.BaseWriter
	serviceMethods []model.ServiceMethod
	transport      model.TransportOption
	enums          *typeutil.Map
}

func (g *jsonRPCJSClient) Prepare(_ context.Context) error {
	return nil
}

func (g *jsonRPCJSClient) Process(_ context.Context) error {
	g.W(jsonRPCClientBase)

	mw := writer.BaseWriter{}

	tdc := typevisitor.NewNamedTypeCollector()

	for _, m := range g.serviceMethods {
		mopt := g.transport.MethodOptions[m.Name]
		mw.W("/**\n")

		if len(m.Comments) > 0 {
			for _, comment := range m.Comments {
				mw.W("* %s\n", comment)
			}
			mw.W("*\n")
		}

		for _, p := range m.Params {
			buf := new(writer.BaseWriter)
			typevisitor.JSTypeVisitor(buf).Visit(p.Type())

			tdc.Visit(p.Type())

			mw.W("* @param {%s} %s\n", buf.String(), p.Name())
		}

		if len(m.Results) > 0 {
			mw.W("* @return {PromiseLike<")
			if m.ResultsNamed {
				if mopt.WrapResponse.Enable {
					mw.W("{%s: ", mopt.WrapResponse.Name)
				} else {
					mw.W("{")
				}
			}
			for i, p := range m.Results {
				if i > 0 {
					mw.W(", ")
				}
				if m.ResultsNamed {
					mw.W("%s: ", p.Name())
				}

				buf := new(writer.BaseWriter)
				typevisitor.JSTypeVisitor(buf).Visit(p.Type())

				tdc.Visit(p.Type())

				mw.W(buf.String())
			}
			if m.ResultsNamed || mopt.WrapResponse.Enable {
				mw.W("}")
			}
			mw.W(">}\n")
		}

		mw.W("**/\n")
		mw.W("%s(", m.LcName)

		for i, p := range m.Params {
			if i > 0 {
				mw.W(",")
			}
			mw.W(p.Name())
		}

		mw.W(") {\n")
		mw.W("return this.__scheduleRequest(\"%s\", {", m.LcName)

		for i, p := range m.Params {
			if i > 0 {
				mw.W(",")
			}
			mw.W("%[1]s:%[1]s", p.Name())
		}

		mw.W("})\n")
		mw.W("}\n")
	}

	buf := new(writer.BaseWriter)

	for _, t := range tdc.TypeDefs() {
		typevisitor.JSTypeDefVisitor(buf).Visit(t)
	}

	g.W(buf.String())

	g.W("export default class extends JSONRPCClient {\n")
	g.W(mw.String())
	g.W("}\n")

	errorKeys := make([]uint32, 0, len(g.transport.Errors))
	for key := range g.transport.Errors {
		errorKeys = append(errorKeys, key)
	}
	sortkeys.Uint32s(errorKeys)

	for _, key := range errorKeys {
		e := g.transport.Errors[key]
		g.W(
			"export class %[1]sError extends JSONRPCError {\nconstructor(message, data) {\nsuper(message, \"%[1]sError\", %d, data);\n}\n}\n",
			e.Named.Obj().Name(), e.Code,
		)
	}

	g.W("function convertError(e) {\n")
	g.W("switch(e.code) {\n")
	g.W("default:\n")
	g.W("return new JSONRPCError(e.message, \"UnknownError\", e.code, e.data);\n")

	for _, key := range errorKeys {
		e := g.transport.Errors[key]
		g.W("case %d:\n", e.Code)
		g.W("return new %sError(e.message, e.data);\n", e.Named.Obj().Name())

	}
	g.W("}\n}\n")

	g.enums.Iterate(func(key stdtypes.Type, value interface{}) {
		if named, ok := key.(*stdtypes.Named); ok {
			b, ok := named.Obj().Type().Underlying().(*stdtypes.Basic)
			if !ok {
				return
			}
			g.W("export const %sEnum = Object.freeze({\n", named.Obj().Name())

			for _, enum := range value.([]model.Enum) {
				value := enum.Value
				if b.Info() == stdtypes.IsString {
					value = strconv.Quote(value)
				}
				g.W("%s: %s,\n", strconv.Quote(enum.Name), value)
			}
			g.W("});\n")
		}
	})

	return nil
}

func (g *jsonRPCJSClient) Imports() []string {
	return nil
}

func (g *jsonRPCJSClient) PkgName() string {
	return ""
}

func (g *jsonRPCJSClient) OutputDir() string {
	return ""
}

func (g *jsonRPCJSClient) Filename() string {
	return "client_jsonrpc_gen.js"
}

func NewJsonRPCJSClient(
	serviceMethods []model.ServiceMethod,
	transport model.TransportOption,
	enums *typeutil.Map,
) generator.Generator {
	return &jsonRPCJSClient{
		serviceMethods: serviceMethods,
		transport:      transport,
		enums:          enums,
	}
}
