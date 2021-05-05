package processor

import (
	"github.com/swipe-io/swipe/v2/internal/domain/model"
	"github.com/swipe-io/swipe/v2/internal/git"
	"github.com/swipe-io/swipe/v2/internal/interface/generator"
	"github.com/swipe-io/swipe/v2/internal/usecase/gateway"
	ug "github.com/swipe-io/swipe/v2/internal/usecase/generator"
	"github.com/swipe-io/swipe/v2/internal/writer"
	"golang.org/x/tools/go/packages"
)

type Service struct {
	ServiceGateway gateway.ServiceGateway
	GIT            *git.GIT
}

func (p *Service) Name() string {
	return "Service"
}

func (p *Service) Generators(pkg *packages.Package, wd string) []ug.Generator {
	generators := []ug.Generator{generator.NewInterface(p.ServiceGateway.Interfaces())}

	if p.ServiceGateway.FoundService() {
		generators = append(generators, generator.NewEndpoint(p.ServiceGateway))
	}
	if p.ServiceGateway.FoundServiceGateway() {
		generators = append(
			generators,
			generator.NewGatewayGenerator(p.ServiceGateway.Interfaces()),
		)
	}
	if p.ServiceGateway.ReadmeEnable() {
		tags, _ := p.GIT.GetTags()
		generators = append(generators,
			generator.NewReadme(
				p.ServiceGateway,
				pkg.PkgPath,
				wd,
				tags,
			),
		)
	}
	if p.ServiceGateway.TransportType() == model.HTTPTransport {
		generators = append(generators, generator.NewHttpTransport(p.ServiceGateway))
		if p.ServiceGateway.LoggingEnable() {
			generators = append(generators, generator.NewLogging(p.ServiceGateway))
		}
		if p.ServiceGateway.InstrumentingEnable() {
			generators = append(generators, generator.NewInstrumenting(p.ServiceGateway))
		}
		if p.ServiceGateway.JSONRPCEnable() {
			if p.ServiceGateway.JSONRPCDocEnable() {
				generators = append(generators, generator.NewJsonrpcDoc(p.ServiceGateway, wd))
			}
			generators = append(generators, generator.NewJsonRPCServer(p.ServiceGateway))
		} else {
			generators = append(generators, generator.NewRestServer(p.ServiceGateway))
		}
	}
	if p.ServiceGateway.ClientEnable() {
		if p.ServiceGateway.GoClientEnable() {
			generators = append(generators,
				generator.NewClientStruct(p.ServiceGateway),
			)
		}
		if p.ServiceGateway.JSONRPCEnable() {
			if p.ServiceGateway.GoClientEnable() {
				generators = append(
					generators,
					generator.NewJsonRPCGoClient(p.ServiceGateway),
				)
			}
			if p.ServiceGateway.JSClientEnable() {
				generators = append(
					generators,
					generator.NewJsonRPCJSClient(p.ServiceGateway),
				)
			}
		} else if p.ServiceGateway.GoClientEnable() {
			generators = append(generators, generator.NewRestGoClient(p.ServiceGateway))
		}
	}
	if p.ServiceGateway.OpenapiEnable() {
		generators = append(generators, generator.NewOpenapi(p.ServiceGateway, wd))
	}
	return generators
}

