package model

import (
	"go/ast"
	stdtypes "go/types"

	"github.com/swipe-io/swipe/v2/internal/openapi"
)

type ReqRespFunc struct {
	Type stdtypes.Type
	Expr ast.Expr
}

type ClientHTTPTransportOption struct {
	Enable bool
}

type OpenapiMethodOption struct {
	Tags []string
}

type OpenapiHTTPTransportOption struct {
	Enable        bool
	Output        string
	Servers       []openapi.Server
	Info          openapi.Info
	Methods       map[string]*OpenapiMethodOption
	DefaultMethod OpenapiMethodOption
}

type WrapResponseHTTPTransportOption struct {
	Enable bool
	Name   string
}

type JsonRPCHTTPTransportOption struct {
	Enable bool
	Path   string
}

type MethodHTTPTransportOption struct {
	MethodName           string
	Expr                 ast.Expr
	Path                 string
	PathVars             map[string]string
	HeaderVars           map[string]string
	QueryVars            map[string]string
	WrapResponse         WrapResponseHTTPTransportOption
	ServerRequestFunc    ReqRespFunc
	ServerResponseFunc   ReqRespFunc
	ClientRequestFunc    ReqRespFunc
	ClientResponseFunc   ReqRespFunc
	LoggingEnable        bool
	LoggingIncludeParams map[string]struct{}
	LoggingExcludeParams map[string]struct{}
	InstrumentingEnable  bool
}

type HTTPError struct {
	Named     *stdtypes.Named
	Code      int64
	IsPointer bool
}

type MarkdownDocHTTPTransportOption struct {
	Enable    bool
	OutputDir string
}

type TransportOption struct {
	Protocol       string
	Prefix         string
	ServerDisabled bool
	Client         ClientHTTPTransportOption
	Openapi        OpenapiHTTPTransportOption
	MarkdownDoc    MarkdownDocHTTPTransportOption
	FastHTTP       bool
	JsonRPC        JsonRPCHTTPTransportOption
	MethodOptions  map[string]MethodHTTPTransportOption
}
