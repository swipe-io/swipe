//+build swipe

package rest

import (
	"net/http"

	"github.com/swipe-io/swipe/fixtures/service"
	. "github.com/swipe-io/swipe/pkg/swipe"
	"github.com/valyala/fasthttp"
)

func Swipe() {
	Build(
		Service((*service.Interface)(nil),
			Transport("http",
				Openapi(
					OpenapiInfo("Service Test", "", "v1.0.0"),
					OpenapiOutput("../../docs"),
					OpenapiServer("", "http://test.api"),
				),

				FastEnable(),

				ClientEnable(),

				MethodOptions(service.Interface.Get,
					Path("/users/{id:[0-9]}/{name:[a-z]}/{fname}"),
					Method(fasthttp.MethodGet),
					HeaderVars([]string{"n", "x-num-n", "b", "x-num-b"}),
					QueryVars([]string{"price", "price", "c", "c"}),
					ServerDecodeRequestFunc(ServerDecodeRequestTest),
				),
				MethodOptions(service.Interface.GetAll,
					Path("/users"),
					Method(fasthttp.MethodGet),
				),
				MethodOptions(service.Interface.Create,
					Path("/users"),
					Method(fasthttp.MethodPost),
				),
				MethodOptions(service.Interface.Delete,
					Method(fasthttp.MethodPost),
				),
				MethodOptions(service.Interface.TestMethod2,
					Path("/{ns}/auth/{utype}/{user}/{restype}/{resource}/{permission}"),
					Method(http.MethodPut),
				),
			),
			Logging(),
			Instrumenting("api", "api"),
		),
	)
}
