//+build swipe ide

package app

import (
	"github.com/swipe-io/swipe/v2"
	"github.com/swipe-io/swipe/v2/fixtures/ServiceJSONRPCMulti/app"
)

func Swipe() {
	swipe.Build(
		swipe.Service(
			swipe.Interface((*app.InterfaceA)(nil), "a"),
			swipe.Interface((*app.InterfaceB)(nil), "b"),

			swipe.HTTPServer(),

			swipe.JSONRPCEnable(),

			swipe.ClientsEnable([]string{"js"}),

			swipe.OpenapiEnable(),

			swipe.MethodOptions(app.InterfaceB.Create,
				swipe.Exclude(false),
			),

			swipe.MethodDefaultOptions(
				swipe.Exclude(true),
			),
		),
	)
}
