package generator

import (
	"context"
	"path"
	"strconv"
	stdstrings "strings"

	"github.com/swipe-io/strcase"

	"github.com/swipe-io/swipe/v3/internal/format"
	"github.com/swipe-io/swipe/v3/internal/plugin"
	"github.com/swipe-io/swipe/v3/internal/plugin/gokit/config"
	"github.com/swipe-io/swipe/v3/option"
	"github.com/swipe-io/swipe/v3/swipe"
	"github.com/swipe-io/swipe/v3/writer"
)

type varType struct {
	p        *option.VarType
	value    string
	required bool
}

type RESTClientGenerator struct {
	w             writer.GoWriter
	Interfaces    []*config.Interface
	UseFast       bool
	MethodOptions map[string]config.MethodOptions
	Output        string
	Pkg           string
}

func (g *RESTClientGenerator) Package() string {
	return g.Pkg
}

func (g *RESTClientGenerator) Generate(ctx context.Context) []byte {
	var (
		kitHTTPPkg string
		httpPkg    string
	)
	importer := ctx.Value(swipe.ImporterKey).(swipe.Importer)

	urlPkg := importer.Import("url", "net/url")
	netPkg := importer.Import("net", "net")
	stringsPkg := importer.Import("strings", "strings")

	if g.UseFast {
		kitHTTPPkg = importer.Import("fasthttp", "github.com/l-vitaly/go-kit/transport/fasthttp")
	} else {
		kitHTTPPkg = importer.Import("http", "github.com/go-kit/kit/transport/http")
	}
	if g.UseFast {
		httpPkg = importer.Import("fasthttp", "github.com/valyala/fasthttp")
	} else {
		httpPkg = importer.Import("http", "net/http")
	}

	g.w.W("type clientErrorWrapper struct {\n")
	g.w.W("Error string `json:\"error\"`\n")
	g.w.W("Code string `json:\"code,omitempty\"`\n")
	g.w.W("Data interface{} `json:\"data,omitempty\"`\n")
	g.w.W("}\n")

	g.writeCreateReqFuncs(importer, httpPkg, urlPkg)

	for _, iface := range g.Interfaces {
		ifaceType := iface.Named.Type.(*option.IfaceType)
		clientType := ClientType(iface)

		constructPostfix := UcNameWithAppPrefix(iface)
		if len(g.Interfaces) == 1 {
			constructPostfix = ""
		}

		g.w.W("func NewClientREST%s(tgt string", constructPostfix)
		g.w.W(" ,options ...ClientOption")
		g.w.W(") (*%s, error) {\n", clientType)
		g.w.W("opts := &clientOpts{}\n")
		g.w.W("c := &%s{}\n", clientType)
		g.w.W("for _, o := range options {\n")
		g.w.W("o(opts)\n")
		g.w.W("}\n")

		g.w.W("if %s.HasPrefix(tgt, \"[\") {\n", stringsPkg)
		g.w.W("host, port, err := %s.SplitHostPort(tgt)\n", netPkg)
		g.w.WriteCheckErr("err", func() {
			g.w.W("return nil, err\n")
		})
		g.w.W("tgt = host + \":\" + port\n")
		g.w.W("}\n")

		g.w.W("u, err := %s.Parse(tgt)\n", urlPkg)

		g.w.WriteCheckErr("err", func() {
			g.w.W("return nil, err")
		})

		g.w.W("if u.Scheme == \"\" {\n")
		g.w.W("u.Scheme = \"https\"\n")
		g.w.W("}\n")

		for _, m := range ifaceType.Methods {
			mopt := g.MethodOptions[iface.Named.Name.Value+m.Name.Value]

			epName := LcNameEndpoint(iface, m)

			httpMethod := mopt.RESTMethod.Take()
			if httpMethod == "" {
				httpMethod = "GET"
			}

			g.w.W("c.%s = %s.NewClient(\n", epName, kitHTTPPkg)
			g.w.W(strconv.Quote(httpMethod))
			g.w.W(",\n")
			g.w.W("u,\n")
			g.w.W("%sReqFn,\n", LcNameIfaceMethod(iface, m))
			g.w.W("%sRespFn,\n", LcNameIfaceMethod(iface, m))
			g.w.W("append(opts.genericOpts.clientOption, opts.%sOpts.clientOption...)...,\n).Endpoint()\n", LcNameIfaceMethod(iface, m))
			g.w.W(
				"c.%[1]s = middlewareChain(append(opts.genericOpts.endpointMiddleware, opts.%[2]sOpts.endpointMiddleware...))(c.%[1]s)\n",
				epName, LcNameIfaceMethod(iface, m),
			)
		}
		g.w.W("return c, nil\n}\n\n")
	}
	return g.w.Bytes()
}

