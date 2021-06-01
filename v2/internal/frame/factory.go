package frame

import (
	"path/filepath"

	"github.com/swipe-io/swipe/v2/internal/importer"
	"github.com/swipe-io/swipe/v2/option"
)

type Framer interface {
	Frame(data []byte) ([]byte, error)
}

func NewFrame(version string, filename string, importer *importer.Importer, pkg *option.PackageType) Framer {
	ext := filepath.Ext(filename)
	switch ext {
	default:
		return NewBytesFrame()
	case ".go":
		return NewGolangFrame(importer, version, pkg.Name)
	case ".js":
		return NewJSFrame(version)
	}
}
