package generator

import (
	"context"
	"fmt"
	stdtypes "go/types"
	"strconv"
	stdstrings "strings"

	"github.com/swipe-io/swipe/v2/internal/usecase/generator"

	"github.com/swipe-io/swipe/v2/internal/domain/model"
	"github.com/swipe-io/swipe/v2/internal/importer"
	"github.com/swipe-io/swipe/v2/internal/strings"
	"github.com/swipe-io/swipe/v2/internal/writer"
)

type restGoClient struct {
	writer.GoLangWriter
	serviceID      string
	serviceType    stdtypes.Type
	serviceMethods []model.ServiceMethod
	transport      model.TransportOption
	i              *importer.Importer
}

func (g *restGoClient) Prepare(ctx context.Context) error {
	return nil
}

func (g *restGoClient) Process(ctx context.Context) error {

	var (
		kitHTTPPkg string
		contextPkg string
		httpPkg    string
		jsonPkg    string
		fmtPkg     string
		urlPkg     string
		netPkg     string
		stringsPkg string
		pkgIO      string
	)

	if len(g.serviceMethods) > 0 {
		if g.transport.FastHTTP {
			kitHTTPPkg = g.i.Import("fasthttp", "github.com/l-vitaly/go-kit/transport/fasthttp")
		} else {
			kitHTTPPkg = g.i.Import("http", "github.com/go-kit/kit/transport/http")
		}
		if g.transport.FastHTTP {
			httpPkg = g.i.Import("fasthttp", "github.com/valyala/fasthttp")
		} else {
			httpPkg = g.i.Import("http", "net/http")
		}
		jsonPkg = g.i.Import("ffjson", "github.com/pquerna/ffjson/ffjson")
		pkgIO = g.i.Import("io", "io")
		fmtPkg = g.i.Import("fmt", "fmt")
		contextPkg = g.i.Import("context", "context")
		urlPkg = g.i.Import("url", "net/url")
		netPkg = g.i.Import("net", "net")
		stringsPkg = g.i.Import("strings", "strings")
	}

	clientType := "client" + g.serviceID
	typeStr := stdtypes.TypeString(g.serviceType, g.i.QualifyPkg)

	g.W("func NewClient%s%s(tgt string", g.transport.Prefix, g.serviceID)

	g.W(" ,opts ...%[1]sClientOption", g.serviceID)

	g.W(") (%s, error) {\n", typeStr)

	g.W("c := &%s{}\n", clientType)

	g.W("for _, o := range opts {\n")
	g.W("o(c)\n")
	g.W("}\n")

	if len(g.serviceMethods) > 0 {
		g.W("if %s.HasPrefix(tgt, \"[\") {\n", stringsPkg)
		g.W("host, port, err := %s.SplitHostPort(tgt)\n", netPkg)
		g.WriteCheckErr(func() {
			g.W("return nil, err")
		})
		g.W("tgt = host + \":\" + port\n")
		g.W("}\n")

		g.W("u, err := %s.Parse(tgt)\n", urlPkg)

		g.WriteCheckErr(func() {
			g.W("return nil, err")
		})

		g.W("if u.Scheme == \"\" {\n")
		g.W("u.Scheme = \"https\"")
		g.W("}\n")
	}

	for _, m := range g.serviceMethods {
		epName := m.LcName + "Endpoint"

		mopt := g.transport.MethodOptions[m.Name]

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

		var (
			pathVars   []string
			queryVars  []string
			headerVars []string
		)

		for _, p := range m.Params {
			if regexp, ok := mopt.PathVars[p.Name()]; ok {
				if regexp != "" {
					regexp = ":" + regexp
				}
				pathStr = stdstrings.Replace(pathStr, "{"+p.Name()+regexp+"}", "%s", -1)
				pathVars = append(pathVars, g.GetFormatType(g.i.Import, "req."+strings.UcFirst(p.Name()), p))
			} else if qName, ok := mopt.QueryVars[p.Name()]; ok {
				queryVars = append(queryVars, strconv.Quote(qName), g.GetFormatType(g.i.Import, "req."+strings.UcFirst(p.Name()), p))
			} else if hName, ok := mopt.HeaderVars[p.Name()]; ok {
				headerVars = append(headerVars, strconv.Quote(hName), g.GetFormatType(g.i.Import, "req."+strings.UcFirst(p.Name()), p))
			}
		}

		g.W("c.%s = %s.NewClient(\n", epName, kitHTTPPkg)
		if mopt.Expr != nil {
			writer.WriteAST(g, g.i, mopt.Expr)
		} else {
			g.W(strconv.Quote(httpMethod))
		}
		g.W(",\n")
		g.W("u,\n")

		if mopt.ClientRequestFunc.Expr != nil {
			writer.WriteAST(g, g.i, mopt.ClientRequestFunc.Expr)
		} else {
			g.W("func(_ %s.Context, r *%s.Request, request interface{}) error {\n", contextPkg, httpPkg)

			if len(m.Params) > 0 {
				g.W("req, ok := request.(%sRequest%s)\n", m.LcName, g.serviceID)
				g.W("if !ok {\n")
				g.W("return %s.Errorf(\"couldn'tpl assert request as %sRequest%s, got %%T\", request)\n", fmtPkg, m.LcName, g.serviceID)
				g.W("}\n")
			}

			if g.transport.FastHTTP {
				g.W("r.Header.SetMethod(")
			} else {
				g.W("r.Method = ")
			}
			if mopt.Expr != nil {
				writer.WriteAST(g, g.i, mopt.Expr)
			} else {
				g.W(strconv.Quote(httpMethod))
			}
			if g.transport.FastHTTP {
				g.W(")")
			}
			g.W("\n")

			if g.transport.FastHTTP {
				g.W("r.SetRequestURI(")
			} else {
				g.W("r.URL.Path += ")
			}
			g.W("%s.Sprintf(%s, %s)", fmtPkg, strconv.Quote(pathStr), stdstrings.Join(pathVars, ","))

			if g.transport.FastHTTP {
				g.W(")")
			}
			g.W("\n")

			if len(queryVars) > 0 {
				if g.transport.FastHTTP {
					g.W("q := r.URI().QueryArgs()\n")
				} else {
					g.W("q := r.URL.Query()\n")
				}

				for i := 0; i < len(queryVars); i += 2 {
					g.W("q.Add(%s, %s)\n", queryVars[i], queryVars[i+1])
				}

				if g.transport.FastHTTP {
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
				g.W("return %s.Errorf(\"couldn'tpl marshal request %%T: %%s\", req, err)\n", fmtPkg)
				g.W("}\n")

				if g.transport.FastHTTP {
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
			writer.WriteAST(g, g.i, mopt.ClientResponseFunc.Expr)
		} else {
			g.W("func(_ %s.Context, r *%s.Response) (interface{}, error) {\n", contextPkg, httpPkg)

			statusCode := "r.StatusCode"
			if g.transport.FastHTTP {
				statusCode = "r.StatusCode()"
			}

			g.W("if statusCode := %s; statusCode != %s.StatusOK {\n", statusCode, httpPkg)
			g.W("return nil, ErrorDecode(statusCode)\n")
			g.W("}\n")

			if len(m.Results) > 0 {
				var responseType string
				if m.ResultsNamed {
					responseType = fmt.Sprintf("%sResponse%s", m.LcName, g.serviceID)
				} else {
					responseType = stdtypes.TypeString(m.Results[0].Type(), g.i.QualifyPkg)
				}
				if mopt.WrapResponse.Enable {
					g.W("var resp struct {\nData %s `json:\"%s\"`\n}\n", responseType, mopt.WrapResponse.Name)
				} else {
					g.W("var resp %s\n", responseType)
				}
				if g.transport.FastHTTP {
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
				g.W("return nil, %s.Errorf(\"couldn'tpl unmarshal body to %sResponse%s: %%s\", err)\n", fmtPkg, m.LcName, g.serviceID)
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
	return "client_gen.go"
}

func (g *restGoClient) SetImporter(i *importer.Importer) {
	g.i = i
}

func NewRestGoClient(
	serviceID string,
	serviceType stdtypes.Type,
	serviceMethods []model.ServiceMethod,
	transport model.TransportOption,
) generator.Generator {
	return &restGoClient{
		serviceID:      serviceID,
		serviceType:    serviceType,
		serviceMethods: serviceMethods,
		transport:      transport,
	}
}
