package generator

import (
	stdtypes "go/types"
	"strconv"
	"strings"

	"github.com/swipe-io/strcase"

	"github.com/swipe-io/swipe/v2/internal/domain/model"
	"github.com/swipe-io/swipe/v2/internal/types"
)

func structKeyValue(vars []*stdtypes.Var, filterFn types.FilterFn) (results []string) {
	return types.Params(
		vars,
		func(p *stdtypes.Var) []string {
			name := p.Name()
			fieldName := strcase.ToCamel(name)
			return []string{fieldName, name}
		},
		filterFn,
	)
}

func makeLogParams(include, exclude map[string]struct{}, data ...*stdtypes.Var) (result []string) {
	return makeLogParamsRecursive(include, exclude, "", data...)
}

func makeLogParamsRecursive(include, exclude map[string]struct{}, parentName string, data ...*stdtypes.Var) (result []string) {
	for _, v := range data {
		if len(include) > 0 {
			if _, ok := include[v.Name()]; !ok {
				continue
			}
		}
		if len(exclude) > 0 {
			if _, ok := exclude[v.Name()]; ok {
				continue
			}
		}
		if logParam := makeLogParam(parentName+v.Name(), v.Type()); len(logParam) > 0 {
			result = append(result, logParam...)
		}
	}
	return
}

func makeLogParam(name string, t stdtypes.Type) []string {
	quoteName := strconv.Quote(name)
	switch t := t.(type) {
	default:
		return []string{quoteName, name}
	case *stdtypes.Named:
		if hasMethodString(t) {
			return []string{quoteName, name}
		}
		if hasMethodLogParams(t) {
			return []string{quoteName, name + ".LogParams()"}
		}
		return nil
	case *stdtypes.Struct:
		return nil
	case *stdtypes.Basic:
		if t.Kind() == stdtypes.Byte {
			return []string{quoteName, "len(" + name + ")"}
		}
		return []string{quoteName, name}
	case *stdtypes.Pointer:
		return makeLogParam(name, t.Elem())
	case *stdtypes.Slice, *stdtypes.Array, *stdtypes.Map, *stdtypes.Chan:
		return []string{quoteName, "len(" + name + ")"}
	}
}

func isGolangNamedType(t stdtypes.Type) bool {
	t = normalizeType(t)
	switch stdtypes.TypeString(t, nil) {
	case "time.Time",
		"time.Location",
		"sql.NullBool",
		"sql.NullFloat64",
		"sql.NullInt32",
		"sql.NullInt64",
		"sql.NullString",
		"sql.NullTime":
		return true
	}
	return false
}

func normalizeType(t stdtypes.Type) stdtypes.Type {
	switch v := t.(type) {
	case *stdtypes.Pointer:
		return normalizeType(v.Elem())
	case *stdtypes.Slice:
		return normalizeType(v.Elem())
	case *stdtypes.Array:
		return normalizeType(v.Elem())
	case *stdtypes.Map:
		return normalizeType(v.Elem())
	default:
		return v
	}
}

func makeEpSetName(iface *model.ServiceInterface, ifaceLen int) (epSetName string) {
	epSetName = "epSet"
	if ifaceLen > 1 {
		epSetName = "epSet" + iface.NameExport()
	}
	return
}

func parseMethodComments(comments []string) (methodComment string, paramsComment map[string]string) {
	paramsComment = make(map[string]string)
	for _, comment := range comments {
		comment = strings.TrimSpace(comment)

		if strings.HasPrefix(comment, "@") {

			matches := paramCommentRegexp.FindAllStringSubmatch(comment, -1)
			if len(matches) == 1 && len(matches[0]) == 3 {
				paramsComment[matches[0][1]] = matches[0][2]
			}
			continue
		}
		methodComment += comment
	}
	return
}

func hasMethodString(named *stdtypes.Named) bool {
	for i := 0; i < named.NumMethods(); i++ {
		m := named.Method(i)
		if m.Name() == "String" {
			sig := m.Type().(*stdtypes.Signature)
			if sig.Params().Len() == 0 && sig.Results().Len() == 1 && stdtypes.TypeString(sig.Results().At(0).Type(), nil) == "string" {
				return true
			}
		}
	}
	return false
}

func hasMethodLogParams(named *stdtypes.Named) bool {
	for i := 0; i < named.NumMethods(); i++ {
		m := named.Method(i)
		if m.Name() == "LogParams" {
			sig := m.Type().(*stdtypes.Signature)
			if sig.Params().Len() == 0 && sig.Results().Len() == 1 && stdtypes.TypeString(sig.Results().At(0).Type(), nil) == "[]string" {
				return true
			}
		}
	}
	return false
}
