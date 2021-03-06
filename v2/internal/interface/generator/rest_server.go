package generator

import (
	"container/list"
	"context"
	"fmt"
	"path"
	"strconv"
	stdstrings "strings"

	"github.com/swipe-io/strcase"
	"github.com/swipe-io/swipe/v2/internal/domain/model"
	"github.com/swipe-io/swipe/v2/internal/importer"
	"github.com/swipe-io/swipe/v2/internal/option"
	"github.com/swipe-io/swipe/v2/internal/usecase/generator"
	"github.com/swipe-io/swipe/v2/internal/writer"
)

type restServerOptionsGateway interface {
	AppID() string
	UseFast() bool
	Interfaces() model.Interfaces
	MethodOption(m model.ServiceMethod) model.MethodOption
	JSONRPCEnable() bool
	DefaultErrorEncoder() option.Value
}

type restServer struct {
	writer.GoLangWriter
	options restServerOptionsGateway
	i       *importer.Importer
}

func (g *restServer) Prepare(_ context.Context) error {
	return nil
}

func (g *restServer) Process(_ context.Context) error {
	var (
		routerPkg  string
		httpPkg    string
		kitHTTPPkg string
	)
	jsonPkg := g.i.Import("ffjson", "github.com/pquerna/ffjson/ffjson")
	contextPkg := g.i.Import("context", "context")

	if g.options.UseFast() {
		httpPkg = g.i.Import("fasthttp", "github.com/valyala/fasthttp")
		kitHTTPPkg = g.i.Import("fasthttp", "github.com/l-vitaly/go-kit/transport/fasthttp")
		routerPkg = g.i.Import("routing", "github.com/qiangxue/fasthttp-routing")
	} else {
		routerPkg = g.i.Import("mux", "github.com/gorilla/mux")
		kitHTTPPkg = g.i.Import("http", "github.com/go-kit/kit/transport/http")
		httpPkg = g.i.Import("http", "net/http")
	}

	if g.options.DefaultErrorEncoder().Expr() == nil {
		g.writeDefaultErrorEncoder(contextPkg, httpPkg, kitHTTPPkg, jsonPkg)
	}
	g.writeEncodeResponseFunc(contextPkg, httpPkg, jsonPkg)

	g.W("// MakeHandler HTTP Transport\n")
	g.W("func MakeHandlerREST(")
	for i := 0; i < g.options.Interfaces().Len(); i++ {
		iface := g.options.Interfaces().At(i)
		typeStr := iface.LcNameWithPrefix() + "Interface"
		if i > 0 {
			g.W(",")
		}
		g.W("svc%s %s", iface.UcName(), typeStr)
	}
	g.W(", options ...ServerOption")
	g.W(") (")
	if g.options.UseFast() {
		g.W("%s.RequestHandler", g.i.Import("fasthttp", "github.com/valyala/fasthttp"))
	} else {
		g.W("%s.Handler", g.i.Import("http", "net/http"))
	}
	g.W(", error) {\n")

	g.W("opts := &serverOpts{}\n")
	g.W("for _, o := range options {\n o(opts)\n }\n")

	if g.options.DefaultErrorEncoder().Expr() != nil {
		g.W("opts.genericServerOption = append(opts.genericServerOption, %s.ServerErrorEncoder(", kitHTTPPkg)
		writer.WriteAST(g, g.i, g.options.DefaultErrorEncoder().Expr())
		g.W("))\n")
	} else {
		g.W("opts.genericServerOption = append(opts.genericServerOption, %s.ServerErrorEncoder(defaultErrorEncoder))\n", kitHTTPPkg)
	}

	for i := 0; i < g.options.Interfaces().Len(); i++ {
		iface := g.options.Interfaces().At(i)
		g.W("%[1]s := Make%[2]sEndpointSet(svc%[2]s)\n", makeEpSetName(iface), iface.UcName())
	}
	for i := 0; i < g.options.Interfaces().Len(); i++ {
		iface := g.options.Interfaces().At(i)
		epSetName := makeEpSetName(iface)
		for _, m := range iface.Methods() {
			g.W(
				"%[3]s.%[2]sEndpoint = middlewareChain(append(opts.genericEndpointMiddleware, opts.%[1]sEndpointMiddleware...))(%[3]s.%[2]sEndpoint)\n",
				m.IfaceLcName, m.Name, epSetName,
			)
		}
	}
	if g.options.UseFast() {
		g.W("r := %s.New()\n", routerPkg)
	} else {
		g.W("r := %s.NewRouter()\n", routerPkg)
	}

	for i := 0; i < g.options.Interfaces().Len(); i++ {
		iface := g.options.Interfaces().At(i)
		var prefix string
		if iface.Namespace() != "" {
			prefix = iface.Namespace()
		}
		epSetName := makeEpSetName(iface)
		for _, m := range iface.Methods() {
			mopt := g.options.MethodOption(m)
			if g.options.UseFast() {
				g.W("r.To(")

				if mopt.MethodName != "" {
					writer.WriteAST(g, g.i, mopt.Expr)
				} else {
					g.W(strconv.Quote("GET"))
				}

				g.W(", ")

				var urlPath string

				if mopt.Path != "" {
					// replace brace indices for fasthttp router
					urlPath = stdstrings.ReplaceAll(mopt.Path, "{", "<")
					urlPath = stdstrings.ReplaceAll(urlPath, "}", ">")
				} else {
					urlPath = "/" + strcase.ToKebab(m.Name)
				}
				if prefix != "" {
					urlPath = prefix + urlPath
				}
				g.W(strconv.Quote(stdstrings.TrimRight(urlPath, "/")))
				g.W(", ")
			} else {
				g.W("r.Methods(")
				if mopt.MethodName != "" {
					writer.WriteAST(g, g.i, mopt.Expr)
				} else {
					g.W(strconv.Quote("GET"))
				}
				g.W(").")
				g.W("Path(")
				if mopt.Path != "" {
					g.W(strconv.Quote(path.Join("/", prefix, mopt.Path)))
				} else {
					g.W(strconv.Quote(path.Join("/", prefix, strcase.ToKebab(m.Name))))
				}

				g.W(").")
				g.W("Handler(")
			}
			g.W(
				"%s.NewServer(\n%s.%sEndpoint,\n",
				kitHTTPPkg,
				epSetName,
				m.Name,
			)
			if mopt.ServerRequestFunc.Expr != nil {
				writer.WriteAST(g, g.i, mopt.ServerRequestFunc.Expr)
			} else {
				g.W("func(ctx %s.Context, r *%s.Request) (interface{}, error) {\n", contextPkg, httpPkg)

				if len(m.Params) > 0 {
					g.W("var req %s\n", m.NameRequest)
					switch stdstrings.ToUpper(mopt.MethodName) {
					case "POST", "PUT", "PATCH":
						jsonPkg := g.i.Import("ffjson", "github.com/pquerna/ffjson/ffjson")
						fmtPkg := g.i.Import("fmt", "fmt")
						pkgIO := g.i.Import("io", "io")
						if g.options.UseFast() {
							g.W("err := %s.Unmarshal(r.Body(), &req)\n", jsonPkg)
						} else {
							ioutilPkg := g.i.Import("ioutil", "io/ioutil")

							g.W("b, err := %s.ReadAll(r.Body)\n", ioutilPkg)
							g.WriteCheckErr(func() {
								g.W("return nil, %s.Errorf(\"couldn't read body for %s: %%w\", err)\n", fmtPkg, m.NameRequest)
							})
							g.W("err = %s.Unmarshal(b, &req)\n", jsonPkg)
						}
						g.W("if err != nil && err != %s.EOF {\n", pkgIO)
						g.W("return nil, %s.Errorf(\"couldn't unmarshal body to %s: %%w\", err)\n", fmtPkg, m.NameRequest)
						g.W("}\n")
					}
					if len(mopt.PathVars) > 0 {
						if g.options.UseFast() {
							fmtPkg := g.i.Import("fmt", "fmt")

							g.W("vars, ok := ctx.Value(%s.ContextKeyRouter).(*%s.Context)\n", kitHTTPPkg, routerPkg)
							g.W("if !ok {\n")
							g.W("return nil, %s.Errorf(\"couldn't assert %s.ContextKeyRouter to *%s.Context\")\n", fmtPkg, kitHTTPPkg, routerPkg)
							g.W("}\n")
						} else {
							g.W("vars := %s.Vars(r)\n", routerPkg)
						}
					}
					if len(mopt.QueryVars) > 0 {
						if g.options.UseFast() {
							g.W("q := r.URI().QueryArgs()\n")
						} else {
							g.W("q := r.URL.Query()\n")
						}
					}

					for _, p := range m.Params {
						if _, ok := mopt.PathVars[p.Name()]; ok {
							var valueID string
							if g.options.UseFast() {
								valueID = "vars.Param(" + strconv.Quote(p.Name()) + ")"
							} else {
								valueID = "vars[" + strconv.Quote(p.Name()) + "]"
							}
							g.WriteConvertType(g.i.Import, "req."+strcase.ToCamel(p.Name()), valueID, p, []string{"nil"}, "", false, "")
						} else if queryName, ok := mopt.QueryVars[p.Name()]; ok {
							var valueID string
							if g.options.UseFast() {
								valueID = "string(q.Peek(" + strconv.Quote(queryName) + "))"
							} else {
								valueID = "q.Get(" + strconv.Quote(queryName) + ")"
							}

							tmpID := "tmp" + p.Name()
							g.W("%s := %s\n", tmpID, valueID)

							g.W("if %s != \"\" {\n", tmpID)
							g.WriteConvertType(g.i.Import, "req."+strcase.ToCamel(p.Name()), tmpID, p, []string{"nil"}, "", false, "")
							g.W("}\n")

						} else if headerName, ok := mopt.HeaderVars[p.Name()]; ok {
							var valueID string
							if g.options.UseFast() {
								valueID = "string(r.Header.Peek(" + strconv.Quote(headerName) + "))"
							} else {
								valueID = "r.Header.Get(" + strconv.Quote(headerName) + ")"
							}
							g.WriteConvertType(g.i.Import, "req."+strcase.ToCamel(p.Name()), valueID, p, []string{"nil"}, "", false, "")
						}
					}
					g.W("return req, nil\n")
				} else {
					g.W("return nil, nil\n")
				}
				g.W("}")
			}
			g.W(",\n")

			if mopt.ServerResponseFunc.Expr != nil {
				writer.WriteAST(g, g.i, mopt.ServerResponseFunc.Expr)
			} else {
				if g.options.JSONRPCEnable() {
					g.W("encodeResponseJSONRPC")
				} else {
					if mopt.WrapResponse.Enable {
						var responseWriterType string
						if g.options.UseFast() {
							responseWriterType = fmt.Sprintf("*%s.Response", httpPkg)
						} else {
							responseWriterType = fmt.Sprintf("%s.ResponseWriter", httpPkg)
						}

						parts := stdstrings.Split(mopt.WrapResponse.Name, ".")

						var fn func(e *list.Element) string
						fn = func(e *list.Element) (out string) {
							if next := e.Next(); next != nil {
								out += " map[string]interface{}{"
								out += strconv.Quote(e.Value.(string)) + ": "
								out += fn(next)
								out += "}"
							} else {
								out += "map[string]interface{}{" + strconv.Quote(e.Value.(string)) + ": response }"
							}
							return out
						}

						l := list.New()
						if len(parts) > 0 {
							e := l.PushFront(parts[0])
							for i := 1; i < len(parts); i++ {
								e = l.InsertAfter(parts[i], e)
							}
						}

						g.W("func (ctx context.Context, w %s, response interface{}) error {\n", responseWriterType)
						g.W("return encodeResponseHTTP(ctx, w, %s)\n", fn(l.Front()))
						g.W("}")
					} else {
						g.W("encodeResponseHTTP")
					}
				}
			}
			g.W(",\n")

			g.W("append(opts.genericServerOption, opts.%sServerOption...)...,\n", m.IfaceLcName)
			g.W(")")

			if g.options.UseFast() {
				g.W(".RouterHandle()")
			}
			g.W(")\n")
		}
	}
	if g.options.UseFast() {
		g.W("return r.HandleRequest, nil")
	} else {
		g.W("return r, nil")
	}
	g.W("}\n\n")
	return nil
}