func ServiceOptions() []byte {
	var w writer.GoWriter
	w.W("// A ServiceOption is an option service.\ntype ServiceOption string\n\n")
	w.W("// A MethodOption is an option method.\ntype MethodOption string\n\n")
	w.W("// A OpenapiServersOption is an openapi servers option.\ntype OpenapiServersOption string\n\n")
	w.W("// A OpenapiServerOption is an openapi concrete server option.\ntype OpenapiServerOption string\n\n")

	w.W("// Service a option that defines the generation of transport, metrics, tracing, and logging for gokit.\n// Given iface is nil pointer interface, for example:\n//  (*pkg.Iface)(nil)\nfunc Service(opts ...ServiceOption) Option {\n\treturn \"implementation not generated, run swipe\"\n}\n\n")
	w.W("func Interface(iface interface{}, ns string) ServiceOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n")
	w.W("func AppName(string) ServiceOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n")

	w.W("// ReadmeEnable enable for generate readme markdown for service.\nfunc ReadmeEnable() ServiceOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n")
	w.W("func ReadmeOutput(string) ServiceOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n")
	w.W("func ReadmeTemplatePath(string) ServiceOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n")

	w.W("// JSONRPCEnable enabled use JSON RPC instead of REST.\nfunc JSONRPCEnable() ServiceOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n")
	w.W("// JSONRPCPath sets the end point for transport.\nfunc JSONRPCPath(string) ServiceOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n")
	w.W("// JSONRPCDocEnable enable for generate markdown JSON RPC doc.\nfunc JSONRPCDocEnable() ServiceOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n")
	w.W("// JSONRPCDocOutput change output dir for generate markdown JSON RPC doc.\nfunc JSONRPCDocOutput(output string) ServiceOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n")

	w.W("// MethodOptions option for defining method settings.\n// Given signature is interface method, for example:\n//  pkg.Iface.Create\nfunc MethodOptions(signature interface{}, opts ...MethodOption) ServiceOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n")
	w.W("// MethodDefaultOptions option for defining for all methods default settings.\nfunc MethodDefaultOptions(...MethodOption) ServiceOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n")

	w.W("// Logging a option enabled/disable logging middleware.\nfunc Logging(enable bool) MethodOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n")
	w.W("func Exclude(enable bool) MethodOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n")
	w.W("func LoggingParams(includes []string, excludes []string) MethodOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n")
	w.W("func LoggingContext(key interface{}, name string) MethodOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n")

	w.W("// InstrumentingEnable a option enabled/disable instrumenting (collect metrics) middleware.\nfunc Instrumenting(enable bool) MethodOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n")
	w.W("// InstrumentingDisable a option disable instrumenting (collect metrics) middleware.\nfunc InstrumentingDisable() MethodOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n")

	w.W("// RESTMethod sets http method, default is GET.\nfunc RESTMethod(string) MethodOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n")
	w.W("// WrapResponse wrap the response from the server to an object, for example if you want to return as:\n//  {data: { you response data }}\n// need to add option:\n//  ...code here...\n//  WrapResponse(\"data\")\n//  ... code here ...\nfunc RESTWrapResponse(string) MethodOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n")
	w.W("// Path sets http path, default is lowecase method name with the prefix \"/\",\n// for example: the Get method will look like \" /get\".\nfunc RESTPath(string) MethodOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n")
	w.W("// HeaderVars sets the key/value array to get method values from headers,\n// where the key is the name of the method parameter,\n// and the value is the name of the header.\nfunc RESTHeaderVars([]string) MethodOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n")
	w.W("// QueryVars sets the key/value array to get method values from query args,\n// where the key is the name of the method parameter,\n// and the value is the name of the query args.\nfunc RESTQueryVars([]string) MethodOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n")

	w.W("// DefaultErrorEncoder is responsible for encoding the server error.\nfunc DefaultErrorEncoder(f interface{}) ServiceOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n")

	w.W("// ServerEncodeResponseFunc sets the encoding function of the passed\n// response object to the response writer.\nfunc ServerEncodeResponseFunc(interface{}) MethodOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n")
	w.W("// ServerDecodeRequestFunc sets a function to extract the user's domain\n// request object from the request object.\nfunc ServerDecodeRequestFunc(interface{}) MethodOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n")

	w.W("// ClientEncodeRequestFunc sets the function to encode the passed\n// request object into an object.\nfunc ClientEncodeRequestFunc(interface{}) MethodOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n")
	w.W("// ClientDecodeResponseFunc sets a function to extract the user's domain\n// response object from the response object.\nfunc ClientDecodeResponseFunc(interface{}) MethodOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n")

	w.W("// ClientsEnable enable generate Golang, JavaScript client.\nfunc ClientsEnable(langs []string) ServiceOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n")

	w.W("// HTTPServer enable generate http server.\nfunc HTTPServer() ServiceOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n")
	w.W("// HTTPFast enable generate fast http server.\nfunc HTTPFast() ServiceOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n")

	w.W("// OpenapiEnable enabled generate openapi documentation.\nfunc OpenapiEnable() ServiceOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n")
	w.W("// Tags sets docs tags for method.\nfunc Tags(methods []interface{}, tags []string) ServiceOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n")
	w.W("// Output sets output directory, path relative to the file, default is \"./\".\nfunc Output(string) ServiceOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n")
	w.W("// Info sets info.\nfunc Info(title, description, version string) ServiceOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n")
	w.W("// OpenapiContact sets openapi contact.\nfunc OpenapiContact(name, email, url string) ServiceOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n")
	w.W("// OpenapiLicence sets openapi licence.\nfunc OpenapiLicence(name, url string) ServiceOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n")
	w.W("// OpenapiServer sets openapi server.\nfunc OpenapiServer(description, url string) ServiceOption {\n\treturn \"implementation not generated, run swipe\"\n}\n\n")
	w.W("\n\n")
	return w.Bytes()
}
