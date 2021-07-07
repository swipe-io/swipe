package generator

import (
	"context"

	"github.com/swipe-io/swipe/v3/internal/plugin/gokit/config"
	"github.com/swipe-io/swipe/v3/option"
	"github.com/swipe-io/swipe/v3/writer"
)

type JSONRPCJSClientGenerator struct {
	w           writer.GoWriter
	Interfaces  []*config.Interface
	IfaceErrors map[string]map[string][]config.Error
}

func (g *JSONRPCJSClientGenerator) Generate(ctx context.Context) []byte {
	g.w.W(jsonRPCClientBase)

	mw := writer.BaseWriter{}

	defTypes := map[string]*option.NamedType{}

	for _, iface := range g.Interfaces {
		ifaceType := iface.Named.Type.(*option.IfaceType)

		mw.W("class JSONRPCClient%s {\n", UcNameJS(iface))
		mw.W("constructor(transport) {\n")
		mw.W("this.scheduler = new JSONRPCScheduler(transport);\n")
		mw.W("}\n\n")

		for _, m := range ifaceType.Methods {
			resultLen := LenWithoutErrors(m.Sig.Results)
			paramLen := LenWithoutContexts(m.Sig.Params)

			mw.W("/**\n")
			if m.Comment != "" {
				mw.W("* %s\n", m.Comment)
				mw.W("*\n")
			}
			for _, p := range m.Sig.Params {
				if IsContext(p) {
					continue
				}
				if t, ok := p.Type.(*option.NamedType); ok {
					key := t.ID()
					if _, ok := defTypes[key]; !ok {
						defTypes[key] = t
					}
				}
				if p.IsVariadic {
					mw.W("* @param {...%s} %s\n", jsDocType(p.Type), p.Name)
				} else {
					mw.W("* @param {%s} %s\n", jsDocType(p.Type), p.Name)
				}
			}
			if len(m.Sig.Results) > 0 {
				mw.W("* @return {PromiseLike<")
				if m.Sig.IsNamed {
					mw.W("{")
				}
				for i, p := range m.Sig.Results {
					if IsError(p) {
						continue
					}
					if t, ok := p.Type.(*option.NamedType); ok {
						key := t.ID()
						if _, ok := defTypes[key]; !ok {
							defTypes[key] = t
						}
					}
					if i != resultLen-1 {
						mw.W(", ")
					}
					if m.Sig.IsNamed {
						mw.W("%s: ", p.Name)
					}
					mw.W(jsDocType(p.Type))
				}
				if m.Sig.IsNamed {
					mw.W("}")
				}
				mw.W(">}\n")

			}

			mw.W("**/\n")
			mw.W("%s(", m.Name.Lower())

			for i, p := range m.Sig.Params {
				if IsContext(p) {
					continue
				}
				if p.IsVariadic {
					mw.W("...")
				}
				mw.W(p.Name.Value)

				if i != paramLen {
					mw.W(",")
				}
			}

			var prefix string
			if iface.Namespace != "" {
				prefix = iface.Namespace + "."
			}

			mw.W(") {\n")
			mw.W("return this.scheduler.__scheduleRequest(\"%s\", {", prefix+m.Name.Lower())

			for i, p := range m.Sig.Params {
				if IsContext(p) {
					continue
				}
				mw.W("%[1]s:%[1]s", p.Name)
				if i != paramLen {
					mw.W(",")
				}
			}

			mw.W("}).catch(e => { throw ")
			mw.W("%s%sConvertError(e)", LcNameWithAppPrefix(iface), m.Name)
			mw.W("; })\n")

			mw.W("}\n")
		}
		mw.W("}\n\n")
	}

	g.w.W(mw.String())

	if len(g.Interfaces) > 1 {
		g.w.W("class JSONRPCClient {\n")
		g.w.W("constructor(transport) {\n")

		for _, iface := range g.Interfaces {
			g.w.W("this.%s = new JSONRPCClient%s(transport);\n", LcNameJS(iface), UcNameJS(iface))
		}
		g.w.W("}\n")
		g.w.W("}\n")

		g.w.W("export default JSONRPCClient\n\n")
	} else if len(g.Interfaces) == 1 {
		g.w.W("export default JSONRPCClient%s\n\n", UcNameJS(g.Interfaces[0]))
	}

	httpErrorsDub := map[string]struct{}{}

	for _, iface := range g.Interfaces {
		ifaceType := iface.Named.Type.(*option.IfaceType)

		ifaceErrors := g.IfaceErrors[iface.Named.Name.Value]

		for _, method := range ifaceType.Methods {
			methodErrors := ifaceErrors[method.Name.Value]
			for _, e := range methodErrors {
				if _, ok := httpErrorsDub[e.Name]; ok {
					continue
				}
				httpErrorsDub[e.Name] = struct{}{}
				g.w.W(
					"export class %[1]s extends JSONRPCError {\nconstructor(message, data) {\nsuper(message, \"%[1]s\", %[2]d, data);\n}\n}\n",
					e.Name, e.Code,
				)
			}
		}
	}
	for _, iface := range g.Interfaces {
		ifaceType := iface.Named.Type.(*option.IfaceType)
		ifaceErrors := g.IfaceErrors[iface.Named.Name.Value]

		for _, method := range ifaceType.Methods {
			methodErrors := ifaceErrors[method.Name.Value]

			g.w.W("function %s%sConvertError(e) {\n", LcNameWithAppPrefix(iface), method.Name)
			g.w.W("switch(e.code) {\n")
			g.w.W("default:\n")
			g.w.W("return new JSONRPCError(e.message, \"UnknownError\", e.code, e.data);\n")
			for _, e := range methodErrors {
				g.w.W("case %d:\n", e.Code)
				g.w.W("return new %s(e.message, e.data);\n", e.Name)
			}
			g.w.W("}\n}\n")
		}
	}
	for _, t := range defTypes {
		switch t.Pkg.Path {
		case "github.com/google/uuid", "github.com/pborman/uuid", "encoding/json", "time":
			continue
		}
		g.w.W(jsTypeDef(t))
	}
	return g.w.Bytes()
}

func (g *JSONRPCJSClientGenerator) OutputDir() string {
	return ""
}

func (g *JSONRPCJSClientGenerator) Filename() string {
	return "jsonrpc_client.js"
}
