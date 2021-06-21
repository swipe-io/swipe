package importer

import (
	"bytes"
	"fmt"
	"sort"
	stdstrings "strings"

	"github.com/swipe-io/swipe/v2/option"
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

func (i *Importer) nameInFileScope(name string) bool {
	for _, other := range i.imports {
		if other.Name == name {
			return true
		}
	}
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
	return i.typeString(v, false)
}

func (i *Importer) TypeSigString(v interface{}) string {
	return i.typeString(v, true)
}

func (i *Importer) typeString(v interface{}, onlySign bool) string {
	switch t := v.(type) {
	case *option.IfaceType:
		return "interface{}"
	case *option.MapType:
		return pointerPrefix(t.IsPointer) + fmt.Sprintf("map[%s]%s", i.typeString(t.Key, onlySign), i.typeString(t.Value, onlySign))
	case *option.ArrayType:
		return pointerPrefix(t.IsPointer) + fmt.Sprintf("[%d]%s", t.Len, i.typeString(t.Value, onlySign))
	case *option.SliceType:
		return pointerPrefix(t.IsPointer) + "[]" + i.typeString(t.Value, onlySign)
	case *option.BasicType:
		return pointerPrefix(t.IsPointer) + t.Name
	case *option.VarType:
		return t.Name.Value + " " + i.typeString(t.Type, onlySign)
	case option.VarsType:
		var buf bytes.Buffer
		buf.WriteByte('(')
		for j, param := range t {
			typ := param.Type
			if j > 0 {
				buf.WriteString(", ")
			}
			if !onlySign && param.Name.Value != "" {
				buf.WriteString(param.Name.Value)
				buf.WriteByte(' ')
			}
			if param.IsVariadic {
				buf.WriteString("...")
				if s, ok := typ.(*option.SliceType); ok {
					typ = s.Value
				}
			}
			buf.WriteString(i.typeString(typ, onlySign))
		}
		buf.WriteByte(')')
		return buf.String()
	case *option.SignType:
		var buf bytes.Buffer
		buf.WriteString(i.typeString(t.Params, onlySign))
		n := len(t.Results)
		if n == 0 {
			return buf.String()
		}
		buf.WriteByte(' ')
		if n == 1 && t.Results[0].Name.Value == "" {
			buf.WriteString(i.typeString(t.Results[0].Type, onlySign))
			return buf.String()
		}
		buf.WriteString(i.typeString(t.Results, onlySign))
		return buf.String()
	case *option.FuncType:
		var buf bytes.Buffer
		buf.WriteString(t.Name.Value)
		buf.WriteString(i.typeString(t.Sig, onlySign))
		return buf.String()
	case *option.NamedType:
		if t.Pkg == nil {
			return pointerPrefix(t.IsPointer) + t.Name.Value
		}
		pkg := t.Pkg.Name
		if !onlySign {
			pkg = i.Import(t.Pkg.Name, t.Pkg.Path)
		}
		if pkg != "" {
			pkg = pkg + "."
		}
		return pointerPrefix(t.IsPointer) + pkg + t.Name.Value
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
