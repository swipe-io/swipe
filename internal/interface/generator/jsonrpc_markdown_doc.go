package generator

import (
	"context"
	"encoding/json"
	"fmt"
	stdtypes "go/types"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/swipe-io/swipe/v2/internal/usecase/generator"

	"github.com/fatih/structtag"
	"github.com/swipe-io/swipe/v2/internal/domain/model"
	"github.com/swipe-io/swipe/v2/internal/writer"
	"golang.org/x/tools/go/types/typeutil"
)

var paramCommentRegexp = regexp.MustCompile(`(?s)@([a-zA-Z0-9_]*) (.*)`)

type jsonrpcMarkdownDoc struct {
	writer.BaseWriter
	serviceID      string
	serviceMethods []model.ServiceMethod
	transport      model.TransportOption
	commentMap     *typeutil.Map
	enums          *typeutil.Map
	workDir        string
	outputDir      string
}

func (g *jsonrpcMarkdownDoc) Prepare(ctx context.Context) error {
	outputDir, err := filepath.Abs(filepath.Join(g.workDir, g.transport.MarkdownDoc.OutputDir))
	if err != nil {
		return err
	}
	g.outputDir = outputDir
	return nil
}

func (g *jsonrpcMarkdownDoc) appendExistsTypes(m *typeutil.Map, tpl stdtypes.Type) {
	tpl = normalizeType(tpl)
	if isGolangNamedType(tpl) {
		return
	}
	if v := m.At(tpl); v != nil {
		return
	}
	if named, ok := tpl.(*stdtypes.Named); ok {
		if st, ok := named.Obj().Type().Underlying().(*stdtypes.Struct); ok {
			m.Set(named, st)
			for i := 0; i < st.NumFields(); i++ {
				f := st.Field(i)
				if !f.Embedded() {
					g.appendExistsTypes(m, f.Type())
				}
			}
		}
	}
}

func (g *jsonrpcMarkdownDoc) Process(ctx context.Context) error {
	var pkgImport string
	pkgJsonFilepath := filepath.Join(g.workDir, "package.json")
	data, err := ioutil.ReadFile(pkgJsonFilepath)
	if err == nil {
		var packageJSON map[string]interface{}
		if err := json.Unmarshal(data, &packageJSON); err != nil {
			return err
		}
		if name, ok := packageJSON["name"].(string); ok {
			pkgImport = name
		}
	}

	g.W("# %s JSONRPC Client\n\n", g.serviceID)

	if pkgImport != "" {
		g.W("## Getting Started\n\n")
		g.W("You can install this with:\n\n```shell script\nnpm install --save-dev %s\n```\n\n", pkgImport)
		g.W("Import the package with the client:\n\n")
		g.W("```javascript\nimport API from \"%s\"\n```\n\n", pkgImport)
		g.W("Create a transport, only one method needs to be implemented: `doRequest(Array.<Object>) PromiseLike<Object>`.\n\n")
		g.W("For example:\n\n```javascript\nclass FetchTransport {\n    constructor(url) {\n      this.url = url;\n    }\n\n    doRequest(requests) {\n        return fetch(this.url, {method: \"POST\", body: JSON.stringify(requests)})\n    }\n}\n```\n\n")
		g.W("Now for a complete example:\n\n```javascript\nimport API from \"%s\"\nimport Transport from \"transport\"\n\nconst api = new API(new Transport(\"http://127.0.0.1\"))\n\n// call method here.\n```\n\n", pkgImport)
		g.W("## API\n## Methods\n\n")
	}

	existsTypes := new(typeutil.Map)

	for _, method := range g.serviceMethods {
		for _, param := range method.Params {
			g.appendExistsTypes(existsTypes, param.Type())
		}
		for _, result := range method.Results {
			g.appendExistsTypes(existsTypes, result.Type())
		}
		g.W("<a href=\"#%[1]s\">%[1]s</a>\n\n", method.Name)
	}

	for _, method := range g.serviceMethods {
		g.W("### <a name=\"%[1]s\"></a> %[1]s(", method.Name)

		for i, param := range method.Params {
			if i > 0 {
				g.W(", ")
			}
			g.W("%s", param.Name())
		}

		g.W(") ⇒")

		if len(method.Results) > 0 {
			if len(method.Results) == 1 {
				g.W("<code>%s</code>", g.getJSType(method.Results[0].Type()))
			}
		} else {
			g.W("<code>void</code>")
		}

		g.W("\n\n")

		paramsComment := make(map[string]string, len(method.Params))
		for _, comment := range method.Comments {
			comment = strings.TrimSpace(comment)
			if strings.HasPrefix(comment, "@") {
				matches := paramCommentRegexp.FindAllStringSubmatch(comment, -1)
				if len(matches) == 1 && len(matches[0]) == 3 {
					paramsComment[matches[0][1]] = matches[0][2]
				}
				continue
			}
			g.W("%s\n\n", strings.Replace(comment, method.Name, "", len(method.Name)))
		}

		g.W("\n\n")

		g.W("**Throws**:\n\n")

		for _, e := range method.Errors {
			g.W("<code>%sException</code>\n\n", e.Named.Obj().Name())
		}

		g.W("\n\n")

		if len(method.Params) > 0 {
			g.W("| Param | Type | Description |\n|------|------|------|\n")
			for _, param := range method.Params {
				comment := paramsComment[param.Name()]
				g.W("|%s|<code>%s</code>|%s|\n", param.Name(), g.getJSType(param.Type()), comment)
			}
		}
	}

	if existsTypes.Len() > 0 {
		g.W("## Members\n\n")

		existsTypes.Iterate(func(key stdtypes.Type, value interface{}) {
			if named, ok := key.(*stdtypes.Named); ok {
				st := named.Obj().Type().Underlying().(*stdtypes.Struct)
				comments, ok := g.commentMap.At(st).(map[string]string)
				if !ok {
					comments = map[string]string{}
				}

				g.W("### %s\n\n", named.Obj().Name())

				g.W("| Field | Type | Description |\n|------|------|------|\n")

				var writeStructInfo func(st *stdtypes.Struct)
				writeStructInfo = func(st *stdtypes.Struct) {
					for i := 0; i < st.NumFields(); i++ {
						f := st.Field(i)
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
						if tags, err := structtag.Parse(st.Tag(i)); err == nil {
							if tag, err := tags.Get("json"); err == nil {
								name = tag.Name
							}
						}
						if !f.Embedded() {
							g.W("|%s|<code>%s</code>|%s|\n", name, g.getJSType(f.Type()), comments[f.Name()])
						} else {
							var st *stdtypes.Struct
							if ptr, ok := f.Type().(*stdtypes.Pointer); ok {
								st = ptr.Elem().Underlying().(*stdtypes.Struct)
							} else {
								st = f.Type().Underlying().(*stdtypes.Struct)
							}
							writeStructInfo(st)
						}
					}
				}
				writeStructInfo(st)
			}
		})
	}

	if g.enums.Len() > 0 {
		g.W("## Enums\n")
		g.enums.Iterate(func(key stdtypes.Type, value interface{}) {
			if named, ok := key.(*stdtypes.Named); ok {
				typeName := ""
				if b, ok := named.Obj().Type().Underlying().(*stdtypes.Basic); ok {
					switch b.Info() {
					default:
						typeName = "string"
					case stdtypes.IsUnsigned | stdtypes.IsInteger, stdtypes.IsInteger:
						typeName = "number"
					}
				}
				g.W("### <a name=\"%[1]s\"></a> %[1]sEnum <code>%[2]s</code>\n\n", named.Obj().Name(), typeName)
				g.W("| Name | Value | Description |\n|------|------|------|\n")
				for _, enum := range value.([]model.Enum) {
					g.W("|%s|<code>%s</code>|%s|\n", enum.Name, enum.Value, "")
				}
			}
		})
	}

	return nil
}

