package gateway

import (
	stdtypes "go/types"

	"github.com/swipe-io/swipe/v2/internal/domain/model"
)

type ServiceGateway interface {
	ID() string
	RawID() string
	Transport() model.TransportOption
	Instrumenting() model.InstrumentingOption
	EnableLogging() bool
	Methods() []model.ServiceMethod
	Type() stdtypes.Type
	TypeName() *stdtypes.Named
	Interface() *stdtypes.Interface
	Readme() model.ServiceReadme
}
