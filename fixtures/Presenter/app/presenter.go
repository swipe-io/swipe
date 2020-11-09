package app

import (
	"github.com/swipe-io/swipe/v2/fixtures/Presenter/app/dto"
	"github.com/swipe-io/swipe/v2/fixtures/Presenter/app/model"
)

type UserPresenter interface {
	Response(user *model.User) *dto.User
}
