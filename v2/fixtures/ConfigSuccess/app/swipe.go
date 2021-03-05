//+build swipe

package app

import swipe "github.com/swipe-io/swipe/v2"

func Swipe() {
	swipe.Build(
		swipe.ConfigEnv(
			&Config{},
			swipe.ConfigEnvDocEnable(),
			swipe.ConfigEnvDocOutput("./"),
		),
	)
}
