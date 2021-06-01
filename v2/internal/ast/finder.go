package ast

import (
	"fmt"

	"github.com/swipe-io/swipe/v2/option"

	"golang.org/x/tools/go/packages"
)

type Finder struct {
	packages []*packages.Package
}

func (f *Finder) FindImplIface(ifaceType option.IfaceType) {
	for _, p := range f.packages {
		for _, syntax := range p.Syntax {
			for _, decl := range syntax.Decls {
				fmt.Println(decl)
				//stdtypes.Implements()
			}
		}
	}
}

func NewFinder(packages []*packages.Package) *Finder {
	return &Finder{packages: packages}
}
