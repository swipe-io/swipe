//+build swipe

package jsonrpc

import (
	"github.com/swipe-io/swipe/fixtures/service"

	. "github.com/swipe-io/swipe/pkg/swipe"
)

func Swipe() {
	Build(
		Service((*service.Interface)(nil),
			Transport("http",
				ClientEnable(),

				JSONRPC(
					JSONRPCPath("/rpc/{method}"),
				),
				Openapi(
					OpenapiOutput("/../../docs"),
					OpenapiVersion("1.0.0"),
				),
			),
			Logging(),
			Instrumenting(),
		),
	)
}
