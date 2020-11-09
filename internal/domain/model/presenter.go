package model

import stdtypes "go/types"

type PresenterMethod struct {
	Method *stdtypes.Func
	From   *stdtypes.Named
	To     *stdtypes.Named
}
