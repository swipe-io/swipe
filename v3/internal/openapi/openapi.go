package openapi

import (
	"encoding/json"
	"net/http"
	"path"
	"strconv"
	stdstrings "strings"

	"github.com/swipe-io/strcase"

	"github.com/swipe-io/swipe/v3/internal/plugin"
	"github.com/swipe-io/swipe/v3/option"
)

type InterfaceMethod struct {
	Name             option.String
	RESTMethod       string
	RESTPath         string
	RESTQueryVars    []string
	RESTPathVars     map[string]string
	Tags             []string
	Func             *option.FuncType
	Description      string
	RESTWrapResponse string
	RESTQueryValues  []string
	RESTHeaderVars   []string
	BearerAuth       bool
}

type Interface struct {
	Name      option.String
	Namespace string
	Methods   []InterfaceMethod
}

type Error struct {
	PkgName string
	PkgPath string
	Name    string
	Code    int64
	ErrCode string
}

type Openapi struct {
	info       Info
	servers    []Server
	interfaces []Interface
	errors     map[string]map[string][]Error
	useJSONRPC bool
	defTypes   map[string]*option.NamedType
}

func (g *Openapi) Build() OpenAPI {
	g.defTypes = make(map[string]*option.NamedType, 1024)
	o := OpenAPI{
		OpenAPI: "3.0.0",
		Paths:   map[string]*Path{},
		Components: Components{
			Schemas: Schemas{},
		},
	}

	o.Info = g.info
	o.Servers = g.servers

	if g.useJSONRPC {
		o.Components.Schemas = getOpenapiJSONRPCErrorSchemas()
	}

	var hasBearerAuth bool

	for _, iface := range g.interfaces {
		for _, m := range iface.Methods {
			var (
				pathStr        string
				op             *Operation
				httpMethodName = m.RESTMethod
			)
			if g.useJSONRPC {
				op = g.makeJSONRPCPath(m, iface.Namespace)
				pathStr = "/" + m.Name.Lower()
				if iface.Namespace != "" {
					pathStr = "/" + iface.Namespace + "." + m.Name.Lower()
				}
				httpMethodName = "POST"
			} else {
				op = g.makeRestPath(m)

				if m.RESTPath != "" {
					pathStr = m.RESTPath
				} else {
					pathStr = strcase.ToKebab(m.Name.Value)
				}
				for _, p := range m.Func.Sig.Params {
					if plugin.IsContext(p) {
						continue
					}
					if regexp, ok := m.RESTPathVars[p.Name.Value]; ok {
						pathStr = stdstrings.Replace(pathStr, ":"+regexp, "", -1)
					}
				}
				if iface.Namespace != "" {
					pathStr = path.Join(iface.Namespace, pathStr)
				}
			}

			if m.BearerAuth {
				op.Security = append(op.Security, map[string][]interface{}{
					"bearerAuth": []interface{}{},
				})
				hasBearerAuth = true
			}

			if methodErrors, ok := g.errors[iface.Name.Value]; ok {
				for _, errors := range methodErrors {
					for _, e := range errors {
						codeStr := strconv.FormatInt(e.Code, 10)
						errResponse := &Response{
							Content: Content{
								"application/json": {
									Schema: &Schema{
										Ref: "#/components/schemas/" + e.Name,
									},
								},
							},
						}
						if g.useJSONRPC {
							codeStr = "x-" + codeStr
							o.Components.Schemas[e.Name] = makeOpenapiSchemaJRPCError(e.Code)
							errResponse.Description = e.Name
						} else {
							errResponse.Description = http.StatusText(int(e.Code))
							o.Components.Schemas[e.Name] = makeOpenapiSchemaRESTError(e.ErrCode)
						}

						op.Responses[codeStr] = errResponse
					}
				}
			}

			if iface.Namespace != "" {
				op.Tags = append(op.Tags, iface.Namespace)
			}
			op.Tags = append(op.Tags, m.Tags...)
			op.Description = m.Description

			if _, ok := o.Paths[pathStr]; !ok {
				o.Paths[pathStr] = &Path{}
			}

			switch httpMethodName {
			default:
				o.Paths[pathStr].Get = op
			case "POST":
				o.Paths[pathStr].Post = op
			case "PUT":
				o.Paths[pathStr].Put = op
			case "PATCH":
				o.Paths[pathStr].Patch = op
			case "DELETE":
				o.Paths[pathStr].Delete = op
			}
		}
	}

	if hasBearerAuth {
		o.Components.SecuritySchemes = map[string]interface{}{
			"bearerAuth": BearerAuthSecuritySchema{
				Type:         "http",
				Scheme:       "bearer",
				BearerFormat: "JWT",
			},
		}
	}
	for _, namedType := range g.defTypes {
		o.Components.Schemas[namedType.Name.Value] = g.schemaByType(namedType.Type)
	}
	return o
}

