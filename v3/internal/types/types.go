package types

import (
	"go/ast"

	"golang.org/x/tools/go/packages"
)

func Inspect(pkgs []*packages.Package, f func(p *packages.Package, n ast.Node) bool) {
	for _, p := range pkgs {
		for _, syntax := range p.Syntax {
			ast.Inspect(syntax, func(n ast.Node) bool {
				return f(p, n)
			})
		}
	}
}
