package generator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	stdtypes "go/types"
	"path/filepath"
	"strconv"
	stdstrings "strings"

	"github.com/iancoleman/strcase"
	"github.com/pquerna/ffjson/ffjson"

	"github.com/swipe-io/swipe/pkg/domain/model"
	"github.com/swipe-io/swipe/pkg/openapi"
	"github.com/swipe-io/swipe/pkg/strings"
	"github.com/swipe-io/swipe/pkg/types"
)

type openapiDoc struct {
	bytes.Buffer
	info      model.GenerateInfo
	o         model.ServiceOption
	outputDir string
}

func (g *openapiDoc) Process(ctx context.Context) error {
	opt := g.o.Transport.Openapi

	outputDir, err := filepath.Abs(filepath.Join(g.info.BasePath, opt.Output))
	if err != nil {
		return err
	}

	g.outputDir = outputDir

	swg := openapi.OpenAPI{
		OpenAPI: "3.0.0",
		Info:    opt.Info,
		Servers: opt.Servers,
		Paths:   map[string]*openapi.Path{},
		Components: openapi.Components{
			Schemas: openapi.Schemas{},
		},
	}

	if g.o.Transport.JsonRPC.Enable {
		swg.Components.Schemas = getOpenapiJSONRPCErrorSchemas()
	} else {
		swg.Components.Schemas["Error"] = getOpenapiRestErrorSchema()
	}

	for name, ei := range g.o.Transport.MapCodeErrors {
		var s *openapi.Schema
		if g.o.Transport.JsonRPC.Enable {
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
								Example: ei.Code,
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

	for _, m := range g.o.Methods {
		mopt := g.o.Transport.MethodOptions[m.Name]

		var (
			o       *openapi.Operation
			pathStr string
			errors  = opt.DefaultMethod.Errors
			tags    = opt.DefaultMethod.Tags
		)

		if openapiMethodOpt, ok := opt.Methods[m.Name]; ok {
			errors = append(errors, openapiMethodOpt.Errors...)
			tags = append(tags, openapiMethodOpt.Tags...)
		}

		if g.o.Transport.JsonRPC.Enable {
			o = g.makeJSONRPCPath(opt, m)
			pathStr = "/" + strings.LcFirst(m.Name)
			mopt.MethodName = "POST"
			for _, name := range errors {
				if ei, ok := g.o.Transport.MapCodeErrors[name]; ok {
					codeStr := strconv.FormatInt(ei.Code, 10)
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
			o = g.makeRestPath(opt, m)
			pathStr = mopt.Path
			for _, regexp := range mopt.PathVars {
				pathStr = stdstrings.Replace(pathStr, ":"+regexp, "", -1)
			}
		}

		o.Tags = tags

		if _, ok := swg.Paths[pathStr]; !ok {
			swg.Paths[pathStr] = &openapi.Path{}
		}

		switch mopt.MethodName {
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
	if err := ffjson.NewEncoder(g).Encode(swg); err != nil {
		return err
	}
	return nil
}

func (g *openapiDoc) PkgName() string {
	return ""
}

func (g *openapiDoc) OutputDir() string {
	return g.outputDir
}

func (g *openapiDoc) Filename() string {
	typeName := "rest"
	if g.o.Transport.JsonRPC.Enable {
		typeName = "jsonrpc"
	}
	return fmt.Sprintf("openapi_%s.json", typeName)
}

func (g *openapiDoc) Imports() []string {
	return nil
}

func (g *openapiDoc) makeJSONRPCPath(opt model.OpenapiHTTPTransportOption, m model.ServiceMethod) *openapi.Operation {
	mopt := g.o.Transport.MethodOptions[m.Name]

	responseSchema := &openapi.Schema{
		Type:       "object",
		Properties: map[string]*openapi.Schema{},
	}

	requestSchema := &openapi.Schema{
		Type:       "object",
		Properties: map[string]*openapi.Schema{},
	}

	if len(m.Params) > 0 {
		for _, p := range m.Params {
			requestSchema.Properties[strcase.ToLowerCamel(p.Name())] = g.makeSwaggerSchema(p.Type())
		}
	} else {
		requestSchema.Example = json.RawMessage("null")
	}

	if len(m.Results) > 1 {
		for _, r := range m.Results {
			responseSchema.Properties[strcase.ToLowerCamel(r.Name())] = g.makeSwaggerSchema(r.Type())
		}
	} else if len(m.Results) == 1 {
		responseSchema = g.makeSwaggerSchema(m.Results[0].Type())
	} else {
		responseSchema.Example = json.RawMessage("null")
	}

	if mopt.WrapResponse.Enable {
		properties := openapi.Properties{}
		properties[mopt.WrapResponse.Name] = responseSchema
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
				Enum: []string{strcase.ToLowerCamel(m.Name)},
			},
			"params": requestSchema,
		},
	}

	return &openapi.Operation{
		Description: stdstrings.Join(m.Comments, "\n"),
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

func (g *openapiDoc) makeSwaggerSchema(t stdtypes.Type) (schema *openapi.Schema) {
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
		schema.Properties = openapi.Properties{
			"string": g.makeSwaggerSchema(v.Elem()),
		}
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

func (g *openapiDoc) makeRestPath(opt model.OpenapiHTTPTransportOption, m model.ServiceMethod) *openapi.Operation {
	mopt := g.o.Transport.MethodOptions[m.Name]

	responseSchema := &openapi.Schema{
		Type:       "object",
		Properties: map[string]*openapi.Schema{},
	}

	requestSchema := &openapi.Schema{
		Type:       "object",
		Properties: map[string]*openapi.Schema{},
	}

	for _, p := range m.Params {
		if _, ok := mopt.PathVars[p.Name()]; ok {
			continue
		}
		if _, ok := mopt.QueryVars[p.Name()]; ok {
			continue
		}
		if _, ok := mopt.HeaderVars[p.Name()]; ok {
			continue
		}
		if types.IsContext(p.Type()) {
			continue
		}
		requestSchema.Properties[strcase.ToLowerCamel(p.Name())] = g.makeSwaggerSchema(p.Type())
	}

	if len(m.Results) > 1 {
		for _, r := range m.Results {
			responseSchema.Properties[strcase.ToLowerCamel(r.Name())] = g.makeSwaggerSchema(r.Type())
		}
	} else if len(m.Results) == 1 {
		responseSchema = g.makeSwaggerSchema(m.Results[0].Type())
	}

	if mopt.WrapResponse.Enable {
		properties := openapi.Properties{}
		properties[mopt.WrapResponse.Name] = responseSchema
		responseSchema = &openapi.Schema{
			Properties: properties,
		}
	}

	o := &openapi.Operation{
		Summary: m.Name,
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

	for name := range mopt.PathVars {
		var schema *openapi.Schema
		if fld := m.Params.LookupField(name); fld != nil {
			schema = g.makeSwaggerSchema(fld.Type())
		}
		o.Parameters = append(o.Parameters, openapi.Parameter{
			In:       "path",
			Name:     name,
			Required: true,
			Schema:   schema,
		})
	}

	for argName, name := range mopt.QueryVars {
		var schema *openapi.Schema
		if fld := m.Params.LookupField(argName); fld != nil {
			schema = g.makeSwaggerSchema(fld.Type())
		}
		o.Parameters = append(o.Parameters, openapi.Parameter{
			In:     "query",
			Name:   name,
			Schema: schema,
		})
	}

	for argName, name := range mopt.HeaderVars {
		var schema *openapi.Schema
		if fld := m.Params.LookupField(argName); fld != nil {
			schema = g.makeSwaggerSchema(fld.Type())
		}
		o.Parameters = append(o.Parameters, openapi.Parameter{
			In:     "header",
			Name:   name,
			Schema: schema,
		})
	}

	switch mopt.MethodName {
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

func NewOpenapi(info model.GenerateInfo, o model.ServiceOption) Generator {
	return &openapiDoc{info: info, o: o}
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
