package generator

import (
	"fmt"
	"strconv"
	stdstrings "strings"

	"github.com/swipe-io/strcase"
	"github.com/swipe-io/swipe/v2/internal/option"
	"github.com/swipe-io/swipe/v2/internal/plugin/gokit/openapi"
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

func NameRequest(m *option.FuncType, named *option.NamedType) string {
	var prefix string
	if named != nil {
		prefix = named.Name.LowerCase
	}
	return prefix + m.Name.UpperCase + "Request"
}

func NameResponse(m *option.FuncType, named *option.NamedType) string {
	var prefix string
	if named != nil {
		prefix = named.Name.LowerCase
	}
	return prefix + m.Name.UpperCase + "Response"
}

func NameMakeEndpoint(m *option.FuncType, named *option.NamedType) string {
	return fmt.Sprintf("Make%s%sEndpoint", named.Name.UpperCase, m.Name.UpperCase)
}

func LcNameWithAppPrefix(named *option.NamedType) string {
	if named.Pkg.Module.External {
		return strcase.ToLowerCamel(stdstrings.Split(named.Pkg.Path, "/")[:3][2]) + named.Name.UpperCase
	}
	return named.Name.LowerCase
}

func UcNameWithAppPrefix(named *option.NamedType) string {
	if named.Pkg.Module.External {
		return strcase.ToCamel(stdstrings.Split(named.Pkg.Path, "/")[:3][2]) + named.Name.UpperCase
	}
	return named.Name.UpperCase
}

func NameInterface(named *option.NamedType) string {
	return LcNameWithAppPrefix(named) + "Interface"
}

func NameLoggingMiddleware(named *option.NamedType) string {
	return LcNameWithAppPrefix(named) + "LoggingMiddleware"
}

func NameInstrumentingMiddleware(named *option.NamedType) string {
	return LcNameWithAppPrefix(named) + "InstrumentingMiddleware"
}

func NameEndpointSetName(named *option.NamedType) string {
	return LcNameWithAppPrefix(named) + "EpSet"
}

func LcNameEndpoint(named *option.NamedType, fn *option.FuncType) string {
	return named.Name.LowerCase + fn.Name.Origin + "Endpoint"
}

func UcNameIfaceMethod(named *option.NamedType, fn *option.FuncType) string {
	return named.Name.UpperCase + fn.Name.UpperCase
}

func LcNameIfaceMethod(named *option.NamedType, fn *option.FuncType) string {
	return named.Name.LowerCase + fn.Name.UpperCase
}

func IsContext(v *option.VarType) bool {
	if named, ok := v.Type.(*option.NamedType); ok {
		if _, ok := named.Type.(*option.IfaceType); ok {
			return named.Name.Origin == "Context" && named.Pkg.Path == "context"
		}
	}
	return false
}

func IsError(v *option.VarType) bool {
	if named, ok := v.Type.(*option.NamedType); ok {
		if _, ok := named.Type.(*option.IfaceType); ok && named.Name.Origin == "error" {
			return true
		}
	}
	return false
}

func Errors(vars option.VarsType) (result []*option.VarType) {
	for _, v := range vars {
		if IsError(v) {
			result = append(result, v)
		}
	}
	return
}

func Contexts(vars option.VarsType) (result []*option.VarType) {
	for _, v := range vars {
		if IsContext(v) {
			result = append(result, v)
		}
	}
	return
}

func LenWithoutErrors(vars option.VarsType) int {
	return len(vars) - len(Errors(vars))
}

func makeLogParams(include, exclude map[string]struct{}, data ...*option.VarType) (result []string) {
	return makeLogParamsRecursive(include, exclude, "", data...)
}

func makeLogParamsRecursive(include, exclude map[string]struct{}, parentName string, data ...*option.VarType) (result []string) {
	for _, v := range data {
		if IsContext(v) {
			continue
		}
		if len(include) > 0 {
			if _, ok := include[v.Name.Origin]; !ok {
				continue
			}
		}
		if len(exclude) > 0 {
			if _, ok := exclude[v.Name.Origin]; ok {
				continue
			}
		}
		if logParam := makeLogParam(parentName+v.Name.Origin, v.Type); len(logParam) > 0 {
			result = append(result, logParam...)
		}
	}
	return
}

func makeLogParam(name string, t interface{}) []string {
	quoteName := strconv.Quote(name)
	switch t := t.(type) {
	default:
		return []string{quoteName, name}
	case *option.NamedType:
		if hasMethodString(t) {
			return []string{quoteName, name + ".String()"}
		}
		return nil
	case *option.StructType:
		return nil
	case *option.BasicType:
		return []string{quoteName, name}
	case *option.SliceType, *option.ArrayType, *option.MapType:
		return []string{quoteName, "len(" + name + ")"}
	}
}

func hasMethodString(v *option.NamedType) bool {
	for _, method := range v.Methods {
		if method.Name.Origin != "String" {
			continue
		}
		if len(method.Sig.Params) == 0 && len(method.Sig.Results) == 1 {
			if t, ok := method.Sig.Results[0].Type.(*option.BasicType); ok {
				return t.IsString()
			}
		}
	}
	return false
}

func findContextVar(vars option.VarsType) (v *option.VarType) {
	for _, p := range vars {
		if IsContext(p) {
			v = p
			break
		}
	}
	return
}

func findErrorVar(vars option.VarsType) (v *option.VarType) {
	for _, p := range vars {
		if IsError(p) {
			v = p
			break
		}
	}
	return
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

func getOpenapiRESTErrorSchema() *openapi.Schema {
	return &openapi.Schema{
		Type: "object",
		Properties: openapi.Properties{
			"error": &openapi.Schema{
				Type: "string",
			},
		},
	}
}

func jsTypeDef(i interface{}) string {
	return jsTypeDefRecursive(i, 0, map[string]struct{}{})
}

func jsTypeDefRecursive(i interface{}, nested int, visited map[string]struct{}) string {
	switch t := i.(type) {
	case *option.NamedType:
		switch t.Pkg.Path {
		case "github.com/google/uuid", "github.com/pborman/uuid":
			switch t.Name.Origin {
			case "UUID":
				return "string"
			}
		case "encoding/json":
			switch t.Name.Origin {
			case "RawMessage":
				return "*"
			}
		case "time":
			switch t.Name.Origin {
			case "Time":
				return "string"
			}
		}
		return t.Name.Origin
	case *option.StructType:
		out := ""
		for _, f := range t.Fields {
			if jsonTag, err := f.Tags.Get("json"); err == nil {
				if jsonTag.Name == "-" {
					continue
				}
			}
			if em, ok := f.Var.Type.(*option.StructType); ok && f.Var.Embedded {
				out += jsTypeDefRecursive(em, nested, visited)
				continue
			}
			out += "* @property {" + jsDocType(f.Var.Type) + "} " + f.Var.Name.LowerCase
			out += "\n"
		}
		return out
	}
	return ""
}

func jsDocType(i interface{}) string {
	return jsDocTypeRecursive(i, 0)
}

func jsDocTypeRecursive(i interface{}, nested int) string {
	switch t := i.(type) {
	case *option.StructType:
		out := ""
		for _, f := range t.Fields {
			if jsonTag, err := f.Tags.Get("json"); err == nil {
				if jsonTag.Name == "-" {
					continue
				}
			}
			if em, ok := f.Var.Type.(*option.StructType); ok && f.Var.Embedded {
				out += jsDocTypeRecursive(em, nested)
				continue
			}
			out += stdstrings.Repeat(" ", nested) + f.Var.Name.LowerCase + ": " + jsDocTypeRecursive(f.Var.Type, nested+1)
			out += "\n"
		}
		return out
	case *option.IfaceType:
		return "object"
	case *option.MapType:
		return "Object<string, " + jsDocTypeRecursive(t.Value, nested) + ">"
	case *option.SliceType:
		if b, ok := t.Value.(*option.BasicType); ok {
			if b.IsByte() {
				return "string"
			}
		}
		return "Array<" + jsDocTypeRecursive(t.Value, nested) + ">"
	case *option.ArrayType:
		return "Array<" + jsDocTypeRecursive(t.Value, nested) + ">"
	case *option.NamedType:
		switch t.Pkg.Path {
		case "github.com/google/uuid", "github.com/pborman/uuid":
			switch t.Name.Origin {
			case "UUID":
				return "string"
			}
		case "encoding/json":
			switch t.Name.Origin {
			case "RawMessage":
				return "*"
			}
		case "time":
			switch t.Name.Origin {
			case "Time":
				return "string"
			}
		}
		return t.Name.Origin
	case *option.BasicType:
		if t.IsString() {
			return "string"
		}
		if t.IsNumeric() {
			return "number"
		}
		if t.IsBool() {
			return "boolean"
		}
		return t.Name
	}
	return ""
}
