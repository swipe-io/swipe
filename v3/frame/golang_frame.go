package frame

import (
	"bytes"
	"fmt"

	"github.com/swipe-io/swipe/v3/format"
)

type GolangFrame struct {
	imports      []string
	pkgName      string
	version      string
	useDoNotEdit bool
}

func (f *GolangFrame) Frame(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	if f.useDoNotEdit {
		buf.WriteString("// Code generated by Swipe " + f.version + ". DO NOT EDIT.\n\n")
	}
	buf.WriteString("package ")
	buf.WriteString(f.pkgName)
	buf.WriteString("\n\n")

	if len(f.imports) > 0 {
		buf.WriteString("import (\n")
		for _, imp := range f.imports {
			_, _ = fmt.Fprint(&buf, imp)
		}
		buf.WriteString(")\n\n")
	}
	buf.Write(data)

	goSrc := buf.Bytes()
	fmtSrc, err := format.Source(goSrc)
	if err != nil {
		return nil, fmt.Errorf("error: %w\n ***\n%s\n***\n\n", err, string(goSrc))
	}
	return fmtSrc, nil
}

func NewGolangFrame(imports []string, version, pkgName string, useDoNotEdit bool) *GolangFrame {
	return &GolangFrame{imports: imports, version: version, pkgName: pkgName, useDoNotEdit: useDoNotEdit}
}
