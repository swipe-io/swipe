package generator

import (
	"github.com/swipe-io/strcase"
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
