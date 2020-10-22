package generator

import (
	"context"
	"fmt"
	stdtypes "go/types"
	"path"
	"strconv"
	stdstrings "strings"

	"github.com/swipe-io/strcase"

	"github.com/swipe-io/swipe/v2/internal/domain/model"
	"github.com/swipe-io/swipe/v2/internal/importer"
	"github.com/swipe-io/swipe/v2/internal/strings"
	"github.com/swipe-io/swipe/v2/internal/usecase/generator"
	"github.com/swipe-io/swipe/v2/internal/writer"
)

type restServerOptionsGateway interface {
	AppID() string
	Prefix() string
	UseFast() bool
	Interfaces() model.Interfaces
	MethodOption(m model.ServiceMethod) model.MethodOption
	JSONRPCEnable() bool
}

type restServer struct {
	writer.GoLangWriter
	options restServerOptionsGateway
	i       *importer.Importer
}

func (g *restServer) Prepare(ctx context.Context) error {
	return nil
}

func (g *restServer) Process(ctx context.Context) error {
	var (
		routerPkg  string
		httpPkg    string
		kitHTTPPkg string
	)
	kitEndpointPkg := g.i.Import("endpoint", "github.com/go-kit/kit/endpoint")
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

	g.writeEncodeResponseFunc(contextPkg, httpPkg, kitEndpointPkg, jsonPkg)

	g.W("// MakeHandler%[1]s HTTP %[1]s Transport\n", g.options.Prefix())
	g.W("func MakeHandler%s(", g.options.Prefix())
	for i := 0; i < g.options.Interfaces().Len(); i++ {
		iface := g.options.Interfaces().At(i)
		typeStr := stdtypes.TypeString(iface.Type(), g.i.QualifyPkg)
		if i > 0 {
			g.W(",")
		}
		g.W("svc%s %s", iface.NameExport(), typeStr)
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

	for i := 0; i < g.options.Interfaces().Len(); i++ {
		iface := g.options.Interfaces().At(i)
		g.W("%[1]s := Make%[2]sEndpointSet(svc%[2]s)\n", makeEpSetName(iface, g.options.Interfaces().Len()), iface.NameExport())
	}
	for i := 0; i < g.options.Interfaces().Len(); i++ {
		iface := g.options.Interfaces().At(i)
		epSetName := makeEpSetName(iface, g.options.Interfaces().Len())
		for _, m := range iface.Methods() {
			g.W(
				"%[3]s.%[2]sEndpoint = middlewareChain(append(opts.genericEndpointMiddleware, opts.%[1]sEndpointMiddleware...))(%[3]s.%[2]sEndpoint)\n",
				m.NameUnExport, m.Name, epSetName,
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

		svcPrefix := ""
		if g.options.Interfaces().Len() > 1 {
			prefix := strcase.ToKebab(iface.Name())
			if iface.Prefix() != "" {
				prefix = iface.Prefix()
			}
			svcPrefix = prefix + "/"
		}
		epSetName := makeEpSetName(iface, g.options.Interfaces().Len())
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

				if mopt.Path != "" {
					// replace brace indices for fasthttp router
					urlPath := stdstrings.ReplaceAll(mopt.Path, "{", "<")
					urlPath = stdstrings.ReplaceAll(urlPath, "}", ">")
					g.W(strconv.Quote(urlPath))
				} else {
					g.W(strconv.Quote("/" + m.LcName))
				}
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
					g.W(strconv.Quote(path.Join(svcPrefix, mopt.Path)))
				} else {
					g.W(strconv.Quote(path.Join("/", svcPrefix, "/", stdstrings.ToLower(m.Name))))
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
						fmtPkg := g.i.Import("fmt", "fmt")
						jsonPkg := g.i.Import("ffjson", "github.com/pquerna/ffjson/ffjson")
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
							g.WriteConvertType(g.i.Import, "req."+strings.UcFirst(p.Name()), valueID, p, []string{"nil"}, "", false, "")
						} else if queryName, ok := mopt.QueryVars[p.Name()]; ok {
							var valueID string
							if g.options.UseFast() {
								valueID = "string(q.Peek(" + strconv.Quote(queryName) + "))"
							} else {
								valueID = "q.Get(" + strconv.Quote(queryName) + ")"
							}
							g.WriteConvertType(g.i.Import, "req."+strings.UcFirst(p.Name()), valueID, p, []string{"nil"}, "", false, "")
						} else if headerName, ok := mopt.HeaderVars[p.Name()]; ok {
							var valueID string
							if g.options.UseFast() {
								valueID = "string(r.Header.Peek(" + strconv.Quote(headerName) + "))"
							} else {
								valueID = "r.Header.Get(" + strconv.Quote(headerName) + ")"
							}
							g.WriteConvertType(g.i.Import, "req."+strings.UcFirst(p.Name()), valueID, p, []string{"nil"}, "", false, "")
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
						g.W("func (ctx context.Context, w %s, response interface{}) error {\n", responseWriterType)
						g.W("return encodeResponseHTTP(ctx, w, map[string]interface{}{\"%s\": response})\n", mopt.WrapResponse.Name)
						g.W("}")
					} else {
						g.W("encodeResponseHTTP")
					}
				}
			}
			g.W(",\n")

			g.W("append(opts.genericServerOption, opts.%sServerOption...)...,\n", m.NameUnExport)
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

func (g *restServer) writeEncodeResponseFunc(contextPkg, httpPkg, kitEndpointPkg, jsonPkg string) {
	g.W("type errorWrapper struct {\n")
	g.W("Error string `json:\"error\"`\n")
	g.W("}\n")

	g.W("func encodeResponseHTTP(ctx %s.Context, ", contextPkg)

	if g.options.UseFast() {
		g.W("w *%s.Response", httpPkg)
	} else {
		g.W("w %s.ResponseWriter", httpPkg)
	}

	g.W(", response interface{}) error {\n")

	if g.options.UseFast() {
		g.W("h := w.Header\n")
	} else {
		g.W("h := w.Header()\n")
	}

	g.W("h.Set(\"Content-Iface\", \"application/json; charset=utf-8\")\n")
	g.W("if e, ok := response.(%s.Failer); ok && e.Failed() != nil {\n", kitEndpointPkg)
	g.W("data, err := %s.Marshal(errorWrapper{Error: e.Failed().Error()})\n", jsonPkg)
	g.W("if err != nil {\n")
	g.W("return err\n")
	g.W("}\n")

	if g.options.UseFast() {
		g.W("w.SetBody(data)\n")
	} else {
		g.W("w.Write(data)\n")
	}

	g.W("return nil\n")
	g.W("}\n")

	g.W("data, err := %s.Marshal(response)\n", jsonPkg)
	g.W("if err != nil {\n")
	g.W("return err\n")
	g.W("}\n")

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
