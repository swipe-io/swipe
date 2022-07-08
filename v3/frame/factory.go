package frame

import (
	"path/filepath"
)

type Framer interface {
	Frame(data []byte) ([]byte, error)
}

func NewFrame(version string, filename string, imports []string, pkgName string, useDoNotEdit bool) Framer {
	ext := filepath.Ext(filename)
	switch ext {
	default:
		return NewBytesFrame()
	case ".go":
		return NewGolangFrame(imports, version, pkgName, useDoNotEdit)
	case ".js":
		return NewJSFrame(version)
	}
}
