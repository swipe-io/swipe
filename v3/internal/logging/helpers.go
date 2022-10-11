package logging

import (
	"strconv"

	"github.com/swipe-io/swipe/v3/internal/plugin"
	"github.com/swipe-io/swipe/v3/option"
)

func makeLogParams(include, exclude map[string]struct{}, data ...*option.VarType) (result []string) {
	return makeLogParamsRecursive(include, exclude, "", data...)
}

func makeLogParamsRecursive(include, exclude map[string]struct{}, parentName string, data ...*option.VarType) (result []string) {
	for _, v := range data {
		if plugin.IsContext(v) {
			continue
		}
		if len(include) > 0 {
			if _, ok := include[v.Name.Value]; !ok {
				continue
			}
		}
		if len(exclude) > 0 {
			if _, ok := exclude[v.Name.Value]; ok {
				continue
			}
		}
		if logParam := makeLogParam(parentName+v.Name.Value, v.Type); len(logParam) > 0 {
			result = append(result, logParam...)
		}
	}
	return
}

func makeLogParam(name string, t interface{}) []string {
	quoteName := strconv.Quote(name)
	switch t := t.(type) {
	default:
		return []string{quoteName, name}
	case *option.NamedType:
		if plugin.IsFileDownloadType(t) {
			return []string{quoteName, "len(" + name + ".Data())"}
		}
		if hasMethodString(t) {
			return []string{quoteName, name + ".String()"}
		}
		return nil
	case *option.StructType:
		return nil
	case *option.BasicType:
		return []string{quoteName, name}
	case *option.SliceType, *option.ArrayType, *option.MapType:
		return []string{quoteName, "len(" + name + ")"}
	}
}

func hasMethodString(v *option.NamedType) bool {
	for _, method := range v.Methods {
		if method.Name.Value != "String" {
			continue
		}
		if len(method.Sig.Params) == 0 && len(method.Sig.Results) == 1 {
			if t, ok := method.Sig.Results[0].Type.(*option.BasicType); ok {
				return t.IsString()
			}
		}
	}
	return false
}
