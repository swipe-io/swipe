package fixtures

import swipe "github.com/swipe-io/swipe/v2"

func Swipe() {
	swipe.Build(
		swipe.Service(
			swipe.Interface((*Service)(nil), ""),
		),
	)
}
