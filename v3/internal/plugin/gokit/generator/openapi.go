package generator

import (
	"context"
	"encoding/json"
	"net/http"
	"path"
	"strconv"
	stdstrings "strings"

	"github.com/pquerna/ffjson/ffjson"
	"github.com/swipe-io/strcase"

	"github.com/swipe-io/swipe/v3/internal/plugin/gokit/config"
	"github.com/swipe-io/swipe/v3/internal/plugin/gokit/openapi"
	"github.com/swipe-io/swipe/v3/option"
	"github.com/swipe-io/swipe/v3/writer"
)

type Openapi struct {
	w             writer.BaseWriter
	JSONRPCEnable bool
	Contact       config.OpenapiContact
	Info          config.OpenapiInfo
	MethodTags    map[string][]string
	Servers       []config.OpenapiServer
	Licence       config.OpenapiLicence
	Output        string
	Interfaces    []*config.Interface
	MethodOptions map[string]config.MethodOption
	IfaceErrors   map[string]map[string][]config.Error
	defTypes      map[string]*option.NamedType
}

func (g *Openapi) Generate(ctx context.Context) []byte {
	g.defTypes = make(map[string]*option.NamedType, 1024)

	o := openapi.OpenAPI{
		OpenAPI: "3.0.0",
		Info: openapi.Info{
			Title:          g.Info.Title,
			Description:    g.Info.Description,
			TermsOfService: "",
			Contact: &openapi.Contact{
				Name:  g.Contact.Name,
				URL:   g.Contact.Url,
				Email: g.Contact.Email,
			},
			License: &openapi.License{
				Name: g.Licence.Name,
				URL:  g.Licence.Url,
			},
			Version: g.Info.Version,
		},
		Paths: map[string]*openapi.Path{},
		Components: openapi.Components{
			Schemas: openapi.Schemas{},
		},
	}
	if g.JSONRPCEnable {
		o.Components.Schemas = getOpenapiJSONRPCErrorSchemas()
	} else {
		o.Components.Schemas["Error"] = getOpenapiRESTErrorSchema()
	}
	for _, s := range g.Servers {
		o.Servers = append(o.Servers, openapi.Server{
			URL:         s.Url,
			Description: s.Description,
			Variables:   nil,
		})
	}
	for _, iface := range g.Interfaces {
		ifaceType := iface.Named.Type.(*option.IfaceType)
		for _, m := range ifaceType.Methods {
			mopt := g.MethodOptions[iface.Named.Name.Value+m.Name.Value]

			var (
				pathStr        string
				op             *openapi.Operation
				httpMethodName = mopt.RESTMethod.Value
			)
			tags := g.MethodTags[iface.Named.Name.Value+m.Name.Value]

			if g.JSONRPCEnable {
				op = g.makeJSONRPCPath(m, iface.Namespace, mopt)
				pathStr = "/" + m.Name.Lower()
				if iface.Namespace != "" {
					pathStr = "/" + iface.Namespace + "." + m.Name.Lower()
				}
				httpMethodName = "POST"
			} else {
				op = g.makeRestPath(m, mopt)

				if mopt.RESTPath != nil {
					pathStr = mopt.RESTPath.Value
				} else {
					pathStr = strcase.ToKebab(m.Name.Value)
				}
				for _, p := range m.Sig.Params {
					if IsContext(p) {
						continue
					}
					if regexp, ok := mopt.RESTPathVars[p.Name.Value]; ok {
						pathStr = stdstrings.Replace(pathStr, ":"+regexp, "", -1)
					}
				}
				if iface.Namespace != "" {
					pathStr = path.Join(iface.Namespace, pathStr)
				}
			}

			errType := config.RESTErrorType
			if g.JSONRPCEnable {
				errType = config.JRPCErrorType
			}
			if methodErrors, ok := g.IfaceErrors[iface.Named.Name.Value]; ok {
				for _, errors := range methodErrors {
					for _, e := range errors {
						if e.Type != errType {
							continue
						}
						codeStr := strconv.FormatInt(e.Code, 10)
						errResponse := &openapi.Response{
							Content: openapi.Content{
								"application/json": {
									Schema: &openapi.Schema{
										Ref: "#/components/schemas/" + e.Name,
									},
								},
							},
						}
						if e.Type == config.JRPCErrorType {
							codeStr = "x-" + codeStr
							o.Components.Schemas[e.Name] = makeOpenapiSchemaJRPCError(e.Code)
							errResponse.Description = e.Name
						} else {
							errResponse.Description = http.StatusText(int(e.Code))
							o.Components.Schemas[e.Name] = makeOpenapiSchemaRESTError()
						}

						op.Responses[codeStr] = errResponse
					}
				}
			}

			ifaceTag := iface.Named.Name.Upper()
			if iface.Namespace != "" {
				ifaceTag = iface.Namespace
			}
			tags = append(tags, ifaceTag)

			op.Description = m.Comment
			op.Tags = tags

			if _, ok := o.Paths[pathStr]; !ok {
				o.Paths[pathStr] = &openapi.Path{}
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

	for _, namedType := range g.defTypes {
		o.Components.Schemas[namedType.Name.Value] = g.schemaByType(namedType.Type)
	}

	data, _ := ffjson.Marshal(o)
	return data
}

func makeOpenapiSchemaRESTError() *openapi.Schema {
	return &openapi.Schema{
		Type: "object",
		Properties: openapi.Properties{
			"error": &openapi.Schema{
				Type: "string",
			},
		},
	}
}

func makeOpenapiSchemaJRPCError(code int64) *openapi.Schema {
	return &openapi.Schema{
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
						Example: code,
					},
					"message": &openapi.Schema{
						Type: "string",
					},
				},
			},
		},
	}
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
		}
		if st, ok := t.Type.(*option.StructType); ok {
			for _, field := range st.Fields {
				g.fillTypeDefRecursive(field.Var.Type)
			}
		}
	}
}

