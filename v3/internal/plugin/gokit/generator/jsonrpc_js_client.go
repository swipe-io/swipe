package generator

import (
	"context"
	"fmt"
	"strings"

	"github.com/swipe-io/swipe/v3/internal/plugin"

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

	mw := writer.TextWriter{}

	defTypes := map[string]*option.NamedType{}

	for _, iface := range g.Interfaces {
		ifaceType := iface.Named.Type.(*option.IfaceType)

		mw.W("class JSONRPCClient%s {\n", UcNameJS(iface))
		mw.W("constructor(transport) {\n")
		mw.W("this.scheduler = new JSONRPCScheduler(transport);\n")
		mw.W("}\n\n")

		for _, m := range ifaceType.Methods {
			mw.W("/**\n")
			if m.Comment != "" {
				mw.W("* %s\n", m.Comment)
				mw.W("*\n")
			}
			for _, p := range m.Sig.Params {
				if plugin.IsContext(p) {
					continue
				}
				nameds := extractNamed(p.Type)
				for _, named := range nameds {
					if _, ok := defTypes[named.ID()]; !ok {
						defTypes[named.ID()] = named
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

				results := make([]string, 0, len(m.Sig.Results))
				for _, p := range m.Sig.Results {
					if plugin.IsError(p) {
						continue
					}
					nameds := extractNamed(p.Type)
					for _, named := range nameds {
						if _, ok := defTypes[named.ID()]; !ok {
							defTypes[named.ID()] = named
						}
					}
					if m.Sig.IsNamed {
						results = append(results, fmt.Sprintf("%s: %s", p.Name, jsDocType(p.Type)))
					} else {
						mw.W(jsDocType(p.Type))
					}
				}
				if m.Sig.IsNamed {
					mw.W("{%s}", strings.Join(results, ","))
				}
				mw.W(">}\n")
			}

			ifaceErrors := g.IfaceErrors[iface.Named.Name.Value]
			methodErrors := ifaceErrors[m.Name.Value]

			errorsDub := map[int64]struct{}{}
			for _, e := range methodErrors {
				if _, ok := errorsDub[e.Code]; !ok {
					errorsDub[e.Code] = struct{}{}
					mw.W("* @throws {%s}\n", jsErrorName(iface, e))
				}
			}

			mw.W("**/\n")
			mw.W("%s(", m.Name.Lower())

			params := make([]string, 0, len(m.Sig.Params))
			for _, p := range m.Sig.Params {
				if plugin.IsContext(p) {
					continue
				}
				name := p.Name.Value
				if p.IsVariadic {
					name = "..." + name
				}
				params = append(params, name)
			}
			mw.W(strings.Join(params, ","))

			var prefix string
			if iface.Namespace != "" {
				prefix = iface.Namespace + "."
			}

			mw.W(") {\n")
			mw.W("return this.scheduler.__scheduleRequest(\"%s\", {", prefix+m.Name.Lower())

			requestParams := make([]string, 0, len(m.Sig.Params))
			for _, p := range m.Sig.Params {
				if plugin.IsContext(p) {
					continue
				}
				requestParams = append(requestParams, fmt.Sprintf("%[1]s:%[1]s", p.Name))
			}
			mw.W(strings.Join(requestParams, ","))

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

	errorsDub := map[int64]struct{}{}
	for _, iface := range g.Interfaces {
		ifaceType := iface.Named.Type.(*option.IfaceType)
		ifaceErrors := g.IfaceErrors[iface.Named.Name.Value]
		for _, method := range ifaceType.Methods {
			methodErrors := ifaceErrors[method.Name.Value]
			for _, e := range methodErrors {
				if _, ok := errorsDub[e.Code]; ok {
					continue
				}
				errorsDub[e.Code] = struct{}{}
				g.w.W(
					"export class %[1]s extends JSONRPCError {\nconstructor(message, data) {\nsuper(message, \"%[1]s\", %[2]d, data);\n}\n}\n",
					jsErrorName(iface, e), e.Code,
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
			g.w.W("return new JSONRPCError(\"%s: \"+e.message, \"UnknownError\", e.code, e.data);\n", LcNameIfaceMethod(iface, method))

			errorsDub := map[int64]struct{}{}
			for _, e := range methodErrors {
				if _, ok := errorsDub[e.Code]; ok {
					continue
				}
				errorsDub[e.Code] = struct{}{}

				g.w.W("case %d:\n", e.Code)
				g.w.W("return new %s(e.message, e.data);\n", jsErrorName(iface, e))
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
