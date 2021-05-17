package generator

import (
	"context"

	"github.com/swipe-io/swipe/v2/internal/writer"
)

type JSONRPCDocGenerator struct {
	w               writer.BaseWriter
	AppName         string
	JSPkgImportPath string
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
		g.w.W("Now for a complete example:\n\n```javascript\nimport API from \"%s\"\nimport Helpers from \"transport\"\n\nconst api = new API(new Helpers(\"http://127.0.0.1\"))\n\n// call method here.\n```\n\n", g.JSPkgImportPath)
		g.w.W("## API\n## Methods\n\n")
	}

	return g.w.Bytes()
}

func (g *JSONRPCDocGenerator) OutputDir() string {
	return ""
}

func (g *JSONRPCDocGenerator) Filename() string {
	return "jsonrpc_doc.md"
}
