//+build swipe

package app

import swipe "github.com/swipe-io/swipe/v2"

func Swipe() {
	swipe.Build(
		swipe.Service(
			swipe.Interface((*InterfaceA)(nil), "a"),
			swipe.Interface((*InterfaceB)(nil), "b"),

			swipe.HTTPServer(),

			swipe.ClientsEnable([]string{"go", "js"}),

			swipe.OpenapiEnable(),
			swipe.OpenapiOutput("./"),

			swipe.JSONRPCEnable(),
			swipe.JSONRPCDocEnable(),
			swipe.JSONRPCDocOutput("./"),

			swipe.ReadmeEnable(),

			swipe.MethodOptions(InterfaceB.Create,
				swipe.Logging(true),
				swipe.LoggingParams([]string{}, []string{"newData"}),
			),

			swipe.MethodDefaultOptions(
				swipe.Logging(false),
				swipe.Instrumenting(true),
			),
		),
	)
}
