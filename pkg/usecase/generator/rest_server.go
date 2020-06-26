package generator

import (
	"context"
	"fmt"
	stdtypes "go/types"
	"strconv"
	stdstrings "strings"

	"github.com/swipe-io/swipe/pkg/domain/model"
	"github.com/swipe-io/swipe/pkg/importer"
	"github.com/swipe-io/swipe/pkg/strings"
	"github.com/swipe-io/swipe/pkg/writer"
)

type restServer struct {
	*writer.GoLangWriter

	info model.GenerateInfo
	o    model.ServiceOption
	i    *importer.Importer
}

func (g *restServer) Process(ctx context.Context) error {
	var (
		routerPkg  string
		httpPkg    string
		kithttpPkg string
	)
	kitEndpointPkg := g.i.Import("endpoint", "github.com/go-kit/kit/endpoint")
	jsonPkg := g.i.Import("ffjson", "github.com/pquerna/ffjson/ffjson")
	contextPkg := g.i.Import("context", "context")
	typeStr := stdtypes.TypeString(g.o.Type, g.i.QualifyPkg)

	transportOpt := g.o.Transport

	if transportOpt.FastHTTP {
		httpPkg = g.i.Import("fasthttp", "github.com/valyala/fasthttp")
		kithttpPkg = g.i.Import("fasthttp", "github.com/l-vitaly/go-kit/transport/fasthttp")
		routerPkg = g.i.Import("routing", "github.com/qiangxue/fasthttp-routing")
	} else {
		routerPkg = g.i.Import("mux", "github.com/gorilla/mux")
		kithttpPkg = g.i.Import("http", "github.com/go-kit/kit/transport/http")
		httpPkg = g.i.Import("http", "net/http")
	}

	g.W("type errorWrapper struct {\n")
	g.W("Error string `json:\"error\"`\n")
	g.W("}\n")

	g.W("func encodeResponseHTTP%s(ctx %s.Context, ", g.o.ID, contextPkg)

	if transportOpt.FastHTTP {
		g.W("w *%s.Response", httpPkg)
	} else {
		g.W("w %s.ResponseWriter", httpPkg)
	}

	g.W(", response interface{}) error {\n")

	if transportOpt.FastHTTP {
		g.W("h := w.Header\n")
	} else {
		g.W("h := w.Header()\n")
	}

	g.W("h.Set(\"Content-Interface\", \"application/json; charset=utf-8\")\n")
	g.W("if e, ok := response.(%s.Failer); ok && e.Failed() != nil {\n", kitEndpointPkg)
	g.W("data, err := %s.Marshal(errorWrapper{Error: e.Failed().Error()})\n", jsonPkg)
	g.W("if err != nil {\n")
	g.W("return err\n")
	g.W("}\n")

	if transportOpt.FastHTTP {
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

	if transportOpt.FastHTTP {
		g.W("w.SetBody(data)\n")
	} else {
		g.W("w.Write(data)\n")
	}

	g.W("return nil\n")
	g.W("}\n\n")

	g.W("// HTTP %s Transport\n", transportOpt.Prefix)
	g.W("func MakeHandler%s%s(s %s", transportOpt.Prefix, g.o.ID, typeStr)
	g.W(", opts ...%sServerOption", g.o.ID)
	g.W(") (")
	if transportOpt.FastHTTP {
		g.W("%s.RequestHandler", g.i.Import("fasthttp", "github.com/valyala/fasthttp"))
	} else {
		g.W("%s.Handler", g.i.Import("http", "net/http"))
	}

	g.W(", error) {\n")

	g.W("sopt := &server%sOpts{}\n", g.o.ID)

	g.W("for _, o := range opts {\n o(sopt)\n }\n")

	g.W("ep := MakeEndpointSet(s)\n")

	for _, m := range g.o.Methods {
		g.W("ep.%[1]sEndpoint = middlewareChain(append(sopt.genericEndpointMiddleware, sopt.%[2]sEndpointMiddleware...))(ep.%[1]sEndpoint)\n", m.Name, m.LcName)
	}

	if transportOpt.FastHTTP {
		g.W("r := %s.New()\n", routerPkg)
	} else {
		g.W("r := %s.NewRouter()\n", routerPkg)
	}
	for _, m := range g.o.Methods {
		mopt := transportOpt.MethodOptions[m.Name]

		if transportOpt.FastHTTP {
			g.W("r.To(")

			if mopt.MethodName != "" {
				g.WriteAST(mopt.Expr)
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
				g.WriteAST(mopt.Expr)
			} else {
				g.W(strconv.Quote("GET"))
			}
			g.W(").")
			g.W("Path(")
			if mopt.Path != "" {
				g.W(strconv.Quote(mopt.Path))
			} else {
				g.W(strconv.Quote("/" + stdstrings.ToLower(m.Name)))
			}
			g.W(").")

			g.W("Handler(")
		}

		g.W(
			"%s.NewServer(\nep.%sEndpoint,\n",
			kithttpPkg,
			m.Name,
		)

		if mopt.ServerRequestFunc.Expr != nil {
			g.WriteAST(mopt.ServerRequestFunc.Expr)
		} else {
			g.W("func(ctx %s.Context, r *%s.Request) (interface{}, error) {\n", contextPkg, httpPkg)

			if len(m.Params) > 0 {
				g.W("var req %sRequest%s\n", m.LcName, g.o.ID)
				switch stdstrings.ToUpper(mopt.MethodName) {
				case "POST", "PUT", "PATCH":
					fmtPkg := g.i.Import("fmt", "fmt")
					jsonPkg := g.i.Import("ffjson", "github.com/pquerna/ffjson/ffjson")
					pkgIO := g.i.Import("io", "io")

					if transportOpt.FastHTTP {
						g.W("err := %s.Unmarshal(r.Body(), &req)\n", jsonPkg)
					} else {
						ioutilPkg := g.i.Import("ioutil", "io/ioutil")

						g.W("b, err := %s.ReadAll(r.Body)\n", ioutilPkg)
						g.WriteCheckErr(func() {
							g.W("return nil, %s.Errorf(\"couldn't read body for %sRequest%s: %%s\", err)\n", fmtPkg, m.LcName, g.o.ID)
						})
						g.W("err = %s.Unmarshal(b, &req)\n", jsonPkg)
					}

					g.W("if err != nil && err != %s.EOF {\n", pkgIO)
					g.W("return nil, %s.Errorf(\"couldn't unmarshal body to %sRequest%s: %%s\", err)\n", fmtPkg, m.LcName, g.o.ID)
					g.W("}\n")
				}
				if len(mopt.PathVars) > 0 {
					if transportOpt.FastHTTP {
						fmtPkg := g.i.Import("fmt", "fmt")

						g.W("vars, ok := ctx.Value(%s.ContextKeyRouter).(*%s.Context)\n", kithttpPkg, routerPkg)
						g.W("if !ok {\n")
						g.W("return nil, %s.Errorf(\"couldn't assert %s.ContextKeyRouter to *%s.Context\")\n", fmtPkg, kithttpPkg, routerPkg)
						g.W("}\n")
					} else {
						g.W("vars := %s.Vars(r)\n", routerPkg)
					}
					for pathVarName := range mopt.PathVars {
						if f := m.Params.LookupField(pathVarName); f != nil {
							var valueID string
							if transportOpt.FastHTTP {
								valueID = "vars.Param(" + strconv.Quote(pathVarName) + ")"
							} else {
								valueID = "vars[" + strconv.Quote(pathVarName) + "]"
							}
							g.WriteConvertType("req."+strings.UcFirst(f.Name()), valueID, f, "", false, "")
						}
					}
				}
				if len(mopt.QueryVars) > 0 {
					if transportOpt.FastHTTP {
						g.W("q := r.URI().QueryArgs()\n")
					} else {
						g.W("q := r.URL.Query()\n")
					}
					for argName, queryName := range mopt.QueryVars {
						if f := m.Params.LookupField(argName); f != nil {
							var valueID string
							if transportOpt.FastHTTP {
								valueID = "string(q.Peek(" + strconv.Quote(queryName) + "))"
							} else {
								valueID = "q.Get(" + strconv.Quote(queryName) + ")"
							}
							g.WriteConvertType("req."+strings.UcFirst(f.Name()), valueID, f, "", false, "")
						}
					}
				}
				for argName, headerName := range mopt.HeaderVars {
					if f := m.Params.LookupField(argName); f != nil {
						var valueID string
						if transportOpt.FastHTTP {
							valueID = "string(r.Header.Peek(" + strconv.Quote(headerName) + "))"
						} else {
							valueID = "r.Header.Get(" + strconv.Quote(headerName) + ")"
						}
						g.WriteConvertType("req."+strings.UcFirst(f.Name()), valueID, f, "", false, "")
					}
				}
				g.W("return req, nil\n")
			} else {
				g.W("return nil, nil\n")
			}
			g.W("},\n")
		}
		if mopt.ServerResponseFunc.Expr != nil {
			g.WriteAST(mopt.ServerResponseFunc.Expr)
		} else {
			if transportOpt.JsonRPC.Enable {
				g.W("encodeResponseJSONRPC%s", g.o.ID)
			} else {
				if mopt.WrapResponse.Enable {
					var responseWriterType string
					if transportOpt.FastHTTP {
						responseWriterType = fmt.Sprintf("*%s.Response", httpPkg)
					} else {
						responseWriterType = fmt.Sprintf("%s.ResponseWriter", httpPkg)
					}
					g.W("func (ctx context.Context, w %s, response interface{}) error {\n", responseWriterType)
					g.W("return encodeResponseHTTP%s(ctx, w, map[string]interface{}{\"%s\": response})\n", g.o.ID, mopt.WrapResponse.Name)
					g.W("}")
				} else {
					g.W("encodeResponseHTTP%s", g.o.ID)
				}
			}
		}

		g.W(",\n")

		g.W("append(sopt.genericServerOption, sopt.%sServerOption...)...,\n", m.LcName)
		g.W(")")

		if transportOpt.FastHTTP {
			g.W(".RouterHandle()")
		}
		g.W(")\n")
	}
	if transportOpt.FastHTTP {
		g.W("return r.HandleRequest, nil")
	} else {
		g.W("return r, nil")
	}

	g.W("}\n\n")

	return nil
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

func (g *restServer) Imports() []string {
	return g.i.SortedImports()
}

func NewRestServer(info model.GenerateInfo, o model.ServiceOption, i *importer.Importer) Generator {
	return &restServer{info: info, o: o, i: i, GoLangWriter: writer.NewGoLangWriter(i)}
}
