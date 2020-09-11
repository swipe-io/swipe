//+build swipe

package app

import "github.com/swipe-io/swipe/v2"

func Swipe() {
	swipe.Build(
		swipe.Service((*Interface)(nil),
			swipe.Transport("http",
				swipe.JSONRPC(),
				swipe.MarkdownDoc("./fixtures/service/app"),
			),
		),
	)
}
