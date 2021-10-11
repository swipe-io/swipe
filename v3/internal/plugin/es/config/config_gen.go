package config

func (*Config) Options() []byte {
	return []byte("// EventSourcing\nfunc EventSourcing(opts ...EventSourcingOption) {}\n\n// EventSourcingOption ...\ntype EventSourcingOption string\n\n// Entity ...\nfunc Entity(value interface{}) EventSourcingOption { return \"implementation not generated, run swipe\" }\n")
}
