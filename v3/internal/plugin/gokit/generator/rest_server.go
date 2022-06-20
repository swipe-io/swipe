package generator

import (
	"context"
	"fmt"
	"path"
	"strconv"
	stdstrings "strings"

	"github.com/swipe-io/swipe/v3/internal/convert"

	"github.com/swipe-io/swipe/v3/internal/plugin"

	"github.com/swipe-io/strcase"
	"github.com/swipe-io/swipe/v3/internal/plugin/gokit/config"
	"github.com/swipe-io/swipe/v3/option"
	"github.com/swipe-io/swipe/v3/swipe"
	"github.com/swipe-io/swipe/v3/writer"
)

type RESTServerGenerator struct {
	w             writer.GoWriter
	UseFast       bool
	JSONRPCEnable bool
	Interfaces    []*config.Interface
	MethodOptions map[string]config.MethodOptions
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

	g.writeDefaultErrorEncoder(contextPkg, httpPkg, kitHTTPPkg, jsonPkg)
	g.writeEncodeResponseFunc(contextPkg, httpPkg, jsonPkg)

	g.w.W("// MakeHandlerREST make REST HTTP transport\n")
	g.w.W("func MakeHandlerREST(")

	var external bool

	for i, iface := range g.Interfaces {
		typeStr := NameInterface(iface)
		if i > 0 {
			g.w.W(",")
		}
		if iface.Gateway != nil {
			external = true
			g.w.W("%s %sOption", LcNameWithAppPrefix(iface, true), UcNameWithAppPrefix(iface, true))
		} else {
			g.w.W("svc%s %s", iface.Named.Name.Upper(), typeStr)
		}
	}

	if external {
		g.w.W(", logger %s.Logger", importer.Import("log", "github.com/go-kit/log"))
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

	g.w.W("if opts.errorEncoder == nil {\n")
	g.w.W("opts.genericOpts.serverOption = append(opts.genericOpts.serverOption, %s.ServerErrorEncoder(defaultErrorEncoder))\n", kitHTTPPkg)
	g.w.W("} else {\n")
	g.w.W("opts.genericOpts.serverOption = append(opts.genericOpts.serverOption, %s.ServerErrorEncoder(opts.errorEncoder))\n", kitHTTPPkg)
	g.w.W("}\n\n")

	for _, iface := range g.Interfaces {
		optName := LcNameWithAppPrefix(iface, iface.Gateway != nil)
		ifaceType := iface.Named.Type.(*option.IfaceType)

		epSetName := NameEndpointSetNameVar(iface)

		if iface.Gateway != nil {
			epEndpointSetName := NameEndpointSetName(iface)

			sdPkg := importer.Import("sd", "github.com/go-kit/kit/sd")
			lbPkg := importer.Import("sd", "github.com/go-kit/kit/sd/lb")

			g.w.W("%s := %s{}\n", epSetName, epEndpointSetName)

			for _, m := range ifaceType.Methods {

				epFactoryName := "endpointFactory"
				kitEndpointPkg := importer.Import("endpoint", "github.com/go-kit/kit/endpoint")
				stdLogPkg := importer.Import("log", "log")

				ioPkg := importer.Import("io", "io")

				g.w.W("{\n")

				g.w.W("if %s.%s.Balancer == nil {\n", optName, m.Name)
				g.w.W("%s.%s.Balancer = %s.NewRoundRobin\n", optName, m.Name, lbPkg)
				g.w.W("}\n")

				g.w.W("if %s.%s.RetryMax == 0 {\n", optName, m.Name)
				g.w.W("%s.%s.RetryMax = DefaultRetryMax\n", optName, m.Name)
				g.w.W("}\n")

				g.w.W("if %s.%s.RetryTimeout == 0 {\n", optName, m.Name)
				g.w.W("%s.%s.RetryTimeout = DefaultRetryTimeout\n", optName, m.Name)
				g.w.W("}\n")

				g.w.W("if %s.Factory == nil {\n", optName)
				g.w.W("%s.Panic(\"%s.Factory is not set\")\n", stdLogPkg, optName)
				g.w.W("}\n")

				g.w.W("%s := func (instance string) (%s.Endpoint, %s.Closer, error) {\n", epFactoryName, kitEndpointPkg, ioPkg)
				g.w.W("c, err := %s.Factory(instance)\n", optName)

				g.w.WriteCheckErr("err", func() {
					g.w.W("return nil, nil, err\n")
				})

				g.w.W("return ")
				g.w.W("Make%sEndpoint(c), nil, nil\n", UcNameWithAppPrefix(iface)+m.Name.Upper())
				g.w.W("\n}\n\n")

				g.w.W("endpointer := %s.NewEndpointer(%s.Instancer, %s, logger)\n", sdPkg, optName, epFactoryName)
				g.w.W(
					"%[4]s.%[3]sEndpoint = %[1]s.RetryWithCallback(%[2]s.%[3]s.RetryTimeout, %[2]s.%[3]s.Balancer(endpointer), retryMax(%[2]s.%[3]s.RetryMax))\n",
					lbPkg, optName, m.Name, epSetName,
				)
				g.w.W(
					"%[2]s.%[1]sEndpoint = RetryErrorExtractor()(%[2]s.%[1]sEndpoint)\n",
					m.Name, epSetName,
				)
				g.w.W(
					"%[3]s.%[2]sEndpoint = middlewareChain(append(opts.genericOpts.endpointMiddleware, opts.%[1]sOpts.endpointMiddleware...))(%[3]s.%[2]sEndpoint)\n",
					LcNameWithAppPrefix(iface)+m.Name.Upper(), m.Name, epSetName,
				)
				g.w.W("}\n")
			}
		} else {
			g.w.W("%s := Make%s(svc%s)\n", NameEndpointSetNameVar(iface), NameEndpointSetName(iface), iface.Named.Name.Upper())
			for _, m := range ifaceType.Methods {
				g.w.W(
					"%[3]s.%[2]sEndpoint = middlewareChain(append(opts.genericOpts.endpointMiddleware, opts.%[1]sOpts.endpointMiddleware...))(%[3]s.%[2]sEndpoint)\n",
					LcNameWithAppPrefix(iface)+m.Name.Upper(), m.Name, epSetName,
				)
			}
		}
	}
	if g.UseFast {
		g.w.W("r := %s.New()\n", routerPkg)
	} else {
		g.w.W("r := %s.NewRouter()\n", routerPkg)
	}

	for _, iface := range g.Interfaces {
		ifaceType := iface.Named.Type.(*option.IfaceType)
		epSetName := NameEndpointSetNameVar(iface)
		for _, m := range ifaceType.Methods {
			mopt := g.MethodOptions[iface.Named.Name.Value+m.Name.Value]

			encRespFuncName := LcNameWithAppPrefix(iface) + m.Name.Upper()

			g.w.W("%s := encodeResponseHTTP\n", encRespFuncName)

			bodyType := mopt.RESTBodyType.Take()
			if bodyType == "" {
				bodyType = "json"
			}

			queryVars := make([]plugin.VarType, 0, len(mopt.RESTQueryVars.Value))
			headerVars := make([]plugin.VarType, 0, len(mopt.RESTHeaderVars.Value))
			pathVars := make([]plugin.VarType, 0, len(mopt.RESTPathVars))
			paramVars := make([]*option.VarType, 0, len(m.Sig.Params))

			for _, p := range m.Sig.Params {
				if plugin.IsContext(p) {
					continue
				}
				if v, ok := plugin.FindParam(p, mopt.RESTQueryVars.Value); ok {
					queryVars = append(queryVars, v)
					continue
				}
				if v, ok := plugin.FindParam(p, mopt.RESTHeaderVars.Value); ok {
					headerVars = append(headerVars, v)
					continue
				}
				if regexp, ok := mopt.RESTPathVars[p.Name.Value]; ok {
					pathVars = append(pathVars, plugin.VarType{
						Param: p,
						Value: regexp,
					})
					continue
				}
				paramVars = append(paramVars, p)
			}

			var urlPath string
			if mopt.RESTPath.IsValid() {
				urlPath = mopt.RESTPath.Take()
			} else {
				urlPath = strcase.ToKebab(m.Name.Value)
			}
			if iface.Namespace != "" {
				urlPath = path.Join(iface.Namespace, urlPath)
			}
			if !stdstrings.HasPrefix(urlPath, "/") {
				urlPath = "/" + urlPath
			}

			if g.UseFast {
				g.w.W("r.To(")
				if mopt.RESTMethod.Take() != "" {
					g.w.W(strconv.Quote(mopt.RESTMethod.Take()))
				} else {
					g.w.W(strconv.Quote("GET"))
				}

				g.w.W(", ")

				// replace brace indices for fasthttp router
				urlPath = stdstrings.ReplaceAll(urlPath, "{", "<")
				urlPath = stdstrings.ReplaceAll(urlPath, "}", ">")

				g.w.W(strconv.Quote(urlPath))

				g.w.W(", ")
			} else {
				g.w.W("r.Methods(%s,", strconv.Quote("OPTIONS"))
				if mopt.RESTMethod.Take() != "" {
					g.w.W(strconv.Quote(mopt.RESTMethod.Take()))
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

			g.w.W("func(ctx %s.Context, r *%s.Request) (_ interface{}, err error) {\n", contextPkg, httpPkg)

			nameRequest := NameRequest(m, iface)
			if len(m.Sig.Params) > 0 {
				g.w.W("var req %s\n", nameRequest)
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

				if len(pathVars) > 0 {
					errorPkg := importer.Import("errors", "errors")
					for _, pathVar := range pathVars {
						var valueVar string
						if g.UseFast {
							valueVar = "vars.Param(" + strconv.Quote(pathVar.Param.Name.Value) + ")"
						} else {
							valueVar = "vars[" + strconv.Quote(pathVar.Param.Name.Value) + "]"
						}

						convert.NewBuilder(importer).
							SetAssignVar("req." + strcase.ToCamel(pathVar.Param.Name.Value)).
							SetValueVar(valueVar).
							SetFieldName(pathVar.Param.Name).
							SetFieldType(pathVar.Param.Type).
							SetErrorReturn(fmt.Sprintf("return nil, %s.New(%s)", errorPkg, strconv.Quote("convert error"))).
							Write(&g.w)
					}
				}
				if len(queryVars) > 0 {
					errorPkg := importer.Import("errors", "errors")
					for _, queryVar := range queryVars {
						var valueVar string
						if g.UseFast {
							valueVar = "string(q.Peek(" + strconv.Quote(queryVar.Value) + "))"
						} else {
							valueVar = "q.Get(" + strconv.Quote(queryVar.Value) + ")"
						}
						if queryVar.IsRequired {
							fmtPkg := importer.Import("fmt", "fmt")
							if g.UseFast {
								g.w.W("if !q.Has(\"%[1]s\") {\nreturn nil, %[2]s.Errorf(\"%[1]s required\")\n}\n", queryVar.Value, fmtPkg)
							} else {
								g.w.W("if _, ok := q[\"%[1]s\"]; !ok {\nreturn nil, %[2]s.Errorf(\"%[1]s required\")\n}\n", queryVar.Value, fmtPkg)
							}
						}
						tmpID := "tmp" + queryVar.Param.Name.Value
						g.w.W("%s := %s\n", tmpID, valueVar)
						g.w.W("if %s != \"\" {\n", tmpID)

						convert.NewBuilder(importer).
							SetAssignVar("req." + queryVar.Param.Name.Upper()).
							SetValueVar(tmpID).
							SetFieldName(queryVar.Param.Name).
							SetFieldType(queryVar.Param.Type).
							SetErrorReturn(fmt.Sprintf("return nil, %s.New(%s)", errorPkg, strconv.Quote("convert error"))).
							Write(&g.w)
						g.w.W("}\n")
					}
				}
				if len(headerVars) > 0 {
					errorPkg := importer.Import("errors", "errors")
					for _, headerVar := range headerVars {
						var valueVar string
						if g.UseFast {
							valueVar = "string(r.Header.Peek(" + strconv.Quote(headerVar.Value) + "))"
						} else {
							valueVar = "r.Header.Get(" + strconv.Quote(headerVar.Value) + ")"
						}
						convert.NewBuilder(importer).
							SetAssignVar("req." + headerVar.Param.Name.Upper()).
							SetValueVar(valueVar).
							SetFieldName(headerVar.Param.Name).
							SetFieldType(headerVar.Param.Type).
							SetErrorReturn(fmt.Sprintf("return nil, %s.New(%s)", errorPkg, strconv.Quote("convert error"))).
							Write(&g.w)
					}
				}
				if len(paramVars) > 0 {
					switch stdstrings.ToUpper(mopt.RESTMethod.Take()) {
					case "POST", "PUT", "PATCH":
						switch bodyType {
						case "json":
							jsonPkg := importer.Import("ffjson", "github.com/pquerna/ffjson/ffjson")
							fmtPkg := importer.Import("fmt", "fmt")
							pkgIO := importer.Import("io", "io")

							g.w.W("var data []byte\n")
							if g.UseFast {
								g.w.W("data = r.Body()\n")
							} else {
								ioutilPkg := importer.Import("ioutil", "io/ioutil")
								g.w.W("data, err = %s.ReadAll(r.Body)\n", ioutilPkg)
								g.w.WriteCheckErr("err", func() {
									g.w.W("return nil, %s.Errorf(\"couldn't read body for %s: %%w\", err)\n", fmtPkg, nameRequest)
								})
							}
							if len(paramVars) == 1 {
								if s, ok := paramVars[0].Type.(*option.SliceType); ok {
									if b, ok := s.Value.(*option.BasicType); ok && b.IsByte() {
										g.w.W("req%s = data\n", "."+paramVars[0].Name.Upper())
									} else {
										g.w.W("err = %s.Unmarshal(data, &req)\n", jsonPkg)
									}
								} else {
									g.w.W("err = %s.Unmarshal(data, &req)\n", jsonPkg)
								}
							} else {
								g.w.W("err = %s.Unmarshal(data, &req)\n", jsonPkg)
							}
							g.w.W("if err != nil && err != %s.EOF {\n", pkgIO)
							g.w.W("return nil, %s.Errorf(\"couldn't unmarshal body to %s: %%w\", err)\n", fmtPkg, nameRequest)
							g.w.W("}\n")
						case "urlencoded":
							if g.UseFast {
							} else {
								errorPkg := importer.Import("errors", "errors")

								g.w.W("r.ParseForm()\n")
								for _, paramVar := range paramVars {
									valueVar := "r.Form.Get(" + strconv.Quote(paramVar.Name.Value) + ")"

									convert.NewBuilder(importer).
										SetAssignVar("req." + "req." + paramVar.Name.Upper()).
										SetValueVar(valueVar).
										SetFieldName(paramVar.Name).
										SetFieldType(paramVar.Type).
										SetErrorReturn(fmt.Sprintf("return nil, %s.New(%s)", errorPkg, strconv.Quote("convert error"))).
										Write(&g.w)
								}
							}
						case "multipart":
							multipartMaxMemory := mopt.RESTMultipartMaxMemory.Take()
							if multipartMaxMemory == 0 {
								multipartMaxMemory = 67108864
							}
							if g.UseFast {
								g.w.W("form, err := r.MultipartForm()\n")
							} else {
								g.w.W("err = r.ParseMultipartForm(%d)\n", multipartMaxMemory)
							}
							g.w.WriteCheckErr("err", func() {
								g.w.W("return nil, err\n")
							})
							for _, paramVar := range paramVars {
								if isFileUploadType(paramVar.Type, importer) {
									osPkg := importer.Import("os", "os")

									if g.UseFast {
										g.w.W("parts := form.File[%s]\n", strconv.Quote(paramVar.Name.Value))
										g.w.W("var (\nf *%s.File\n)\n", osPkg)
										g.w.W("if len(parts) > 0 {\n")
										g.w.W("f, err = %s.Open(parts[0].Filename)\n", osPkg)
										g.w.WriteCheckErr("err", func() {
											g.w.W("return nil, err\n")
										})
										g.w.W("}\n")
									} else {
										g.w.W("_, h, err := r.FormFile(%s)\n", strconv.Quote(paramVar.Name.Value))
										g.w.WriteCheckErr("err", func() {
											g.w.W("return nil, err\n")
										})
										g.w.W("f, err := %s.Open(h.Filename)\n", osPkg)
										g.w.WriteCheckErr("err", func() {
											g.w.W("return nil, err\n")
										})
									}
									g.w.W("req.%s = f\n", paramVar.Name.Upper())
									continue
								}
								var valueVar string
								if g.UseFast {
									valueVar = "form" + paramVar.Name.Upper()
									g.w.W("var %s string\n", valueVar)
									g.w.W("if fv, ok := form.Value[%s]; ok && len(fv) > 0 {\n", strconv.Quote(paramVar.Name.Value))
									g.w.W("%s = fv[0]\n", valueVar)
									g.w.W("}\n")
								} else {
									valueVar = "r.FormValue(" + strconv.Quote(paramVar.Name.Value) + ")"
								}
								errorPkg := importer.Import("errors", "errors")

								convert.NewBuilder(importer).
									SetAssignVar("req." + "req." + "req." + paramVar.Name.Upper()).
									SetValueVar(valueVar).
									SetFieldName(paramVar.Name).
									SetFieldType(paramVar.Type).
									SetErrorReturn(fmt.Sprintf("return nil, %s.New(%s)", errorPkg, strconv.Quote("convert error"))).
									Write(&g.w)
							}
						}
					}
				}

				g.w.W("return req, nil\n")
			} else {
				g.w.W("return nil, nil\n")
			}
			g.w.W("},\n")

			if mopt.RESTWrapResponse.Take() != "" {
				var responseWriterType string
				if g.UseFast {
					responseWriterType = fmt.Sprintf("*%s.Response", httpPkg)
				} else {
					responseWriterType = fmt.Sprintf("%s.ResponseWriter", httpPkg)
				}
				g.w.W("func (ctx context.Context, w %s, response interface{}) error {\n", responseWriterType)
				g.w.W("return %s(ctx, w, %s)\n", encRespFuncName, wrapDataServer(stdstrings.Split(mopt.RESTWrapResponse.Take(), ".")))
				g.w.W("}")
			} else {
				g.w.W(encRespFuncName)
			}

			g.w.W(",\n")
			g.w.W("append(opts.genericOpts.serverOption, opts.%sOpts.serverOption...)...,\n", LcNameWithAppPrefix(iface)+m.Name.Value)
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
	return "rest_server.go"
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

	g.w.W("var data []byte\n")

	g.w.W("if response != nil {\n")

	g.w.W("if cookie, ok := response.(interface{ HTTPCookies() []%s.Cookie }); ok {\n", httpPkg)
	g.w.W("for _, c := range cookie.HTTPCookies() {\n")
	g.w.W("%s.SetCookie(w, &c)\n", httpPkg)
	g.w.W("}\n")
	g.w.W("}\n")

	g.w.W("if download, ok := response.(downloader); ok {\n")
	g.w.W("contentType = download.ContentType()\n")
	g.w.W("data = download.Data()\n")
	g.w.W("} else {\n")
	g.w.W("data, err = %s.Marshal(response)\n", jsonPkg)
	g.w.W("if err != nil {\n")
	g.w.W("return err\n")
	g.w.W("}\n")
	g.w.W("}\n")
	g.w.W("} else {\n")
	g.w.W("contentType = \"text/plain; charset=utf-8\"\n")
	g.w.W("statusCode = 201\n")
	g.w.W("}\n")

	if g.UseFast {
		g.w.W("w.Header.Set(\"Content-Type\", contentType)\n")
		g.w.W("w.SetStatusCode(statusCode)\n")
		g.w.W("w.SetBody(data)\n")
	} else {
		g.w.W("w.Header().Set(\"Content-Type\", contentType)\n")
		g.w.W("w.WriteHeader(statusCode)\n")
		g.w.W("w.Write(data)\n")
	}

	g.w.W("return nil\n")
	g.w.W("}\n\n")
}

func (g *RESTServerGenerator) writeDefaultErrorEncoder(contextPkg string, httpPkg string, kitHTTPPkg string, jsonPkg string) {
	g.w.W("type errorWrapper struct {\n")
	g.w.W("Error string `json:\"error\"`\n")
	g.w.W("Code string `json:\"code,omitempty\"`\n")
	g.w.W("Data interface{} `json:\"data,omitempty\"`\n")
	g.w.W("}\n")

	g.w.W("func defaultErrorEncoder(ctx %s.Context, err error, ", contextPkg)
	if g.UseFast {
		g.w.W("w *%s.RequestCtx) {\n", httpPkg)
	} else {
		g.w.W("w %s.ResponseWriter) {\n", httpPkg)
	}

	g.w.W("var (\nerrData interface{}\nerrCode string\n)\n")
	g.w.W("if e, ok := err.(interface{ Data() interface{} }); ok {\n")
	g.w.W("errData = e.Data()\n")
	g.w.W("}\n")

	g.w.W("if e, ok := err.(interface{ Code() string }); ok {\n")
	g.w.W("errCode = e.Code()\n")
	g.w.W("}\n")

	g.w.W("data, jsonErr := %s.Marshal(errorWrapper{Error: err.Error(), Code: errCode, Data: errData})\n", jsonPkg)
	g.w.W("if jsonErr != nil {\n")
	if g.UseFast {
		g.w.W("w.SetBody([]byte(")
	} else {
		g.w.W("_, _ = w.Write([]byte(")
	}
	g.w.W("%s))\n", strconv.Quote("unexpected marshal error"))
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
