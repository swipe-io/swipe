package swipe_test

import (
	"net/http"

	"github.com/valyala/fasthttp"

	"github.com/swipe-io/swipe/fixtures/service"
	"github.com/swipe-io/swipe/fixtures/transport/jsonrpc"
	"github.com/swipe-io/swipe/fixtures/transport/rest"
	. "github.com/swipe-io/swipe/pkg/swipe"
)

func ExampleTransport() {
	Build(
		Service((*service.Interface)(nil),
			Transport("http"),
		),
	)
}

// Example enabled valyala/fasthttp. Supported in both REST and JSON RPC.
func ExampleFastEnable() {
	Build(
		Service((*service.Service)(nil),
			Transport("http",
				FastEnable(),
			),
		),
	)
}

func ExampleTransport_restListener() {
	h, err := rest.MakeHandlerRESTServiceInterface(&service.Service{})
	if err != nil {
		panic(err)
	}
	go func() {
		_ = fasthttp.ListenAndServe(":80", h)
	}()
}

func ExampleTransport_jsonRPCListener() {
	h, err := jsonrpc.MakeHandlerJSONRPCServiceInterface(&service.Service{})
	if err != nil {
		panic(err)
	}
	go func() {
		_ = http.ListenAndServe(":80", h)
	}()
}

// Example basic use Service option.
func ExampleService() {
	Build(
		Service((*service.Service)(nil)),
	)
}

// Example basic use logging.
func ExampleLogging() {
	Build(
		Service((*service.Service)(nil),
			Transport("http"),
			Logging(),
		),
	)
}

// Example basic use instrumenting.
func ExampleInstrumenting() {
	Build(
		Service((*service.Service)(nil),
			Transport("http"),
			Instrumenting(
				Namespace("api"),
				Subsystem("api"),
			),
		),
	)
}

// Use the swipe.MethodOptions option to specify settings for generating the service method.
func ExamplePath() {
	Build(
		Service((*service.Service)(nil),
			Transport("http",
				MethodOptions(service.Interface.Get,
					Path("/users"),
				),
			),
		),
	)
}

// Use the swipe.MethodOptions option to specify settings for generating the service method.
func ExampleMethod() {
	Build(
		Service((*service.Service)(nil),
			Transport("http",
				MethodOptions(service.Interface.Get,
					Method(http.MethodGet),
				),
			),
		),
	)
}

// A parameter is a key pair, where the key is the name of the method parameter,
// and the value is the name of the parameter in the header.
func ExampleHeaderVars() {
	Build(
		Service((*service.Service)(nil),
			Transport("http",
				MethodOptions(service.Interface.Get,
					HeaderVars([]string{"name", "x-name"}),
				),
			),
		),
	)
}

// A parameter is a key pair, where the key is the name of the method parameter,
// and the value is the name of the parameter in the url query arguments.
func ExampleQueryVars() {
	Build(
		Service((*service.Service)(nil),
			Transport("http",
				MethodOptions(service.Interface.Get,
					QueryVars([]string{"name", "x-name"}),
				),
			),
		),
	)
}

func ExampleOpenapi() {
	Build(
		Service((*service.Service)(nil),
			Transport("http",
				Openapi(),
			),
		),
	)
}

func ExampleOpenapiOutput() {
	Build(
		Service((*service.Service)(nil),
			Transport("http",
				Openapi(
					OpenapiOutput("../../docs"),
				),
			),
		),
	)
}

func ExampleOpenapiInfo() {
	Build(
		Service((*service.Service)(nil),
			Transport("http",
				Openapi(
					OpenapiInfo("Openapi doc title", "1.0.0", "description"),
				),
			),
		),
	)
}

func ExampleOpenapiServer() {
	Build(
		Service((*service.Service)(nil),
			Transport("http",
				Openapi(
					OpenapiServer("Description for server", "http://server.domain"),
				),
			),
		),
	)
}

func ExampleOpenapiContact() {
	Build(
		Service((*service.Service)(nil),
			Transport("http",
				Openapi(
					OpenapiContact("name", "your_email@mail.com", "http://contact.url"),
				),
			),
		),
	)
}

func ExampleOpenapiLicence() {
	Build(
		Service((*service.Service)(nil),
			Transport("http",
				Openapi(
					OpenapiLicence("MIT", "http://licence.url"),
				),
			),
		),
	)
}

type Config struct {
	BindAddr string
}

func ExampleConfigEnv() {
	Build(
		ConfigEnv(
			&Config{
				BindAddr: ":9000",
			},
			FuncName("LoadConfig"),
		),
	)
}
