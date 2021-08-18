package config

func (*Config) Options() []byte {
	return []byte("// Gokit\nfunc Gokit(opts ...GokitOption) {}\n\n// GokitOption ...\ntype GokitOption string\n\n// HTTPServer ...\nfunc HTTPServer() GokitOption { return \"implementation not generated, run swipe\" }\n\n// HTTPFast ...\nfunc HTTPFast() GokitOption { return \"implementation not generated, run swipe\" }\n\n// ClientsEnable ...\nfunc ClientsEnable(langs []string) GokitOption { return \"implementation not generated, run swipe\" }\n\n// JSONRPCEnable ...\nfunc JSONRPCEnable() GokitOption { return \"implementation not generated, run swipe\" }\n\n// JSONRPCPath ...\nfunc JSONRPCPath(value string) GokitOption { return \"implementation not generated, run swipe\" }\n\n// JSONRPCDocEnable ...\nfunc JSONRPCDocEnable() GokitOption { return \"implementation not generated, run swipe\" }\n\n// JSONRPCDocOutput ...\nfunc JSONRPCDocOutput(value string) GokitOption { return \"implementation not generated, run swipe\" }\n\n// InterfaceOption ...\ntype InterfaceOption string\n\n// ClientName ...\nfunc ClientName(value string) InterfaceOption { return \"implementation not generated, run swipe\" }\n\n// Gateway ...\nfunc Gateway() InterfaceOption { return \"implementation not generated, run swipe\" }\n\n// Interface ...\n// @type:\"repeat\"\nfunc Interface(iface interface{}, ns string, opts ...InterfaceOption) GokitOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// OpenapiEnable ...\nfunc OpenapiEnable() GokitOption { return \"implementation not generated, run swipe\" }\n\n// OpenapiTags ...\n// @type:\"repeat\"\nfunc OpenapiTags(methods []interface{}, tags []string) GokitOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// OpenapiOutput ...\nfunc OpenapiOutput(value string) GokitOption { return \"implementation not generated, run swipe\" }\n\n// OpenapiInfo ...\nfunc OpenapiInfo(title string, description string, version string) GokitOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// OpenapiContact ...\nfunc OpenapiContact(name string, email string, url string) GokitOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// OpenapiLicence ...\nfunc OpenapiLicence(name string, url string) GokitOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// OpenapiServer ...\n// @type:\"repeat\"\nfunc OpenapiServer(description string, url string) GokitOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// MethodDefaultOption ...\ntype MethodDefaultOption string\n\n// Instrumenting ...\nfunc Instrumenting(value bool) MethodDefaultOption { return \"implementation not generated, run swipe\" }\n\n// Logging ...\nfunc Logging(value bool) MethodDefaultOption { return \"implementation not generated, run swipe\" }\n\n// LoggingParams ...\nfunc LoggingParams(includes []string, excludes []string) MethodDefaultOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// LoggingContext ...\n// @type:\"repeat\"\nfunc LoggingContext(key interface{}, name string) MethodDefaultOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// RESTMethod ...\nfunc RESTMethod(value string) MethodDefaultOption { return \"implementation not generated, run swipe\" }\n\n// RESTWrapResponse ...\nfunc RESTWrapResponse(value string) MethodDefaultOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// RESTPath ...\nfunc RESTPath(value string) MethodDefaultOption { return \"implementation not generated, run swipe\" }\n\n// RESTMultipartMaxMemory ...\nfunc RESTMultipartMaxMemory(value int64) MethodDefaultOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// RESTHeaderVars ...\nfunc RESTHeaderVars(value []string) MethodDefaultOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// RESTQueryVars ...\nfunc RESTQueryVars(value []string) MethodDefaultOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// RESTQueryValues ...\nfunc RESTQueryValues(value []string) MethodDefaultOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// RESTBodyType ...\nfunc RESTBodyType(value string) MethodDefaultOption { return \"implementation not generated, run swipe\" }\n\n// ServerEncodeResponse ...\nfunc ServerEncodeResponse(value interface{}) MethodDefaultOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// ServerDecodeRequest ...\nfunc ServerDecodeRequest(value interface{}) MethodDefaultOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// ClientEncodeRequest ...\nfunc ClientEncodeRequest(value interface{}) MethodDefaultOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// ClientDecodeResponse ...\nfunc ClientDecodeResponse(value interface{}) MethodDefaultOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// MethodOptions ...\n// @type:\"repeat\"\nfunc MethodOptions(signature interface{}, opts ...MethodDefaultOption) GokitOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// MethodDefaultOptions ...\nfunc MethodDefaultOptions(opts ...MethodDefaultOption) GokitOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// DefaultErrorEncoder ...\nfunc DefaultErrorEncoder(value interface{}) GokitOption {\n\treturn \"implementation not generated, run swipe\"\n}\n")
}
