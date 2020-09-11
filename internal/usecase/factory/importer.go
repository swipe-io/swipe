package factory

import (
	"github.com/swipe-io/swipe/v2/internal/importer"
	"golang.org/x/tools/go/packages"
)

type ImporterFactory interface {
	NewImporter(ns string, pkg *packages.Package) *importer.Importer
}
