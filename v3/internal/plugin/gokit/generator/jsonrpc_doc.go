package generator

import (
	"context"

	"github.com/swipe-io/swipe/v3/internal/plugin"

	"github.com/swipe-io/swipe/v3/internal/plugin/gokit/config"
	"github.com/swipe-io/swipe/v3/option"
	"github.com/swipe-io/swipe/v3/writer"
)

type JSONRPCDocGenerator struct {
	w               writer.TextWriter
	AppName         string
	JSPkgImportPath string
	Interfaces      []*config.Interface
	IfaceErrors     map[string]map[string][]config.Error
	Output          string
}

func (g *JSONRPCDocGenerator) Generate(ctx context.Context) []byte {
	g.w.W("# %s JSONRPC Client\n\n", g.AppName)

	if g.JSPkgImportPath != "" {
		g.w.W("## Getting Started\n\n")
		g.w.W("You can install this with:\n\n```shell script\nnpm install --save-dev %s\n```\n\n", g.JSPkgImportPath)
		g.w.W("Import the package with the client:\n\n")
		g.w.W("```javascript\nimport API from \"%s\"\n```\n\n", g.JSPkgImportPath)
		g.w.W("Create a transport, only one method needs to be implemented: `doRequest(Array.<Object>) PromiseLike<Object>`.\n\n")
		g.w.W("For example:\n\n```javascript\nclass FetchTransport {\n    constructor(url) {\n      this.url = url;\n    }\n\n    doRequest(requests) {\n        return fetch(this.url, {method: \"POST\", body: JSON.stringify(requests)})\n    }\n}\n```\n\n")
		g.w.W("Now for a complete example:\n\n```javascript\nimport API from \"%s\"\nimport Transport from \"transport\"\n\nconst api = new API(new Transport(\"http://127.0.0.1\"))\n\n// call method here.\n```\n\n", g.JSPkgImportPath)
		g.w.W("## API\n## Methods\n\n")
	}

	defTypes := map[string]*option.NamedType{}
	responseTypes := map[string]option.VarsType{}

	for _, iface := range g.Interfaces {
		ifaceType := iface.Named.Type.(*option.IfaceType)

		for _, m := range ifaceType.Methods {
			name := docMethodName(iface, m)
			g.w.W("<a href=\"#%[1]s\">%[1]s</a>\n\n", name)
		}

		methodErrors := g.IfaceErrors[iface.Named.Name.Value]

		for _, m := range ifaceType.Methods {
			name := docMethodName(iface, m)

			errors := methodErrors[m.Name.Value]

			g.w.W("### <a name=\"%[1]s\"></a>%[1]s(", name)
			for i, p := range m.Sig.Params {
				if plugin.IsContext(p) {
					continue
				}
				nameds := extractNamed(p.Type)
				for _, named := range nameds {
					if _, ok := defTypes[named.ID()]; ok {
						continue
					}
					defTypes[named.ID()] = named
				}
				if p.IsVariadic {
					g.w.W(", ...%s", p.Name.Value)
				} else {
					g.w.W("%s", p.Name.Value)
				}
				if i > 0 && i != len(m.Sig.Params)-1 {
					g.w.W(", ")
				}
			}
			g.w.W(") ⇒")

			resultRen := plugin.LenWithoutErrors(m.Sig.Results)

			if resultRen == 0 {
				g.w.W("<code>void</code>")
			} else if resultRen > 0 {
				if resultRen == 1 {
					nameds := extractNamed(m.Sig.Results[0].Type)
					for _, named := range nameds {
						if _, ok := defTypes[named.ID()]; ok {
							continue
						}
						defTypes[named.ID()] = named
					}
					g.w.W("<code>%s</code>", jsDocType(m.Sig.Results[0].Type))
				} else if resultRen > 1 {
					responseName := m.Name.Value + "Response"
					g.w.W("<code>%s</code>", responseName)
					_, ok := responseTypes[responseName]
					if !ok {
						responseTypes[responseName] = m.Sig.Results
					}
					for _, p := range m.Sig.Results {
						if plugin.IsError(p) {
							continue
						}
						nameds := extractNamed(p.Type)
						for _, named := range nameds {
							if _, ok := defTypes[named.ID()]; ok {
								continue
							}
							defTypes[named.ID()] = named
						}
					}
				}
			}

			g.w.W("\n\n")

			if m.Comment != "" {
				g.w.W("%s\n", m.Comment)
				g.w.W("\n")
			}

			if len(errors) > 0 {
				errorsDub := map[int64]struct{}{}

				g.w.W("**Throws**:\n\n")
				for _, e := range errors {
					if _, ok := errorsDub[e.Code]; ok {
						continue
					}
					errorsDub[e.Code] = struct{}{}
					g.w.W("<code>%s</code>\n\n", jsErrorName(iface, e))
				}
				g.w.W("\n\n")
			}

			if plugin.LenWithoutContexts(m.Sig.Params) > 0 {
				g.w.W("| Param | Type | Description |\n|------|------|------|\n")
				for _, p := range m.Sig.Params {
					g.w.W("|%s|<code>%s</code>|%s|\n", p.Name.Value, jsDocType(p.Type), p.Comment)
				}
				g.w.W("\n")
			}
		}
	}

	g.w.W("\n")

	if len(defTypes) > 0 || len(responseTypes) > 0 {
		g.w.W("## Members\n\n")
	}

	for name, results := range responseTypes {
		g.w.W("### %s\n\n", name)
		g.w.W("| Field | Type | Description |\n|------|------|------|\n")
		for _, p := range results {
			if plugin.IsError(p) {
				continue
			}
			g.w.W("|%s|<code>%s</code>|%s|\n", p.Name.Value, jsDocType(p.Type), p.Comment)
		}
	}

	g.w.W("\n")

	for _, named := range defTypes {
		st, ok := named.Type.(*option.StructType)
		if !ok {
			continue
		}
		g.w.W("### %s\n\n", named.Name)
		g.w.W("| Field | Type | Description |\n|------|------|------|\n")
		for _, f := range st.Fields {
			if tag, err := f.Tags.Get("json"); err == nil {
				if tag.Name == "-" {
					continue
				}
			}
			g.w.W("|%s|<code>%s</code>|%s|\n", f.Var.Name.Value, jsDocType(f.Var.Type), f.Var.Comment)
		}
	}

	return g.w.Bytes()
}

func (g *JSONRPCDocGenerator) OutputPath() string {
	return g.Output
}

func (g *JSONRPCDocGenerator) Filename() string {
	return "jsonrpc_doc.md"
}
