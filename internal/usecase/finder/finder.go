package finder

import (
	stdtypes "go/types"

	"github.com/swipe-io/swipe/v2/internal/usecase/gateway"
)

type ServiceFinder interface {
	Find(named *stdtypes.Named) (gateway.ServiceGateway, []error)
}