func (g *RESTClientGenerator) writeCreateReqFuncs(importer swipe.Importer, httpPkg, urlPkg string) {
	fmtPkg := importer.Import("fmt", "fmt")
	contextPkg := importer.Import("context", "context")

	for _, iface := range g.Interfaces {
		ifaceType := iface.Named.Type.(*option.IfaceType)
		for _, m := range ifaceType.Methods {
			mopt := g.MethodOptions[iface.Named.Name.Value+m.Name.Value]

			g.w.W("func %sRespFn(_ %s.Context, r *%s.Response) (response interface{}, err error) {\n", LcNameIfaceMethod(iface, m), contextPkg, httpPkg)
			statusCode := "r.StatusCode"
			if g.UseFast {
				statusCode = "r.StatusCode(r)"
			}
			g.w.W("if %s > 299 {\n", statusCode)

			if mopt.ErrorDecode.Fn != nil {
				pkgName := importer.Import(mopt.ErrorDecode.Fn.Pkg.Name, mopt.ErrorDecode.Fn.Pkg.Path)
				if pkgName != "" {
					pkgName = pkgName + "."
				}
				g.w.W("return %s%s(r)\n", pkgName, mopt.ErrorDecode.Fn.Name)
			} else {
				jsonPkg := importer.Import("json", "encoding/json")
				g.w.W("var errorData clientErrorWrapper\n")
				g.w.W("if err := %s.NewDecoder(r.Body).Decode(&errorData); err != nil {\nreturn nil, err\n}\n", jsonPkg)
				g.w.W("return nil, ")
				g.w.W("%sErrorDecode(%s, errorData.Code)", LcNameWithAppPrefix(iface)+m.Name.Value, statusCode)
			}
			g.w.W("\n}\n")

			resultsLen := plugin.LenWithoutErrors(m.Sig.Results)
			if resultsLen > 0 {
				var responseType string
				if m.Sig.IsNamed && resultsLen > 1 {
					responseType = NameResponse(m, iface)
				} else {
					responseType = swipe.TypeString(m.Sig.Results[0].Type, false, importer)
				}

				var (
					wrapData, structPath string
				)

				if mopt.RESTWrapResponse.Take() != "" {
					wrapData, structPath = wrapDataClient(stdstrings.Split(mopt.RESTWrapResponse.Take(), "."), responseType)
					g.w.W("var resp %s\n", wrapData)
				} else {
					g.w.W("var resp %s\n", responseType)
				}

				g.w.W("var b []byte\n")

				if g.UseFast {
					g.w.W("b = r.Body()\n")
				} else {
					pkgIO := importer.Import("io", "io")
					g.w.W("b, err = %s.ReadAll(r.Body)\n", pkgIO)
					g.w.WriteCheckErr("err", func() {
						g.w.W("return nil, err\n")
					})
				}

				jsonPkg := importer.Import("ffjson", "github.com/pquerna/ffjson/ffjson")

				g.w.W("if len(b) == 0 {\nreturn nil, nil\n}\n")
				g.w.W("err = %s.Unmarshal(b, &resp)\n", jsonPkg)
				g.w.W("if err != nil {\n")
				g.w.W("return nil, %s.Errorf(\"couldn't unmarshal body to %s: %%s\", err)\n", fmtPkg, responseType)
				g.w.W("}\n")

				if mopt.RESTWrapResponse.Take() != "" {
					g.w.W("return resp.%s, nil\n", structPath)
				} else {
					g.w.W("return resp, nil\n")
				}
			} else {
				g.w.W("return nil, nil\n")
			}
			g.w.W("}\n")

			g.w.W("func %sReqFn(_ %s.Context, r *%s.Request, request interface{}) error {\n", LcNameIfaceMethod(iface, m), contextPkg, httpPkg)

			nameRequest := NameRequest(m, iface)
			httpMethod := mopt.RESTMethod.Take()
			if httpMethod == "" {
				httpMethod = "GET"
			}

			bodyType := mopt.RESTBodyType.Take()
			if bodyType == "" {
				bodyType = "json"
			}

			var pathStr string
			if mopt.RESTPath.IsValid() {
				pathStr = mopt.RESTPath.Take()
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

			methodQueryVars := make(map[string]varType, len(mopt.RESTQueryVars.Value))
			for i := 0; i < len(mopt.RESTQueryVars.Value); i += 2 {
				queryName := mopt.RESTQueryVars.Value[i]
				fieldName := mopt.RESTQueryVars.Value[i+1]
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

			methodQueryValues := make(map[string]string, len(mopt.RESTQueryValues.Value))
			for i := 0; i < len(mopt.RESTQueryValues.Value); i += 2 {
				queryName := mopt.RESTQueryValues.Value[i]
				value := mopt.RESTQueryValues.Value[i+1]
				methodQueryValues[queryName] = value
			}

			methodHeaderVars := make(map[string]string, len(mopt.RESTHeaderVars.Value))
			for i := 0; i < len(mopt.RESTHeaderVars.Value); i += 2 {
				headerName := mopt.RESTHeaderVars.Value[i]
				fieldName := mopt.RESTHeaderVars.Value[i+1]
				methodHeaderVars[fieldName] = headerName
			}

			for _, p := range m.Sig.Params {
				if plugin.IsContext(p) {
					continue
				}
				if regexp, ok := mopt.RESTPathVars[p.Name.Value]; ok {
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

			paramsLen := plugin.LenWithoutContexts(m.Sig.Params)
			if paramsLen > 0 {
				g.w.W("req, ok := request.(%s)\n", nameRequest)
				g.w.W("if !ok {\n")
				g.w.W("return %s.Errorf(\"couldn't assert request as %s, got %%T\", request)\n", fmtPkg, nameRequest)
				g.w.W("}\n")
			}

			if g.UseFast {
				g.w.W("r.Header.SetMethod(")
			} else {
				g.w.W("r.Method = ")
			}
			g.w.W(strconv.Quote(httpMethod))

			if g.UseFast {
				g.w.W(")")
			}
			g.w.W("\n")

			pathVarNames := make([]string, 0, len(pathVars))
			for _, p := range pathVars {
				name := p.Name.Value + "Str"
				pathVarNames = append(pathVarNames, name)

				format.NewBuilder(importer).
					SetAssignVar(name).
					SetValueVar("req." + p.Name.Upper()).
					SetFieldType(p.Type).
					Write(&g.w)
			}
			if g.UseFast {
				g.w.W("r.URI().SetPath(")
			} else {
				g.w.W("r.URL.Path += ")
			}
			if len(pathVars) > 0 {
				g.w.W("%s.Sprintf(%s, %s)", fmtPkg, strconv.Quote(pathStr), stdstrings.Join(pathVarNames, ", "))
			} else {
				g.w.W(strconv.Quote(pathStr))
			}
			if g.UseFast {
				g.w.W(")")
			}
			g.w.W("\n")

			if len(queryVars) > 0 || len(methodQueryValues) > 0 {
				if g.UseFast {
					g.w.W("q := r.URI().QueryArgs()\n")
				} else {
					g.w.W("q := r.URL.Query()\n")
				}
				for _, p := range queryVars {
					var isPointer bool
					valueVar := "req." + strcase.ToCamel(p.Name.Value)
					name := p.Name.Value + "Str"
					if t, ok := p.Type.(*option.BasicType); ok {
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
						SetFieldType(p.Type).
						Write(&g.w)

					g.w.W("q.Add(%s, %s)\n", strconv.Quote(methodQueryVars[p.Name.Value].value), name)

					if isPointer {
						g.w.W("}\n")
					}
				}

				if len(methodQueryValues) > 0 {
					for k, v := range methodQueryValues {
						g.w.W("q.Add(%s, %s)\n", strconv.Quote(k), strconv.Quote(v))
					}
				}

				if g.UseFast {
					g.w.W("r.URI().SetQueryString(q.String())\n")
				} else {
					g.w.W("r.URL.RawQuery = q.Encode()\n")
				}
			}

			if paramsLen > 0 {
				switch stdstrings.ToUpper(httpMethod) {
				case "POST", "PUT", "PATCH":
					switch bodyType {
					case "json":
						jsonPkg := importer.Import("ffjson", "github.com/pquerna/ffjson/ffjson")
						g.w.W("r.Header.Set(\"Content-Type\", \"application/json\")\n")

						g.w.W("var reqData interface{}\n")

						if wrapRequest := mopt.RESTWrapRequest.Take(); wrapRequest != "" {
							reqData, structPath := wrapDataClient(stdstrings.Split(wrapRequest, "."), nameRequest)

							g.w.W("var wrapReq %s\n", reqData)
							g.w.W("wrapReq.%s = req\n", structPath)
							g.w.W("reqData = wrapReq\n")
						} else {
							g.w.W("reqData = request\n")
						}

						g.w.W("data, err := %s.Marshal(reqData)\n", jsonPkg)
						g.w.W("if err != nil  {\n")
						g.w.W("return %s.Errorf(\"couldn't marshal request %%T: %%s\", req, err)\n", fmtPkg)
						g.w.W("}\n")
						if g.UseFast {
							g.w.W("r.SetBody(data)\n")
						} else {
							pkgIO := importer.Import("io", "io")
							bytesPkg := importer.Import("bytes", "bytes")
							g.w.W("r.Body = %s.NopCloser(%s.NewBuffer(data))\n", pkgIO, bytesPkg)
						}
					case "urlencoded":
						pkgIO := importer.Import("io", "io")
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
						g.w.W("r.Body = %s.NopCloser(%s.NewBufferString(params.Encode()))\n", pkgIO, bytesPkg)
					case "multipart":
						bytesPkg := importer.Import("bytes", "bytes")
						multipartPkg := importer.Import("multipart", "mime/multipart")
						pkgIO := importer.Import("io", "io")

						g.w.W("body := new(%s.Buffer)\n", bytesPkg)
						g.w.W("writer := %s.NewWriter(body)\n", multipartPkg)

						for _, p := range paramVars {
							if isFileUploadType(p.Type) {
								g.w.W("part, err := writer.CreateFormFile(%s, req.%s.Name())\n", strconv.Quote(p.Name.Value), p.Name.Upper())
								g.w.WriteCheckErr("err", func() {
									g.w.W("return err\n")
								})
								g.w.W("data, err := %s.ReadAll(req.%s)\n", pkgIO, p.Name.Upper())
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

						if g.UseFast {
							g.w.W("r.SetBody(body.Bytes())\n")
						} else {
							g.w.W("r.Body = %s.NopCloser(body)\n", pkgIO)
						}
						g.w.W("r.Header.Set(\"Content-Type\", writer.FormDataContentType())\n")
					}
				}
			}
			for _, p := range headerVars {
				name := p.Name.Value + "Str"
				format.NewBuilder(importer).
					SetAssignVar(name).
					SetValueVar("req." + strcase.ToCamel(p.Name.Value)).
					SetFieldType(p.Type).
					Write(&g.w)

				g.w.W("r.Header.Add(%s, %s)\n", strconv.Quote(methodHeaderVars[p.Name.Value]), name)
			}

			g.w.W("return nil\n")
			g.w.W("}\n")
		}
	}
}

func (g *RESTClientGenerator) OutputPath() string {
	return g.Output
}

func (g *RESTClientGenerator) Filename() string {
	return "rest_client.go"
}
