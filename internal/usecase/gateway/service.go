package gateway

import (
	"github.com/swipe-io/swipe/v2/internal/domain/model"
	"github.com/swipe-io/swipe/v2/internal/openapi"
	"github.com/swipe-io/swipe/v2/internal/option"
)

type ServiceGateway interface {
	AppID() string
	AppName() string
	Interfaces() model.Interfaces
	Error(key uint32) *model.HTTPError
	ErrorKeys() []uint32
	TransportType() model.Transport
	UseFast() bool
	MethodOption(m model.ServiceMethod) model.MethodOption
	Prefix() string
	DefaultErrorEncoder() option.Value

	ReadmeEnable() bool
	ReadmeOutput() string
	ReadmeTemplatePath() string

	LoggingEnable() bool
	InstrumentingEnable() bool

	JSONRPCEnable() bool
	JSONRPCDocEnable() bool
	JSONRPCDocOutput() string
	JSONRPCPath() string

	ClientEnable() bool
	GoClientEnable() bool
	JSClientEnable() bool

	OpenapiEnable() bool
	OpenapiOutput() string
	OpenapiInfo() openapi.Info
	OpenapiServers() []openapi.Server
	OpenapiMethodTags(name string) []string
	OpenapiDefaultMethodTags() []string
}
