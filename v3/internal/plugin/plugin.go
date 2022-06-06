package plugin

import (
	stdstrings "strings"

	"github.com/swipe-io/swipe/v3/option"
)

func IsContext(v *option.VarType) bool {
	if named, ok := v.Type.(*option.NamedType); ok {
		if _, ok := named.Type.(*option.IfaceType); ok {
			return named.Name.Value == "Context" && named.Pkg.Path == "context"
		}
	}
	return false
}

func IsError(v *option.VarType) bool {
	if named, ok := v.Type.(*option.NamedType); ok {
		if _, ok := named.Type.(*option.IfaceType); ok && named.Name.Value == "error" {
			return true
		}
	}
	return false
}

func Error(vars option.VarsType) *option.VarType {
	for _, v := range vars {
		if IsError(v) {
			return v
		}
	}
	return nil
}

type VarType struct {
	Param      *option.VarType
	Value      string
	IsRequired bool
}

func FindParam(p *option.VarType, vars []string) (VarType, bool) {
	for i := 0; i < len(vars); i += 2 {
		paramName := vars[i+1]
		if paramName == p.Name.Value {
			varName := vars[i]
			var required bool
			if stdstrings.HasPrefix(varName, "!") {
				varName = varName[1:]
				required = true
			}
			return VarType{
				Param:      p,
				Value:      varName,
				IsRequired: required,
			}, true
		}
	}
	return VarType{}, false
}
