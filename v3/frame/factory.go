package frame

import (
	"path/filepath"

	"github.com/swipe-io/swipe/v3/internal/importer"
)

type Framer interface {
	Frame(data []byte) ([]byte, error)
}

func NewFrame(version string, filename string, importer *importer.Importer, pkgName string, useDoNotEdit bool) Framer {
	ext := filepath.Ext(filename)
	switch ext {
	default:
		return NewBytesFrame()
	case ".go":
		return NewGolangFrame(importer, version, pkgName, useDoNotEdit)
	case ".js":
		return NewJSFrame(version)
	}
}