func (g *restServer) writeDefaultErrorEncoder(contextPkg, httpPkg, kitHTTPPkg, jsonPkg string) {
	g.W("type errorWrapper struct {\n")
	g.W("Error string `json:\"error\"`\n")
	g.W("Data interface{} `json:\"data,omitempty\"`\n")
	g.W("}\n")

	g.W("func defaultErrorEncoder(ctx %s.Context, err error, ", contextPkg)
	if g.options.UseFast() {
		g.W("w *%s.RequestCtx) {\n", httpPkg)
	} else {
		g.W("w %s.ResponseWriter) {\n", httpPkg)
	}

	g.W("var errData interface{}\n")
	g.W("if e, ok := err.(interface{ ErrorData() interface{} }); ok {\n")
	g.W("errData = e.ErrorData()\n")
	g.W("}\n")

	g.W("data, merr := %s.Marshal(errorWrapper{Error: err.Error(), Data: errData})\n", jsonPkg)
	g.W("if merr != nil {\n")
	if g.options.UseFast() {
		g.W("w.SetBody([]byte(")
	} else {
		g.W("_, _ = w.Write([]byte(")
	}
	g.W("%s))\n", strconv.Quote("unexpected error"))
	g.W("return\n")
	g.W("}\n")

	if g.options.UseFast() {
		g.W("w.Response.Header")
	} else {
		g.W("w.Header()")
	}
	g.W(".Set(\"Content-Type\", \"application/json; charset=utf-8\")\n")

	g.W("if headerer, ok := err.(%s.Headerer); ok {\n", kitHTTPPkg)

	if g.options.UseFast() {
		g.W("for k, v := range headerer.Headers() {\n")
		g.W("w.Response.Header.Add(k, v)")
		g.W("}\n")
	} else {
		g.W("for k, values := range headerer.Headers() {\n")
		g.W("for _, v := range values {\n")
		g.W("w.Header().Add(k, v)")
		g.W("}\n}\n")
	}
	g.W("}\n")
	g.W("code := %s.StatusInternalServerError\n", httpPkg)
	g.W("if sc, ok := err.(%s.StatusCoder); ok {\n", kitHTTPPkg)
	g.W("code = sc.StatusCode()\n")
	g.W("}\n")

	if g.options.UseFast() {
		g.W("w.SetStatusCode(code)\n")
		g.W("w.SetBody(data)\n")
	} else {
		g.W("w.WriteHeader(code)\n")
		g.W("_, _ = w.Write(data)\n")
	}
	g.W("}\n\n")
}

