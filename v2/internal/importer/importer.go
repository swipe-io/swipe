package importer

import (
	"bytes"
	"fmt"
	"sort"
	stdstrings "strings"

	"github.com/swipe-io/swipe/v2/internal/option"
)

type ImportInfo struct {
	Name    string
	Differs bool
}

type Importer struct {
	pkg     *option.PackageType
	imports map[string]ImportInfo
}

func (i *Importer) Import(name, path string) string {
	if path == i.pkg.Path {
		return ""
	}
	const vendorPart = "vendor/"
	unVendor := path
	if i := stdstrings.LastIndex(path, vendorPart); i != -1 && (i == 0 || path[i-1] == '/') {
		unVendor = path[i+len(vendorPart):]
	}
	if info, ok := i.imports[unVendor]; ok {
		return info.Name
	}
	newName := disambiguate(name, func(n string) bool {
		return n == "err" || i.nameInFileScope(n)
	})
	i.imports[unVendor] = ImportInfo{
		Name:    newName,
		Differs: newName != name,
	}
	return newName
}

//func (i *Importer) RewritePkgRefs(node ast.Node) ast.Node {
//	start, end := node.Pos(), node.End()
//
//	node = ast2.Copy(node)
//
//	node = astutil.Apply(node, func(c *astutil.Cursor) bool {
//		switch node := c.Node().(type) {
//		case *ast.Ident:
//			obj := i.pkg.TypesInfo.ObjectOf(node)
//			if obj == nil {
//				return false
//			}
//			if pkg := obj.Pkg(); pkg != nil && obj.Parent() == pkg.Scope() && pkg.Path() != i.pkg.PkgPath {
//				newPkgID := i.Import(pkg.Name(), pkg.Path())
//				c.Replace(&ast.SelectorExpr{
//					X:   ast.NewIdent(newPkgID),
//					Sel: ast.NewIdent(node.Name),
//				})
//				return false
//			}
//			return true
//		case *ast.SelectorExpr:
//			pkgIdent, ok := node.X.(*ast.Ident)
//			if !ok {
//				return true
//			}
//			pkgName, ok := i.pkg.TypesInfo.ObjectOf(pkgIdent).(*stdtypes.PkgName)
//			if !ok {
//				return true
//			}
//			imported := pkgName.Imported()
//			newPkgID := i.Import(imported.Name(), imported.Path())
//			c.Replace(&ast.SelectorExpr{
//				X:   ast.NewIdent(newPkgID),
//				Sel: ast.NewIdent(node.Sel.Name),
//			})
//			return false
//		default:
//			return true
//		}
//	}, nil)
//	newNames := make(map[stdtypes.Object]string)
//	inNewNames := func(n string) bool {
//		for _, other := range newNames {
//			if other == n {
//				return true
//			}
//		}
//		return false
//	}
//	var scopeStack []*stdtypes.Scope
//	pkgScope := i.pkg.Types.Scope()
//	node = astutil.Apply(node, func(c *astutil.Cursor) bool {
//		if scope := i.pkg.TypesInfo.Scopes[c.Node()]; scope != nil {
//			scopeStack = append(scopeStack, scope)
//		}
//		id, ok := c.Node().(*ast.Ident)
//		if !ok {
//			return true
//		}
//		obj := i.pkg.TypesInfo.ObjectOf(id)
//		if obj == nil {
//			return true
//		}
//		if n, ok := newNames[obj]; ok {
//			c.Replace(ast.NewIdent(n))
//			return false
//		}
//		if par := obj.Parent(); par == nil || par == pkgScope {
//			return true
//		}
//		objName := obj.Name()
//		if pos := obj.Pos(); pos < start || end <= pos || !(i.nameInFileScope(objName) || inNewNames(objName)) {
//			return true
//		}
//		newName := disambiguate(objName, func(n string) bool {
//			if i.nameInFileScope(n) || inNewNames(n) {
//				return true
//			}
//			if len(scopeStack) > 0 {
//				_, obj := scopeStack[len(scopeStack)-1].LookupParent(n, token.NoPos)
//				if obj != nil {
//					return true
//				}
//			}
//			return false
//		})
//		newNames[obj] = newName
//		c.Replace(ast.NewIdent(newName))
//		return false
//	}, func(c *astutil.Cursor) bool {
//		if i.pkg.TypesInfo.Scopes[c.Node()] != nil {
//			scopeStack = scopeStack[:len(scopeStack)-1]
//		}
//		return true
//	})
//	return node
//}

func (i *Importer) nameInFileScope(name string) bool {
	for _, other := range i.imports {
		if other.Name == name {
			return true
		}
	}
	//_, obj := i.pkg.Types.Scope().LookupParent(name, token.NoPos)
	//return obj != nil
	return false
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

func (i *Importer) TypeString(v interface{}) string {
	switch t := v.(type) {
	case *option.MapType:
		return pointerPrefix(t.IsPointer) + fmt.Sprintf("map[%s]%s", i.TypeString(t.KeyType), i.TypeString(t.ValueType))
	case *option.ArrayType:
		return pointerPrefix(t.IsPointer) + fmt.Sprintf("[%d]%s", t.Len, i.TypeString(t.ValueType))
	case *option.SliceType:
		return pointerPrefix(t.IsPointer) + "[]" + i.TypeString(t.ValueType)
	case *option.BasicType:
		return pointerPrefix(t.IsPointer) + t.Name
	case *option.VarType:
		return t.Name.Origin + " " + i.TypeString(t.Type)
	case option.VarsType:
		var buf bytes.Buffer
		buf.WriteByte('(')
		for j, param := range t {
			typ := param.Type
			if j > 0 {
				buf.WriteString(", ")
			}
			if param.Name.Origin != "" {
				buf.WriteString(param.Name.Origin)
				buf.WriteByte(' ')
			}
			if param.IsVariadic {
				buf.WriteString("...")
				if s, ok := typ.(*option.SliceType); ok {
					typ = s.ValueType
				}
			}
			buf.WriteString(i.TypeString(typ))
		}
		buf.WriteByte(')')
		return buf.String()
	case *option.SignType:
		var buf bytes.Buffer
		buf.WriteString(i.TypeString(t.Params))
		n := len(t.Results)
		if n == 0 {
			return buf.String()
		}
		buf.WriteByte(' ')
		if n == 1 && t.Results[0].Name.Origin == "" {
			buf.WriteString(i.TypeString(t.Results[0].Type))
			return buf.String()
		}
		buf.WriteString(i.TypeString(t.Results))
		return buf.String()
	case *option.NamedType:
		if t.Pkg == nil {
			return t.Name.Origin
		}
		pkg := i.Import(t.Pkg.Name, t.Pkg.Path)
		return pointerPrefix(t.IsPointer) + pkg + "." + t.Name.Origin
	}
	return ""
}

func NewImporter(pkg *option.PackageType) *Importer {
	return &Importer{
		pkg:     pkg,
		imports: map[string]ImportInfo{},
	}
}

func pointerPrefix(isPointer bool) string {
	if isPointer {
		return "*"
	}
	return ""
}