func (g *Openapi) makeRef(named *option.NamedType) string {
	return "#/components/schemas/" + named.Name.Upper()
}

func (g *Openapi) fillTypeDefRecursive(t interface{}) {
	switch t := t.(type) {
	case *option.SliceType:
		g.fillTypeDefRecursive(t.Value)
	case *option.ArrayType:
		g.fillTypeDefRecursive(t.Value)
	case *option.MapType:
		g.fillTypeDefRecursive(t.Value)
	case *option.NamedType:
		switch t.Pkg.Path {
		case "time", "error", "github.com/pborman/uuid", "github.com/google/uuid":
			return
		}
		if _, ok := g.defTypes[t.Pkg.Path+t.Name.Value]; !ok {
			g.defTypes[t.Pkg.Path+t.Name.Value] = t

			switch tt := t.Type.(type) {
			case *option.SliceType:
				g.fillTypeDefRecursive(tt.Value)
			case *option.ArrayType:
				g.fillTypeDefRecursive(tt.Value)
			case *option.MapType:
				g.fillTypeDefRecursive(tt.Value)
			case *option.StructType:
				for _, field := range tt.Fields {
					g.fillTypeDefRecursive(field.Var.Type)
				}
			}
		}

	}
}

func (g *Openapi) fillTypeDef(t interface{}) {
	g.fillTypeDefRecursive(t)
}

