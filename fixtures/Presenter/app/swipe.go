package app

import (
	"github.com/swipe-io/swipe/v2"
)

func Swipe() {
	swipe.Build(
		swipe.Presenter(
			swipe.PresenterInterface((*UserPresenter)(nil)),
		),
	)
}
