// Package swipe is a code generation tool that automates the creation of repetitively used code.
// Configuration parameters are presented in Swipe as parameters of the Golang function, using explicit initialization instead of
// global variables or reflections.
//
// Swipe generates code using an option: a function that calls functions that define the generation parameters.
// Using Swipe, you describe the generation parameters in the option, and then Swipe generates the code.
//
// 1. The "function as an option" approach is used to configure generation.
//
// 2. All code that is not associated with the generation parameters will not be copied to the generated file.
//
// 3. Function with a `swipe.Build` option inserted in the body. `swipe.Build` will not be transferred to the generated code.
//
// If you want the generate code, you can run:
//  swipe ./pkg/...
//
// Full example:
//  // +build swipe
//
//  package jsonrpc
//
//  import (
//	  "github.com/swipe-io/swipe/v2/fixtures/service"
//
// 	  . "github.com/swipe-io/swipe/v2/pkg/swipe"
//  )
//
//  func Swipe() {
// 	Build(
// 		Service(
// 			(*service.Interface)(nil),
// 			Transport("http",
// 				ClientEnable(),
// 				Openapi(
// 					OpenapiOutput("/../../docs"),
// 					OpenapiVersion("1.0.0"),
// 				),
// 			),
// 			Logging(),
// 			Instrumenting(),
// 		),
// 	  )
//  }
package swipe

const Version = "v2.0.0-rc1"

// A Option is an option for a Swipe.
type Option string

// A ServiceOption is an option service.
type ServiceOption string

// A MethodOption is an option method.
type MethodOption string

// A ConfigEnvOption is an option env config.
type ConfigEnvOption string

// A OpenapiServersOption is an openapi servers option.
type OpenapiServersOption string

// A OpenapiServerOption is an openapi concrete server option.
type OpenapiServerOption string

type ReadmeOption string

type InterfaceOption string

// Build the basic option for defining the generation.
func Build(Option) {
}

// ConfigEnv option for config generation.
//
// To generate a configuration loader, you can use the swipe.ConfigEnv option.
// The optionsStruct parameter is a pointer to the configuration structure.
//
//  The option can work with all primitives, including datetime, and an array of primitives.
//
//  The option supports nested structures.
//
//  To use the default value, just specify it as a value in the structure.
//
// Default func name is `LoadConfig`.
//
// You can use structure tags to control generation:
//
//  env  - name of environment var, options: `required`.
//
//  flag - name of flag, enable as the console flag.
//
//  desc - description for String function.
func ConfigEnv(optionsStruct interface{}, opts ...ConfigEnvOption) Option {
	return "implementation not generated, run swipe"
}

func ConfigEnvFuncName(name string) ConfigEnvOption {
	return "implementation not generated, run swipe"
}

// ConfigEnvDocEnable enable markdown doc generate.
func ConfigEnvDocEnable() ConfigEnvOption {
	return "implementation not generated, run swipe"
}

// ConfigEnvDocOutput output path markdown doc generate.
func ConfigEnvDocOutput(output string) ConfigEnvOption {
	return "implementation not generated, run swipe"
}

// Service a option that defines the generation of transport, metrics, tracing, and logging for gokit.
// Given iface is nil pointer interface, for example:
//  (*pkg.Iface)(nil)
func Service(opts ...ServiceOption) Option {
	return "implementation not generated, run swipe"
}

func Interface(iface interface{}, name string) ServiceOption {
	return "implementation not generated, run swipe"
}

// Name override service name prefix.
func ServiceNamePrefix(string) ServiceOption {
	return "implementation not generated, run swipe"
}

// ReadmeEnable enable for generate readme markdown for service.
func ReadmeEnable() ServiceOption {
	return "implementation not generated, run swipe"
}

func ReadmeOutput(string) ServiceOption {
	return "implementation not generated, run swipe"
}

func ReadmeTemplatePath(string) ServiceOption {
	return "implementation not generated, run swipe"
}

// JSONRPCEnable enabled use JSON RPC instead of REST.
func JSONRPCEnable() ServiceOption {
	return "implementation not generated, run swipe"
}

// JSONRPCPath sets the end point for transport.
func JSONRPCPath(string) ServiceOption {
	return "implementation not generated, run swipe"
}

// JSONRPCDocEnable enable for generate markdown JSON RPC doc.
func JSONRPCDocEnable() ServiceOption {
	return "implementation not generated, run swipe"
}

// JSONRPCDocOutput change output dir for generate markdown JSON RPC doc.
func JSONRPCDocOutput(output string) ServiceOption {
	return "implementation not generated, run swipe"
}

