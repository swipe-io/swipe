package generator

import (
	"context"
	"fmt"
	"path"
	"strconv"
	stdstrings "strings"

	"github.com/swipe-io/strcase"

	"github.com/swipe-io/swipe/v2/internal/option"
	"github.com/swipe-io/swipe/v2/internal/plugin/gokit/config"
	"github.com/swipe-io/swipe/v2/internal/swipe"
	"github.com/swipe-io/swipe/v2/internal/writer"
)

type RESTServerGenerator struct {
	w                    writer.GoWriter
	UseFast              bool
	JSONRPCEnable        bool
	DefaultErrorEncoder  *option.FuncType
	Interfaces           []*config.Interface
	MethodOptions        map[string]*config.MethodOption
	DefaultMethodOptions config.MethodOption
}

func (g *RESTServerGenerator) Generate(ctx context.Context) []byte {
	var (
		routerPkg  string
		httpPkg    string
		kitHTTPPkg string
	)

	importer := ctx.Value(swipe.ImporterKey).(swipe.Importer)

	jsonPkg := importer.Import("ffjson", "github.com/pquerna/ffjson/ffjson")
	contextPkg := importer.Import("context", "context")

	if g.UseFast {
		httpPkg = importer.Import("fasthttp", "github.com/valyala/fasthttp")
		kitHTTPPkg = importer.Import("fasthttp", "github.com/l-vitaly/go-kit/transport/fasthttp")
		routerPkg = importer.Import("routing", "github.com/qiangxue/fasthttp-routing")
	} else {
		routerPkg = importer.Import("mux", "github.com/gorilla/mux")
		kitHTTPPkg = importer.Import("http", "github.com/go-kit/kit/transport/http")
		httpPkg = importer.Import("http", "net/http")
	}
	if g.DefaultErrorEncoder == nil {
		g.writeDefaultErrorEncoder(contextPkg, httpPkg, kitHTTPPkg, jsonPkg)
	}
	g.writeEncodeResponseFunc(contextPkg, httpPkg, jsonPkg)

	g.w.W("// MakeHandlerREST make REST HTTP transport\n")
	g.w.W("func MakeHandlerREST(")
	for i, iface := range g.Interfaces {
		typeStr := NameInterface(iface.Named)
		if i > 0 {
			g.w.W(",")
		}
		g.w.W("svc%s %s", iface.Named.Name.Origin, typeStr)
	}
	g.w.W(", options ...ServerOption")
	g.w.W(") (")
	if g.UseFast {
		g.w.W("%s.RequestHandler", importer.Import("fasthttp", "github.com/valyala/fasthttp"))
	} else {
		g.w.W("%s.Handler", importer.Import("http", "net/http"))
	}
	g.w.W(", error) {\n")

	g.w.W("opts := &serverOpts{}\n")
	g.w.W("for _, o := range options {\n o(opts)\n }\n")
	if g.DefaultErrorEncoder != nil {
		g.w.W("opts.genericServerOption = append(opts.genericServerOption, %s.ServerErrorEncoder(", kitHTTPPkg)
		g.w.W(importer.TypeString(g.DefaultErrorEncoder))
		g.w.W("))\n")
	} else {
		g.w.W("opts.genericServerOption = append(opts.genericServerOption, %s.ServerErrorEncoder(defaultErrorEncoder))\n", kitHTTPPkg)
	}

	for _, iface := range g.Interfaces {
		g.w.W("%[1]s := Make%[2]sEndpointSet(svc%[2]s)\n", NameEndpointSetName(iface.Named), iface.Named.Name.Origin)
	}

	for _, iface := range g.Interfaces {
		ifaceType := iface.Named.Type.(*option.IfaceType)
		epSetName := NameEndpointSetName(iface.Named)
		for _, m := range ifaceType.Methods {
			g.w.W(
				"%[3]s.%[2]sEndpoint = middlewareChain(append(opts.genericEndpointMiddleware, opts.%[1]sEndpointMiddleware...))(%[3]s.%[2]sEndpoint)\n",
				LcNameWithAppPrefix(iface.Named)+m.Name.Origin, m.Name, epSetName,
			)
		}
	}
	if g.UseFast {
		g.w.W("r := %s.New()\n", routerPkg)
	} else {
		g.w.W("r := %s.NewRouter()\n", routerPkg)
	}

	for _, iface := range g.Interfaces {
		ifaceType := iface.Named.Type.(*option.IfaceType)
		epSetName := NameEndpointSetName(iface.Named)
		for _, m := range ifaceType.Methods {
			mopt := &g.DefaultMethodOptions
			if opt, ok := g.MethodOptions[iface.Named.Name.Origin+m.Name.Origin]; ok {
				mopt = opt
			}

			queryVars := make(map[string]string, len(mopt.RESTQueryVars.Value))
			for i := 0; i < len(mopt.RESTQueryVars.Value); i += 2 {
				queryVars[mopt.RESTQueryVars.Value[i]] = mopt.RESTQueryVars.Value[i+1]
			}

			headerVars := make(map[string]string, len(mopt.RESTHeaderVars.Value))
			for i := 0; i < len(mopt.RESTQueryVars.Value); i += 2 {
				headerVars[mopt.RESTHeaderVars.Value[i]] = mopt.RESTHeaderVars.Value[i+1]
			}

			var urlPath string
			if mopt.RESTPath.Value != "" {
				urlPath = stdstrings.TrimLeft(mopt.RESTPath.Value, "/")
			} else {
				urlPath = strcase.ToKebab(m.Name.Origin)
			}
			if iface.Namespace != "" {
				urlPath = path.Join(iface.Namespace, urlPath)
			}

			if g.UseFast {
				g.w.W("r.To(")
				if mopt.RESTMethod.Value != "" {
					g.w.W(strconv.Quote(mopt.RESTMethod.Value))
				} else {
					g.w.W(strconv.Quote("GET"))
				}

				g.w.W(", ")

				// replace brace indices for fasthttp router
				urlPath = stdstrings.ReplaceAll(mopt.RESTPath.Value, "{", "<")
				urlPath = stdstrings.ReplaceAll(urlPath, "}", ">")

				g.w.W(strconv.Quote(urlPath))

				g.w.W(", ")
			} else {
				g.w.W("r.Methods(")
				if mopt.RESTMethod.Value != "" {
					g.w.W(strconv.Quote(mopt.RESTMethod.Value))
				} else {
					g.w.W(strconv.Quote("GET"))
				}
				g.w.W(").")
				g.w.W("Path(")

				g.w.W(strconv.Quote(urlPath))

				g.w.W(").")
				g.w.W("Handler(")
			}
			g.w.W(
				"%s.NewServer(\n%s.%sEndpoint,\n",
				kitHTTPPkg,
				epSetName,
				m.Name,
			)
			if mopt.ServerDecodeRequest.Value != nil {
				g.w.W(importer.TypeString(mopt.ServerDecodeRequest.Value))
			} else {
				g.w.W("func(ctx %s.Context, r *%s.Request) (interface{}, error) {\n", contextPkg, httpPkg)

				nameRequest := NameRequest(m, iface.Named)

				if len(m.Sig.Params) > 0 {
					g.w.W("var req %s\n", nameRequest)
					switch stdstrings.ToUpper(mopt.RESTMethod.Value) {
					case "POST", "PUT", "PATCH":
						jsonPkg := importer.Import("ffjson", "github.com/pquerna/ffjson/ffjson")
						fmtPkg := importer.Import("fmt", "fmt")
						pkgIO := importer.Import("io", "io")
						if g.UseFast {
							g.w.W("err := %s.Unmarshal(r.Body(), &req)\n", jsonPkg)
						} else {
							ioutilPkg := importer.Import("ioutil", "io/ioutil")

							g.w.W("b, err := %s.ReadAll(r.Body)\n", ioutilPkg)
							g.w.WriteCheckErr("err", func() {
								g.w.W("return nil, %s.Errorf(\"couldn't read body for %s: %%w\", err)\n", fmtPkg, nameRequest)
							})
							g.w.W("err = %s.Unmarshal(b, &req)\n", jsonPkg)
						}
						g.w.W("if err != nil && err != %s.EOF {\n", pkgIO)
						g.w.W("return nil, %s.Errorf(\"couldn't unmarshal body to %s: %%w\", err)\n", fmtPkg, nameRequest)
						g.w.W("}\n")
					}
					if len(mopt.RESTPathVars) > 0 {
						if g.UseFast {
							fmtPkg := importer.Import("fmt", "fmt")

							g.w.W("vars, ok := ctx.Value(%s.ContextKeyRouter).(*%s.Context)\n", kitHTTPPkg, routerPkg)
							g.w.W("if !ok {\n")
							g.w.W("return nil, %s.Errorf(\"couldn't assert %s.ContextKeyRouter to *%s.Context\")\n", fmtPkg, kitHTTPPkg, routerPkg)
							g.w.W("}\n")
						} else {
							g.w.W("vars := %s.Vars(r)\n", routerPkg)
						}
					}
					if len(mopt.RESTQueryVars.Value) > 0 {
						if g.UseFast {
							g.w.W("q := r.URI().QueryArgs()\n")
						} else {
							g.w.W("q := r.URL.Query()\n")
						}
					}

					for _, p := range m.Sig.Params {
						if _, ok := mopt.RESTPathVars[p.Name.Origin]; ok {
							var valueID string
							if g.UseFast {
								valueID = "vars.Param(" + strconv.Quote(p.Name.Origin) + ")"
							} else {
								valueID = "vars[" + strconv.Quote(p.Name.Origin) + "]"
							}
							g.w.WriteConvertType(importer, "req."+strcase.ToCamel(p.Name.Origin), valueID, p, []string{"nil"}, "", false, "")
						} else if queryName, ok := queryVars[p.Name.Origin]; ok {
							var valueID string
							if g.UseFast {
								valueID = "string(q.Peek(" + strconv.Quote(queryName) + "))"
							} else {
								valueID = "q.Get(" + strconv.Quote(queryName) + ")"
							}

							tmpID := "tmp" + p.Name.Origin
							g.w.W("%s := %s\n", tmpID, valueID)

							g.w.W("if %s != \"\" {\n", tmpID)
							g.w.WriteConvertType(importer, "req."+p.Name.UpperCase, tmpID, p, []string{"nil"}, "", false, "")
							g.w.W("}\n")

						} else if headerName, ok := headerVars[p.Name.Origin]; ok {
							var valueID string
							if g.UseFast {
								valueID = "string(r.Header.Peek(" + strconv.Quote(headerName) + "))"
							} else {
								valueID = "r.Header.Get(" + strconv.Quote(headerName) + ")"
							}
							g.w.WriteConvertType(importer, "req."+p.Name.UpperCase, valueID, p, []string{"nil"}, "", false, "")
						}
					}
					g.w.W("return req, nil\n")
				} else {
					g.w.W("return nil, nil\n")
				}
				g.w.W("}")
			}
			g.w.W(",\n")

			if mopt.ServerEncodeResponse.Value != nil {
				g.w.W(importer.TypeString(mopt.ServerEncodeResponse.Value))
			} else {
				if g.JSONRPCEnable {
					g.w.W("encodeResponseJSONRPC")
				} else {
					if mopt.RESTWrapResponse.Value != "" {
						var responseWriterType string
						if g.UseFast {
							responseWriterType = fmt.Sprintf("*%s.Response", httpPkg)
						} else {
							responseWriterType = fmt.Sprintf("%s.ResponseWriter", httpPkg)
						}
						g.w.W("func (ctx context.Context, w %s, response interface{}) error {\n", responseWriterType)
						g.w.W("return encodeResponseHTTP(ctx, w, map[string]interface{}{\"%s\": response})\n", mopt.RESTWrapResponse.Value)
						g.w.W("}")
					} else {
						g.w.W("encodeResponseHTTP")
					}
				}
			}
			g.w.W(",\n")

			g.w.W("append(opts.genericServerOption, opts.%sServerOption...)...,\n", iface.Named.Name.LowerCase+m.Name.Origin)
			g.w.W(")")

			if g.UseFast {
				g.w.W(".RouterHandle()")
			}
			g.w.W(")\n")
		}
	}
	if g.UseFast {
		g.w.W("return r.HandleRequest, nil\n")
	} else {
		g.w.W("return r, nil\n")
	}
	g.w.W("}\n\n")

	return g.w.Bytes()
}