func (g *restServer) writeEncodeResponseFunc(contextPkg, httpPkg, jsonPkg string) {
	g.W("func encodeResponseHTTP(ctx %s.Context, ", contextPkg)

	if g.options.UseFast() {
		g.W("w *%s.Response", httpPkg)
	} else {
		g.W("w %s.ResponseWriter", httpPkg)
	}

	g.W(", response interface{}) (err error) {\n")
	g.W("contentType := \"application/json; charset=utf-8\"\n")
	g.W("statusCode := 200\n")

	g.W("var data []byte\n")
	g.W("if response != nil {\n")
	g.W("data, err = %s.Marshal(response)\n", jsonPkg)
	g.W("if err != nil {\n")
	g.W("return err\n")
	g.W("}\n")
	g.W("} else {\n")
	g.W("contentType = \"text/plain; charset=utf-8\"\n")
	g.W("statusCode = 201\n")
	g.W("}\n")

	if g.options.UseFast() {
		g.W("w.Header.Set(\"Content-Type\", contentType)\n")
	} else {
		g.W("w.Header().Set(\"Content-Type\", contentType)\n")
	}

	if g.options.UseFast() {
		g.W("w.SetStatusCode(statusCode)\n")
	} else {
		g.W("w.WriteHeader(statusCode)\n")
	}
	if g.options.UseFast() {
		g.W("w.SetBody(data)\n")
	} else {
		g.W("w.Write(data)\n")
	}
	g.W("return nil\n")
	g.W("}\n\n")
}

func (g *restServer) PkgName() string {
	return ""
}

func (g *restServer) OutputDir() string {
	return ""
}

func (g *restServer) Filename() string {
	return "server_gen.go"
}

func (g *restServer) SetImporter(i *importer.Importer) {
	g.i = i
}

func NewRestServer(
	options restServerOptionsGateway,
) generator.Generator {
	return &restServer{
		options: options,
	}
}
