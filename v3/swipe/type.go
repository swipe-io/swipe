package swipe

import (
	"bytes"
	"fmt"

	"github.com/swipe-io/swipe/v3/option"
)

func TypeStringWithoutImport(v interface{}, onlySign bool) string {
	return typeString(v, onlySign, nil)
}

func TypeString(v interface{}, onlySign bool, importer Importer) string {
	return typeString(v, onlySign, importer)
}

func typeString(v interface{}, onlySign bool, importer Importer) string {
	switch t := v.(type) {
	case *option.IfaceType:
		return "interface{}"
	case *option.MapType:
		return pointerPrefix(t.IsPointer) + fmt.Sprintf("map[%s]%s", typeString(t.Key, onlySign, importer), typeString(t.Value, onlySign, importer))
	case *option.ArrayType:
		return pointerPrefix(t.IsPointer) + fmt.Sprintf("[%d]%s", t.Len, typeString(t.Value, onlySign, importer))
	case *option.SliceType:
		return pointerPrefix(t.IsPointer) + "[]" + typeString(t.Value, onlySign, importer)
	case *option.BasicType:
		return pointerPrefix(t.IsPointer) + t.Name
	case *option.VarType:
		return t.Name.Value + " " + typeString(t.Type, onlySign, importer)
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
			buf.WriteString(typeString(typ, onlySign, importer))
		}
		buf.WriteByte(')')
		return buf.String()
	case *option.SignType:
		var buf bytes.Buffer
		buf.WriteString(typeString(t.Params, onlySign, importer))
		n := len(t.Results)
		if n == 0 {
			return buf.String()
		}
		buf.WriteByte(' ')
		if n == 1 && t.Results[0].Name.Value == "" {
			buf.WriteString(typeString(t.Results[0].Type, onlySign, importer))
			return buf.String()
		}
		buf.WriteString(typeString(t.Results, onlySign, importer))
		return buf.String()
	case *option.FuncType:
		var buf bytes.Buffer
		buf.WriteString(t.Name.Value)
		if t.Sig != nil {
			buf.WriteString(typeString(t.Sig, onlySign, importer))
		}
		return buf.String()
	case *option.NamedType:
		if t.Pkg == nil {
			return pointerPrefix(t.IsPointer) + t.Name.Value
		}
		pkg := t.Pkg.Name
		if importer != nil {
			pkg = importer.Import(t.Pkg.Name, t.Pkg.Path)
		}
		if pkg != "" {
			pkg = pkg + "."
		}
		return pointerPrefix(t.IsPointer) + pkg + t.Name.Value
	}
	return ""
}

func pointerPrefix(isPointer bool) string {
	if isPointer {
		return "*"
	}
	return ""
}
