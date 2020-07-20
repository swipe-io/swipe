package model

import stdtypes "go/types"

type GatewayMethodOption struct {
	Name         string
	BalancerType string
}

type GatewayServiceOption struct {
	ID            string
	RawID         string
	Type          stdtypes.Type
	TypeName      *stdtypes.Named
	Iface         *stdtypes.Interface
	MethodOptions map[string]GatewayMethodOption
}

type GatewayOption struct {
	Services []GatewayServiceOption
}
