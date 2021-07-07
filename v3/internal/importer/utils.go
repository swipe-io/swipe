package importer

import (
	"go/token"
	"strconv"
)

func disambiguate(name string, collides func(string) bool) string {
	if !token.Lookup(name).IsKeyword() && !collides(name) {
		return name
	}
	buf := []byte(name)
	if len(buf) > 0 && buf[len(buf)-1] >= '0' && buf[len(buf)-1] <= '9' {
		buf = append(buf, '_')
	}
	base := len(buf)
	for n := 2; ; n++ {
		buf = strconv.AppendInt(buf[:base], int64(n), 10)
		sbuf := string(buf)
		if !token.Lookup(sbuf).IsKeyword() && !collides(sbuf) {
			return sbuf
		}
	}
}
