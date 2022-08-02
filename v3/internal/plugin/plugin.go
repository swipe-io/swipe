package plugin

import (
	"fmt"
	"strings"
	stdstrings "strings"

	"github.com/swipe-io/swipe/v3/swipe"

	"github.com/swipe-io/swipe/v3/option"
)

func httpBraceIndices(s string) ([]int, error) {
	var level, idx int
	var idxs []int
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '{':
			if level++; level == 1 {
				idx = i
			}
		case '}':
			if level--; level == 0 {
				idxs = append(idxs, idx, i+1)
			} else if level < 0 {
				return nil, fmt.Errorf("mux: unbalanced braces in %q", s)
			}
		}
	}
	if level != 0 {
		return nil, fmt.Errorf("mux: unbalanced braces in %q", s)
	}
	return idxs, nil
}

func PathVars(path string) (map[string]string, error) {
	idxs, err := httpBraceIndices(path)
	if err != nil {
		return nil, err
	}
	pathVars := make(map[string]string, len(idxs))
	if len(idxs) > 0 {
		var end int
		for i := 0; i < len(idxs); i += 2 {
			end = idxs[i+1]
			parts := strings.SplitN(path[idxs[i]+1:end-1], ":", 2)
			name := parts[0]
			regexp := ""
			if len(parts) == 2 {
				regexp = parts[1]
			}
			pathVars[name] = regexp
		}
	}
	return pathVars, nil
}

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

func Contexts(vars option.VarsType) (result []*option.VarType) {
	for _, v := range vars {
		if IsContext(v) {
			result = append(result, v)
		}
	}
	return
}

func LenWithoutErrors(vars option.VarsType) int {
	if Error(vars) != nil {
		return len(vars) - 1
	}
	return len(vars)
}

func LenWithoutContexts(vars option.VarsType) int {
	return len(vars) - len(Contexts(vars))
}

func DownloadFile(vars option.VarsType) *option.VarType {
	for _, v := range vars {
		if IsFileDownloadType(v.Type) {
			return v
		}
	}
	return nil
}

func IsFileDownloadType(i interface{}) bool {
	if n, ok := i.(*option.NamedType); ok {
		var done int
		for _, method := range n.Methods {
			sigStr := swipe.TypeStringWithoutImport(method, true)
			switch sigStr {
			case "ContentType() (string)", "Name() (string)", "Data() ([]byte)":
				done++
			}
		}
		if done == 3 {
			return true
		}
	}
	return false
}