// MethodOptions option for defining method settings.
// Given signature is interface method, for example:
//  pkg.Iface.Create
func MethodOptions(signature interface{}, opts ...MethodOption) ServiceOption {
	return "implementation not generated, run swipe"
}

// MethodDefaultOptions option for defining for all methods default settings.
func MethodDefaultOptions(...MethodOption) ServiceOption {
	return "implementation not generated, run swipe"
}

// Logging a option enabled/disable logging middleware.
func Logging(enable bool) MethodOption {
	return "implementation not generated, run swipe"
}

func Exclude(enable bool) MethodOption {
	return "implementation not generated, run swipe"
}

func LoggingParams(includes []string, excludes []string) MethodOption {
	return "implementation not generated, run swipe"
}

func LoggingContext(key interface{}, name string) MethodOption {
	return "implementation not generated, run swipe"
}

// InstrumentingEnable a option enabled/disable instrumenting (collect metrics) middleware.
func Instrumenting(enable bool) MethodOption {
	return "implementation not generated, run swipe"
}

// InstrumentingDisable a option disable instrumenting (collect metrics) middleware.
func InstrumentingDisable() MethodOption {
	return "implementation not generated, run swipe"
}

// RESTMethod sets http method, default is GET.
func RESTMethod(string) MethodOption {
	return "implementation not generated, run swipe"
}

// WrapResponse wrap the response from the server to an object, for example if you want to return as:
//  {data: { you response data }}
// need to add option:
//  ...code here...
//  WrapResponse("data")
//  ... code here ...
func RESTWrapResponse(string) MethodOption {
	return "implementation not generated, run swipe"
}

// Path sets http path, default is lowecase method name with the prefix "/",
// for example: the Get method will look like " /get".
func RESTPath(string) MethodOption {
	return "implementation not generated, run swipe"
}

// HeaderVars sets the key/value array to get method values from headers,
// where the key is the name of the method parameter,
// and the value is the name of the header.
func RESTHeaderVars([]string) MethodOption {
	return "implementation not generated, run swipe"
}

// QueryVars sets the key/value array to get method values from query args,
// where the key is the name of the method parameter,
// and the value is the name of the query args.
func RESTQueryVars([]string) MethodOption {
	return "implementation not generated, run swipe"
}

// DefaultErrorEncoder is responsible for encoding the server error.
func DefaultErrorEncoder(f interface{}) ServiceOption {
	return "implementation not generated, run swipe"
}

// ServerEncodeResponseFunc sets the encoding function of the passed
// response object to the response writer.
func ServerEncodeResponseFunc(interface{}) MethodOption {
	return "implementation not generated, run swipe"
}

// ServerDecodeRequestFunc sets a function to extract the user's domain
// request object from the request object.
func ServerDecodeRequestFunc(interface{}) MethodOption {
	return "implementation not generated, run swipe"
}

// ClientEncodeRequestFunc sets the function to encode the passed
// request object into an object.
func ClientEncodeRequestFunc(interface{}) MethodOption {
	return "implementation not generated, run swipe"
}

// ClientDecodeResponseFunc sets a function to extract the user's domain
// response object from the response object.
func ClientDecodeResponseFunc(interface{}) MethodOption {
	return "implementation not generated, run swipe"
}

// ClientsEnable enable generate Golang, JavaScript client.
func ClientsEnable(langs []string) ServiceOption {
	return "implementation not generated, run swipe"
}

// ServerDisabled enable generate http server.
func HTTPServer() ServiceOption {
	return "implementation not generated, run swipe"
}

// HTTPFast enable generate fast http server.
func HTTPFast() ServiceOption {
	return "implementation not generated, run swipe"
}

// OpenapiEnable enabled generate openapi documentation.
func OpenapiEnable() ServiceOption {
	return "implementation not generated, run swipe"
}

// OpenapiTags sets docs tags for method.
func OpenapiTags(methods []interface{}, tags []string) ServiceOption {
	return "implementation not generated, run swipe"
}

// OpenapiOutput sets output directory, path relative to the file, default is "./".
func OpenapiOutput(string) ServiceOption {
	return "implementation not generated, run swipe"
}

// OpenapiInfo sets info.
func OpenapiInfo(title, description, version string) ServiceOption {
	return "implementation not generated, run swipe"
}

// OpenapiContact sets openapi contact.
func OpenapiContact(name, email, url string) ServiceOption {
	return "implementation not generated, run swipe"
}

// OpenapiLicence sets openapi licence.
func OpenapiLicence(name, url string) ServiceOption {
	return "implementation not generated, run swipe"
}

// OpenapiServer sets openapi server.
func OpenapiServer(description, url string) ServiceOption {
	return "implementation not generated, run swipe"
}
