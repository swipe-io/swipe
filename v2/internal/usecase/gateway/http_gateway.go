package gateway

import "github.com/swipe-io/swipe/v2/internal/domain/model"

type HTTPGatewayGateway interface {
	Services() []model.GatewayServiceOption
}
