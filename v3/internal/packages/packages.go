package packages

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/packages"
)

type Packages struct {
	pkgs []*packages.Package
}

func (p *Packages) FindPkgByPath(path string) *packages.Package {
	for _, pkg := range p.pkgs {
		if pkg.PkgPath == path {
			return pkg
		}
	}
	return nil
}

func (p *Packages) ObjectOf(id *ast.Ident) types.Object {
	for _, pkg := range p.pkgs {
		if obj := pkg.TypesInfo.ObjectOf(id); obj != nil {
			return obj
		}
	}
	return nil
}

func (p *Packages) TypeOf(e ast.Expr) types.Type {
	for _, pkg := range p.pkgs {
		if t := pkg.TypesInfo.TypeOf(e); t != nil {
			return t
		}
	}
	return nil
}

func (p *Packages) TraverseTypes(c func(pkg *packages.Package, expr ast.Expr, value types.TypeAndValue) error) error {
	for _, pkg := range p.pkgs {
		for expr, value := range pkg.TypesInfo.Types {
			if err := c(pkg, expr, value); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *Packages) TraverseObjects(c func(pkg *packages.Package, id *ast.Ident, obj types.Object) error) error {
	for _, pkg := range p.pkgs {
		for id, obj := range pkg.TypesInfo.Uses {
			if obj == nil {
				continue
			}
			if err := c(pkg, id, obj); err != nil {
				return err
			}
		}
		for id, obj := range pkg.TypesInfo.Defs {
			if obj == nil {
				continue
			}
			if err := c(pkg, id, obj); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *Packages) TraverseDecls(c func(pkg *packages.Package, file *ast.File, decl ast.Decl) error) error {
	for _, pkg := range p.pkgs {
		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				if err := c(pkg, file, decl); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func NewPackages(pkgs []*packages.Package) *Packages {
	return &Packages{pkgs: pkgs}
}
