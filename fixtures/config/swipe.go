//+build swipe

package config

import (
	. "github.com/swipe-io/swipe/pkg/swipe"
)

func SwipeConfig() {
	Build(
		ConfigEnv(
			&Config{
				Bind: "hohoho",
				Name: "Default MethodName",
			},
			FuncName("LoadConfig"),
		),
	)
}
