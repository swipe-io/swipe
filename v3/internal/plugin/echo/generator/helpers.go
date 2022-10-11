package generator

import (
	"container/list"
	"fmt"
	stdstrings "strings"

	"github.com/swipe-io/swipe/v3/swipe"

	"github.com/swipe-io/strcase"
	"github.com/swipe-io/swipe/v3/internal/plugin"
	"github.com/swipe-io/swipe/v3/internal/plugin/echo/config"
	"github.com/swipe-io/swipe/v3/option"
)

func NameInterface(iface *config.Interface) string {
	return UcNameWithAppPrefix(iface) + "Interface"
}

func UcNameWithAppPrefix(iface *config.Interface, useServicePrefix ...bool) string {
	var isUseServicePrefix bool
	if len(useServicePrefix) > 0 {
		isUseServicePrefix = useServicePrefix[0]
	}
	if isUseServicePrefix {
		if iface.ClientName.Take() != "" {
			return strcase.ToCamel(iface.Named.Pkg.Module.ID) + strcase.ToCamel(iface.ClientName.Take())
		}
		return strcase.ToCamel(iface.Named.Pkg.Module.ID) + iface.Named.Name.Upper()
	}
	if iface.ClientName.Take() != "" {
		return strcase.ToCamel(iface.ClientName.Take())
	}
	return iface.Named.Name.Upper()
}

func LcNameWithAppPrefix(iface *config.Interface, notInternal ...bool) string {
	return strcase.ToLowerCamel(UcNameWithAppPrefix(iface, notInternal...))
}

func LcNameIfaceMethod(iface *config.Interface, fn *option.FuncType) string {
	return LcNameWithAppPrefix(iface) + fn.Name.Upper()
}

func UcNameIfaceMethod(iface *config.Interface, fn *option.FuncType) string {
	return UcNameWithAppPrefix(iface) + fn.Name.Upper()
}

func ClientType(iface *config.Interface) string {
	return UcNameWithAppPrefix(iface) + "Client"
}

func findContextVar(vars option.VarsType) (v *option.VarType) {
	for _, p := range vars {
		if plugin.IsContext(p) {
			v = p
			break
		}
	}
	return
}

func findErrorVar(vars option.VarsType) (v *option.VarType) {
	for _, p := range vars {
		if plugin.IsError(p) {
			v = p
			break
		}
	}
	return
}

func wrapDataClientRecursive(e *list.Element, responseType string) (out string) {
	value := e.Value.(string)
	out += strcase.ToCamel(value)
	if next := e.Next(); next != nil {
		out += " struct {\n"
		out += wrapDataClientRecursive(next, responseType)
		out += "} `json:\"" + value + "\"`"
	} else {
		out += fmt.Sprintf(" %s `json:\"%s\"`\n", responseType, e.Value)
	}
	return
}

func wrapDataClient(parts []string, responseType string) (result, structPath string) {
	paths := make([]string, 0, len(parts))
	l := list.New()
	if len(parts) > 0 {
		paths = append(paths, strcase.ToCamel(parts[0]))
		e := l.PushFront(parts[0])
		for i := 1; i < len(parts); i++ {
			paths = append(paths, strcase.ToCamel(parts[i]))
			e = l.InsertAfter(parts[i], e)
		}
	}
	structPath = stdstrings.Join(paths, ".")
	result += "struct { "
	result += wrapDataClientRecursive(l.Front(), responseType)
	result += "}"
	return

}

func isFileUploadType(i interface{}) bool {
	if n, ok := i.(*option.NamedType); ok {
		if iface, ok := n.Type.(*option.IfaceType); ok {
			var done int
			for _, method := range iface.Methods {
				sigStr := swipe.TypeStringWithoutImport(method, true)
				switch sigStr {
				case "Close() (error)", "Name() (string)", "Read([]byte) (int, error)":
					done++
				}
			}
			if done == 3 {
				return true
			}
		}
	}
	return false
}

func IfaceMiddlewareTypeName(iface *config.Interface) string {
	return UcNameWithAppPrefix(iface) + "Middleware"
}
