package config

type Foo struct {
	Name string `descr:"database connection foo name"`
}

type DB struct {
	Conn string `descr:"database connection"`
	Foo  Foo
}

type Config struct {
	Bind     string `flag:"bind-addr,required"`
	Name     string
	MaxPrice int `env:"MAX_PRICE,required"`
	DB       DB  `env:"DB2"`
	URLs     []int
}
