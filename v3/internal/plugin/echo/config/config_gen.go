package config

func (*Config) Options() []byte {
	return []byte("// Echo\nfunc Echo(opts ...EchoOption) {}\n\n// InterfaceOption ...\ntype InterfaceOption string\n\n// ClientName ...\nfunc ClientName(value string) InterfaceOption { return \"implementation not generated, run swipe\" }\n\n// EchoOption ...\ntype EchoOption string\n\n// Interface ...\n// @type:\"repeat\"\nfunc Interface(iface interface{}, ns string, opts ...InterfaceOption) EchoOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// MethodOptionsOption ...\ntype MethodOptionsOption string\n\n// RESTMethod ...\nfunc RESTMethod(value interface{}) MethodOptionsOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// RESTWrapResponse ...\nfunc RESTWrapResponse(value string) MethodOptionsOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// RESTWrapRequest ...\nfunc RESTWrapRequest(value string) MethodOptionsOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// RESTPath ...\nfunc RESTPath(value interface{}) MethodOptionsOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// RESTMultipartMaxMemory ...\nfunc RESTMultipartMaxMemory(value int64) MethodOptionsOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// RESTHeaderVars ...\nfunc RESTHeaderVars(value []string) MethodOptionsOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// RESTQueryVars ...\nfunc RESTQueryVars(value []string) MethodOptionsOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// RESTQueryValues ...\nfunc RESTQueryValues(value []string) MethodOptionsOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// RESTBodyType ...\nfunc RESTBodyType(value string) MethodOptionsOption { return \"implementation not generated, run swipe\" }\n\n// MethodOptions ...\n// @type:\"repeat\"\nfunc MethodOptions(signature interface{}, opts ...MethodOptionsOption) EchoOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// MethodDefaultOptions ...\nfunc MethodDefaultOptions(opts ...MethodOptionsOption) EchoOption {\n\treturn \"implementation not generated, run swipe\"\n}\n")
}
