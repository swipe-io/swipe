package config

func (*Config) Options() []byte {
	return []byte("// Gokit\nfunc Gokit(opts ...GokitOption) {}\n\n// GokitOption ...\ntype GokitOption string\n\n// HTTPServer ...\nfunc HTTPServer() GokitOption { return \"implementation not generated, run swipe\" }\n\n// HTTPFast ...\nfunc HTTPFast() GokitOption { return \"implementation not generated, run swipe\" }\n\n// ClientsEnable ...\nfunc ClientsEnable(langs []string) GokitOption { return \"implementation not generated, run swipe\" }\n\n// ClientOutput ...\nfunc ClientOutput(value string) GokitOption { return \"implementation not generated, run swipe\" }\n\n// CURLEnable ...\nfunc CURLEnable() GokitOption { return \"implementation not generated, run swipe\" }\n\n// CURLOutput ...\nfunc CURLOutput(value string) GokitOption { return \"implementation not generated, run swipe\" }\n\n// CURLURL ...\nfunc CURLURL(value string) GokitOption { return \"implementation not generated, run swipe\" }\n\n// JSONRPCEnable ...\nfunc JSONRPCEnable() GokitOption { return \"implementation not generated, run swipe\" }\n\n// JSONRPCPath ...\nfunc JSONRPCPath(value string) GokitOption { return \"implementation not generated, run swipe\" }\n\n// JSONRPCDocEnable ...\nfunc JSONRPCDocEnable() GokitOption { return \"implementation not generated, run swipe\" }\n\n// JSONRPCDocOutput ...\nfunc JSONRPCDocOutput(value string) GokitOption { return \"implementation not generated, run swipe\" }\n\n// InterfaceOption ...\ntype InterfaceOption string\n\n// ClientName ...\nfunc ClientName(value string) InterfaceOption { return \"implementation not generated, run swipe\" }\n\n// Gateway ...\nfunc Gateway() InterfaceOption { return \"implementation not generated, run swipe\" }\n\n// Interface ...\n// @type:\"repeat\"\nfunc Interface(iface interface{}, ns string, opts ...InterfaceOption) GokitOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// OpenapiEnable ...\nfunc OpenapiEnable() GokitOption { return \"implementation not generated, run swipe\" }\n\n// OpenapiTags ...\n// @type:\"repeat\"\nfunc OpenapiTags(methods []interface{}, tags []string) GokitOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// OpenapiOutput ...\nfunc OpenapiOutput(value string) GokitOption { return \"implementation not generated, run swipe\" }\n\n// OpenapiInfo ...\nfunc OpenapiInfo(title string, description string, version interface{}) GokitOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// OpenapiContact ...\nfunc OpenapiContact(name string, email string, url string) GokitOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// OpenapiLicence ...\nfunc OpenapiLicence(name string, url string) GokitOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// OpenapiServer ...\n// @type:\"repeat\"\nfunc OpenapiServer(description string, url string) GokitOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// MethodOptionsOption ...\ntype MethodOptionsOption string\n\n// Instrumenting ...\nfunc Instrumenting(value bool) MethodOptionsOption { return \"implementation not generated, run swipe\" }\n\n// Logging ...\nfunc Logging(value bool) MethodOptionsOption { return \"implementation not generated, run swipe\" }\n\n// LoggingParams ...\nfunc LoggingParams(includes []string, excludes []string) MethodOptionsOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// LoggingContext ...\n// @type:\"repeat\"\nfunc LoggingContext(key interface{}, name string) MethodOptionsOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// RESTMethod ...\nfunc RESTMethod(value interface{}) MethodOptionsOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// RESTWrapResponse ...\nfunc RESTWrapResponse(value string) MethodOptionsOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// RESTWrapRequest ...\nfunc RESTWrapRequest(value string) MethodOptionsOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// RESTPath ...\nfunc RESTPath(value interface{}) MethodOptionsOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// RESTMultipartMaxMemory ...\nfunc RESTMultipartMaxMemory(value int64) MethodOptionsOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// RESTHeaderVars ...\nfunc RESTHeaderVars(value []string) MethodOptionsOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// RESTQueryVars ...\nfunc RESTQueryVars(value []string) MethodOptionsOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// RESTQueryValues ...\nfunc RESTQueryValues(value []string) MethodOptionsOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// RESTBodyType ...\nfunc RESTBodyType(value string) MethodOptionsOption { return \"implementation not generated, run swipe\" }\n\n// ServerEncodeResponse ...\nfunc ServerEncodeResponse(value interface{}) MethodOptionsOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// ServerDecodeRequest ...\nfunc ServerDecodeRequest(value interface{}) MethodOptionsOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// ClientEncodeRequest ...\nfunc ClientEncodeRequest(value interface{}) MethodOptionsOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// ClientDecodeResponse ...\nfunc ClientDecodeResponse(value interface{}) MethodOptionsOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// ClientErrorDecode ...\nfunc ClientErrorDecode(value interface{}) MethodOptionsOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// MethodOptions ...\n// @type:\"repeat\"\nfunc MethodOptions(signature interface{}, opts ...MethodOptionsOption) GokitOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// MethodDefaultOptions ...\nfunc MethodDefaultOptions(opts ...MethodOptionsOption) GokitOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// ServerErrorEncoder ...\nfunc ServerErrorEncoder(value interface{}) GokitOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n// Labels ...\n// @type:\"repeat\"\nfunc Labels(key interface{}, name string) GokitOption {\n\treturn \"implementation not generated, run swipe\"\n}\n")
}