func (g *Openapi) fillTypeDef(t interface{}) {
	g.fillTypeDefRecursive(t)
}

func (g *Openapi) schemaByTypeRecursive(schema *openapi.Schema, t interface{}) {
	switch t := t.(type) {
	case *option.NamedType:
		switch t.Pkg.Path {
		default:
			schema.Ref = g.makeRef(t)
			return
		case "encoding/json":
			schema.Type = "object"
			schema.Properties = openapi.Properties{}
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
				name = tag.Value()
			}
			if name == "-" {
				continue
			}
			filedSchema := &openapi.Schema{
				Properties: openapi.Properties{},
			}
			schema.Properties[name] = filedSchema
			g.schemaByTypeRecursive(filedSchema, field.Var.Type)
		}
	case *option.MapType:
		mapSchema := &openapi.Schema{
			Properties: openapi.Properties{},
		}
		schema.Properties = openapi.Properties{"key": mapSchema}
		g.schemaByTypeRecursive(mapSchema, t.Value)
		return
	case *option.ArrayType:
		schema.Type = "array"
		schema.Items = &openapi.Schema{
			Properties: openapi.Properties{},
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
			schema.Items = &openapi.Schema{
				Properties: openapi.Properties{},
			}
			g.schemaByTypeRecursive(schema.Items, t.Value)
		}
		return
	case *option.IfaceType:
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
		return
	case *option.BasicType:
		if t.IsString() {
			schema.Type = "string"
			schema.Format = "string"
			schema.Example = "abc"
			return
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

func (g *Openapi) schemaByType(t interface{}) (schema *openapi.Schema) {
	schema = &openapi.Schema{
		Properties: openapi.Properties{},
	}
	g.schemaByTypeRecursive(schema, t)
	return
}

func (g *Openapi) makeJSONRPCPath(
	m *option.FuncType, prefix string, mopt config.MethodOption,
) *openapi.Operation {
	responseSchema := &openapi.Schema{
		Type:       "object",
		Properties: map[string]*openapi.Schema{},
	}
	requestSchema := &openapi.Schema{
		Type:       "object",
		Properties: map[string]*openapi.Schema{},
	}

	if LenWithoutErrors(m.Sig.Params) > 0 {
		for _, p := range m.Sig.Params {
			if IsContext(p) {
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

	lenResults := LenWithoutErrors(m.Sig.Results)

	if lenResults > 1 {
		for _, r := range m.Sig.Results {
			if IsError(r) {
				continue
			}
			g.fillTypeDef(r.Type)
			schema := g.schemaByType(r.Type)
			responseSchema.Properties[r.Name.Lower()] = schema
		}
	} else if lenResults == 1 {
		responseSchema = g.schemaByType(m.Sig.Results[0].Type)
	} else {
		responseSchema.Example = json.RawMessage("null")
	}

	if mopt.RESTWrapResponse.Value != "" {
		properties := openapi.Properties{}
		properties[mopt.RESTWrapResponse.Value] = responseSchema
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

	restMethod := m.Name.Lower()
	if prefix != "" {
		restMethod = prefix + "." + restMethod
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
				Enum: []string{restMethod},
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
		Responses: map[string]*openapi.Response{
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

func (g *Openapi) makeRestPath(m *option.FuncType, mopt config.MethodOption) *openapi.Operation {
	responseSchema := &openapi.Schema{
		Type:       "object",
		Properties: map[string]*openapi.Schema{},
	}

	requestSchema := &openapi.Schema{
		Type:       "object",
		Properties: map[string]*openapi.Schema{},
	}

	queryVars := make(map[string]queryVar, len(mopt.RESTQueryVars.Value))
	for i := 0; i < len(mopt.RESTQueryVars.Value); i += 2 {
		queryName := mopt.RESTQueryVars.Value[i]
		fieldName := mopt.RESTQueryVars.Value[i+1]
		var required bool
		if stdstrings.HasPrefix(queryName, "!") {
			queryName = queryName[1:]
			required = true
		}
		queryVars[fieldName] = queryVar{
			name:     queryName,
			required: required,
		}
	}

	queryValues := make(map[string]string, len(mopt.RESTQueryValues.Value))
	for i := 0; i < len(mopt.RESTQueryValues.Value); i += 2 {
		queryName := mopt.RESTQueryValues.Value[i]
		value := mopt.RESTQueryValues.Value[i+1]
		queryValues[queryName] = value
	}

	headerVars := make(map[string]string, len(mopt.RESTHeaderVars.Value))
	for i := 0; i < len(mopt.RESTHeaderVars.Value); i += 2 {
		headerVars[mopt.RESTHeaderVars.Value[i]] = mopt.RESTHeaderVars.Value[i+1]
	}

	for _, p := range m.Sig.Params {
		if IsContext(p) {
			continue
		}
		if _, ok := mopt.RESTPathVars[p.Name.Value]; ok {
			continue
		}
		if _, ok := queryVars[p.Name.Value]; ok {
			continue
		}
		if _, ok := headerVars[p.Name.Value]; ok {
			continue
		}

		g.fillTypeDef(p.Type)
		schema := g.schemaByType(p.Type)
		schema.Description = p.Comment
		requestSchema.Properties[p.Name.Lower()] = schema
	}

	lenResults := LenWithoutErrors(m.Sig.Results)
	if lenResults > 1 {
		for _, r := range m.Sig.Results {
			if IsError(r) {
				continue
			}

			g.fillTypeDef(r.Type)
			responseSchema.Properties[r.Name.Lower()] = g.schemaByType(r.Type)
		}
	} else if lenResults == 1 {
		responseSchema = g.schemaByType(m.Sig.Results[0].Type)
	}
	if mopt.RESTWrapResponse.Value != "" {
		properties := openapi.Properties{}
		properties[mopt.RESTWrapResponse.Value] = responseSchema
		responseSchema = &openapi.Schema{
			Properties: properties,
		}
	}
	responses := map[string]*openapi.Response{}
	if lenResults == 0 {
		responses["201"] = &openapi.Response{
			Description: "Created",
			Content: openapi.Content{
				"text/plain": {},
			},
		}
	} else {
		responses["200"] = &openapi.Response{
			Description: "OK",
			Content: openapi.Content{
				"application/json": {
					Schema: responseSchema,
				},
			},
		}
	}

	o := &openapi.Operation{
		Summary:   m.Name.Value,
		Responses: responses,
	}
	for _, p := range m.Sig.Params {
		if name, ok := mopt.RESTPathVars[p.Name.Value]; ok {
			o.Parameters = append(o.Parameters, openapi.Parameter{
				In:          "path",
				Name:        name,
				Description: p.Comment,
				//Required:    true,
				Schema: g.schemaByType(p.Type),
			})
		} else if name, ok := headerVars[p.Name.Value]; ok {
			o.Parameters = append(o.Parameters, openapi.Parameter{
				In:          "header",
				Name:        name,
				Description: p.Comment,
				//Required:    true,
				Schema: g.schemaByType(p.Type),
			})
		} else if queryVar, ok := queryVars[p.Name.Value]; ok {
			o.Parameters = append(o.Parameters, openapi.Parameter{
				In:          "query",
				Name:        queryVar.name,
				Description: p.Comment,
				Required:    queryVar.required,
				Schema:      g.schemaByType(p.Type),
			})
		}
	}
	switch mopt.RESTMethod.Value {
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

func (g *Openapi) OutputDir() string {
	return g.Output
}

func (g *Openapi) Filename() string {
	return "openapi.json"
}
