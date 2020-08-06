package processor

import (
	"github.com/swipe-io/swipe/pkg/importer"

	"golang.org/x/tools/go/packages"
)

type ImporterFactory struct {
	pkg       *packages.Package
	importers map[string]*importer.Importer
}

func (f *ImporterFactory) Instance(name string) *importer.Importer {
	if _, ok := f.importers[name]; !ok {
		f.importers[name] = importer.NewImporter(f.pkg)
	}
	return f.importers[name]
}

func NewImporterFactory(pkg *packages.Package) *ImporterFactory {
	return &ImporterFactory{pkg: pkg, importers: map[string]*importer.Importer{}}
}