func (g *jsonrpcMarkdownDoc) PkgName() string {
	return ""
}

func (g *jsonrpcMarkdownDoc) OutputDir() string {
	return g.outputDir
}

func (g *jsonrpcMarkdownDoc) Filename() string {
	return "jsonrpc_doc_gen.md"
}

func (g *jsonrpcMarkdownDoc) getJSType(tpl stdtypes.Type) string {
	switch v := tpl.(type) {
	default:
		return ""
	case *stdtypes.Interface:
		return "Object"
	case *stdtypes.Pointer:
		return g.getJSType(v.Elem())
	case *stdtypes.Array:
		return fmt.Sprintf("Array.&lt;%s&gt;", g.getJSType(v.Elem()))
	case *stdtypes.Slice:
		return fmt.Sprintf("Array.&lt;%s&gt;", g.getJSType(v.Elem()))
	case *stdtypes.Map:
		return fmt.Sprintf("Object.&lt;string, %s&gt;", g.getJSType(v.Elem()))
	case *stdtypes.Named:
		switch stdtypes.TypeString(v.Obj().Type(), nil) {
		default:
			var postfix string
			if g.enums.At(v) != nil {
				postfix = "Enum"
			}
			return fmt.Sprintf("<a href=\"#%[1]s\">%[1]s%[2]s</a>", v.Obj().Name(), postfix)
		case "encoding/json.RawMessage":
			return "*"
		case "github.com/pborman/uuid.UUID",
			"github.com/google/uuid.UUID":
			return "string"
		case "time.Time", "time.Location":
			return "string"
		}
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

func NewJsonrpcMarkdownDoc(
	serviceID string,
	serviceMethods []model.ServiceMethod,
	transport model.TransportOption,
	commentMap *typeutil.Map,
	enums *typeutil.Map,
	workDir string,
) generator.Generator {
	return &jsonrpcMarkdownDoc{
		serviceID:      serviceID,
		serviceMethods: serviceMethods,
		transport:      transport,
		commentMap:     commentMap,
		enums:          enums,
		workDir:        workDir,
	}
}
