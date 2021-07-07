package stcreator

type Config struct {
	Commands []string `yaml:"commands"`
	Loaders  Loaders  `yaml:"loaders"`
}
