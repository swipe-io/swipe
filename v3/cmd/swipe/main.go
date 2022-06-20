package main

import (
	"fmt"

	v3 "github.com/swipe-io/swipe/v3"
)

var (
	version = "dev"
	date    = ""
	commit  = ""
)

func main() {
	v3.Main(fmt.Sprintf("%s %s %s", version, commit, date))
}
