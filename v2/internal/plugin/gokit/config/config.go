package config

import "github.com/swipe-io/swipe/v2/internal/option"

type SliceStringValue struct {
	Value []string
}

type StringValue struct {
	Value string
}

type BoolValue struct {
	Value bool
}

type Interface struct {
	Named     *option.NamedType `mapstructure:"iface"`
	Namespace string            `mapstructure:"ns"`
}

type OpenapiTag struct {
	Methods []option.NamedType `mapstructure:"methods"`
	Tags    []string           `mapstructure:"tags"`
}

type LoggingParams struct {
	Includes []string
	Excludes []string
}

type LoggingContext struct {
	Key  interface{}
	Name string
}

type MethodOption struct {
	Signature        *option.NamedType
	Instrumenting    BoolValue
	Logging          BoolValue
	Exclude          BoolValue
	LoggingParams    LoggingParams
	LoggingContext   []*LoggingContext
	RESTMethod       StringValue
	RESTWrapResponse StringValue
	RESTPath         StringValue
	RESTHeaderVars   SliceStringValue
	RESTQueryVars    SliceStringValue
	RESTPathVars     map[string]string
}

type OpenapiInfo struct {
	Title, Description, Version string
}

type OpenapiContact struct {
	Name, Email, Url string
}

type OpenapiLicence struct {
	Name, Url string
}

type OpenapiServer struct {
	Description, Url string
}

type Langs []string

type ClientsEnable struct {
	Langs Langs
}

func (a Langs) Contains(v string) bool {
	for _, n := range a {
		if v == n {
			return true
		}
	}
	return false
}

type Config struct {
	HTTPServer           *struct{}
	HTTPFast             *struct{}
	ClientsEnable        *ClientsEnable
	JSONRPCEnable        *struct{}
	JSONRPCPath          StringValue
	JSONRPCDocEnable     *struct{}
	JSONRPCDocOutput     StringValue
	Interfaces           []*Interface `mapstructure:"Interface"`
	OpenapiEnable        *struct{}
	OpenapiTags          []OpenapiTag
	OpenapiOutput        StringValue
	OpenapiInfo          OpenapiInfo
	OpenapiContact       OpenapiContact
	OpenapiLicence       OpenapiLicence
	OpenapiServers       []OpenapiServer `mapstructure:"OpenapiServer"`
	MethodOptions        []*MethodOption
	MethodDefaultOptions *MethodOption

	LoggingEnable       bool                     `mapstructure:"-"`
	InstrumentingEnable bool                     `mapstructure:"-"`
	MethodOptionsMap    map[string]*MethodOption `mapstructure:"-"`
	OpenapiMethodTags   map[string][]string      `mapstructure:"-"`
}