func (g *RESTServerGenerator) OutputDir() string {
	return ""
}

func (g *RESTServerGenerator) Filename() string {
	return "rest.go"
}

func (g *RESTServerGenerator) writeEncodeResponseFunc(contextPkg, httpPkg, jsonPkg string) {
	g.w.W("func encodeResponseHTTP(ctx %s.Context, ", contextPkg)
	if g.UseFast {
		g.w.W("w *%s.Response", httpPkg)
	} else {
		g.w.W("w %s.ResponseWriter", httpPkg)
	}
	g.w.W(", response interface{}) (err error) {\n")
	g.w.W("contentType := \"application/json; charset=utf-8\"\n")
	g.w.W("statusCode := 200\n")
	if g.UseFast {
		g.w.W("h := w.Header\n")
	} else {
		g.w.W("h := w.Header()\n")
	}
	g.w.W("var data []byte\n")
	g.w.W("if response != nil {\n")
	g.w.W("data, err = %s.Marshal(response)\n", jsonPkg)
	g.w.W("if err != nil {\n")
	g.w.W("return err\n")
	g.w.W("}\n")
	g.w.W("} else {\n")
	g.w.W("contentType = \"text/plain; charset=utf-8\"\n")
	g.w.W("statusCode = 201\n")
	g.w.W("}\n")
	g.w.W("h.Set(\"Content-Type\", contentType)\n")
	if g.UseFast {
		g.w.W("w.SetStatusCode(statusCode)\n")
	} else {
		g.w.W("w.WriteHeader(statusCode)\n")
	}
	if g.UseFast {
		g.w.W("w.SetBody(data)\n")
	} else {
		g.w.W("w.Write(data)\n")
	}
	g.w.W("return nil\n")
	g.w.W("}\n\n")
}

