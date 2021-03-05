package strings

import "strings"

func UcFirst(s string) string {
	return strings.ToUpper(s[:1]) + s[1:]
}

func LcFirst(s string) string {
	return strings.ToLower(s[:1]) + s[1:]
}
