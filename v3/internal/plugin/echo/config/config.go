package config

import (
	"github.com/swipe-io/swipe/v3/internal/finder"
	"github.com/swipe-io/swipe/v3/option"
)

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
	BearerAuth             *struct{}               `swipe:"option"`
}

type OpenapiInfo struct {
	Title       string
	Description string
	Version     string
}

type OpenapiContact struct {
	Name  string
	Email string
	Url   string
}

type OpenapiLicence struct {
	Name string
	Url  string
}

type OpenapiServer struct {
	Description string
	Url         string
}

type OpenapiTag struct {
	Methods []option.NamedType `mapstructure:"methods"`
	Tags    []string           `mapstructure:"tags"`
}

// Config
// @swipe:"Echo"
type Config struct {
	Interfaces           []*Interface `mapstructure:"Interface"`
	ClientEnable         *struct{}
	ClientOutput         option.StringValue
	MethodOptions        []MethodOption
	MethodDefaultOptions MethodOptions
	OpenapiEnable        *struct{}
	OpenapiTags          []OpenapiTag
	OpenapiOutput        option.StringValue
	OpenapiInfo          OpenapiInfo
	OpenapiContact       OpenapiContact
	OpenapiLicence       OpenapiLicence
	OpenapiServers       []OpenapiServer `mapstructure:"OpenapiServer"`

	MethodOptionsMap  map[string]MethodOptions             `mapstructure:"-"`
	OpenapiMethodTags map[string][]string                  `mapstructure:"-"`
	IfaceErrors       map[string]map[string][]finder.Error `mapstructure:"-"`
}
