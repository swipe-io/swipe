package generator

import (
	"context"
	"fmt"
	"path"
	"strconv"
	stdstrings "strings"

	"github.com/swipe-io/strcase"

	"github.com/swipe-io/swipe/v3/internal/convert"
	"github.com/swipe-io/swipe/v3/internal/plugin"
	"github.com/swipe-io/swipe/v3/internal/plugin/echo/config"
	"github.com/swipe-io/swipe/v3/option"
	"github.com/swipe-io/swipe/v3/swipe"
	"github.com/swipe-io/swipe/v3/writer"
)

type RoutesGenerator struct {
	w             writer.GoWriter
	Interfaces    []*config.Interface
	MethodOptions map[string]config.MethodOptions
}

func (g *RoutesGenerator) Generate(ctx context.Context) []byte {
	importer := ctx.Value(swipe.ImporterKey).(swipe.Importer)

	echoPkg := importer.Import("echo", "github.com/labstack/echo/v4")
	httpPkg := importer.Import("http", "net/http")

	g.writeDefaultErrorEncoder(echoPkg, httpPkg)
	g.writeEncodeResponseFunc(echoPkg, httpPkg)

	g.w.W("type Option func(*opts)\n")
	g.w.W("type ServerOption func(*serverOpts)\n\n")

	for _, iface := range g.Interfaces {
		ifaceType := iface.Named.Type.(*option.IfaceType)
		for _, m := range ifaceType.Methods {
			g.w.W("type %s struct { opts }\n\n", LcNameIfaceMethod(iface, m)+"Opts")
		}
	}

	g.w.W("type serverOpts struct {\ngenericOpts opts\n")
	for _, iface := range g.Interfaces {
		ifaceType := iface.Named.Type.(*option.IfaceType)

		g.w.W("generic%s opts\n", UcNameWithAppPrefix(iface)+"Opts")

		for _, m := range ifaceType.Methods {
			serverOpt := LcNameWithAppPrefix(iface) + m.Name.Value + "Opts"
			g.w.W("%[1]s %[1]s\n", serverOpt)
		}
	}
	g.w.W("}\n")

	for _, iface := range g.Interfaces {
		ifaceType := iface.Named.Type.(*option.IfaceType)

		g.w.W("func %sOptions(opt ...Option) ServerOption {\nreturn func(c *serverOpts) {\nfor _, o := range opt {\no(&c.generic%s)\n}\n}\n}\n\n", UcNameWithAppPrefix(iface), UcNameWithAppPrefix(iface)+"Opts")

		for _, m := range ifaceType.Methods {
			serverOptName := LcNameWithAppPrefix(iface) + m.Name.Value + "Opts"
			serverOptFuncName := UcNameWithAppPrefix(iface) + m.Name.Value
			g.w.W("func %sOptions(opt ...Option) ServerOption {\nreturn func(c *serverOpts) {\nfor _, o := range opt {\no(&c.%s.opts)\n}\n}\n}\n\n", serverOptFuncName, serverOptName)
		}
	}

	g.w.W("type opts struct {\nmiddlewares []%s.MiddlewareFunc\n}\n", echoPkg)

	g.w.W("func Middlewares(middlewares []%s.MiddlewareFunc) Option {\nreturn func(c *opts) {\n c.middlewares = middlewares\n}\n}\n\n", echoPkg)

	g.w.W("func SetupRoutes(e *%s.Echo,", echoPkg)

	for i, iface := range g.Interfaces {
		ifacePkg := importer.Import(iface.Named.Pkg.Name, iface.Named.Pkg.Path)
		paramName := iface.Named.Name.Lower()
		if i > 0 {
			g.w.W(",")
		}
		g.w.W("%s %s.%s", paramName, ifacePkg, iface.Named.Name)
	}

	g.w.W(",options ...ServerOption) {\n")

	g.w.W("opts := &serverOpts{}\nfor _, o := range options {\no(opts)\n}\n")

	var contextParamFound bool

	for _, iface := range g.Interfaces {
		ifaceType := iface.Named.Type.(*option.IfaceType)
		paramName := iface.Named.Name.Lower()

		for _, m := range ifaceType.Methods {
			mopt := g.MethodOptions[iface.Named.Name.Value+m.Name.Value]

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

			// replace brace indices for echo router
			urlPath = stdstrings.ReplaceAll(urlPath, "{", ":")
			urlPath = stdstrings.ReplaceAll(urlPath, "}", "")

			httpMethod := "GET"
			if mopt.RESTMethod.Take() != "" {
				httpMethod = stdstrings.ToUpper(mopt.RESTMethod.Take())
			}

			g.w.W("e.%s(%s, func(ctx %s.Context) (err error) {\n", httpMethod, strconv.Quote(urlPath), echoPkg)

			var (
				paramValues  []string
				paramContext *option.VarType
			)

			if len(m.Sig.Params) > 0 {
				g.w.W("var req struct {\n")
				for _, p := range m.Sig.Params {
					if p.IsContext && paramContext == nil {
						paramContext = p
					}
					g.w.W("%s %s `json:\"%s\"`\n", p.Name.Upper(), swipe.TypeString(p.Type, true, importer), p.Name)
					if p.IsVariadic {
						paramValues = append(paramValues, "req."+p.Name.Upper()+"...")
					} else {
						paramValues = append(paramValues, "req."+p.Name.Upper())
					}
				}
				g.w.W("}\n")
			}

			if len(headerVars) > 0 || len(queryVars) > 0 || len(paramVars) > 0 {
				g.w.W("r := ctx.Request()\n")
			}

			if paramContext != nil {
				contextParamFound = true
				g.w.W("req.%s = &contextWrapper{ctx: ctx, next: ctx.Request().Context()}\n", paramContext.Name.Upper())
			}

			if len(paramVars) > 0 {
				switch stdstrings.ToUpper(mopt.RESTMethod.Take()) {
				case "POST", "PUT", "PATCH":
					switch bodyType {
					case "json":
						jsonPkg := importer.Import("json", "encoding/json")
						fmtPkg := importer.Import("fmt", "fmt")
						pkgIO := importer.Import("io", "io")

						g.w.W("var data []byte\n")

						g.w.W("data, err = %s.ReadAll(r.Body)\n", pkgIO)
						g.w.WriteCheckErr("err", func() {
							g.w.W("return %s.NewHTTPError(500, %s.Sprintf(\"couldn't read body for %s: %%s\", err))\n", echoPkg, fmtPkg, m.Name.Upper())
						})

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
						g.w.W("return %s.NewHTTPError(400, %s.Sprintf(\"couldn't unmarshal body to %s: %%s\", err))\n", echoPkg, fmtPkg, m.Name.Upper())
						g.w.W("}\n")
					}
				}
			}
			if len(pathVars) > 0 {
				for _, pathVar := range pathVars {
					valueVar := "ctx.Param(" + strconv.Quote(pathVar.Param.Name.Value) + ")"

					convert.NewBuilder(importer).
						SetAssignVar("req." + strcase.ToCamel(pathVar.Param.Name.Value)).
						SetValueVar(valueVar).
						SetFieldName(pathVar.Param.Name).
						SetFieldType(pathVar.Param.Type).
						SetErrorReturn(func() string {
							fmtPkg := importer.Import("fmt", "fmt")
							return fmt.Sprintf("return %s.NewHTTPError(500, %s.Sprintf(\"convert error: %%v\", %s))", echoPkg, fmtPkg, valueVar)
						}).
						Write(&g.w)
				}
			}
			for _, headerVar := range headerVars {
				textProtoPkg := importer.Import("textproto", "net/textproto")
				g.w.W("if _, ok := r.Header[%s.CanonicalMIMEHeaderKey(%s)]; ok {\n", textProtoPkg, strconv.Quote(headerVar.Value))
				convert.NewBuilder(importer).
					SetFieldName(headerVar.Param.Name).
					SetFieldType(headerVar.Param.Type).
					SetAssignVar("req." + headerVar.Param.Name.Upper()).
					SetValueVar("r.Header.Get(" + strconv.Quote(headerVar.Value) + ")").
					SetErrorReturn(func() string {
						fmtPkg := importer.Import("fmt", "fmt")
						return fmt.Sprintf("return %s.NewHTTPError(500, %s.Sprintf(\"convert error: %%v\", %s))", echoPkg, fmtPkg, "req."+headerVar.Param.Name.Upper())
					}).
					Write(&g.w)
				if headerVar.IsRequired {
					g.w.W("} else {\nreturn %s.NewHTTPError(400, \"header %s is required\")\n}\n", echoPkg, headerVar.Value)
				} else {
					g.w.W("}\n")
				}
			}

			if len(mopt.RESTQueryVars.Value) > 0 {
				g.w.W("q := r.URL.Query()\n")

				for _, queryVar := range queryVars {

					if named, ok := queryVar.Param.Type.(*option.NamedType); ok {
						if st, ok := named.Type.(*option.StructType); ok {
							for _, field := range st.Fields {
								if tag, err := field.Tags.Get("json"); err == nil {

									g.w.W("if q.Has(" + strconv.Quote(tag.Value()) + ") {\n")

									convert.NewBuilder(importer).
										SetFieldName(field.Var.Name).
										SetFieldType(field.Var.Type).
										SetAssignVar("req." + queryVar.Param.Name.Upper() + "." + field.Var.Name.Upper()).
										SetValueVar("q.Get(" + strconv.Quote(tag.Value()) + ")").
										SetErrorReturn(func() string {
											fmtPkg := importer.Import("fmt", "fmt")
											return fmt.Sprintf("return %s.NewHTTPError(500, %s.Sprintf(\"convert error: %%v\", %s))", echoPkg, fmtPkg, "req."+queryVar.Param.Name.Upper())
										}).
										Write(&g.w)
									g.w.W("}\n")
								}
							}
						}
					} else {
						g.w.W("if _, ok := q[%s]; ok {\n", strconv.Quote(queryVar.Value))
						convert.NewBuilder(importer).
							SetFieldName(queryVar.Param.Name).
							SetFieldType(queryVar.Param.Type).
							SetAssignVar("req." + queryVar.Param.Name.Upper()).
							SetValueVar("q.Get(" + strconv.Quote(queryVar.Value) + ")").
							SetErrorReturn(func() string {
								fmtPkg := importer.Import("fmt", "fmt")
								return fmt.Sprintf("return %s.NewHTTPError(500, %s.Sprintf(\"convert error: %%v\", %s))", echoPkg, fmtPkg, "req."+queryVar.Param.Name.Upper())
							}).
							Write(&g.w)
						if queryVar.IsRequired {
							g.w.W("} else {\nreturn %s.NewHTTPError(400, \"query var %s is required\")\n}\n", echoPkg, queryVar.Value)
						} else {
							g.w.W("}\n")
						}
					}
				}
			}

			if len(m.Sig.Results) > 0 {
				for i, p := range m.Sig.Results {
					if i > 0 {
						g.w.W(", ")
					}
					g.w.W(p.Name.Value)
				}
				if len(m.Sig.Results) == 1 {
					g.w.W(" = ")
				} else {
					g.w.W(" := ")
				}
			}

			g.w.W("%s.%s(%s)\n", paramName, m.Name.Upper(), stdstrings.Join(paramValues, ","))

			if len(m.Sig.Results) > 0 {
				var results option.VarsType

				for _, result := range m.Sig.Results {
					if plugin.IsError(result) {
						g.w.WriteCheckErr(result.Name.Value, func() {
							g.w.W("return defaultErrorEncoder(ctx, %s)\n", result.Name)
						})
						continue
					}
					results = append(results, result)
				}
				if len(results) > 1 {
					g.w.W("response := map[string]interface{}{")
					for _, result := range results {
						if plugin.IsError(result) {
							continue
						}
						g.w.W("%s: %s,\n", strconv.Quote(result.Name.Value), result.Name)
					}
					g.w.W("}\n")
					g.w.W("return encodeResponseHTTP(ctx, response)\n")
				} else if len(results) == 1 {
					g.w.W("return encodeResponseHTTP(ctx, %s)\n", results[0].Name)
				} else {
					g.w.W("return nil\n")
				}
			} else {
				g.w.W("return nil\n")
			}

			g.w.W("}, append(append(opts.genericOpts.middlewares, opts.generic%s.middlewares...), opts.%s.middlewares...)...)\n", UcNameWithAppPrefix(iface)+"Opts", LcNameIfaceMethod(iface, m)+"Opts")
		}
	}

	g.w.W("\n}\n")

	if contextParamFound {
		contextPkg := importer.Import("context", "context")
		timePkg := importer.Import("time", "time")
		g.writeContextWrapper(echoPkg, contextPkg, timePkg)
	}

	return g.w.Bytes()
}

