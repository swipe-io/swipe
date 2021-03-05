package model

type GatewayMethodOption struct {
	Name         string
	BalancerType string
}

type GatewayServiceOption struct {
	Iface         *ServiceInterface
	MethodOptions map[string]GatewayMethodOption
}