func (g *Openapi) schemaByTypeRecursive(schema *Schema, t interface{}) {
	switch t := t.(type) {
	case *option.NamedType:
		switch t.Pkg.Path {
		default:
			schema.Ref = g.makeRef(t)
			return
		case "encoding/json":
			schema.Type = "object"
			schema.Properties = Properties{}
			return
		case "time":
			switch t.Name.Value {
			case "Duration":
				schema.Type = "string"
				schema.Example = "1h3m30s"
			case "Time":
				schema.Type = "string"
				schema.Format = "date-time"
				schema.Example = "1985-04-02T01:30:00.00Z"
			}
			return
		case "github.com/pborman/uuid", "github.com/google/uuid":
			schema.Type = "string"
			schema.Format = "uuid"
			schema.Example = "d5c02d83-6fbc-4dd7-8416-9f85ed80de46"
			return
		}
	case *option.StructType:
		for _, field := range t.Fields {
			name := field.Var.Name.Value
			if tag, err := field.Tags.Get("json"); err == nil {
				name = tag.Name
			}
			if name == "-" {
				continue
			}
			filedSchema := &Schema{
				Properties: Properties{},
			}
			filedSchema.Description = field.Var.Comment
			schema.Properties[name] = filedSchema
			g.schemaByTypeRecursive(filedSchema, field.Var.Type)
		}
	case *option.MapType:
		mapSchema := &Schema{
			Properties: Properties{},
		}
		schema.Properties = Properties{"key": mapSchema}
		g.schemaByTypeRecursive(mapSchema, t.Value)
		return
	case *option.ArrayType:
		schema.Type = "array"
		schema.Items = &Schema{
			Properties: Properties{},
		}
		g.schemaByTypeRecursive(schema.Items, t.Value)
		return
	case *option.SliceType:
		if b, ok := t.Value.(*option.BasicType); ok && b.IsByte() {
			schema.Type = "string"
			schema.Format = "byte"
			schema.Example = "U3dhZ2dlciByb2Nrcw=="
		} else {
			schema.Type = "array"
			schema.Items = &Schema{
				Properties: Properties{},
			}
			g.schemaByTypeRecursive(schema.Items, t.Value)
		}
		return
	case *option.IfaceType:
		schema.Type = "object"
		schema.Description = "Can be any value - string, number, boolean, array or object."
		schema.Properties = Properties{}
		schema.Example = json.RawMessage("null")
		schema.AnyOf = []Schema{
			{Type: "string", Example: "abc"},
			{Type: "integer", Example: 1},
			{Type: "number", Format: "float", Example: 1.11},
			{Type: "boolean", Example: true},
			{Type: "array"},
			{Type: "object"},
		}
		return
	case *option.BasicType:
		if t.IsString() {
			schema.Type = "string"
			schema.Example = "abc"
			return
		}
		if t.IsBool() {
			schema.Type = "boolean"
			schema.Example = "true"
		}
		if t.IsNumeric() {
			if t.IsInt32() || t.IsUint32() {
				schema.Type = "integer"
				schema.Format = "int32"
				schema.Example = 1
				return
			}
			if t.IsInt64() || t.IsUint64() {
				schema.Type = "integer"
				schema.Format = "int64"
				schema.Example = 1
				return
			}
			if t.IsFloat32() || t.IsFloat64() {
				schema.Type = "number"
				schema.Format = "float"
				schema.Example = 1.11
				return
			}
			schema.Type = "integer"
			schema.Example = 1
			return
		}
	}
}

func (g *Openapi) schemaByType(t interface{}) (schema *Schema) {
	schema = &Schema{
		Properties: Properties{},
	}
	g.schemaByTypeRecursive(schema, t)
	return
}

