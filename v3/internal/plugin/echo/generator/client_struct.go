package generator

import (
	"context"
	"path"
	"strconv"
	"strings"
	stdstrings "strings"

	"github.com/swipe-io/swipe/v3/internal/format"

	"github.com/swipe-io/strcase"
	"github.com/swipe-io/swipe/v3/internal/plugin"
	"github.com/swipe-io/swipe/v3/internal/plugin/echo/config"
	"github.com/swipe-io/swipe/v3/option"
	"github.com/swipe-io/swipe/v3/swipe"
	"github.com/swipe-io/swipe/v3/writer"
)

type varType struct {
	p        *option.VarType
	value    string
	required bool
}

type ClientStruct struct {
	w             writer.GoWriter
	Interfaces    []*config.Interface
	MethodOptions map[string]config.MethodOptions
	Output        string
	Pkg           string
}

func (g *ClientStruct) Generate(ctx context.Context) []byte {
	importer := ctx.Value(swipe.ImporterKey).(swipe.Importer)

	httpPkg := importer.Import("http", "net/http")
	jsonPkg := importer.Import("json", "encoding/json")
	fmtPkg := importer.Import("fmt", "fmt")

	if len(g.Interfaces) > 1 {
		g.w.W("type AppClient struct {\n")

		for _, iface := range g.Interfaces {
			g.w.W("%s *%s\n", UcNameWithAppPrefix(iface), ClientType(iface))
		}
		g.w.W("}\n\n")

		g.w.W("func NewClientREST(tgt string")

		g.w.W(" ,opts ...ClientOption")
		g.w.W(") (*AppClient, error) {\n")

		for _, iface := range g.Interfaces {
			name := UcNameWithAppPrefix(iface)
			lcName := LcNameWithAppPrefix(iface)
			g.w.W("%s, err := NewClientREST%s(tgt, opts...)\n", lcName, name)
			g.w.WriteCheckErr("err", func() {
				g.w.W("return nil, err")
			})
		}

		g.w.W("return &AppClient{\n")
		for _, iface := range g.Interfaces {
			g.w.W("%[1]s: %[2]s,\n", UcNameWithAppPrefix(iface), LcNameWithAppPrefix(iface))
		}
		g.w.W("}, nil\n")
		g.w.W("}\n\n")
	}

	if len(g.Interfaces) > 0 {
		urlPkg := importer.Import("url", "net/url")

		for _, iface := range g.Interfaces {
			ifaceType := iface.Named.Type.(*option.IfaceType)

			clientType := ClientType(iface)
			g.w.W("type %s struct {\nu *%s.URL}\n\n", clientType, urlPkg)

			for _, m := range ifaceType.Methods {
				mOpt := g.MethodOptions[iface.Named.Name.Value+m.Name.Value]

				httpMethod := mOpt.RESTMethod.Take()
				if httpMethod == "" {
					httpMethod = "GET"
				}

				bodyType := mOpt.RESTBodyType.Take()
				if bodyType == "" {
					bodyType = "json"
				}

				var pathStr string
				if mOpt.RESTPath.IsValid() {
					pathStr = mOpt.RESTPath.Take()
				} else {
					pathStr = path.Join("/", strcase.ToKebab(m.Name.Value))
				}

				if iface.Namespace != "" {
					pathStr = path.Join("/", strcase.ToKebab(iface.Namespace), "/", pathStr)
				}

				var (
					pathVars   []*option.VarType
					queryVars  []*option.VarType
					headerVars []*option.VarType
					paramVars  []*option.VarType
				)

				methodQueryVars := make(map[string]varType, len(mOpt.RESTQueryVars.Value))
				for i := 0; i < len(mOpt.RESTQueryVars.Value); i += 2 {
					queryName := mOpt.RESTQueryVars.Value[i]
					fieldName := mOpt.RESTQueryVars.Value[i+1]
					var required bool
					if stdstrings.HasPrefix(queryName, "!") {
						queryName = queryName[1:]
						required = true
					}
					methodQueryVars[fieldName] = varType{
						value:    queryName,
						required: required,
					}
				}

				methodQueryValues := make(map[string]string, len(mOpt.RESTQueryValues.Value))
				for i := 0; i < len(mOpt.RESTQueryValues.Value); i += 2 {
					queryName := mOpt.RESTQueryValues.Value[i]
					value := mOpt.RESTQueryValues.Value[i+1]
					methodQueryValues[queryName] = value
				}

				methodHeaderVars := make(map[string]string, len(mOpt.RESTHeaderVars.Value))
				for i := 0; i < len(mOpt.RESTHeaderVars.Value); i += 2 {
					headerName := mOpt.RESTHeaderVars.Value[i]
					fieldName := mOpt.RESTHeaderVars.Value[i+1]
					methodHeaderVars[fieldName] = headerName
				}

				for _, p := range m.Sig.Params {
					if plugin.IsContext(p) {
						continue
					}
					if regexp, ok := mOpt.RESTPathVars[p.Name.Value]; ok {
						if regexp != "" {
							regexp = ":" + regexp
						}
						pathStr = stdstrings.Replace(pathStr, "{"+p.Name.Value+regexp+"}", "%s", -1)
						pathVars = append(pathVars, p)
					} else if _, ok := methodQueryVars[p.Name.Value]; ok {
						queryVars = append(queryVars, p)
					} else if _, ok := methodHeaderVars[p.Name.Value]; ok {
						headerVars = append(headerVars, p)
					} else {
						paramVars = append(paramVars, p)
					}
				}

				resultLen := plugin.LenWithoutErrors(m.Sig.Results)

				//var responseType string
				//if m.Sig.IsNamed && resultsLen > 1 {
				//	responseType = NameResponse(m, iface)
				//} else {
				//	responseType = swipe.TypeString(m.Sig.Results[0].Type, false, importer)
				//}

				g.w.W("func (c *%s) %s %s {\n", clientType, m.Name.Value, swipe.TypeString(m.Sig, false, importer))

				if resultLen > 0 {
					g.w.W("var response ")

					if m.Sig.IsNamed && resultLen > 1 {
						g.w.W("struct{\n")
						for _, param := range m.Sig.Results {
							if plugin.IsError(param) {
								continue
							}
							g.w.W("%s %s `json:\"%s\"`\n", param.Name.Upper(), swipe.TypeString(param.Type, false, importer), param.Name)
						}
						g.w.W("}\n")
					} else {
						g.w.W("%s\n", swipe.TypeString(m.Sig.Results[0].Type, false, importer))
					}
				}

				pathVarNames := make([]string, 0, len(pathVars))
				for _, p := range pathVars {
					name := p.Name.Value + "Str"
					pathVarNames = append(pathVarNames, name)
					format.NewBuilder(importer).
						SetAssignVar(name).
						SetValueVar(p.Name.Value).
						SetFieldType(p.Type).
						Write(&g.w)
				}

				if len(paramVars) > 0 {
					g.w.W("var req = map[string]interface{}{\n")
					for _, param := range paramVars {
						if plugin.IsContext(param) {
							continue
						}
						g.w.W("%s: %s,", strconv.Quote(param.Name.Value), param.Name.Value)
					}
					g.w.W("\n}\n\n")
				}

				g.w.W("r, err := %s.NewRequest(%s, ", httpPkg, strconv.Quote(httpMethod))

				g.w.W("c.u.String()+")

				if len(pathVars) > 0 {
					g.w.W("%s.Sprintf(%s, %s)", fmtPkg, strconv.Quote(pathStr), stdstrings.Join(pathVarNames, ", "))
				} else {
					g.w.W(strconv.Quote(pathStr))
				}
				g.w.W(", nil)\n")

				g.w.WriteCheckErr("err", func() {
					g.w.W("return\n")
				})

				if len(queryVars) > 0 || len(methodQueryValues) > 0 {
					g.w.W("q := r.URL.Query()\n")

					for _, queryVar := range queryVars {
						if named, ok := queryVar.Type.(*option.NamedType); ok {
							if st, ok := named.Type.(*option.StructType); ok {
								for _, field := range st.Fields {
									var isPointer bool
									valueVar := field.Var.Name.Value
									name := field.Var.Name.Lower() + "Str"
									if t, ok := field.Var.Type.(*option.BasicType); ok {
										if t.IsPointer {
											isPointer = true
										}
									}
									if isPointer {
										g.w.W("if %s != nil {\n", valueVar)
									}

									format.NewBuilder(importer).
										SetAssignVar(name).
										SetValueVar(queryVar.Name.Value + "." + valueVar).
										SetFieldType(field.Var.Type).
										Write(&g.w)

									g.w.W("q.Add(%s, %s)\n", strconv.Quote(field.Var.Name.Lower()), name)

									if isPointer {
										g.w.W("}\n")
									}
								}
							}
						} else {
							var isPointer bool
							valueVar := queryVar.Name.Value
							name := queryVar.Name.Value + "Str"
							if t, ok := queryVar.Type.(*option.BasicType); ok {
								if t.IsPointer {
									isPointer = true
								}
							}
							if isPointer {
								g.w.W("if %s != nil {\n", valueVar)
							}

							format.NewBuilder(importer).
								SetAssignVar(name).
								SetValueVar(valueVar).
								SetFieldType(queryVar.Type).
								Write(&g.w)

							g.w.W("q.Add(%s, %s)\n", strconv.Quote(methodQueryVars[queryVar.Name.Value].value), name)

							if isPointer {
								g.w.W("}\n")
							}
						}
					}
					if len(methodQueryValues) > 0 {
						for k, v := range methodQueryValues {
							g.w.W("q.Add(%s, %s)\n", strconv.Quote(k), strconv.Quote(v))
						}
					}
					g.w.W("r.URL.RawQuery = q.Encode()\n")
				}

				paramsLen := plugin.LenWithoutContexts(m.Sig.Params)
				if paramsLen > 0 {
					switch stdstrings.ToUpper(httpMethod) {
					case "POST", "PUT", "PATCH":
						switch bodyType {
						case "json":
							g.w.W("r.Header.Set(\"Content-Type\", \"application/json\")\n")

							if wrapRequest := mOpt.RESTWrapRequest.Take(); wrapRequest != "" {
								reqData, structPath := wrapDataClient(stdstrings.Split(wrapRequest, "."), "test")
								g.w.W("var wrapReq %s\n", reqData)
								g.w.W("wrapReq.%s = req\n", structPath)
								g.w.W("data, err := %s.Marshal(wrapReq)\n", jsonPkg)
							} else {
								g.w.W("data, err := %s.Marshal(req)\n", jsonPkg)
							}

							g.w.WriteCheckErr("err", func() {
								g.w.W("err = %s.Errorf(\"couldn't marshal request %%T: %%s\", req, err)\nreturn\n", fmtPkg)
							})

							ioutilPkg := importer.Import("ioutil", "io/ioutil")
							bytesPkg := importer.Import("bytes", "bytes")
							g.w.W("r.Body = %s.NopCloser(%s.NewBuffer(data))\n", ioutilPkg, bytesPkg)

						case "urlencoded":
							ioutilPkg := importer.Import("ioutil", "io/ioutil")
							urlPkg := importer.Import("url", "url")

							bytesPkg := importer.Import("bytes", "bytes")
							g.w.W("r.Header.Add(\"Content-Type\", \"application/x-www-form-urlencoded; charset=utf-8\")\n")
							g.w.W("params := %s.Values{}\n", urlPkg)
							for _, p := range paramVars {
								name := p.Name.Value + "Str"

								format.NewBuilder(importer).
									SetAssignVar(name).
									SetValueVar("req." + p.Name.Upper()).
									SetFieldType(p.Type).
									Write(&g.w)

								g.w.W("params.Set(\"data\", %s)\n", name)
							}
							g.w.W("r.Body = %s.NopCloser(%s.NewBufferString(params.Encode()))\n", ioutilPkg, bytesPkg)
						case "multipart":
							bytesPkg := importer.Import("bytes", "bytes")
							multipartPkg := importer.Import("multipart", "mime/multipart")
							ioutilPkg := importer.Import("ioutil", "io/ioutil")

							g.w.W("body := new(%s.Buffer)\n", bytesPkg)
							g.w.W("writer := %s.NewWriter(body)\n", multipartPkg)

							for _, p := range paramVars {
								if isFileUploadType(p.Type) {
									g.w.W("part, err := writer.CreateFormFile(%s, req.%s.Name())\n", strconv.Quote(p.Name.Value), p.Name.Upper())
									g.w.WriteCheckErr("err", func() {
										g.w.W("return err\n")
									})
									g.w.W("data, err := %s.ReadAll(req.%s)\n", ioutilPkg, p.Name.Upper())
									g.w.WriteCheckErr("err", func() {
										g.w.W("return err\n")
									})
									g.w.W("part.Write(data)\n")
									continue
								}
								name := p.Name.Value + "Str"

								format.NewBuilder(importer).
									SetAssignVar(name).
									SetValueVar("req." + p.Name.Upper()).
									SetFieldType(p.Type).
									Write(&g.w)

								g.w.W("_ = writer.WriteField(%s, %s)\n", strconv.Quote(p.Name.Value), name)
							}
							g.w.W("if err := writer.Close(); err != nil {\n return err\n}\n")

							g.w.W("r.Body = %s.NopCloser(body)\n", ioutilPkg)

							g.w.W("r.Header.Set(\"Content-Type\", writer.FormDataContentType())\n")
						}
					}
				}

				for _, p := range headerVars {
					name := p.Name.Value + "Str"
					format.NewBuilder(importer).
						SetAssignVar(name).
						SetValueVar(p.Name.Value).
						SetFieldType(p.Type).
						Write(&g.w)

					g.w.W("r.Header.Add(%s, %s)\n", strconv.Quote(methodHeaderVars[p.Name.Value]), name)
				}

				g.w.W("resp, err := %s.DefaultClient.Do(r)\n", httpPkg)
				g.w.WriteCheckErr("err", func() {
					g.w.W("return\n")
				})

				g.w.W("if resp.StatusCode > 299 {\n")
				g.w.W("err = %sErrorDecode(resp.StatusCode, resp.Status)\n", LcNameIfaceMethod(iface, m))
				g.w.W("return\n")
				g.w.W("}\n")

				if resultLen > 0 {
					g.w.W("err = %s.NewDecoder(resp.Body).Decode(&response)\n", jsonPkg)
					g.w.WriteCheckErr("err", func() {
						g.w.W("return\n")
					})
					if m.Sig.IsNamed && resultLen > 1 {
						var results []string
						for _, result := range m.Sig.Results {
							if plugin.IsError(result) {
								continue
							}
							results = append(results, "response."+result.Name.Upper())
						}
						g.w.W("return %s, nil\n", strings.Join(results, ","))
					} else {
						g.w.W("return response, nil\n")
					}
				} else {
					g.w.W("return\n")
				}

				g.w.W("}\n\n")
			}
		}
	}
	return g.w.Bytes()
}

func (g *ClientStruct) Package() string {
	return g.Pkg
}

func (g *ClientStruct) OutputPath() string {
	return g.Output
}

func (g *ClientStruct) Filename() string {
	return "client_struct.go"
}
