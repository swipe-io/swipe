package generator

import (
	"bytes"
	"context"
	"fmt"
	stdtypes "go/types"
	"strconv"
	"strings"

	"github.com/fatih/structtag"

	"github.com/swipe-io/swipe/pkg/domain/model"
	"github.com/swipe-io/swipe/pkg/writer"
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
	makeJSONRPCRequest(id, async, method, params) {
	  return {
		jsonrpc: "2.0",
		id: id,
		async: async,
		method: method,
		params: params,
	  };
	}
	__scheduleRequest(method, params, async) {
	  const p = new Promise((resolve, reject) => {
		const request = this.makeJSONRPCRequest(
		  this.__requestIDGenerate(),
		  async,
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
	filename string
	info     model.GenerateInfo
	o        model.ServiceOption
}

func (g *jsonRPCJSClient) Prepare(_ context.Context) error {
	return nil
}

func (g *jsonRPCJSClient) Process(_ context.Context) error {
	g.W(jsonRPCClientBase)

	g.W("export default class extends JSONRPCClient {\n")

	for _, m := range g.o.Methods {
		mopt := g.o.Transport.MethodOptions[m.Name]
		g.W("/**\n")

		if len(m.Comments) > 0 {
			for _, comment := range m.Comments {
				g.W("* %s\n", comment)
			}
			g.W("*\n")
		}

		for _, p := range m.Params {
			g.W("* @param {%s} %s\n", g.getJSDocType(p.Type(), 0), p.Name())
		}

		if len(m.Results) > 0 {
			g.W("* @return {PromiseLike<")
			if m.ResultsNamed {
				if mopt.WrapResponse.Enable {
					g.W("{%s: ", mopt.WrapResponse.Name)
				} else {
					g.W("{")
				}
			}

			for i, p := range m.Results {
				if i > 0 {
					g.W(", ")
				}
				if p.Name() != "" {
					g.W("%s: ", p.Name())
				}
				g.W(g.getJSDocType(p.Type(), 0))
			}
			if m.ResultsNamed || mopt.WrapResponse.Enable {
				g.W("}")
			}
			g.W(">}\n")
		}

		g.W("**/\n")
		renderMethod := func(async bool) {
			suffixAsync := ""
			if async {
				suffixAsync = "Async"
			}
			g.W("%s%s(", m.LcName, suffixAsync)

			for i, p := range m.Params {
				if i > 0 {
					g.W(",")
				}
				g.W(p.Name())
			}

			g.W(") {\n")
			g.W("return this.__scheduleRequest(\"%s\", {", m.LcName)

			firstComma := ""
			if len(m.Params) > 0 {
				firstComma = ","
			}
			g.W("async: %t%s", async, firstComma)

			for i, p := range m.Params {
				if i > 0 {
					g.W(",")
				}
				g.W("%[1]s:%[1]s", p.Name())
			}

			g.W("})\n")
			g.W("}\n")
		}
		renderMethod(false)
		g.W("\n")
		renderMethod(true)
	}

	g.W("}\n")

	for _, e := range g.o.Transport.Errors {
		g.W(
			"export class %[1]sError extends JSONRPCError {\nconstructor(message, data) {\nsuper(message, \"%[1]sError\", %d, data);\n}\n}\n",
			e.Named.Obj().Name(), e.Code,
		)
	}
	g.W("function convertError(e) {\n")
	g.W("switch(e.code) {\n")
	g.W("default:\n")
	g.W("return new JSONRPCError(e.message, \"UnknownError\", e.code, e.data);\n")

	for _, e := range g.o.Transport.Errors {
		g.W("case %d:\n", e.Code)
		g.W("return new %sError(e.message, e.data);\n", e.Named.Obj().Name())

	}
	g.W("}\n}\n")

	g.info.Enums.Iterate(func(key stdtypes.Type, value interface{}) {
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
	return g.filename
}

func (g *jsonRPCJSClient) getJSDocType(t stdtypes.Type, nested int) string {
	switch v := t.(type) {
	default:
		return "*"
	case *stdtypes.Pointer:
		return g.getJSDocType(v.Elem(), nested)
	case *stdtypes.Array:
		return fmt.Sprintf("Array.<%s>", g.getJSDocType(v.Elem(), nested))
	case *stdtypes.Slice:
		return fmt.Sprintf("Array.<%s>", g.getJSDocType(v.Elem(), nested))
	case *stdtypes.Map:
		return fmt.Sprintf("Object.<string, %s>", g.getJSDocType(v.Elem(), nested))
	case *stdtypes.Named:
		switch stdtypes.TypeString(v.Obj().Type(), nil) {
		case "encoding/json.RawMessage":
			return "*"
		case "github.com/pborman/uuid.UUID",
			"github.com/google/uuid.UUID":
			return "string"
		case "time.Time":
			return "string"
		}
		return g.getJSDocType(v.Obj().Type().Underlying(), nested)
	case *stdtypes.Struct:
		buf := new(bytes.Buffer)
		_, _ = fmt.Fprintf(buf, "{\n")

		var writeStruct func(st *stdtypes.Struct)
		writeStruct = func(st *stdtypes.Struct) {
			var j int
			for i := 0; i < st.NumFields(); i++ {
				f := st.Field(i)
				if f.Embedded() {
					var st *stdtypes.Struct
					if ptr, ok := f.Type().(*stdtypes.Pointer); ok {
						st = ptr.Elem().Underlying().(*stdtypes.Struct)
					} else {
						st = f.Type().Underlying().(*stdtypes.Struct)
					}
					writeStruct(st)
					_, _ = fmt.Fprint(buf, ",\n")
					continue
				}
				var (
					skip bool
					name = f.Name()
				)
				if tags, err := structtag.Parse(st.Tag(i)); err == nil {
					if jsonTag, err := tags.Get("json"); err == nil {
						if jsonTag.Name == "-" {
							skip = true
						} else {
							name = jsonTag.Name
						}
					}
				}
				if skip {
					continue
				}
				if j > 0 {
					_, _ = fmt.Fprint(buf, ",\n")
				}
				_, _ = fmt.Fprintf(buf, "* %s %s: %s", strings.Repeat("  ", nested), name, g.getJSDocType(f.Type(), nested+1))
				j++
			}
		}

		writeStruct(v)

		_, _ = fmt.Fprintln(buf)

		endNested := nested - 2
		if endNested < 0 {
			endNested = 0
		}

		_, _ = fmt.Fprintf(buf, "* %s }", strings.Repeat("  ", endNested))
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

func NewJsonRPCJSClient(filename string, info model.GenerateInfo, o model.ServiceOption) Generator {
	return &jsonRPCJSClient{filename: filename, info: info, o: o}
}
