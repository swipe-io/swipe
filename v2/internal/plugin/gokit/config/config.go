package config

import (
	"github.com/swipe-io/swipe/v2/option"
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
	ClientName StringValue
	External   *ExternalInterface `mapstructure:"-"`
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
	Signature              *option.NamedType
	Instrumenting          BoolValue
	Logging                BoolValue
	Exclude                BoolValue
	LoggingParams          LoggingParams
	LoggingContext         []LoggingContext
	RESTMethod             StringValue
	RESTWrapResponse       StringValue
	RESTPath               StringValue
	RESTMultipart          SliceStringValue
	RESTHeaderVars         SliceStringValue
	RESTQueryVars          SliceStringValue
	RESTPathVars           map[string]string
	RESTMultipartMaxMemory IntValue
	ServerEncodeResponse   FuncTypeValue
	ServerDecodeRequest    FuncTypeValue
	ClientEncodeRequest    FuncTypeValue
	ClientDecodeResponse   FuncTypeValue
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
	MethodOptions        []*MethodOption
	MethodDefaultOptions MethodOption
	DefaultErrorEncoder  FuncTypeValue

	// non options params
	LoggingEnable       bool                          `mapstructure:"-"`
	InstrumentingEnable bool                          `mapstructure:"-"`
	MethodOptionsMap    map[string]*MethodOption      `mapstructure:"-"`
	OpenapiMethodTags   map[string][]string           `mapstructure:"-"`
	IfaceErrors         map[string]map[string][]Error `mapstructure:"-"`
	JSPkgImportPath     string                        `mapstructure:"-"`
	AppName             string                        `mapstructure:"-"`
	HasExternal         bool                          `mapstructure:"-"`
}
