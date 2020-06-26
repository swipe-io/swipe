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

type restGoClient struct {
	*writer.GoLangWriter

	info model.GenerateInfo
	o    model.ServiceOption
	i    *importer.Importer
}

func (g *restGoClient) Process(ctx context.Context) error {
	var (
		kithttpPkg string
		httpPkg    string
	)
	transportOpt := g.o.Transport

	if transportOpt.FastHTTP {
		kithttpPkg = g.i.Import("fasthttp", "github.com/l-vitaly/go-kit/transport/fasthttp")
	} else {
		kithttpPkg = g.i.Import("http", "github.com/go-kit/kit/transport/http")
	}
	if transportOpt.FastHTTP {
		httpPkg = g.i.Import("fasthttp", "github.com/valyala/fasthttp")
	} else {
		httpPkg = g.i.Import("http", "net/http")
	}
	jsonPkg := g.i.Import("ffjson", "github.com/pquerna/ffjson/ffjson")
	pkgIO := g.i.Import("io", "io")
	fmtPkg := g.i.Import("fmt", "fmt")
	contextPkg := g.i.Import("context", "context")
	urlPkg := g.i.Import("url", "net/url")

	clientType := "client" + g.o.ID
	typeStr := stdtypes.TypeString(g.o.Type, g.i.QualifyPkg)

	g.W("func NewClient%s%s(tgt string", g.o.Transport.Prefix, g.o.ID)

	g.W(" ,opts ...%[1]sClientOption", g.o.ID)

	g.W(") (%s, error) {\n", typeStr)

	g.W("c := &%s{}\n", clientType)

	g.W("for _, o := range opts {\n")
	g.W("o(c)\n")
	g.W("}\n")

	g.W("u, err := %s.Parse(tgt)\n", urlPkg)

	g.WriteCheckErr(func() {
		g.W("return nil, err")
	})

	for _, m := range g.o.Methods {
		epName := m.LcName + "Endpoint"

		mopt := transportOpt.MethodOptions[m.Name]

		httpMethod := mopt.MethodName
		if httpMethod == "" {
			if len(m.Params) > 0 {
				httpMethod = "POST"
			} else {
				httpMethod = "GET"
			}
		}

		pathStr := mopt.Path
		if pathStr == "" {
			pathStr = "/" + m.LcName
		}

		pathVars := []string{}
		for name, regexp := range mopt.PathVars {
			if p := m.Params.LookupField(name); p != nil {
				if regexp != "" {
					regexp = ":" + regexp
				}
				pathStr = stdstrings.Replace(pathStr, "{"+name+regexp+"}", "%s", -1)
				pathVars = append(pathVars, g.GetFormatType("req."+strings.UcFirst(p.Name()), p))
			}
		}
		queryVars := []string{}
		for fName, qName := range mopt.QueryVars {
			if p := m.Params.LookupField(fName); p != nil {
				queryVars = append(queryVars, strconv.Quote(qName), g.GetFormatType("req."+strings.UcFirst(p.Name()), p))
			}
		}
		headerVars := []string{}
		for fName, hName := range mopt.HeaderVars {
			if p := m.Params.LookupField(fName); p != nil {
				headerVars = append(headerVars, strconv.Quote(hName), g.GetFormatType("req."+strings.UcFirst(p.Name()), p))
			}
		}

		g.W("c.%s = %s.NewClient(\n", epName, kithttpPkg)
		if mopt.Expr != nil {
			g.WriteAST(mopt.Expr)
		} else {
			g.W(strconv.Quote(httpMethod))
		}
		g.W(",\n")
		g.W("u,\n")

		if mopt.ClientRequestFunc.Expr != nil {
			g.WriteAST(mopt.ClientRequestFunc.Expr)
		} else {
			g.W("func(_ %s.Context, r *%s.Request, request interface{}) error {\n", contextPkg, httpPkg)

			if len(m.Params) > 0 {
				g.W("req, ok := request.(%sRequest%s)\n", m.LcName, g.o.ID)
				g.W("if !ok {\n")
				g.W("return %s.Errorf(\"couldn't assert request as %sRequest%s, got %%T\", request)\n", fmtPkg, m.LcName, g.o.ID)
				g.W("}\n")
			}

			if transportOpt.FastHTTP {
				g.W("r.Header.SetMethod(")
			} else {
				g.W("r.Method = ")
			}
			if mopt.Expr != nil {
				g.WriteAST(mopt.Expr)
			} else {
				g.W(strconv.Quote(httpMethod))
			}
			if transportOpt.FastHTTP {
				g.W(")")
			}
			g.W("\n")

			if transportOpt.FastHTTP {
				g.W("r.SetRequestURI(")
			} else {
				g.W("r.URL.Path += ")
			}
			g.W("%s.Sprintf(%s, %s)", fmtPkg, strconv.Quote(pathStr), stdstrings.Join(pathVars, ","))

			if transportOpt.FastHTTP {
				g.W(")")
			}
			g.W("\n")

			if len(queryVars) > 0 {
				if transportOpt.FastHTTP {
					g.W("q := r.URI().QueryArgs()\n")
				} else {
					g.W("q := r.URL.Query()\n")
				}

				for i := 0; i < len(queryVars); i += 2 {
					g.W("q.Add(%s, %s)\n", queryVars[i], queryVars[i+1])
				}

				if transportOpt.FastHTTP {
					g.W("r.URI().SetQueryString(q.String())\n")
				} else {
					g.W("r.URL.RawQuery = q.Encode()\n")
				}
			}

			for i := 0; i < len(headerVars); i += 2 {
				g.W("r.Header.Add(%s, %s)\n", headerVars[i], headerVars[i+1])
			}

			switch stdstrings.ToUpper(httpMethod) {
			case "POST", "PUT", "PATCH":
				jsonPkg := g.i.Import("ffjson", "github.com/pquerna/ffjson/ffjson")

				g.W("data, err := %s.Marshal(req)\n", jsonPkg)
				g.W("if err != nil  {\n")
				g.W("return %s.Errorf(\"couldn't marshal request %%T: %%s\", req, err)\n", fmtPkg)
				g.W("}\n")

				if transportOpt.FastHTTP {
					g.W("r.SetBody(data)\n")
				} else {
					ioutilPkg := g.i.Import("ioutil", "io/ioutil")
					bytesPkg := g.i.Import("bytes", "bytes")

					g.W("r.Body = %s.NopCloser(%s.NewBuffer(data))\n", ioutilPkg, bytesPkg)
				}
			}
			g.W("return nil\n")
			g.W("}")
		}
		g.W(",\n")

		if mopt.ClientResponseFunc.Expr != nil {
			g.WriteAST(mopt.ClientResponseFunc.Expr)
		} else {
			g.W("func(_ %s.Context, r *%s.Response) (interface{}, error) {\n", contextPkg, httpPkg)

			statusCode := "r.StatusCode"
			if transportOpt.FastHTTP {
				statusCode = "r.StatusCode()"
			}

			g.W("if statusCode := %s; statusCode != %s.StatusOK {\n", statusCode, httpPkg)
			g.W("return nil, ErrorDecode(statusCode)\n")
			g.W("}\n")

			if len(m.Results) > 0 {
				var responseType string
				if m.ResultsNamed {
					responseType = fmt.Sprintf("%sResponse%s", m.LcName, g.o.ID)
				} else {
					responseType = stdtypes.TypeString(m.Results[0].Type(), g.i.QualifyPkg)
				}
				if mopt.WrapResponse.Enable {
					g.W("var resp struct {\nData %s `json:\"%s\"`\n}\n", responseType, mopt.WrapResponse.Name)
				} else {
					g.W("var resp %s\n", responseType)
				}
				if transportOpt.FastHTTP {
					g.W("err := %s.Unmarshal(r.Body(), ", jsonPkg)
				} else {
					ioutilPkg := g.i.Import("ioutil", "io/ioutil")

					g.W("b, err := %s.ReadAll(r.Body)\n", ioutilPkg)
					g.WriteCheckErr(func() {
						g.W("return nil, err\n")
					})
					g.W("err = %s.Unmarshal(b, ", jsonPkg)
				}

				g.W("&resp)\n")

				g.W("if err != nil && err != %s.EOF {\n", pkgIO)
				g.W("return nil, %s.Errorf(\"couldn't unmarshal body to %sResponse%s: %%s\", err)\n", fmtPkg, m.LcName, g.o.ID)
				g.W("}\n")

				if mopt.WrapResponse.Enable {
					g.W("return resp.Data, nil\n")
				} else {
					g.W("return resp, nil\n")
				}
			} else {
				g.W("return nil, nil\n")
			}

			g.W("}")
		}

		g.W(",\n")

		g.W("append(c.genericClientOption, c.%sClientOption...)...,\n", m.LcName)

		g.W(").Endpoint()\n")

		g.W(
			"c.%[1]sEndpoint = middlewareChain(append(c.genericEndpointMiddleware, c.%[1]sEndpointMiddleware...))(c.%[1]sEndpoint)\n",
			m.LcName,
		)
	}

	g.W("return c, nil\n")
	g.W("}\n")
	return nil
}

func (g *restGoClient) PkgName() string {
	return ""
}

func (g *restGoClient) OutputDir() string {
	return ""
}

func (g *restGoClient) Filename() string {
	return "client_rest_gen.go"
}

func (g *restGoClient) Imports() []string {
	return g.i.SortedImports()
}

func NewRestGoClient(info model.GenerateInfo, o model.ServiceOption, i *importer.Importer) Generator {
	return &restGoClient{info: info, o: o, i: i, GoLangWriter: writer.NewGoLangWriter(i)}
}