func (g *Openapi) makeJSONRPCPath(m InterfaceMethod, prefix string) *Operation {
	responseSchema := &Schema{
		Type:       "object",
		Properties: map[string]*Schema{},
	}
	requestSchema := &Schema{
		Type:       "object",
		Properties: map[string]*Schema{},
	}

	if plugin.LenWithoutErrors(m.Func.Sig.Params) > 0 {
		for _, p := range m.Func.Sig.Params {
			if plugin.IsContext(p) {
				continue
			}
			g.fillTypeDef(p.Type)

			schema := g.schemaByType(p.Type)
			schema.Description = p.Comment
			requestSchema.Properties[p.Name.Lower()] = schema
		}
	} else {
		requestSchema.Type = "object"
		requestSchema.Nullable = true
		requestSchema.Example = json.RawMessage("null")
	}

	lenResults := plugin.LenWithoutErrors(m.Func.Sig.Results)

	if lenResults > 1 {
		for _, r := range m.Func.Sig.Results {
			if plugin.IsError(r) {
				continue
			}
			if plugin.IsFileDownloadType(r.Type) {
				continue
			}
			g.fillTypeDef(r.Type)
			schema := g.schemaByType(r.Type)
			responseSchema.Properties[r.Name.Lower()] = schema
		}
	} else if lenResults == 1 {
		if !plugin.IsFileDownloadType(m.Func.Sig.Results[0].Type) {
			g.fillTypeDef(m.Func.Sig.Results[0].Type)
			responseSchema = g.schemaByType(m.Func.Sig.Results[0].Type)
		}
	} else {
		responseSchema.Example = json.RawMessage("null")
	}

	if m.RESTWrapResponse != "" {
		properties := Properties{}
		properties[m.RESTWrapResponse] = responseSchema
		responseSchema = &Schema{
			Properties: properties,
		}
	}

	response := &Schema{
		Type: "object",
		Properties: Properties{
			"jsonrpc": &Schema{
				Type:    "string",
				Example: "2.0",
			},
			"id": &Schema{
				Type:    "string",
				Example: "c9b14c57-7503-447a-9fb9-be6f8920f31f",
			},
			"result": responseSchema,
		},
	}

	restMethod := m.Name.Lower()
	if prefix != "" {
		restMethod = prefix + "." + restMethod
	}

	request := &Schema{
		Type: "object",
		Properties: Properties{
			"jsonrpc": &Schema{
				Type:    "string",
				Example: "2.0",
			},
			"id": &Schema{
				Type:    "string",
				Example: "c9b14c57-7503-447a-9fb9-be6f8920f31f",
			},
			"method": &Schema{
				Type: "string",
				Enum: []string{restMethod},
			},
			"params": requestSchema,
		},
	}

	return &Operation{
		RequestBody: &RequestBody{
			Required: true,
			Content: map[string]Media{
				"application/json": {
					Schema: request,
				},
			},
		},
		Responses: map[string]*Response{
			"200": {
				Description: "OK",
				Content: Content{
					"application/json": {
						Schema: response,
					},
				},
			},
			"x-32700": {
				Description: "Parse error. Invalid JSON was received by the server. An error occurred on the server while parsing the JSON text.",
				Content: Content{
					"application/json": {
						Schema: &Schema{
							Ref: "#/components/schemas/ParseError",
						},
					},
				},
			},
			"x-32600": {
				Description: "Invalid Request. The JSON sent is not a valid Request object.",
				Content: Content{
					"application/json": {
						Schema: &Schema{
							Ref: "#/components/schemas/InvalidRequestError",
						},
					},
				},
			},
			"x-32601": {
				Description: "Method not found. The method does not exist / is not available.",
				Content: Content{
					"application/json": {
						Schema: &Schema{
							Ref: "#/components/schemas/MethodNotFoundError",
						},
					},
				},
			},
			"x-32602": {
				Description: "Invalid params. Invalid method parameters.",
				Content: Content{
					"application/json": {
						Schema: &Schema{
							Ref: "#/components/schemas/InvalidParamsError",
						},
					},
				},
			},
			"x-32603": {
				Description: "Internal error. Internal JSON-RPC error.",
				Content: Content{
					"application/json": {
						Schema: &Schema{
							Ref: "#/components/schemas/InternalError",
						},
					},
				},
			},
		},
	}
}

