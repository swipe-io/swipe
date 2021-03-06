package generator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	stdstrings "strings"

	"github.com/pquerna/ffjson/ffjson"
	"github.com/swipe-io/strcase"
	"github.com/swipe-io/swipe/v2/internal/domain/model"
	iftypevisitor "github.com/swipe-io/swipe/v2/internal/interface/typevisitor"

	"github.com/swipe-io/swipe/v2/internal/openapi"
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

type openapiDocOptionsGateway interface {
	Interfaces() model.Interfaces
	MethodOption(m model.ServiceMethod) model.MethodOption
	JSONRPCEnable() bool
	OpenapiOutput() string
	OpenapiInfo() openapi.Info
	OpenapiServers() []openapi.Server
	OpenapiMethodTags(name string) []string
	OpenapiDefaultMethodTags() []string
}

type openapiDoc struct {
	bytes.Buffer
	options   openapiDocOptionsGateway
	workDir   string
	outputDir string
}

func (g *openapiDoc) Prepare(ctx context.Context) error {
	outputDir, err := filepath.Abs(filepath.Join(g.workDir, g.options.OpenapiOutput()))
	if err != nil {
		return err
	}
	g.outputDir = outputDir
	return nil
}

func (g *openapiDoc) Process(ctx context.Context) error {
	swg := openapi.OpenAPI{
		OpenAPI: "3.0.0",
		Info:    g.options.OpenapiInfo(),
		Servers: g.options.OpenapiServers(),
		Paths:   map[string]*openapi.Path{},
		Components: openapi.Components{
			Schemas: openapi.Schemas{},
		},
	}

	ntc := iftypevisitor.NewNamedTypeCollector()

	if g.options.JSONRPCEnable() {
		swg.Components.Schemas = getOpenapiJSONRPCErrorSchemas()
	} else {
		swg.Components.Schemas["Error"] = getOpenapiRESTErrorSchema()
	}

	for i := 0; i < g.options.Interfaces().Len(); i++ {
		iface := g.options.Interfaces().At(i)

		for _, method := range iface.Methods() {
			for _, e := range method.Errors {
				var s *openapi.Schema
				if g.options.JSONRPCEnable() {
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
										Example: e.Code,
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
				swg.Components.Schemas[e.Named.Obj().Name()] = s
			}
		}
	}

	for i := 0; i < g.options.Interfaces().Len(); i++ {
		iface := g.options.Interfaces().At(i)
		for _, m := range iface.Methods() {
			mopt := g.options.MethodOption(m)

			var (
				o       *openapi.Operation
				pathStr string
			)

			methodTags := g.options.OpenapiMethodTags(m.Name)
			tags := append(g.options.OpenapiDefaultMethodTags(), methodTags...)

			methodName := mopt.MethodName
			methodComment, paramsComment := parseMethodComments(m.Comments)
			methodComment = stdstrings.Replace(methodComment, m.Name, "", len(m.Name))

			var prefix string
			if iface.Namespace() != "" {
				prefix = iface.Namespace()
			}

			if g.options.JSONRPCEnable() {
				o = g.makeJSONRPCPath(m, ntc, paramsComment, prefix)
				pathStr = "/" + m.LcName
				if prefix != "" {
					pathStr = "/" + prefix + "." + m.LcName
				}
				methodName = "POST"
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
				o = g.makeRestPath(m, ntc, paramsComment)
				pathStr = strcase.ToKebab(m.Name)
				if mopt.Path != "" {
					pathStr = stdstrings.TrimLeft(mopt.Path, "/")
				}
				for _, p := range m.Params {
					if regexp, ok := mopt.PathVars[p.Name()]; ok {
						pathStr = stdstrings.Replace(pathStr, ":"+regexp, "", -1)
					}
				}
				if iface.Namespace() != "" {
					pathStr = iface.Namespace() + "/" + pathStr
				}
				pathStr = "/" + stdstrings.TrimLeft(pathStr, "/")
				for _, ei := range m.Errors {
					codeStr := strconv.FormatInt(ei.Code, 10)
					o.Responses[codeStr] = openapi.Response{
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
			}

			ifaceTag := strcase.ToLowerCamel(iface.UcName())
			if iface.Namespace() != "" {
				ifaceTag = iface.Namespace()
			}
			tags = append(tags, ifaceTag)

			o.Description = methodComment
			o.Tags = tags

			if _, ok := swg.Paths[pathStr]; !ok {
				swg.Paths[pathStr] = &openapi.Path{}
			}

			switch methodName {
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
	if g.options.JSONRPCEnable() {
		typeName = "jsonrpc"
	}
	return fmt.Sprintf("openapi_%s_gen.json", typeName)
}

func (g *openapiDoc) Imports() []string {
	return nil
}

func (g *openapiDoc) makeJSONRPCPath(
	m model.ServiceMethod, ntc ustypevisitor.NamedTypeCollector, paramsComment map[string]string, prefix string,
) *openapi.Operation {
	mopt := g.options.MethodOption(m)
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
			ntc.Visit(p.Type())

			schema := &openapi.Schema{}
			iftypevisitor.OpenapiVisitor(schema).Visit(p.Type())

			schema.Description = paramsComment[p.Name()]
			requestSchema.Properties[strcase.ToLowerCamel(p.Name())] = schema
		}
		if m.ParamVariadic != nil {
			schema := &openapi.Schema{}
			iftypevisitor.OpenapiVisitor(schema).Visit(m.ParamVariadic.Type())
			requestSchema.Properties[strcase.ToLowerCamel(m.ParamVariadic.Name())] = schema
		}
	} else {
		requestSchema.Type = "object"
		requestSchema.Nullable = true
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

	methodName := strcase.ToLowerCamel(m.Name)
	if prefix != "" {
		methodName = prefix + "." + methodName
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
				Enum: []string{methodName},
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

func (g *openapiDoc) makeRestPath(m model.ServiceMethod, ntc ustypevisitor.NamedTypeCollector, paramsComment map[string]string) *openapi.Operation {
	mopt := g.options.MethodOption(m)
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

		ntc.Visit(p.Type())

		schema := &openapi.Schema{}
		iftypevisitor.OpenapiVisitor(schema).Visit(p.Type())

		schema.Description = paramsComment[p.Name()]

		requestSchema.Properties[strcase.ToLowerCamel(p.Name())] = schema
	}
	if m.ParamVariadic != nil {
		schema := &openapi.Schema{}
		iftypevisitor.OpenapiVisitor(schema).Visit(m.ParamVariadic.Type())
		requestSchema.Properties[strcase.ToLowerCamel(m.ParamVariadic.Name())] = schema
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
	}

	if mopt.WrapResponse.Enable {
		properties := openapi.Properties{}
		properties[mopt.WrapResponse.Name] = responseSchema
		responseSchema = &openapi.Schema{
			Properties: properties,
		}
	}

	responses := map[string]openapi.Response{}

	if len(m.Results) == 0 {
		responses["201"] = openapi.Response{
			Description: "Created",
			Content: openapi.Content{
				"text/plain": {},
			},
		}
	} else {
		responses["200"] = openapi.Response{
			Description: "OK",
			Content: openapi.Content{
				"application/json": {
					Schema: responseSchema,
				},
			},
		}
	}

	responses["500"] = openapi.Response{
		Description: "Internal Server Error",
		Content: openapi.Content{
			"application/json": {
				Schema: &openapi.Schema{
					Ref: "#/components/schemas/Error",
				},
			},
		},
	}

	o := &openapi.Operation{
		Summary:   m.Name,
		Responses: responses,
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
			schema := &openapi.Schema{}
			iftypevisitor.OpenapiVisitor(schema).Visit(p.Type())
			o.Parameters = append(o.Parameters, openapi.Parameter{
				In:       in,
				Name:     p.Name(),
				Required: true,
				Schema:   schema,
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
	options openapiDocOptionsGateway,
	workDir string,
) generator.Generator {
	return &openapiDoc{
		options: options,
		workDir: workDir,
	}
}
