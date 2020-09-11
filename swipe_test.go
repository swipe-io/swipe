package swipe_test

//import (
//	"net/http"
//
//	"github.com/swipe-io/swipe/v2"
//
//	"github.com/swipe-io/swipe/v2/fixtures/service"
//)
//
//func ExampleTransport() {
//	swipe.Build(
//		swipe.Service((*service.Interface)(nil),
//			swipe.Transport("http"),
//		),
//	)
//}
//
//// Example enabled valyala/fasthttp. Supported in both REST and JSON RPC.
//func ExampleFastEnable() {
//	swipe.Build(
//		swipe.Service((*service.Service)(nil),
//			swipe.Transport("http",
//				swipe.FastEnable(),
//			),
//		),
//	)
//}
//
//// Example basic use Service option.
//func ExampleService() {
//	swipe.Build(
//		swipe.Service((*service.Service)(nil)),
//	)
//}
//
//// Example basic use logging.
//func ExampleLogging() {
//	swipe.Build(
//		swipe.Service((*service.Service)(nil),
//			swipe.Transport("http"),
//			swipe.Logging(),
//		),
//	)
//}
//
//// Example basic use instrumenting.
//func ExampleInstrumenting() {
//	swipe.Build(
//		swipe.Service((*service.Service)(nil),
//			swipe.Transport("http"),
//			swipe.Instrumenting("api", "api"),
//		),
//	)
//}
//
//// Use the swipe.MethodOptions option to specify settings for generating the service method.
//func ExamplePath() {
//	swipe.Build(
//		swipe.Service((*service.Service)(nil),
//			swipe.Transport("http",
//				swipe.MethodOptions(service.Interface.Get,
//					swipe.Path("/users"),
//				),
//			),
//		),
//	)
//}
//
//// Use the swipe.MethodOptions option to specify settings for generating the service method.
//func ExampleMethod() {
//	swipe.Build(
//		swipe.Service((*service.Service)(nil),
//			swipe.Transport("http",
//				swipe.MethodOptions(service.Interface.Get,
//					swipe.Method(http.MethodGet),
//				),
//			),
//		),
//	)
//}
//
//// A parameter is a key pair, where the key is the name of the method parameter,
//// and the value is the name of the parameter in the header.
//func ExampleHeaderVars() {
//	swipe.Build(
//		swipe.Service((*service.Service)(nil),
//			swipe.Transport("http",
//				swipe.MethodOptions(service.Interface.Get,
//					swipe.HeaderVars([]string{"name", "x-name"}),
//				),
//			),
//		),
//	)
//}
//
//// A parameter is a key pair, where the key is the name of the method parameter,
//// and the value is the name of the parameter in the url query arguments.
//func ExampleQueryVars() {
//	swipe.Build(
//		swipe.Service((*service.Service)(nil),
//			swipe.Transport("http",
//				swipe.MethodOptions(service.Interface.Get,
//					swipe.QueryVars([]string{"name", "x-name"}),
//				),
//			),
//		),
//	)
//}
//
//func ExampleOpenapi() {
//	swipe.Build(
//		swipe.Service((*service.Service)(nil),
//			swipe.Transport("http",
//				swipe.Openapi(),
//			),
//		),
//	)
//}
//
//func ExampleOpenapiOutput() {
//	swipe.Build(
//		swipe.Service((*service.Service)(nil),
//			swipe.Transport("http",
//				swipe.Openapi(
//					swipe.OpenapiOutput("../../docs"),
//				),
//			),
//		),
//	)
//}
//
//func ExampleOpenapiInfo() {
//	swipe.Build(
//		swipe.Service((*service.Service)(nil),
//			swipe.Transport("http",
//				swipe.Openapi(
//					swipe.OpenapiInfo("Openapi doc title", "1.0.0", "description"),
//				),
//			),
//		),
//	)
//}
//
//func ExampleOpenapiServer() {
//	swipe.Build(
//		swipe.Service((*service.Service)(nil),
//			swipe.Transport("http",
//				swipe.Openapi(
//					swipe.OpenapiServer("Description for server", "http://server.domain"),
//				),
//			),
//		),
//	)
//}
//
//func ExampleOpenapiContact() {
//	swipe.Build(
//		swipe.Service((*service.Service)(nil),
//			swipe.Transport("http",
//				swipe.Openapi(
//					swipe.OpenapiContact("name", "your_email@mail.com", "http://contact.url"),
//				),
//			),
//		),
//	)
//}
//
//func ExampleOpenapiLicence() {
//	swipe.Build(
//		swipe.Service((*service.Service)(nil),
//			swipe.Transport("http",
//				swipe.Openapi(
//					swipe.OpenapiLicence("MIT", "http://licence.url"),
//				),
//			),
//		),
//	)
//}
//
//type Config struct {
//	BindAddr string
//}
//
//func ExampleConfigEnv() {
//	swipe.Build(
//		swipe.ConfigEnv(
//			&Config{
//				BindAddr: ":9000",
//			},
//			swipe.FuncName("LoadConfig"),
//		),
//	)
//}
