//+build swipe

package app

import (
	"net/http"

	"github.com/swipe-io/swipe/v2/fixture/ServiceRESTMultiIdenticalInterface/app/controller/app1"
	"github.com/swipe-io/swipe/v2/fixture/ServiceRESTMultiIdenticalInterface/app/controller/app2"

	swipe "github.com/swipe-io/swipe/v2"
)

func Swipe() {
	swipe.Build(
		swipe.Service(
			swipe.Interface((*app1.App)(nil), "app1"),
			swipe.Interface((*app2.App)(nil), "app2"),

			swipe.HTTPServer(),

			swipe.ClientsEnable([]string{"go", "js"}),

			swipe.OpenapiEnable(),
			swipe.OpenapiOutput("./"),

			swipe.ReadmeEnable(),

			swipe.MethodOptions(app1.App.Create,
				swipe.RESTMethod(http.MethodPost),
				swipe.Logging(true),
				swipe.LoggingParams([]string{}, []string{"newData"}),
				swipe.LoggingContext("123", "123"),
			),

			swipe.MethodOptions(app2.App.Create,
				swipe.RESTMethod(http.MethodPost),
				swipe.Logging(true),
				swipe.LoggingParams([]string{}, []string{"newData"}),
				swipe.LoggingContext("123", "123"),
			),

			swipe.MethodDefaultOptions(
				swipe.Logging(false),
				swipe.Instrumenting(true),
			),
		),
	)
}
