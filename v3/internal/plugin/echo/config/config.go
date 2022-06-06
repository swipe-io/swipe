package config

import "github.com/swipe-io/swipe/v3/option"

type Interface struct {
	Named      *option.NamedType  `mapstructure:"iface"`
	Namespace  string             `mapstructure:"ns"`
	ClientName option.StringValue `swipe:"option"`
}

type MethodOption struct {
	Signature     *option.NamedType
	MethodOptions `mapstructure:",squash"`
}

type MethodOptions struct {
	RESTMethod             option.ExprStringValue  `swipe:"option"`
	RESTWrapResponse       option.StringValue      `swipe:"option"`
	RESTWrapRequest        option.StringValue      `swipe:"option"`
	RESTPath               option.ExprStringValue  `swipe:"option"`
	RESTMultipartMaxMemory option.Int64Value       `swipe:"option"`
	RESTHeaderVars         option.SliceStringValue `swipe:"option"`
	RESTQueryVars          option.SliceStringValue `swipe:"option"`
	RESTQueryValues        option.SliceStringValue `swipe:"option"`
	RESTPathVars           map[string]string       `swipe:"option"`
	RESTBodyType           option.StringValue      `swipe:"option"`
}

// Config
// @swipe:"Echo"
type Config struct {
	Interfaces           []*Interface `mapstructure:"Interface"`
	MethodOptions        []MethodOption
	MethodDefaultOptions MethodOptions

	MethodOptionsMap map[string]MethodOptions `mapstructure:"-"`
}
