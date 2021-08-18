package config

import (
	"github.com/swipe-io/swipe/v3/option"
)

type ErrorType string

const (
	RESTErrorType ErrorType = "rest"
	JRPCErrorType ErrorType = "jrpc"
)

type Error struct {
	PkgName   string
	PkgPath   string
	IsPointer bool
	Name      string
	Type      ErrorType
	Code      int64
}

type FuncTypeValue struct {
	Value *option.FuncType
}

type SliceStringValue struct {
	Value []string
}

type StringValue struct {
	Value string
}

type IntValue struct {
	Value int
}

type Int64Value struct {
	Value int64
}

type BoolValue struct {
	Value bool
}

type ExternalInterface struct {
	Iface  *Interface
	Config *Config
	Build  *option.Build
}

type Interface struct {
	Named      *option.NamedType `mapstructure:"iface"`
	Namespace  string            `mapstructure:"ns"`
	ClientName StringValue       `swipe:"option"`
	Gateway    *struct{}         `swipe:"option"`
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

type RESTMultipart struct {
	MaxMemory int64
}

type Aggregate struct {
	Method  *option.NamedType
	Params  SliceStringValue `swipe:"option"`
	Results SliceStringValue `swipe:"option"`
}

type MethodDefaultOption struct {
	Instrumenting          BoolValue         `swipe:"option"`
	Logging                BoolValue         `swipe:"option"`
	LoggingParams          LoggingParams     `swipe:"option"`
	LoggingContext         []LoggingContext  `swipe:"option"`
	RESTMethod             StringValue       `swipe:"option"`
	RESTWrapResponse       StringValue       `swipe:"option"`
	RESTPath               *StringValue      `swipe:"option"`
	RESTMultipartMaxMemory Int64Value        `swipe:"option"`
	RESTHeaderVars         SliceStringValue  `swipe:"option"`
	RESTQueryVars          SliceStringValue  `swipe:"option"`
	RESTQueryValues        SliceStringValue  `swipe:"option"`
	RESTPathVars           map[string]string `swipe:"option"`
	RESTBodyType           StringValue       `swipe:"option"`
	//Aggregate              []Aggregate       `swipe:"option"`
	ServerEncodeResponse FuncTypeValue `swipe:"option"`
	ServerDecodeRequest  FuncTypeValue `swipe:"option"`
	ClientEncodeRequest  FuncTypeValue `swipe:"option"`
	ClientDecodeResponse FuncTypeValue `swipe:"option"`
}

type MethodOption struct {
	Signature *option.NamedType
	MethodDefaultOption
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

// Config
// @swipe:"Gokit"
type Config struct {
	HTTPServer           *struct{}
	HTTPFast             *struct{}
	ClientsEnable        ClientsEnable
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
	MethodOptions        []MethodOption
	MethodDefaultOptions MethodDefaultOption
	DefaultErrorEncoder  FuncTypeValue

	// non options params
	LoggingEnable       bool                           `mapstructure:"-"`
	InstrumentingEnable bool                           `mapstructure:"-"`
	MethodOptionsMap    map[string]MethodDefaultOption `mapstructure:"-"`
	OpenapiMethodTags   map[string][]string            `mapstructure:"-"`
	IfaceErrors         map[string]map[string][]Error  `mapstructure:"-"`
	JSPkgImportPath     string                         `mapstructure:"-"`
	AppName             string                         `mapstructure:"-"`
	HasExternal         bool                           `mapstructure:"-"`
}
