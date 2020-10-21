//+build swipe

package app

import (
	"net/http"

	"github.com/swipe-io/swipe/v2"
)

func Swipe() {
	swipe.Build(
		swipe.Service(
			swipe.Interface((*InterfaceA)(nil), "a"),
			swipe.Interface((*InterfaceB)(nil), "b"),

			swipe.HTTPServer(),

			swipe.ClientsEnable([]string{"go", "js"}),

			swipe.OpenapiEnable(),
			swipe.OpenapiOutput("./"),

			swipe.ReadmeEnable(),

			swipe.MethodOptions(InterfaceB.Create,
				swipe.RESTMethod(http.MethodPost),
				swipe.Logging(true),
				swipe.LoggingParams([]string{}, []string{"newData"}),
			),
			swipe.MethodOptions(InterfaceB.Get,
				swipe.RESTMethod(http.MethodPost),
				swipe.RESTQueryVars([]string{"cc", "cc"}),
			),
			swipe.MethodDefaultOptions(
				swipe.Logging(false),
				swipe.Instrumenting(true),
			),
		),
	)
}
