package generator

import (
	"context"
	stdtypes "go/types"
	"strconv"

	"github.com/swipe-io/swipe/v2/internal/interface/typevisitor"
	"golang.org/x/tools/go/types/typeutil"

	"github.com/swipe-io/swipe/v2/internal/domain/model"
	"github.com/swipe-io/swipe/v2/internal/usecase/generator"
	"github.com/swipe-io/swipe/v2/internal/writer"
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
              const schedule = scheduleRequests[responses[i].id];
			  if (responses[i].error) {
				schedule.reject(responses[i].error);
				continue;
			  }
			  schedule.resolve(responses[i].result);
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
	Interfaces() model.Interfaces
	MethodOption(m model.ServiceMethod) model.MethodOption
	CommentFields() map[string]map[string]string
	Enums() *typeutil.Map
	Errors() map[uint32]*model.HTTPError
}

type jsonRPCJSClient struct {
	writer.BaseWriter
	options jsonRPCJSClientOptionsGateway
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
			if iface.IsNameChange() || g.options.Interfaces().Len() > 1 {
				prefix = iface.NameUnExport() + "."
			}

			mw.W(") {\n")
			mw.W("return this.scheduler.__scheduleRequest(\"%s\", {", prefix+m.LcName)

			for i, p := range m.Params {
				if i > 0 {
					mw.W(",")
				}
				mw.W("%[1]s:%[1]s", p.Name())
			}

			mw.W("}).catch(e => { return ")
			if iface.External() {
				mw.W("%s%sConvertError(e)", iface.AppName(), m.Name)
			} else {
				mw.W("%s%sConvertError(e)", iface.LoweName(), m.Name)
			}
			mw.W("; })\n")

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
		g.W("export default JSONRPCClient%s\n\n", iface.NameExport())
	}

	httpErrorsDub := map[uint32]struct{}{}

	for i := 0; i < g.options.Interfaces().Len(); i++ {
		iface := g.options.Interfaces().At(i)
		for _, method := range iface.Methods() {
			for _, e := range method.Errors {
				if _, ok := httpErrorsDub[e.ID]; ok {
					continue
				}
				httpErrorsDub[e.ID] = struct{}{}

				g.W("export class ")
				if iface.External() {
					g.W(iface.AppName())
				}
				g.W(
					"%[1]sError extends JSONRPCError {\nconstructor(message, data) {\nsuper(message, \"%[1]sError\", %[2]d, data);\n}\n}\n",
					e.Named.Obj().Name(), e.Code,
				)
			}
		}
	}

	for i := 0; i < g.options.Interfaces().Len(); i++ {
		iface := g.options.Interfaces().At(i)
		for _, method := range iface.Methods() {
			if iface.External() {
				g.W("function %s%sConvertError(e) {\n", iface.AppName(), method.Name)
			} else {
				g.W("function %s%sConvertError(e) {\n", iface.LoweName(), method.Name)
			}
			g.W("switch(e.code) {\n")
			g.W("default:\n")
			g.W("return new JSONRPCError(e.message, \"UnknownError\", e.code, e.data);\n")
			for _, e := range method.Errors {
				g.W("case %d:\n", e.Code)
				if iface.External() {
					g.W("return new %s%sError(e.message, e.data);\n", iface.AppName(), e.Named.Obj().Name())
				} else {
					g.W("return new %sError(e.message, e.data);\n", e.Named.Obj().Name())
				}
			}
			g.W("}\n}\n")
		}
	}

	g.options.Enums().Iterate(func(key stdtypes.Type, value interface{}) {
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

func NewJsonRPCJSClient(options jsonRPCJSClientOptionsGateway) generator.Generator {
	return &jsonRPCJSClient{
		options: options,
	}
}
