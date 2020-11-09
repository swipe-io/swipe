package gateway

import (
	"github.com/swipe-io/swipe/v2/internal/domain/model"
)

type PresenterGateway interface {
	Name() string
	Methods() []model.PresenterMethod
}
