package gokit

import (
	"fmt"
	"strings"
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

func pathVars(path string) (map[string]string, error) {
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
