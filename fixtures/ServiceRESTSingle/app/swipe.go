//+build swipe

package app

import (
	"net/http"

	"github.com/swipe-io/swipe/v2"
)

func Swipe() {
	swipe.Build(
		swipe.Service(
			swipe.Interface((*AppInterface)(nil), ""),

			swipe.HTTPServer(),

			swipe.ClientsEnable([]string{"go", "js"}),

			swipe.OpenapiEnable(),
			swipe.OpenapiOutput("./"),

			swipe.ReadmeEnable(),

			swipe.MethodOptions(AppInterface.Create,
				swipe.RESTMethod(http.MethodPost),
				swipe.RESTQueryVars([]string{"date", "date"}),
				swipe.Logging(true),
				swipe.LoggingParams([]string{}, []string{"newData"}),
			),
			swipe.MethodOptions(AppInterface.Get,
				swipe.RESTPath("/get/{fname}"),
				swipe.RESTMethod(http.MethodPost),
				swipe.RESTQueryVars([]string{"cc", "cc"}),
				swipe.RESTHeaderVars([]string{"fname", "fname"}),
			),
			swipe.MethodDefaultOptions(
				swipe.Logging(false),
				swipe.Instrumenting(true),
			),
		),
	)
}
