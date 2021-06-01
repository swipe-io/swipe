package generator

import (
	"context"
	"path"
	"strconv"
	stdstrings "strings"

	"github.com/swipe-io/strcase"
	"github.com/swipe-io/swipe/v2/internal/plugin/gokit/config"
	"github.com/swipe-io/swipe/v2/option"
	"github.com/swipe-io/swipe/v2/swipe"
	"github.com/swipe-io/swipe/v2/writer"
)

type RESTClientGenerator struct {
	w                    writer.GoWriter
	Interfaces           []*config.Interface
	UseFast              bool
	MethodOptions        map[string]*config.MethodOption
	DefaultMethodOptions config.MethodOption
}

func (g *RESTClientGenerator) Generate(ctx context.Context) []byte {
	var (
		kitHTTPPkg string
		httpPkg    string
	)
	importer := ctx.Value(swipe.ImporterKey).(swipe.Importer)

	jsonPkg := importer.Import("ffjson", "github.com/pquerna/ffjson/ffjson")
	fmtPkg := importer.Import("fmt", "fmt")
	contextPkg := importer.Import("context", "context")
	urlPkg := importer.Import("url", "net/url")
	netPkg := importer.Import("net", "net")
	stringsPkg := importer.Import("strings", "strings")

	for _, iface := range g.Interfaces {
		ifaceType := iface.Named.Type.(*option.IfaceType)
		name := UcNameWithAppPrefix(iface)
		clientType := ClientType(iface)

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

		g.w.W("func NewClientREST%s(tgt string", name)
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
			mopt := &g.DefaultMethodOptions
			if opt, ok := g.MethodOptions[iface.Named.Name.Value+m.Name.Value]; ok {
				mopt = opt
			}

			epName := LcNameEndpoint(iface, m)

			httpMethod := mopt.RESTMethod.Value
			if httpMethod == "" {
				httpMethod = "GET"
			}

			pathStr := mopt.RESTPath.Value
			if pathStr == "" {
				pathStr = path.Join("/", strcase.ToKebab(m.Name.Value))
			}

			if iface.Namespace != "" {
				pathStr = path.Join("/", strcase.ToKebab(iface.Namespace), "/", pathStr)
			}

			var (
				pathVars      []*option.VarType
				queryVars     []*option.VarType
				headerVars    []*option.VarType
				multipartVars []*option.VarType
			)

			methodMultipartVars := make(map[string]string, len(mopt.RESTMultipart.Value))
			for i := 0; i < len(mopt.RESTMultipart.Value); i += 2 {
				methodMultipartVars[mopt.RESTMultipart.Value[i]] = mopt.RESTMultipart.Value[i+1]
			}

			methodQueryVars := make(map[string]string, len(mopt.RESTQueryVars.Value))
			methodQueryValues := make(map[string]string)
			for i := 0; i < len(mopt.RESTQueryVars.Value); i += 2 {
				fieldName := mopt.RESTQueryVars.Value[i]
				if stdstrings.HasPrefix(fieldName, ":") {
					methodQueryValues[mopt.RESTQueryVars.Value[i+1]] = fieldName[1:]
				} else {
					methodQueryVars[mopt.RESTQueryVars.Value[i]] = mopt.RESTQueryVars.Value[i+1]
				}
			}

			methodHeaderVars := make(map[string]string, len(mopt.RESTHeaderVars.Value))
			for i := 0; i < len(mopt.RESTHeaderVars.Value); i += 2 {
				methodHeaderVars[mopt.RESTHeaderVars.Value[i]] = mopt.RESTHeaderVars.Value[i+1]
			}

			for _, p := range m.Sig.Params {
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
				} else if _, ok := methodMultipartVars[p.Name.Value]; ok {
					multipartVars = append(multipartVars, p)
				}
			}

			remainingParams := len(m.Sig.Params) - (len(pathVars) + len(queryVars) + len(headerVars) + len(multipartVars))

			g.w.W("c.%s = %s.NewClient(\n", epName, kitHTTPPkg)
			g.w.W(strconv.Quote(httpMethod))
			g.w.W(",\n")
			g.w.W("u,\n")

			if mopt.ClientEncodeRequest.Value != nil {
				g.w.W(importer.TypeString(mopt.ClientEncodeRequest.Value))
			} else {
				g.w.W("func(_ %s.Context, r *%s.Request, request interface{}) error {\n", contextPkg, httpPkg)

				if len(m.Sig.Params) > 0 {
					nameRequest := NameRequest(m, iface)

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
					g.w.WriteFormatType(importer, name, "req."+p.Name.Upper(), p)
				}
				if g.UseFast {
					g.w.W("r.SetRequestURI(")
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

				if remainingParams > 0 {
					g.w.W("r.Header.Set(\"Content-Type\", \"application/json\")\n")
				}

				if len(queryVars) > 0 || len(methodQueryValues) > 0 {
					if g.UseFast {
						g.w.W("q := r.URI().QueryArgs()\n")
					} else {
						g.w.W("q := r.URL.Query()\n")
					}

					for _, p := range queryVars {
						name := p.Name.Value + "Str"
						g.w.WriteFormatType(importer, name, "req."+strcase.ToCamel(p.Name.Value), p)
						g.w.W("q.Add(%s, %s)\n", strconv.Quote(methodQueryVars[p.Name.Value]), name)
					}

					if len(methodQueryValues) > 0 {
						for k, v := range methodQueryValues {
							g.w.W("q.Add(%s, %s)\n", strconv.Quote(k), strconv.Quote(v))
						}
					}

					if len(multipartVars) > 0 {
						bytesPkg := importer.Import("bytes", "bytes")
						multipartPkg := importer.Import("multipart", "mime/multipart")
						ioutilPkg := importer.Import("ioutil", "io/ioutil")

						g.w.W("body := new(%s.Buffer)\n", bytesPkg)
						g.w.W("writer := %s.NewWriter(body)\n", multipartPkg)

						for _, p := range multipartVars {
							if isBytes(p.Type) {
								g.w.W("part, err := writer.CreateFormFile(%s, \"\")\n", strconv.Quote(methodMultipartVars[p.Name.Value]))
								g.w.WriteCheckErr("err", func() {
									g.w.W("return err\n")
								})
								g.w.W("part.Write(req.%s)\n", p.Name.Upper())
								continue
							}
							name := p.Name.Value + "Str"
							g.w.WriteFormatType(importer, name, "req."+p.Name.Upper(), p)
							g.w.W("_ = writer.WriteField(%s, %s)\n", strconv.Quote(methodMultipartVars[p.Name.Value]), name)
						}
						g.w.W("if err := writer.Close(); err != nil {\n return err\n}\n")

						if g.UseFast {
							g.w.W("r.SetBody(body.Bytes())\n")
						} else {
							g.w.W("r.Body = %s.NopCloser(body)\n", ioutilPkg)
						}
					}

					if g.UseFast {
						g.w.W("r.URI().SetQueryString(q.String())\n")
					} else {
						g.w.W("r.URL.RawQuery = q.Encode()\n")
					}
				}

				for _, p := range headerVars {
					name := p.Name.Value + "Str"
					g.w.WriteFormatType(importer, name, "req."+strcase.ToCamel(p.Name.Value), p)
					g.w.W("r.Header.Add(%s, %s)\n", strconv.Quote(methodHeaderVars[p.Name.Value]), name)
				}
				switch stdstrings.ToUpper(httpMethod) {
				case "POST", "PUT", "PATCH":
					if remainingParams > 0 {
						jsonPkg := importer.Import("ffjson", "github.com/pquerna/ffjson/ffjson")

						g.w.W("data, err := %s.Marshal(req)\n", jsonPkg)
						g.w.W("if err != nil  {\n")
						g.w.W("return %s.Errorf(\"couldn't marshal request %%T: %%s\", req, err)\n", fmtPkg)
						g.w.W("}\n")

						if g.UseFast {
							g.w.W("r.SetBody(data)\n")
						} else {
							ioutilPkg := importer.Import("ioutil", "io/ioutil")
							bytesPkg := importer.Import("bytes", "bytes")

							g.w.W("r.Body = %s.NopCloser(%s.NewBuffer(data))\n", ioutilPkg, bytesPkg)
						}
					}
				}
				g.w.W("return nil\n")
				g.w.W("}")
			}

			g.w.W(",\n")

			if mopt.ClientDecodeResponse.Value != nil {
				g.w.W(importer.TypeString(mopt.ClientDecodeResponse.Value))
			} else {
				g.w.W("func(_ %s.Context, r *%s.Response) (interface{}, error) {\n", contextPkg, httpPkg)
				statusCode := "r.StatusCode"
				if g.UseFast {
					statusCode = "r.StatusCode()"
				}
				g.w.W("if %s > 299 {\n", statusCode)
				g.w.W("return nil, %sErrorDecode(%s)\n", LcNameWithAppPrefix(iface)+m.Name.Value, statusCode)
				g.w.W("}\n")

				if len(m.Sig.Results) > 0 {
					var responseType string
					if m.Sig.IsNamed {
						responseType = NameResponse(m, iface)
					} else {
						responseType = importer.TypeString(m.Sig.Results[0].Type)
					}
					if mopt.RESTWrapResponse.Value != "" {
						g.w.W("var resp struct {\nData %s `json:\"%s\"`\n}\n", responseType, mopt.RESTWrapResponse.Value)
					} else {
						g.w.W("var resp %s\n", responseType)
					}

					g.w.W("var b []byte\n")

					if g.UseFast {
						g.w.W("b = r.Body()\n")
					} else {
						ioutilPkg := importer.Import("ioutil", "io/ioutil")
						g.w.W("b, err = %s.ReadAll(r.Body)\n", ioutilPkg)
						g.w.WriteCheckErr("err", func() {
							g.w.W("return nil, err\n")
						})
					}

					g.w.W("if len(b) == 0 {\nreturn nil, nil\n}\n")

					g.w.W("err = %s.Unmarshal(b, ", jsonPkg)

					g.w.W("&resp)\n")

					g.w.W("if err != nil {\n")
					g.w.W("return nil, %s.Errorf(\"couldn't unmarshal body to %s: %%s\", err)\n", fmtPkg, responseType)
					g.w.W("}\n")

					if mopt.RESTWrapResponse.Value != "" {
						g.w.W("return resp.Data, nil\n")
					} else {
						g.w.W("return resp, nil\n")
					}
				} else {
					g.w.W("return nil, nil\n")
				}

				g.w.W("}")
			}

			g.w.W(",\n")

			g.w.W("append(opts.genericClientOption, opts.%sClientOption...)...,\n", LcNameWithAppPrefix(iface)+m.Name.Value)

			g.w.W(").Endpoint()\n")

			g.w.W(
				"c.%[1]s = middlewareChain(append(opts.genericEndpointMiddleware, opts.%[1]sMiddleware...))(c.%[1]s)\n",
				epName,
			)
		}
		g.w.W("return c, nil\n")
		g.w.W("}\n\n")
	}

	return g.w.Bytes()
}

func (g *RESTClientGenerator) OutputDir() string {
	return ""
}

func (g *RESTClientGenerator) Filename() string {
	return "rest_client.go"
}
