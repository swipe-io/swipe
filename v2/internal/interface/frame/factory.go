package frame

import (
	"path/filepath"

	option2 "github.com/swipe-io/swipe/v2/internal/option"

	"github.com/swipe-io/swipe/v2/internal/importer"
	uf "github.com/swipe-io/swipe/v2/internal/usecase/frame"
)

func NewFrame(version string, filename string, importer *importer.Importer, pkg *option2.PackageType) uf.Frame {
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
