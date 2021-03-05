package writer

import (
	"go/ast"
	"go/printer"
	stdtypes "go/types"
	"io"

	"github.com/swipe-io/swipe/v2/internal/importer"
)

func WriteAST(w io.Writer, i *importer.Importer, node ast.Node) {
	node = i.RewritePkgRefs(node)
	if err := printer.Fprint(w, i.Pkg().Fset, node); err != nil {
		panic(err)
	}
}

func isNumeric(kind stdtypes.BasicKind) bool {
	switch kind {
	default:
		return false
	case stdtypes.Uint,
		stdtypes.Uint8,
		stdtypes.Uint16,
		stdtypes.Uint32,
		stdtypes.Uint64,
		stdtypes.Int,
		stdtypes.Int8,
		stdtypes.Int16,
		stdtypes.Int32,
		stdtypes.Int64,
		stdtypes.Float32,
		stdtypes.Float64:
		return true
	}
}