func (g *RoutesGenerator) writeContextWrapper(echoPkg, contextPkg, timePkg string) {
	g.w.W("type contextWrapper struct {\n")
	g.w.W("ctx %s.Context\n", echoPkg)
	g.w.W("next %s.Context\n}\n", contextPkg)
	g.w.W("func (e *contextWrapper) Deadline() (deadline %s.Time, ok bool) {\nreturn e.next.Deadline()\n}\n", timePkg)
	g.w.W("func (e *contextWrapper) Done() <-chan struct{} {\nreturn e.next.Done()\n}\n")
	g.w.W("func (e *contextWrapper) Err() error {\nreturn e.next.Err()\n}\n")
	g.w.W("func (e *contextWrapper) Value(key any) any {\nif k, ok := key.(string); ok {\nreturn e.ctx.Get(k)\n}\nreturn e.next.Value(key)\n}\n")
}

func (g *RoutesGenerator) writeDefaultErrorEncoder(echoPkg, httpPkg string) {
	g.w.W("type errorWrapper struct {\n")
	g.w.W("Error string `json:\"error\"`\n")
	g.w.W("Code string `json:\"code,omitempty\"`\n")
	g.w.W("Data interface{} `json:\"data,omitempty\"`\n")
	g.w.W("}\n")

	g.w.W("func defaultErrorEncoder(ctx %s.Context, err error) error {\n", echoPkg)

	g.w.W("var (\nerrData interface{}\nerrCode string\n)\n")
	g.w.W("if e, ok := err.(interface{ Data() interface{} }); ok {\n")
	g.w.W("errData = e.Data()\n")
	g.w.W("}\n")

	g.w.W("if e, ok := err.(interface{ Code() string }); ok {\n")
	g.w.W("errCode = e.Code()\n")
	g.w.W("}\n")

	g.w.W("ctx.Response().Header().Set(\"Content-Type\", \"application/json; charset=utf-8\")\n")
	g.w.W("if headerer, ok := err.(interface{ Headers() %s.Header }); ok {\n", httpPkg)
	g.w.W("for k, values := range headerer.Headers() {\n")
	g.w.W("for _, v := range values {\n")
	g.w.W("ctx.Response().Header().Add(k, v)")
	g.w.W("}\n}\n")

	g.w.W("}\n")
	g.w.W("code := %s.StatusInternalServerError\n", httpPkg)
	g.w.W("if sc, ok := err.(interface { StatusCode() int }); ok {\n")
	g.w.W("code = sc.StatusCode()\n")
	g.w.W("}\n")

	g.w.W("return ctx.JSON(code, errorWrapper{Error: err.Error(), Code: errCode, Data: errData})\n")
	g.w.W("}\n\n")
}

func (g *RoutesGenerator) writeEncodeResponseFunc(echoPkg, httpPkg string) {
	g.w.W("func encodeResponseHTTP(ctx %s.Context, response interface{}) (err error) {\n", echoPkg)
	g.w.W("if response != nil {\n")
	g.w.W("if cookie, ok := response.(interface{ HTTPCookies() []%s.Cookie }); ok {\n", httpPkg)
	g.w.W("for _, c := range cookie.HTTPCookies() {\n")
	g.w.W("ctx.SetCookie(&c)\n")
	g.w.W("}\n")
	g.w.W("}\n")
	g.w.W("if download, ok := response.(interface{\nContentType() string\nData() []byte\n}); ok {\n")
	g.w.W("return ctx.Blob(200, download.ContentType(), download.Data())\n")
	g.w.W("}")
	g.w.W("} else {\n")
	g.w.W("return ctx.Blob(204, \"text/plain; charset=utf-8\", nil)\n")
	g.w.W("}\n")
	g.w.W("return ctx.JSON(200, response)\n")
	g.w.W("}\n\n")
}

func (g *RoutesGenerator) OutputPath() string {
	return ""
}

func (g *RoutesGenerator) Filename() string {
	return "routes.go"
}
