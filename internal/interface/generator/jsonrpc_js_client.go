package generator

import (
	"context"
	stdtypes "go/types"
	"strconv"

	"github.com/swipe-io/swipe/v2/internal/interface/typevisitor"

	"github.com/swipe-io/swipe/v2/internal/domain/model"
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

class JSONRPCScheduler {
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
	/**
    * @param {string} method
    * @param {Object} params
    * @returns {Promise<*>}
    */
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

type jsonRPCJSClientOptionsGateway interface {
	Error(key uint32) *model.HTTPError
	ErrorKeys() []uint32
	Interfaces() model.Interfaces
	MethodOption(m model.ServiceMethod) model.MethodOption
}

type jsonRPCJSClient struct {
	writer.BaseWriter
	options jsonRPCJSClientOptionsGateway
	enums   *typeutil.Map
}

func (g *jsonRPCJSClient) Prepare(_ context.Context) error {
	return nil
}

func (g *jsonRPCJSClient) Process(_ context.Context) error {
	g.W(jsonRPCClientBase)

	mw := writer.BaseWriter{}

	tdc := typevisitor.NewNamedTypeCollector()

	for i := 0; i < g.options.Interfaces().Len(); i++ {
		iface := g.options.Interfaces().At(i)

		mw.W("class JSONRPCClient%s {\n", iface.NameExport())
		mw.W("constructor(transport) {\n")
		mw.W("this.scheduler = new JSONRPCScheduler(transport);\n")
		mw.W("}\n\n")

		for _, m := range iface.Methods() {
			mopt := g.options.MethodOption(m)
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
			var prefix string
			if g.options.Interfaces().Len() > 1 {
				prefix = iface.LoweName() + "."
				if iface.NameUnExport() != "" {
					prefix = iface.NameUnExport() + "."
				}
			}

			mw.W(") {\n")
			mw.W("return this.scheduler.__scheduleRequest(\"%s\", {", prefix+m.LcName)

			for i, p := range m.Params {
				if i > 0 {
					mw.W(",")
				}
				mw.W("%[1]s:%[1]s", p.Name())
			}

			mw.W("})\n")
			mw.W("}\n")
		}
		mw.W("}\n\n")
	}

	buf := new(writer.BaseWriter)
	for _, t := range tdc.TypeDefs() {
		typevisitor.JSTypeDefVisitor(buf).Visit(t)
	}

	g.W(buf.String())
	g.W(mw.String())

	if g.options.Interfaces().Len() > 1 {
		g.W("class JSONRPCClient {\n")
		g.W("constructor(transport) {\n")
		for i := 0; i < g.options.Interfaces().Len(); i++ {
			iface := g.options.Interfaces().At(i)
			g.W("this.%[1]s = new JSONRPCClient%[1]s(transport);\n", iface.NameExport())
		}
		g.W("}\n")
		g.W("}\n")

		g.W("export default JSONRPCClient\n\n")
	} else if g.options.Interfaces().Len() == 1 {
		iface := g.options.Interfaces().At(0)
		g.W("export default JSONRPCClient%s\n\n", iface.Name())
	}

	for _, key := range g.options.ErrorKeys() {
		e := g.options.Error(key)
		g.W(
			"export class %[1]sError extends JSONRPCError {\nconstructor(message, data) {\nsuper(message, \"%[1]sError\", %d, data);\n}\n}\n",
			e.Named.Obj().Name(), e.Code,
		)
	}

	g.W("function convertError(e) {\n")
	g.W("switch(e.code) {\n")
	g.W("default:\n")
	g.W("return new JSONRPCError(e.message, \"UnknownError\", e.code, e.data);\n")

	for _, key := range g.options.ErrorKeys() {
		e := g.options.Error(key)
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
	options jsonRPCJSClientOptionsGateway,
	enums *typeutil.Map,
) generator.Generator {
	return &jsonRPCJSClient{
		options: options,
		enums:   enums,
	}
}
