package config

import "strings"

func normalizeCamelCase(s string) string {
	n := ""
	for i, v := range s {
		if i > 0 && (v >= 'A' && v <= 'Z') && (s[i-1] >= 'A' && s[i-1] <= 'Z') {
			n += strings.ToLower(string(v))
			continue
		}
		n += string(v)
	}
	return n
}
