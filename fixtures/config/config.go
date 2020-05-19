//+build swipe

package config

import (
	. "github.com/swipe-io/swipe/pkg/swipe"
)

type Foo struct {
	Name string `descr:"database connection foo name"`
}

type DB struct {
	Conn string `descr:"database connection"`
	Foo  Foo
}

type Config struct {
	Bind     string `flag:"bind-addr"`
	Name     string
	MaxPrice int `env:"MAX_PRICE"`
	DB       DB  `env:"DB2"`
	URLs     []int
}

func SwipeConfig() {
	Build(
		ConfigEnv(
			&Config{
				Name: "Default Name",
			},
			FuncName("LoadConfig"),
		),
	)
}
