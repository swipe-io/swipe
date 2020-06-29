//+build swipe

package rest

import (
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
					Path("/users/{name:[a-z]}"),
					Method(fasthttp.MethodGet),
					HeaderVars([]string{"n", "x-num"}),
					QueryVars([]string{"price", "price"}),
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
			),
			Logging(),
			Instrumenting("api", "api"),
		),
	)
}
