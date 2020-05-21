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
					OpenapiInfo("Service Test", "description", "v1.0.0"),
				),
			),
			Logging(),
			Instrumenting(),
		),
	)
}
