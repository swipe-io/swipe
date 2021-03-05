package frame

import (
	"path/filepath"

	"github.com/swipe-io/swipe/v2/internal/importer"
	uf "github.com/swipe-io/swipe/v2/internal/usecase/frame"
	"golang.org/x/tools/go/packages"
)

type frameFactory struct {
	version string
}

func (f *frameFactory) NewFrame(
	filename string,
	importer *importer.Importer,
	pkg *packages.Package,
) uf.Frame {
	ext := filepath.Ext(filename)
	switch ext {
	default:
		return NewBytesFrame()
	case ".go":
		return NewGolangFrame(importer, f.version, pkg.Name)
	case ".js":
		return NewJSFrame(f.version)
	}
}

func NewFrameFactory(version string) uf.Factory {
	return &frameFactory{version: version}
}
