package generator

import (
	"fmt"
	"strconv"
	stdstrings "strings"

	"github.com/gertd/go-pluralize"

	"github.com/swipe-io/swipe/v2/internal/plugin/gokit/config"

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

func NameRequest(m *option.FuncType, iface *config.Interface) string {
	return UcNameWithAppPrefix(iface) + m.Name.UpperCase + "Request"
}

func NameResponse(m *option.FuncType, iface *config.Interface) string {
	return UcNameWithAppPrefix(iface) + m.Name.UpperCase + "Response"
}

func NameMakeEndpoint(m *option.FuncType, iface *config.Interface) string {
	return fmt.Sprintf("Make%s%sEndpoint", UcNameWithAppPrefix(iface), m.Name.UpperCase)
}

func LcNameWithAppPrefix(iface *config.Interface) string {
	if iface.Named.Pkg.Module.External {
		return LcAppName(iface) + iface.Named.Name.UpperCase
	}
	return iface.Named.Name.LowerCase
}

func UcNameWithAppPrefix(iface *config.Interface) string {
	if iface.Named.Pkg.Module.External {
		return UcAppName(iface) + iface.Named.Name.UpperCase
	}
	return iface.Named.Name.UpperCase
}

func UcAppName(iface *config.Interface) string {
	return strcase.ToCamel(stdstrings.Split(iface.Named.Pkg.Path, "/")[:3][2])
}

func LcAppName(iface *config.Interface) string {
	return strcase.ToLowerCamel(stdstrings.Split(iface.Named.Pkg.Path, "/")[:3][2])
}

func UcNameJS(iface *config.Interface) string {
	if iface.ClientName.Value != "" {
		return strcase.ToCamel(iface.ClientName.Value)
	}
	return iface.Named.Name.UpperCase
}

func LcNameJS(iface *config.Interface) string {
	if iface.ClientName.Value != "" {
		return strcase.ToLowerCamel(iface.ClientName.Value)
	}
	return iface.Named.Name.LowerCase
}

func NameInterface(iface *config.Interface) string {
	return LcNameWithAppPrefix(iface) + "Interface"
}

func NameLoggingMiddleware(iface *config.Interface) string {
	return LcNameWithAppPrefix(iface) + "LoggingMiddleware"
}

func NameInstrumentingMiddleware(iface *config.Interface) string {
	return LcNameWithAppPrefix(iface) + "InstrumentingMiddleware"
}

func NameEndpointSetNameVar(iface *config.Interface) string {
	return LcNameWithAppPrefix(iface) + "EpSet"
}

func NameEndpointSetName(iface *config.Interface) string {
	return UcNameWithAppPrefix(iface) + "EndpointSet"
}

func LcNameEndpoint(iface *config.Interface, fn *option.FuncType) string {
	return iface.Named.Name.LowerCase + fn.Name.Origin + "Endpoint"
}

func UcNameIfaceMethod(iface *config.Interface, fn *option.FuncType) string {
	return iface.Named.Name.UpperCase + fn.Name.UpperCase
}

func LcNameIfaceMethod(iface *config.Interface, fn *option.FuncType) string {
	return iface.Named.Name.LowerCase + fn.Name.UpperCase
}

func ClientType(iface *config.Interface) string {
	return UcNameWithAppPrefix(iface) + "Client"
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

func LenWithoutContexts(vars option.VarsType) int {
	return len(vars) - len(Contexts(vars))
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
		if t.Pkg != nil {
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

func docMethodName(iface *config.Interface, method *option.FuncType) string {
	if iface.ClientName.Value != "" {
		return "JSONRPCClient" + strcase.ToCamel(iface.ClientName.Value) + "." + method.Name.LowerCase
	}
	return "JSONRPCClient" + iface.Named.Name.UpperCase + "." + method.Name.LowerCase
}

func jsErrorName(iface *config.Interface, e config.Error) (errorName string) {
	return UcNameWithAppPrefix(iface) + singular(e.Name)
}

func singular(word string) string {
	return pluralize.NewClient().Singular(word)
}

func isGolangNamedType(t *option.NamedType) bool {
	switch t.Pkg.Path {
	case "time":
		switch t.Name.Origin {
		case "Time", "Location":
			return true
		}
	case "sql":
		switch t.Name.Origin {
		case "NullBool", "NullFloat64", "NullInt32", "NullInt64", "NullString", "NullTime":
			return true
		}
	}
	return false
}

func fillType(i interface{}, visited map[string]*option.NamedType) {
	switch t := i.(type) {
	case *option.NamedType:
		if _, ok := t.Type.(*option.StructType); ok {
			key := t.Pkg.Path + t.Name.Origin
			_, ok := visited[key]
			if !ok {
				visited[key] = t
				fillType(t.Type, visited)
			}
		}
	case *option.SliceType:
		fillType(t.Value, visited)
	case *option.ArrayType:
		fillType(t.Value, visited)
	case *option.MapType:
		fillType(t.Value, visited)
	}
}
