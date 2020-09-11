package writer

import (
	"go/ast"
	"go/printer"
	"io"

	"github.com/swipe-io/swipe/v2/internal/importer"
)

func WriteAST(w io.Writer, i *importer.Importer, node ast.Node) {
	node = i.RewritePkgRefs(node)
	if err := printer.Fprint(w, i.Pkg().Fset, node); err != nil {
		panic(err)
	}
}
