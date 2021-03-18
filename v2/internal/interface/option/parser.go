package option

import (
	"fmt"
	"go/ast"
	"reflect"

	"golang.org/x/tools/go/ast/astutil"

	"golang.org/x/tools/go/packages"

	"github.com/swipe-io/swipe/v2/internal/astloader"
)

type Parser struct {
}

func (p *Parser) Parse(s interface{}, data *astloader.Data) interface{} {

	p.loadPackages(data.Pkgs)

	v := reflect.ValueOf(s).Elem()
	for i := 0; i < v.NumField(); i++ {
		valueField := v.Field(i)
		typeField := v.Type().Field(i)
		tag := typeField.Tag

		fmt.Printf("Field Name: %s,\t Field Value: %v,\t Tag Value: %s\n", typeField.Name, valueField.Interface(), tag.Get("swipe"))
	}

	return nil
}

func (p Parser) parseAST(pkg *packages.Package, expr ast.Expr) error {
	//exprPos := pkg.Fset.Position(expr.Pos())

	expr = astutil.Unparen(expr)

	switch v := expr.(type) {
	case *ast.CallExpr:

		obj := qualifiedObject(pkg, v.Fun)
		fmt.Println(v, obj.Name())

	}

	return nil
}

func (p *Parser) loadDecl(pkg *packages.Package, decl ast.Decl) {
	switch decl := decl.(type) {
	case *ast.FuncDecl:
		call, err := findInjector(pkg.TypesInfo, decl)
		if err != nil {
			//return nil, err
		}
		if call != nil {
			p.parseAST(pkg, call.Args[0])
			//opt, err := NewParser(pkg).Parse(call.Args[0])
			//if err != nil {
			//	return nil, err
			//}
			//return &ResultOption{
			//	Pkg:    pkg,
			//	Option: opt,
			//}, nil
		}
	}
	//return nil, nil
}

func (p *Parser) loadPackages(pkgs []*packages.Package) {
	//outCh := make(chan *ResultOption)
	//errCh := make(chan error)
	//go func() {
	//	var wg sync.WaitGroup
	for _, pkg := range pkgs {
		for _, f := range pkg.Syntax {
			for _, decl := range f.Decls {

				p.loadDecl(pkg, decl)

				//				wg.Add(1)
				//				go func(pkg *packages.Package, decl ast.Decl) {
				//					defer wg.Done()
				//					result, err := l.declProcess(pkg, decl)
				//					if err != nil {
				//						errCh <- err
				//						return
				//					}
				//					if result != nil {
				//						outCh <- result
				//					}
				//				}(pkg, decl)
			}
		}
	}
	//	wg.Wait()
	//	close(errCh)
	//	close(outCh)
	//}()
	//return outCh, errCh
}

func NewParser() *Parser {
	return &Parser{}
}
