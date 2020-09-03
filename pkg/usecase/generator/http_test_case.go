package generator

import (
	"context"

	"github.com/swipe-io/swipe/pkg/domain/model"
	"github.com/swipe-io/swipe/pkg/importer"
	"github.com/swipe-io/swipe/pkg/writer"
)

type httpTestCase struct {
	*writer.GoLangWriter
	filename string
	info     model.GenerateInfo
	o        model.ServiceOption
	i        *importer.Importer
}

func (g *httpTestCase) Prepare(ctx context.Context) error {
	return nil
}

func (g *httpTestCase) Process(ctx context.Context) error {
	g.W("// test case here :-)\n")
	return nil
}

func (g *httpTestCase) PkgName() string {
	return ""
}

func (g *httpTestCase) OutputDir() string {
	return ""
}

func (g *httpTestCase) Filename() string {
	return g.filename
}

func (g *httpTestCase) SetImporter(i *importer.Importer) {
	g.i = i
}

func NewHTTPTestCase(filename string, info model.GenerateInfo, o model.ServiceOption) Generator {
	return &httpTestCase{GoLangWriter: writer.NewGoLangWriter(), filename: filename, info: info, o: o}
}
