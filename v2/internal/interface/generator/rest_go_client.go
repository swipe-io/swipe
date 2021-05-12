package generator

import (
	"context"
	stdtypes "go/types"
	"path"
	"strconv"
	stdstrings "strings"

	"github.com/swipe-io/strcase"
	"github.com/swipe-io/swipe/v2/internal/domain/model"
	"github.com/swipe-io/swipe/v2/internal/importer"
	"github.com/swipe-io/swipe/v2/internal/usecase/generator"
	"github.com/swipe-io/swipe/v2/internal/writer"
)

type restGoClientOptionsGateway interface {
	Interfaces() model.Interfaces
	MethodOption(m model.ServiceMethod) model.MethodOption
	UseFast() bool
}

type restGoClient struct {
	writer.GoLangWriter
	options restGoClientOptionsGateway
	i       *importer.Importer
}

func (g *restGoClient) Prepare(_ context.Context) error {
	return nil
}

func (g *restGoClient) Process(_ context.Context) error {
	for i := 0; i < g.options.Interfaces().Len(); i++ {
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
		iface := g.options.Interfaces().At(i)

		name := iface.UcName()
		if iface.Namespace() != "" {
			name = strcase.ToCamel(iface.Namespace())
		}
		clientType := name + "Client"

		if g.options.UseFast() {
			kitHTTPPkg = g.i.Import("fasthttp", "github.com/l-vitaly/go-kit/transport/fasthttp")
		} else {
			kitHTTPPkg = g.i.Import("http", "github.com/go-kit/kit/transport/http")
		}
		if g.options.UseFast() {
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

		if g.options.Interfaces().Len() == 1 {
			g.W("// Deprecated\nfunc NewClientREST(tgt string")
			g.W(" ,options ...ClientOption")
			g.W(") (*%s, error) {\n", clientType)
			g.W("return NewClientREST%s(tgt, options...)", name)
			g.W("}\n")
		}

		g.W("func NewClientREST%s(tgt string", name)
		g.W(" ,options ...ClientOption")
		g.W(") (*%s, error) {\n", clientType)
		g.W("opts := &clientOpts{}\n")
		g.W("c := &%s{}\n", clientType)
		g.W("for _, o := range options {\n")
		g.W("o(opts)\n")
		g.W("}\n")

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

		for _, m := range iface.Methods() {
			epName := m.IfaceLcName + "Endpoint"
			mopt := g.options.MethodOption(m)

			httpMethod := mopt.MethodName
			if httpMethod == "" {
				httpMethod = "GET"
			}

			pathStr := mopt.Path
			if pathStr == "" {
				pathStr = path.Join("/", strcase.ToKebab(m.Name))
			}

			if iface.Namespace() != "" {
				pathStr = path.Join("/", strcase.ToKebab(iface.Namespace()), "/", pathStr)
			}

			var (
				pathVars   []*stdtypes.Var
				queryVars  []*stdtypes.Var
				headerVars []*stdtypes.Var
			)

			for _, p := range m.Params {
				if regexp, ok := mopt.PathVars[p.Name()]; ok {
					if regexp != "" {
						regexp = ":" + regexp
					}
					pathStr = stdstrings.Replace(pathStr, "{"+p.Name()+regexp+"}", "%s", -1)
					pathVars = append(pathVars, p)
				} else if _, ok := mopt.QueryVars[p.Name()]; ok {
					queryVars = append(queryVars, p)
				} else if _, ok := mopt.HeaderVars[p.Name()]; ok {
					headerVars = append(headerVars, p)
				}
			}

			remainingParams := len(m.Params) - (len(pathVars) + len(queryVars) + len(headerVars))

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
					g.W("req, ok := request.(%s)\n", m.NameRequest)
					g.W("if !ok {\n")
					g.W("return %s.Errorf(\"couldn't assert request as %s, got %%T\", request)\n", fmtPkg, m.NameRequest)
					g.W("}\n")
				}

				if g.options.UseFast() {
					g.W("r.Header.SetMethod(")
				} else {
					g.W("r.Method = ")
				}
				if mopt.Expr != nil {
					writer.WriteAST(g, g.i, mopt.Expr)
				} else {
					g.W(strconv.Quote(httpMethod))
				}
				if g.options.UseFast() {
					g.W(")")
				}
				g.W("\n")

				pathVarNames := make([]string, 0, len(pathVars))
				for _, p := range pathVars {
					name := p.Name() + "Str"
					pathVarNames = append(pathVarNames, name)
					g.WriteFormatType(g.i.Import, name, "req."+strcase.ToCamel(p.Name()), p)
				}
				if g.options.UseFast() {
					g.W("r.SetRequestURI(")
				} else {
					g.W("r.URL.Path += ")
				}
				if len(pathVars) > 0 {
					g.W("%s.Sprintf(%s, %s)", fmtPkg, strconv.Quote(pathStr), stdstrings.Join(pathVarNames, ", "))
				} else {
					g.W(strconv.Quote(pathStr))
				}

				if g.options.UseFast() {
					g.W(")")
				}

				g.W("\n")

				if len(queryVars) > 0 {
					if g.options.UseFast() {
						g.W("q := r.URI().QueryArgs()\n")
					} else {
						g.W("q := r.URL.Query()\n")
					}
					for _, p := range queryVars {
						name := p.Name() + "Str"
						g.WriteFormatType(g.i.Import, name, "req."+strcase.ToCamel(p.Name()), p)
						g.W("q.Add(%s, %s)\n", strconv.Quote(mopt.QueryVars[p.Name()]), name)
					}
					if g.options.UseFast() {
						g.W("r.URI().SetQueryString(q.String())\n")
					} else {
						g.W("r.URL.RawQuery = q.Encode()\n")
					}
				}
				for _, p := range headerVars {
					name := p.Name() + "Str"
					g.WriteFormatType(g.i.Import, name, "req."+strcase.ToCamel(p.Name()), p)
					g.W("r.Header.Add(%s, %s)\n", strconv.Quote(mopt.HeaderVars[p.Name()]), name)
				}
				switch stdstrings.ToUpper(httpMethod) {
				case "POST", "PUT", "PATCH":
					if remainingParams > 0 {
						jsonPkg := g.i.Import("ffjson", "github.com/pquerna/ffjson/ffjson")

						g.W("data, err := %s.Marshal(req)\n", jsonPkg)
						g.W("if err != nil  {\n")
						g.W("return %s.Errorf(\"couldn't marshal request %%T: %%s\", req, err)\n", fmtPkg)
						g.W("}\n")

						if g.options.UseFast() {
							g.W("r.SetBody(data)\n")
						} else {
							ioutilPkg := g.i.Import("ioutil", "io/ioutil")
							bytesPkg := g.i.Import("bytes", "bytes")

							g.W("r.Body = %s.NopCloser(%s.NewBuffer(data))\n", ioutilPkg, bytesPkg)
						}
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
				if g.options.UseFast() {
					statusCode = "r.StatusCode()"
				}

				g.W("if statusCode := %s; statusCode != %s.StatusOK {\n", statusCode, httpPkg)
				g.W("return nil, %sErrorDecode(statusCode)\n", m.IfaceLcName)
				g.W("}\n")

				if len(m.Results) > 0 {
					var responseType string
					if m.ResultsNamed {
						responseType = m.NameRequest
					} else {
						responseType = stdtypes.TypeString(m.Results[0].Type(), g.i.QualifyPkg)
					}
					if mopt.WrapResponse.Enable {
						g.W("var resp struct {\nData %s `json:\"%s\"`\n}\n", responseType, mopt.WrapResponse.Name)
					} else {
						g.W("var resp %s\n", responseType)
					}
					if g.options.UseFast() {
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
					g.W("return nil, %s.Errorf(\"couldn't unmarshal body to %s: %%s\", err)\n", fmtPkg, m.NameRequest)
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

			g.W("append(opts.genericClientOption, opts.%sClientOption...)...,\n", m.IfaceLcName)

			g.W(").Endpoint()\n")

			g.W(
				"c.%[1]sEndpoint = middlewareChain(append(opts.genericEndpointMiddleware, opts.%[2]sEndpointMiddleware...))(c.%[1]sEndpoint)\n",
				m.IfaceLcName,
				m.IfaceLcName,
			)
		}
		g.W("return c, nil\n")
		g.W("}\n\n")
	}
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
	options restGoClientOptionsGateway,
) generator.Generator {
	return &restGoClient{
		options: options,
	}
}
