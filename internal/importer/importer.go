package importer

import (
	"fmt"
	"go/ast"
	"go/token"
	stdtypes "go/types"
	"sort"
	stdstrings "strings"

	"github.com/swipe-io/swipe/v2/internal/astcopy"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

type ImportInfo struct {
	Name    string
	Differs bool
}

type Importer struct {
	pkg     *packages.Package
	imports map[string]ImportInfo
}

func (i *Importer) Pkg() *packages.Package {
	return i.pkg
}

func (i *Importer) Import(name, path string) string {
	if path == i.pkg.PkgPath {
		return ""
	}
	const vendorPart = "vendor/"
	unvendored := path
	if i := stdstrings.LastIndex(path, vendorPart); i != -1 && (i == 0 || path[i-1] == '/') {
		unvendored = path[i+len(vendorPart):]
	}
	if info, ok := i.imports[unvendored]; ok {
		return info.Name
	}
	newName := disambiguate(name, func(n string) bool {
		return n == "err" || i.nameInFileScope(n)
	})
	i.imports[unvendored] = ImportInfo{
		Name:    newName,
		Differs: newName != name,
	}
	return newName
}

func (i *Importer) RewritePkgRefs(node ast.Node) ast.Node {
	start, end := node.Pos(), node.End()

	node = astcopy.CopyAST(node)

	node = astutil.Apply(node, func(c *astutil.Cursor) bool {
		switch node := c.Node().(type) {
		case *ast.Ident:
			obj := i.pkg.TypesInfo.ObjectOf(node)
			if obj == nil {
				return false
			}
			if pkg := obj.Pkg(); pkg != nil && obj.Parent() == pkg.Scope() && pkg.Path() != i.pkg.PkgPath {
				newPkgID := i.Import(pkg.Name(), pkg.Path())
				c.Replace(&ast.SelectorExpr{
					X:   ast.NewIdent(newPkgID),
					Sel: ast.NewIdent(node.Name),
				})
				return false
			}
			return true
		case *ast.SelectorExpr:
			pkgIdent, ok := node.X.(*ast.Ident)
			if !ok {
				return true
			}
			pkgName, ok := i.pkg.TypesInfo.ObjectOf(pkgIdent).(*stdtypes.PkgName)
			if !ok {
				return true
			}
			imported := pkgName.Imported()
			newPkgID := i.Import(imported.Name(), imported.Path())
			c.Replace(&ast.SelectorExpr{
				X:   ast.NewIdent(newPkgID),
				Sel: ast.NewIdent(node.Sel.Name),
			})
			return false
		default:
			return true
		}
	}, nil)
	newNames := make(map[stdtypes.Object]string)
	inNewNames := func(n string) bool {
		for _, other := range newNames {
			if other == n {
				return true
			}
		}
		return false
	}
	var scopeStack []*stdtypes.Scope
	pkgScope := i.pkg.Types.Scope()
	node = astutil.Apply(node, func(c *astutil.Cursor) bool {
		if scope := i.pkg.TypesInfo.Scopes[c.Node()]; scope != nil {
			scopeStack = append(scopeStack, scope)
		}
		id, ok := c.Node().(*ast.Ident)
		if !ok {
			return true
		}
		obj := i.pkg.TypesInfo.ObjectOf(id)
		if obj == nil {
			return true
		}
		if n, ok := newNames[obj]; ok {
			c.Replace(ast.NewIdent(n))
			return false
		}
		if par := obj.Parent(); par == nil || par == pkgScope {
			return true
		}
		objName := obj.Name()
		if pos := obj.Pos(); pos < start || end <= pos || !(i.nameInFileScope(objName) || inNewNames(objName)) {
			return true
		}
		newName := disambiguate(objName, func(n string) bool {
			if i.nameInFileScope(n) || inNewNames(n) {
				return true
			}
			if len(scopeStack) > 0 {
				_, obj := scopeStack[len(scopeStack)-1].LookupParent(n, token.NoPos)
				if obj != nil {
					return true
				}
			}
			return false
		})
		newNames[obj] = newName
		c.Replace(ast.NewIdent(newName))
		return false
	}, func(c *astutil.Cursor) bool {
		if i.pkg.TypesInfo.Scopes[c.Node()] != nil {
			scopeStack = scopeStack[:len(scopeStack)-1]
		}
		return true
	})
	return node
}

func (i *Importer) nameInFileScope(name string) bool {
	for _, other := range i.imports {
		if other.Name == name {
			return true
		}
	}
	_, obj := i.pkg.Types.Scope().LookupParent(name, token.NoPos)
	return obj != nil
}

func (i *Importer) HasImports() bool {
	return len(i.imports) > 0
}

func (i *Importer) SortedImports() (result []string) {
	imps := make([]string, 0, len(i.imports))
	for impPath := range i.imports {
		imps = append(imps, impPath)
	}
	sort.Strings(imps)
	result = make([]string, len(imps))
	for j, impPath := range imps {
		info := i.imports[impPath]
		if info.Differs {
			result[j] = fmt.Sprintf("\t%s %q\n", info.Name, impPath)
		} else {
			result[j] = fmt.Sprintf("\t%q\n", impPath)
		}
	}
	return
}

func (i *Importer) QualifyPkg(pkg *stdtypes.Package) string {
	return i.Import(pkg.Name(), pkg.Path())
}

func NewImporter(pkg *packages.Package) *Importer {
	return &Importer{
		pkg:     pkg,
		imports: map[string]ImportInfo{},
	}
}