func (g *RESTServerGenerator) writeDefaultErrorEncoder(contextPkg string, httpPkg string, kitHTTPPkg string, jsonPkg string) {
	g.w.W("type errorWrapper struct {\n")
	g.w.W("Error string `json:\"error\"`\n")
	g.w.W("Data interface{} `json:\"data,omitempty\"`\n")
	g.w.W("}\n")

	g.w.W("func defaultErrorEncoder(ctx %s.Context, err error, ", contextPkg)
	if g.UseFast {
		g.w.W("w *%s.RequestCtx) {\n", httpPkg)
	} else {
		g.w.W("w %s.ResponseWriter) {\n", httpPkg)
	}

	g.w.W("var errData interface{}\n")
	g.w.W("if e, ok := err.(interface{ ErrorData() interface{} }); ok {\n")
	g.w.W("errData = e.ErrorData()\n")
	g.w.W("}\n")

	g.w.W("data, merr := %s.Marshal(errorWrapper{Error: err.Error(), Data: errData})\n", jsonPkg)
	g.w.W("if merr != nil {\n")
	if g.UseFast {
		g.w.W("w.SetBody([]byte(")
	} else {
		g.w.W("_, _ = w.Write([]byte(")
	}
	g.w.W("%s))\n", strconv.Quote("unexpected error"))
	g.w.W("return\n")
	g.w.W("}\n")

	if g.UseFast {
		g.w.W("w.Response.Header")
	} else {
		g.w.W("w.Header()")
	}
	g.w.W(".Set(\"Content-Type\", \"application/json; charset=utf-8\")\n")

	g.w.W("if headerer, ok := err.(%s.Headerer); ok {\n", kitHTTPPkg)

	if g.UseFast {
		g.w.W("for k, v := range headerer.Headers() {\n")
		g.w.W("w.Response.Header.Add(k, v)")
		g.w.W("}\n")
	} else {
		g.w.W("for k, values := range headerer.Headers() {\n")
		g.w.W("for _, v := range values {\n")
		g.w.W("w.Header().Add(k, v)")
		g.w.W("}\n}\n")
	}
	g.w.W("}\n")
	g.w.W("code := %s.StatusInternalServerError\n", httpPkg)
	g.w.W("if sc, ok := err.(%s.StatusCoder); ok {\n", kitHTTPPkg)
	g.w.W("code = sc.StatusCode()\n")
	g.w.W("}\n")

	if g.UseFast {
		g.w.W("w.SetStatusCode(code)\n")
		g.w.W("w.SetBody(data)\n")
	} else {
		g.w.W("w.WriteHeader(code)\n")
		g.w.W("_, _ = w.Write(data)\n")
	}
	g.w.W("}\n\n")
}
