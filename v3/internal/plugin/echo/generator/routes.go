package generator

import (
	"context"
	"fmt"
	"path"
	"strconv"
	stdstrings "strings"

	"github.com/swipe-io/swipe/v3/internal/convert"

	"github.com/swipe-io/swipe/v3/internal/plugin"
	"github.com/swipe-io/swipe/v3/option"

	"github.com/swipe-io/strcase"
	"github.com/swipe-io/swipe/v3/internal/plugin/echo/config"
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

	echoPkg := importer.Import("echo", "github.com/labstack/echo")

	g.w.W("func SetupRoutes(e *%s.Echo,", echoPkg)

	for i, iface := range g.Interfaces {
		ifacePkg := importer.Import(iface.Named.Pkg.Name, iface.Named.Pkg.Path)
		paramName := iface.Named.Name.Lower()
		g.w.W("%s %s.%s", paramName, ifacePkg, iface.Named.Name)
		if i > 0 {
			g.w.W(",")
		}
	}

	g.w.W(") {\n")

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
			httpMethod := "GET"
			if mopt.RESTMethod.Take() != "" {
				httpMethod = stdstrings.ToUpper(mopt.RESTMethod.Take())
			}
			g.w.W("e.%s(%s, func(ctx %s.Context) (err error) {\n", httpMethod, strconv.Quote(urlPath), echoPkg)

			var paramValues []string

			g.w.W("r := ctx.Request()\n")

			if len(m.Sig.Params) > 0 {
				g.w.W("var req struct {\n")
				for _, p := range m.Sig.Params {
					g.w.W("%[1]s %[2]s `json:\"%[1]s\"`\n", p.Name.Lower(), swipe.TypeString(p.Type, true, importer))
					paramValues = append(paramValues, "req."+p.Name.Lower())
				}
				g.w.W("}\n")
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

						ioutilPkg := importer.Import("ioutil", "io/ioutil")
						g.w.W("data, err := %s.ReadAll(r.Body)\n", ioutilPkg)
						g.w.WriteCheckErr("err", func() {
							g.w.W("return %s.Errorf(\"couldn't read body for %s: %%w\", err)\n", fmtPkg, m.Name)
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
						g.w.W("return %s.Errorf(\"couldn't unmarshal body to %s: %%w\", err)\n", fmtPkg, m.Name)
						g.w.W("}\n")
					}
				}
			}

			for _, headerVar := range headerVars {
				convert.NewBuilder(importer).
					SetFieldName(headerVar.Param.Name).
					SetFieldType(headerVar.Param.Type).
					SetAssignVar("req." + headerVar.Param.Name.Lower()).
					SetValueVar("r.Header.Get(" + strconv.Quote(headerVar.Value) + ")").
					SetErrorReturn(fmt.Sprintf("return %s.Errorf(\"convert error: %%v\", %s)", importer.Import("fmt", "fmt"), "req."+headerVar.Param.Name.Lower())).
					Write(&g.w)

				//valueID := "r.Header.Get(" + strconv.Quote(headerVar.Value) + ")"
				//g.w.WriteConvertType(importer, "req."+headerVar.Param.Name.Lower(), valueID, headerVar.Param, []string{"nil"}, "", false, "")
			}

			//if len(mopt.RESTQueryVars.Value) > 0 {
			//	g.w.W("q := r.URL.Query()\n")
			//}

			//for _, queryVar := range queryVars {
			//	valueID := "q.Get(" + strconv.Quote(queryVar.Value) + ")"
			//	if queryVar.IsRequired {
			//		fmtPkg := importer.Import("fmt", "fmt")
			//		g.w.W("if _, ok := q[\"%[1]s\"]; !ok {\nreturn %[2]s.Errorf(\"%[1]s required\")\n}\n", queryVar.Value, fmtPkg)
			//	}
			//	tmpID := queryVar.Param.Name.Value + "Str"
			//	g.w.W("%s := %s\n", tmpID, valueID)
			//	g.w.W("if %s != \"\" {\n", tmpID)
			//	g.w.WriteConvertType(importer, "req."+queryVar.Param.Name.Lower(), tmpID, queryVar.Param, []string{"nil"}, "", false, "")
			//	g.w.W("}\n")
			//}

			g.w.W("%s.%s(%s)\n", paramName, m.Name, stdstrings.Join(paramValues, ","))
			g.w.W("return nil\n})\n")
		}
	}

	g.w.W("\n}\n")

	return g.w.Bytes()

}

func (g *RoutesGenerator) OutputDir() string {
	return ""
}

func (g *RoutesGenerator) Filename() string {
	return "routes.go"
}
