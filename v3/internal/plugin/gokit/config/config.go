package config

import (
	"github.com/swipe-io/swipe/v3/option"
)

type ErrorType string

const (
	RESTErrorType ErrorType = "rest"
	JRPCErrorType ErrorType = "jrpc"
)

type FuncTypeValue struct {
	Value *option.FuncType
}

type Error struct {
	PkgName string
	PkgPath string
	Name    string
	Type    ErrorType
	Code    int64
}

type ExternalInterface struct {
	Iface  *Interface
	Config *Config
	Build  *option.Inject
}

type Interface struct {
	Named      *option.NamedType `mapstructure:"iface"`
	Namespace  string            `mapstructure:"ns"`
	ClientName option.StringValue
	Gateway    *struct{} `swipe:"option"`
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
	Params  option.SliceStringValue `swipe:"option"`
	Results option.SliceStringValue `swipe:"option"`
}

type MethodDefaultOption struct {
	Instrumenting          option.BoolValue        `swipe:"option"`
	Logging                option.BoolValue        `swipe:"option"`
	LoggingParams          LoggingParams           `swipe:"option"`
	LoggingContext         []LoggingContext        `swipe:"option"`
	RESTMethod             option.StringValue      `swipe:"option"`
	RESTWrapResponse       option.StringValue      `swipe:"option"`
	RESTPath               option.StringValue      `swipe:"option"`
	RESTMultipartMaxMemory option.Int64Value       `swipe:"option"`
	RESTHeaderVars         option.SliceStringValue `swipe:"option"`
	RESTQueryVars          option.SliceStringValue `swipe:"option"`
	RESTQueryValues        option.SliceStringValue `swipe:"option"`
	RESTPathVars           map[string]string       `swipe:"option"`
	RESTBodyType           option.StringValue      `swipe:"option"`
	//Aggregate              []Aggregate       `swipe:"option"`
	ServerEncodeResponse FuncTypeValue `swipe:"option"`
	ServerDecodeRequest  FuncTypeValue `swipe:"option"`
	ClientEncodeRequest  FuncTypeValue `swipe:"option"`
	ClientDecodeResponse FuncTypeValue `swipe:"option"`
}

type MethodOption struct {
	Signature           *option.NamedType
	MethodDefaultOption `mapstructure:",squash"`
}

type OpenapiInfo struct {
	Title       string
	Description string
	Version     interface{}
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
	JSONRPCPath          option.StringValue
	JSONRPCDocEnable     *struct{}
	JSONRPCDocOutput     option.StringValue
	Interfaces           []*Interface `mapstructure:"Interface"`
	OpenapiEnable        *struct{}
	OpenapiTags          []OpenapiTag
	OpenapiOutput        option.StringValue
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
