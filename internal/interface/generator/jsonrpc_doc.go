package generator

import (
	"context"
	"encoding/json"
	"fmt"
	stdtypes "go/types"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/fatih/structtag"

	"github.com/swipe-io/swipe/v2/internal/domain/model"
	"github.com/swipe-io/swipe/v2/internal/usecase/generator"
	"github.com/swipe-io/swipe/v2/internal/writer"

	"golang.org/x/tools/go/types/typeutil"
)

type NamedSlice []*stdtypes.Named

func (n NamedSlice) Len() int           { return len(n) }
func (n NamedSlice) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }
func (n NamedSlice) Less(i, j int) bool { return n[i].Obj().Name() < n[j].Obj().Name() }

var paramCommentRegexp = regexp.MustCompile(`(?s)@([a-zA-Z0-9_]*) (.*)`)

type jsonrpcDocOptionsGateway interface {
	AppID() string
	Interfaces() model.Interfaces
	JSONRPCDocOutput() string
	CommentFields() map[string]map[string]string
	Enums() *typeutil.Map
}

type jsonrpcDoc struct {
	writer.BaseWriter
	options   jsonrpcDocOptionsGateway
	workDir   string
	outputDir string
}

func (g *jsonrpcDoc) Prepare(ctx context.Context) error {
	outputDir, err := filepath.Abs(filepath.Join(g.workDir, g.options.JSONRPCDocOutput()))
	if err != nil {
		return err
	}
	g.outputDir = outputDir
	return nil
}

func (g *jsonrpcDoc) appendExistsTypes(m *typeutil.Map, tpl stdtypes.Type) {
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

func (g *jsonrpcDoc) Process(ctx context.Context) error {
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

	g.W("# %s JSONRPC Client\n\n", g.options.AppID())

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

	for i := 0; i < g.options.Interfaces().Len(); i++ {
		iface := g.options.Interfaces().At(i)

		for _, method := range iface.Methods() {
			for _, param := range method.Params {
				g.appendExistsTypes(existsTypes, param.Type())
			}
			for _, result := range method.Results {
				g.appendExistsTypes(existsTypes, result.Type())
			}
			name := method.Name
			if g.options.Interfaces().Len() > 1 {
				name = iface.NameUnExport() + "." + method.Name
			}
			g.W("<a href=\"#%[1]s\">%[1]s</a>\n\n", name)
		}

		for _, method := range iface.Methods() {
			name := method.Name
			if g.options.Interfaces().Len() > 1 {
				name = iface.NameUnExport() + "." + method.Name
			}

			g.W("### <a name=\"%[1]s\"></a>%[1]s(", name)

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

			methodComment, paramsComment := parseMethodComments(method.Comments)

			g.W("%s\n\n", strings.Replace(methodComment, method.Name, "", len(method.Name)))

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
	}

	if existsTypes.Len() > 0 {
		nameds := make(NamedSlice, 0, existsTypes.Len())
		existsTypes.Iterate(func(key stdtypes.Type, value interface{}) {
			if named, ok := key.(*stdtypes.Named); ok {
				nameds = append(nameds, named)
			}
		})
		sort.Sort(nameds)

		g.W("## Members\n\n")

		for _, named := range nameds {
			st := named.Obj().Type().Underlying().(*stdtypes.Struct)

			comments, ok := g.options.CommentFields()[named.Obj().String()]
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
	}

	if g.options.Enums().Len() > 0 {
		g.W("## Enums\n")

		g.options.Enums().Iterate(func(key stdtypes.Type, value interface{}) {
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

func (g *jsonrpcDoc) PkgName() string {
	return ""
}

func (g *jsonrpcDoc) OutputDir() string {
	return g.outputDir
}

func (g *jsonrpcDoc) Filename() string {
	return "jsonrpc_doc_gen.md"
}

func (g *jsonrpcDoc) getJSType(tpl stdtypes.Type) string {
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
			if g.options.Enums().At(v) != nil {
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

func NewJsonrpcDoc(options jsonrpcDocOptionsGateway, workDir string) generator.Generator {
	return &jsonrpcDoc{
		options: options,
		workDir: workDir,
	}
}