func (g *Openapi) makeRestPath(m InterfaceMethod) *Operation {
	responseSchema := &Schema{
		Type:       "object",
		Properties: map[string]*Schema{},
	}

	requestSchema := &Schema{
		Type:       "object",
		Properties: map[string]*Schema{},
	}

	queryVars := make([]plugin.VarType, 0, len(m.RESTQueryVars))
	queryValues := make([]plugin.VarType, 0, len(m.RESTQueryValues))
	headerVars := make([]plugin.VarType, 0, len(m.RESTHeaderVars))
	pathVars := make([]plugin.VarType, 0, len(m.RESTPathVars))
	paramVars := make([]*option.VarType, 0, len(m.Func.Sig.Params))

	for _, p := range m.Func.Sig.Params {
		if plugin.IsContext(p) {
			continue
		}
		if v, ok := plugin.FindParam(p, m.RESTQueryVars); ok {
			queryVars = append(queryVars, v)
			continue
		}
		if v, ok := plugin.FindParam(p, m.RESTQueryValues); ok {
			queryValues = append(queryValues, v)
			continue
		}
		if v, ok := plugin.FindParam(p, m.RESTHeaderVars); ok {
			headerVars = append(headerVars, v)
			continue
		}
		if regexp, ok := m.RESTPathVars[p.Name.Value]; ok {
			pathVars = append(pathVars, plugin.VarType{
				Param: p,
				Value: regexp,
			})
			continue
		}
		paramVars = append(paramVars, p)
	}

	for _, p := range paramVars {
		g.fillTypeDef(p.Type)
		schema := g.schemaByType(p.Type)
		schema.Description = p.Comment
		requestSchema.Properties[p.Name.Lower()] = schema
	}

	lenResults := plugin.LenWithoutErrors(m.Func.Sig.Results)
	if lenResults > 1 {
		for _, r := range m.Func.Sig.Results {
			if plugin.IsError(r) {
				continue
			}
			if plugin.IsFileDownloadType(r.Type) {
				responseSchema.Type = "string"
				responseSchema.Format = "binary"
				continue
			}
			g.fillTypeDef(r.Type)
			responseSchema.Properties[r.Name.Lower()] = g.schemaByType(r.Type)
		}
	} else if lenResults == 1 {
		if !plugin.IsFileDownloadType(m.Func.Sig.Results[0].Type) {
			g.fillTypeDef(m.Func.Sig.Results[0].Type)
			responseSchema = g.schemaByType(m.Func.Sig.Results[0].Type)
		} else {
			responseSchema.Type = "string"
			responseSchema.Format = "binary"
		}
	}
	if m.RESTWrapResponse != "" {
		properties := Properties{}
		properties[m.RESTWrapResponse] = responseSchema
		responseSchema = &Schema{
			Properties: properties,
		}
	}
	responses := map[string]*Response{}
	if lenResults == 0 {
		responses["201"] = &Response{
			Description: "Created",
			Content: Content{
				"text/plain": {},
			},
		}
	} else {

		if responseSchema.Type == "string" && responseSchema.Format == "binary" {
			responses["200"] = &Response{
				Description: "OK",
				Content: Content{
					"application/file": {
						Schema: responseSchema,
					},
				},
			}
		} else {
			responses["200"] = &Response{
				Description: "OK",
				Content: Content{
					"application/json": {
						Schema: responseSchema,
					},
				},
			}
		}

	}

	o := &Operation{
		Summary:   m.Name.Value,
		Responses: responses,
	}

	for _, pathVar := range pathVars {
		o.Parameters = append(o.Parameters, Parameter{
			In:          "path",
			Name:        pathVar.Param.Name.Lower(),
			Description: pathVar.Param.Comment,
			Required:    pathVar.IsRequired,
			Schema:      g.schemaByType(pathVar.Param.Type),
		})
	}

	for _, headerVar := range headerVars {
		o.Parameters = append(o.Parameters, Parameter{
			In:          "header",
			Name:        headerVar.Value,
			Description: headerVar.Param.Comment,
			Required:    headerVar.IsRequired,
			Schema:      g.schemaByType(headerVar.Param.Type),
		})
	}

	for _, queryVar := range queryVars {
		if named, ok := queryVar.Param.Type.(*option.NamedType); ok {
			if st, ok := named.Type.(*option.StructType); ok {
				for _, field := range st.Fields {
					o.Parameters = append(o.Parameters, Parameter{
						In:          "query",
						Name:        field.Var.Name.Lower(),
						Description: field.Var.Comment,
						Required:    queryVar.IsRequired,
						Schema: &Schema{
							Type:       "string",
							Properties: Properties{},
						},
					})
				}
			}
		} else {
			o.Parameters = append(o.Parameters, Parameter{
				In:          "query",
				Name:        queryVar.Param.Name.Lower(),
				Description: queryVar.Param.Comment,
				Required:    queryVar.IsRequired,
				Schema: &Schema{
					Type:       "string",
					Properties: Properties{},
				},
			})
		}
	}

	switch m.RESTMethod {
	case "POST", "PUT", "PATCH":
		o.RequestBody = &RequestBody{
			Required: true,
			Content: map[string]Media{
				"application/json": {
					Schema: requestSchema,
				},
			},
		}
	}
	return o
}

func NewOpenapi(info Info, servers []Server, interfaces []Interface, errors map[string]map[string][]Error, useJSONRPC bool) *Openapi {
	return &Openapi{info: info, servers: servers, interfaces: interfaces, errors: errors, useJSONRPC: useJSONRPC}
}
