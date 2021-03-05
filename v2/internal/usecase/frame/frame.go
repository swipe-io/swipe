package frame

import (
	"github.com/swipe-io/swipe/v2/internal/importer"
	"golang.org/x/tools/go/packages"
)

type Frame interface {
	Frame(data []byte) ([]byte, error)
}

type Factory interface {
	NewFrame(filename string, importer *importer.Importer, pkg *packages.Package) Frame
}
