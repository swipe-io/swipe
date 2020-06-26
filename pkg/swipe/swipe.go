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
//	  "github.com/swipe-io/swipe/fixtures/service"
//
// 	  . "github.com/swipe-io/swipe/pkg/swipe"
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

// A Option is an option for a Swipe.
type Option string

// A ServiceOption is an option service.
type ServiceOption string

// A InstrumentingOption is an option metrics.
type InstrumentingOption string

// A TransportOption is an option gokit transport.
type TransportOption string

// A JSONRPCOption is an option JSON RPC.
type JSONRPCOption interface{}

// A MethodOption is an option method.
type MethodOption string

// A OpenapiOption is an option for openapi doc.
type OpenapiOption string

// A ConfigEnvOption is an option env config.
type ConfigEnvOption string

// A OpenapiServersOption is an openapi servers option.
type OpenapiServersOption string

// A OpenapiServerOption is an openapi concrete server option.
type OpenapiServerOption string

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

// FuncName sets name of the function to load the configuration, default is "LoadConfig".
func FuncName(string) ConfigEnvOption {
	return "implementation not generated, run swipe"
}

// Service a option that defines the generation of transport, metrics, tracing, and logging for gokit.
// Given iface is nil pointer interface, for example:
//  (*pkg.Interface)(nil)
func Service(iface interface{}, opts ...ServiceOption) Option {
	return "implementation not generated, run swipe"
}

// Logging a option enabled logging middleware.
func Logging() ServiceOption {
	return "implementation not generated, run swipe"
}

// Instrumenting a option enabled instrumenting (collect metrics) middleware.
func Instrumenting(...InstrumentingOption) ServiceOption {
	return "implementation not generated, run swipe"
}

// Namespace a namespace metric.
func Namespace(string) InstrumentingOption {
	return "implementation not generated, run swipe"
}

// Subsystem a subsystem metric.
func Subsystem(string) InstrumentingOption {
	return "implementation not generated, run swipe"
}

// Transport a option that defines the transport generation settings.
//
// Swipe generates a method for creating an transport handler using the
// following template:
//
//  MakeHandler<transportType><packageName><serviceName>
//
// <transportType> is REST or JSONRPC.
func Transport(protocol string, opts ...TransportOption) ServiceOption {
	return "implementation not generated, run swipe"
}

// FastEnable enable use valyala/fasthttp instead net/http package.
//
// Supported in both REST and JSON RPC.
func FastEnable() TransportOption {
	return "implementation not generated, run swipe"
}

// MethodOptions option for defining method settings.
// Given signature is interface method, for example:
//  pkg.Interface.Create
func MethodOptions(signature interface{}, opts ...MethodOption) TransportOption {
	return "implementation not generated, run swipe"
}

// MethodDefaultOptions option for defining for all methods default settings.
func MethodDefaultOptions(...MethodOption) TransportOption {
	return "implementation not generated, run swipe"
}

// JSONRPC enabled use JSON RPC instead of REST.
func JSONRPC(...JSONRPCOption) TransportOption {
	return "implementation not generated, run swipe"
}

// JSONRPCPath sets the end point for transport.
func JSONRPCPath(string) JSONRPCOption {
	return "implementation not generated, run swipe"
}

// WrapResponse wrap the response from the server to an object, for example if you want to return as:
//  {data: { you responce data }}
// need to add option:
//  ...code here...
//  WrapResponse("data")
//  ... code here ...
func WrapResponse(string) MethodOption {
	return "implementation not generated, run swipe"
}

// Method sets http method, default is GET.
func Method(string) MethodOption {
	return "implementation not generated, run swipe"
}

// Path sets http path, default is lowecase method name with the prefix "/",
// for example: the Get method will look like " /get".
func Path(string) MethodOption {
	return "implementation not generated, run swipe"
}

// HeaderVars sets the key/value array to get method values from headers,
// where the key is the name of the method parameter,
// and the value is the name of the header.
func HeaderVars([]string) MethodOption {
	return "implementation not generated, run swipe"
}

// QueryVars sets the key/value array to get method values from query args,
// where the key is the name of the method parameter,
// and the value is the name of the query args.
func QueryVars([]string) MethodOption {
	return "implementation not generated, run swipe"
}

// ServerEncodeResponseFunc it is intended for use in servers, for server-side endpoints.
// One of the simple functions of encoding a response is that JSON encodes
// an object directly into the response body.
func ServerEncodeResponseFunc(interface{}) MethodOption {
	return "implementation not generated, run swipe"
}

// ServerDecodeRequestFunc it is intended for use in servers, for server-side endpoints.
// One of the simple methods of DecodeRequestFunc can be that JSON decodes
// from the request body to a specific request type.
func ServerDecodeRequestFunc(interface{}) MethodOption {
	return "implementation not generated, run swipe"
}

// ClientEncodeRequestFunc it is intended for use in clients, for client-side endpoints.
// One of the simple functions of an encoding request is that JSON encodes
// an object directly into the request body.
func ClientEncodeRequestFunc(interface{}) MethodOption {
	return "implementation not generated, run swipe"
}

// ClientDecodeResponseFunc it is intended for use in clients, for client-side endpoints.
// One of the simple response decoding functions can be that JSON decodes
// from the response body to a specific response type.
func ClientDecodeResponseFunc(interface{}) MethodOption {
	return "implementation not generated, run swipe"
}

// ClientEnable enable generate client for the selected transport.
func ClientEnable() TransportOption {
	return "implementation not generated, run swipe"
}

// ServerDisabled disable generate http server.
func ServerDisabled() TransportOption {
	return "implementation not generated, run swipe"
}

// Openapi generate openapi documentation.
func Openapi(...OpenapiOption) TransportOption {
	return "implementation not generated, run swipe"
}

// OpenapiErrors sets binding errors for method in docs generation.
func OpenapiErrors(methods []interface{}, errors []error) OpenapiOption {
	return "implementation not generated, run swipe"
}

// OpenapiTags sets docs tags for method.
func OpenapiTags(methods []interface{}, tags []string) OpenapiOption {
	return "implementation not generated, run swipe"
}

// OpenapiOutput sets output directory, path relative to the file, default is "./".
func OpenapiOutput(string) OpenapiOption {
	return "implementation not generated, run swipe"
}

// OpenapiInfo sets info.
func OpenapiInfo(title, description, version string) OpenapiOption {
	return "implementation not generated, run swipe"
}

// OpenapiContact sets openapi contact.
func OpenapiContact(name, email, url string) OpenapiOption {
	return "implementation not generated, run swipe"
}

// OpenapiLicence sets openapi licence.
func OpenapiLicence(name, url string) OpenapiOption {
	return "implementation not generated, run swipe"
}

// OpenapiServer sets openapi server.
func OpenapiServer(description, url string) OpenapiOption {
	return "implementation not generated, run swipe"
}
