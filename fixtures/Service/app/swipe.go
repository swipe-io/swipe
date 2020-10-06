//+build swipe

package app

import "github.com/swipe-io/swipe/v2"

func Swipe() {
	swipe.Build(
		swipe.Service((*Interface)(nil),
			swipe.Transport("http",
				swipe.JSONRPC(),
				swipe.MarkdownDoc("./"),
				swipe.ClientEnable(),
				swipe.Openapi(
					swipe.OpenapiOutput("./"),
				),

				swipe.MethodOptions(Interface.Create,
					swipe.Logging(true),
					swipe.LoggingParams([]string{}, []string{"newData"}),
				),

				swipe.MethodDefaultOptions(
					swipe.Logging(false),
					swipe.Instrumenting(true),
				),
			),
		),
	)
}
