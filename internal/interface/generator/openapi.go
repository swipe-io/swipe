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

	"github.com/pquerna/ffjson/ffjson"

	"github.com/swipe-io/strcase"
	"github.com/swipe-io/swipe/v2/internal/domain/model"
	iftypevisitor "github.com/swipe-io/swipe/v2/internal/interface/typevisitor"
	"github.com/swipe-io/swipe/v2/internal/openapi"
	"github.com/swipe-io/swipe/v2/internal/strings"
	"github.com/swipe-io/swipe/v2/internal/types"
	"github.com/swipe-io/swipe/v2/internal/usecase/generator"
	ustypevisitor "github.com/swipe-io/swipe/v2/internal/usecase/typevisitor"
)

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

type openapiDoc struct {
	bytes.Buffer
	serviceMethods []model.ServiceMethod
	transport      model.TransportOption
	workDir        string
	outputDir      string
}

func (g *openapiDoc) Prepare(ctx context.Context) error {
	outputDir, err := filepath.Abs(filepath.Join(g.workDir, g.transport.Openapi.Output))
	if err != nil {
		return err
	}
	g.outputDir = outputDir
	return nil
}

func (g *openapiDoc) Process(ctx context.Context) error {
	opt := g.transport.Openapi
	swg := openapi.OpenAPI{
		OpenAPI: "3.0.0",
		Info:    opt.Info,
		Servers: opt.Servers,
		Paths:   map[string]*openapi.Path{},
		Components: openapi.Components{
			Schemas: openapi.Schemas{},
		},
	}

	ntc := iftypevisitor.NewNamedTypeCollector()

	if g.transport.JsonRPC.Enable {
		swg.Components.Schemas = getOpenapiJSONRPCErrorSchemas()
	} else {
		swg.Components.Schemas["Error"] = getOpenapiRestErrorSchema()
	}

	for _, ei := range g.transport.Errors {
		var s *openapi.Schema
		if g.transport.JsonRPC.Enable {
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
		swg.Components.Schemas[ei.Named.Obj().Name()] = s
	}

	for _, m := range g.serviceMethods {
		mopt := g.transport.MethodOptions[m.Name]

		var (
			o       *openapi.Operation
			pathStr string
			tags    = opt.DefaultMethod.Tags
		)

		if openapiMethodOpt, ok := opt.Methods[m.Name]; ok {
			tags = append(tags, openapiMethodOpt.Tags...)
		}

		if g.transport.JsonRPC.Enable {
			o = g.makeJSONRPCPath(m, ntc)
			pathStr = "/" + strings.LcFirst(m.Name)
			mopt.MethodName = "POST"

			for _, ei := range m.Errors {
				codeStr := strconv.FormatInt(ei.Code, 10)
				o.Responses["x"+codeStr] = openapi.Response{
					Description: ei.Named.Obj().Name(),
					Content: openapi.Content{
						"application/json": {
							Schema: &openapi.Schema{
								Ref: "#/components/schemas/" + ei.Named.Obj().Name(),
							},
						},
					},
				}
			}
		} else {
			o = g.makeRestPath(opt, m)
			pathStr = mopt.Path
			for _, p := range m.Params {
				if regexp, ok := mopt.PathVars[p.Name()]; ok {
					pathStr = stdstrings.Replace(pathStr, ":"+regexp, "", -1)
				}
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

	for _, t := range ntc.TypeDefs() {
		schema := &openapi.Schema{}
		iftypevisitor.OpenapiDefVisitor(schema).Visit(t)
		swg.Components.Schemas[t.Obj().Name()] = schema
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
	if g.transport.JsonRPC.Enable {
		typeName = "jsonrpc"
	}
	return fmt.Sprintf("openapi_%s_gen.json", typeName)
}

func (g *openapiDoc) Imports() []string {
	return nil
}

func (g *openapiDoc) makeJSONRPCPath(m model.ServiceMethod, ntc ustypevisitor.NamedTypeCollector) *openapi.Operation {
	mopt := g.transport.MethodOptions[m.Name]

	responseSchema := &openapi.Schema{
		Type:       "object",
		Properties: map[string]*openapi.Schema{},
	}

	requestSchema := &openapi.Schema{
		Type:       "object",
		Properties: map[string]*openapi.Schema{},
	}

	//ntc := typevisitor.NewNamedTypeCollector()

	if len(m.Params) > 0 {
		for _, p := range m.Params {
			ntc.Visit(p.Type())

			schema := &openapi.Schema{}
			iftypevisitor.OpenapiVisitor(schema).Visit(p.Type())
			requestSchema.Properties[strcase.ToLowerCamel(p.Name())] = schema
		}
	} else {
		requestSchema.Example = json.RawMessage("null")
	}

	if len(m.Results) > 1 {
		for _, r := range m.Results {
			ntc.Visit(r.Type())
			schema := &openapi.Schema{}
			iftypevisitor.OpenapiVisitor(schema).Visit(r.Type())
			responseSchema.Properties[strcase.ToLowerCamel(r.Name())] = schema
		}
	} else if len(m.Results) == 1 {
		ntc.Visit(m.Results[0].Type())
		responseSchema = &openapi.Schema{}
		iftypevisitor.OpenapiVisitor(responseSchema).Visit(m.Results[0].Type())
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

func (g *openapiDoc) makeSwaggerSchema(tpl stdtypes.Type) (schema *openapi.Schema) {
	schema = &openapi.Schema{}
	//switch v := tpl.(type) {
	//case *stdtypes.Pointer:
	//	return g.makeSwaggerSchema(v.Elem())
	//case *stdtypes.Interface:
	//	// TODO: not anyOf works in SwaggerUI, so the object type is used to display the field.
	//	schema.Type = "object"
	//	schema.Description = "Can be any value - string, number, boolean, array or object."
	//	schema.Properties = openapi.Properties{}
	//	schema.Example = json.RawMessage("null")
	//	schema.AnyOf = []openapi.Schema{
	//		{Type: "string", Example: "abc"},
	//		{Type: "integer", Example: 1},
	//		{Type: "number", Format: "float", Example: 1.11},
	//		{Type: "boolean", Example: true},
	//		{Type: "array"},
	//		{Type: "object"},
	//	}
	//case *stdtypes.Map:
	//	schema.Type = "object"
	//	schema.Properties = openapi.Properties{
	//		"key": g.makeSwaggerSchema(v.Elem()),
	//	}
	//case *stdtypes.Slice:
	//	if vv, ok := v.Elem().(*stdtypes.Basic); ok && vv.Kind() == stdtypes.Byte {
	//		schema.Type = "string"
	//		schema.Format = "byte"
	//		schema.Example = "U3dhZ2dlciByb2Nrcw=="
	//	} else {
	//		schema.Type = "array"
	//		schema.Items = g.makeSwaggerSchema(v.Elem())
	//	}
	//case *stdtypes.Basic:
	//	switch v.Kind() {
	//	case stdtypes.String:
	//		schema.Type = "string"
	//		schema.Format = "string"
	//		schema.Example = "abc"
	//	case stdtypes.Bool:
	//		schema.Type = "boolean"
	//		schema.Example = true
	//	case stdtypes.Int,
	//		stdtypes.Uint,
	//		stdtypes.Uint8,
	//		stdtypes.Uint16,
	//		stdtypes.Int8,
	//		stdtypes.Int16:
	//		schema.Type = "integer"
	//		schema.Example = 1
	//	case stdtypes.Uint32, stdtypes.Int32:
	//		schema.Type = "integer"
	//		schema.Format = "int32"
	//		schema.Example = 1
	//	case stdtypes.Uint64, stdtypes.Int64:
	//		schema.Type = "integer"
	//		schema.Format = "int64"
	//		schema.Example = 1
	//	case stdtypes.Float32, stdtypes.Float64:
	//		schema.Type = "number"
	//		schema.Format = "float"
	//		schema.Example = 1.11
	//	}
	//case *stdtypes.Struct:
	//	schema.Type = "object"
	//	schema.Properties = openapi.Properties{}
	//
	//	var populateSchema func(st *stdtypes.Struct)
	//	populateSchema = func(st *stdtypes.Struct) {
	//		for i := 0; i < st.NumFields(); i++ {
	//			f := st.Field(i)
	//			if !f.Embedded() {
	//				name := f.Name()
	//				if tags, err := structtag.Parse(st.Tag(i)); err == nil {
	//					if tag, err := tags.Get("json"); err == nil {
	//						name = tag.Value()
	//					}
	//				}
	//				schema.Properties[name] = g.makeSwaggerSchema(f.Type())
	//			} else {
	//				var st *stdtypes.Struct
	//				if ptr, ok := f.Type().(*stdtypes.Pointer); ok {
	//					st = ptr.Elem().Underlying().(*stdtypes.Struct)
	//				} else {
	//					st = f.Type().Underlying().(*stdtypes.Struct)
	//				}
	//				populateSchema(st)
	//			}
	//		}
	//	}
	//	populateSchema(v)
	//case *stdtypes.Named:
	//	switch stdtypes.TypeString(v, nil) {
	//	case "encoding/json.RawMessage":
	//		schema.Type = "object"
	//		schema.Properties = openapi.Properties{}
	//		return
	//	case "time.Time":
	//		schema.Type = "string"
	//		schema.Format = "date-time"
	//		schema.Example = "1985-04-02T01:30:00.00Z"
	//		return
	//	case "github.com/pborman/uuid.UUID",
	//		"github.com/google/uuid.UUID":
	//		schema.Type = "string"
	//		schema.Format = "uuid"
	//		schema.Example = "d5c02d83-6fbc-4dd7-8416-9f85ed80de46"
	//		return
	//	}
	//	return g.makeSwaggerSchema(v.Obj().Type().Underlying())
	//}
	return
}

func (g *openapiDoc) makeRestPath(opt model.OpenapiHTTPTransportOption, m model.ServiceMethod) *openapi.Operation {
	mopt := g.transport.MethodOptions[m.Name]

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
	for _, p := range m.Params {
		var in string
		if _, ok := mopt.PathVars[p.Name()]; ok {
			in = "path"
		} else if _, ok := mopt.HeaderVars[p.Name()]; ok {
			in = "header"
		} else if _, ok := mopt.QueryVars[p.Name()]; ok {
			in = "query"
		}
		if in != "" {
			o.Parameters = append(o.Parameters, openapi.Parameter{
				In:       in,
				Name:     p.Name(),
				Required: true,
				Schema:   g.makeSwaggerSchema(p.Type()),
			})
		}
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

func NewOpenapi(
	serviceMethods []model.ServiceMethod,
	transport model.TransportOption,
	workDir string,
) generator.Generator {
	return &openapiDoc{
		serviceMethods: serviceMethods,
		transport:      transport,
		workDir:        workDir,
	}
}
