//+build swipe

package app

import (
	"github.com/swipe-io/swipe/v2"
	"github.com/swipe-io/swipe/v2/fixtures/Service/app"
)

func Swipe() {
	swipe.Build(
		swipe.Gateway(
			swipe.GatewayService((*app.Interface)(nil)),
		),
	)
}
