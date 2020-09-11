package strings

import "strings"

func UcFirst(s string) string {
	return strings.ToUpper(s[:1]) + s[1:]
}

func LcFirst(s string) string {
	return strings.ToLower(s[:1]) + s[1:]
}

func NormalizeCamelCase(s string) string {
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
