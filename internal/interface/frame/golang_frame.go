package frame

import (
	"bytes"
	"fmt"
	"go/format"

	"github.com/swipe-io/swipe/v2/internal/importer"

	"github.com/swipe-io/swipe/v2/internal/usecase/frame"
)

type golangFrame struct {
	importer *importer.Importer
	pkgName  string
	version  string
}

func (f *golangFrame) Frame(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("//+build !swipe\n\n")
	buf.WriteString("// Code generated by Swipe " + f.version + ". DO NOT EDIT.\n\n")
	buf.WriteString("//go:generate swipe\n")
	buf.WriteString("package ")
	buf.WriteString(f.pkgName)
	buf.WriteString("\n\n")

	if f.importer.HasImports() {
		buf.WriteString("import (\n")
		for _, imp := range f.importer.SortedImports() {
			_, _ = fmt.Fprint(&buf, imp)
		}
		buf.WriteString(")\n\n")
	}
	buf.Write(data)

	goSrc := buf.Bytes()
	fmtSrc, err := format.Source(goSrc)
	if err == nil {
		goSrc = fmtSrc
	}
	return goSrc, err
}

func NewGolangFrame(importer *importer.Importer, version, pkgName string) frame.Frame {
	return &golangFrame{importer: importer, version: version, pkgName: pkgName}

}
