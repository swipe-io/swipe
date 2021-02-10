package gateway

import (
	"github.com/swipe-io/swipe/v2/internal/domain/model"
	"github.com/swipe-io/swipe/v2/internal/openapi"
	"github.com/swipe-io/swipe/v2/internal/option"
	"golang.org/x/tools/go/types/typeutil"
)

type ServiceGateway interface {
	AppID() string
	AppName() string
	Interfaces() model.Interfaces
	TransportType() model.Transport
	UseFast() bool
	MethodOption(m model.ServiceMethod) model.MethodOption
	Prefix() string
	DefaultErrorEncoder() option.Value
	Errors() map[uint32]*model.HTTPError

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

	CommentFields() map[string]map[string]string
	Enums() *typeutil.Map

	FoundService() bool
	FoundServiceGateway() bool
}
