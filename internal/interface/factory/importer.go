package factory

import (
	"sync"

	"github.com/swipe-io/swipe/v2/internal/importer"
	"github.com/swipe-io/swipe/v2/internal/usecase/factory"
	"golang.org/x/tools/go/packages"
)

type importerFactory struct {
	importers *sync.Map
}

func (f *importerFactory) NewImporter(ns string, pkg *packages.Package) *importer.Importer {
	v, ok := f.importers.Load(ns)
	if !ok {
		v = importer.NewImporter(pkg)
		f.importers.Store(ns, v)
	}
	return v.(*importer.Importer)
}

func NewImporterFactory() factory.ImporterFactory {
	return &importerFactory{importers: new(sync.Map)}
}
