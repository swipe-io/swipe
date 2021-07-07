package config

func (*Config) Options() []byte {
	return []byte("// Config\nfunc Config(opts ...ConfigOption) {}\n\n// EnvironmentOption ...\ntype EnvironmentOption string\n\n// FuncName ...\nfunc FuncName(value string) EnvironmentOption { return \"implementation not generated, run swipe\" }\n\n// ConfigOption ...\ntype ConfigOption string\n\n// Environment ...\nfunc Environment(structType interface{}, opts ...EnvironmentOption) ConfigOption {\n\treturn \"implementation not generated, run swipe\"\n}\n")
}
