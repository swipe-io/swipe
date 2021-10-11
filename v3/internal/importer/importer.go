package importer

import (
	"fmt"
	"go/token"
	"sort"
	stdstrings "strings"

	"github.com/swipe-io/swipe/v3/option"
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

func NewImporter(pkg *option.PackageType) *Importer {
	return &Importer{
		pkg:     pkg,
		imports: map[string]ImportInfo{},
	}
}
