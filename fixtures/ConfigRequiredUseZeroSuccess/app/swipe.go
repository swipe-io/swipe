//+build swipe

package app

import "github.com/swipe-io/swipe/v2"

func Swipe() {
	swipe.Build(
		swipe.ConfigEnv(
			&Config{},
			swipe.ConfigEnvDocEnable(),
			swipe.ConfigEnvDocOutput("./"),
		),
	)
}
