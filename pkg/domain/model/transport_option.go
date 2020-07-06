package model

import (
	"go/ast"
	stdtypes "go/types"

	"github.com/swipe-io/swipe/pkg/openapi"
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
	MethodName         string
	Expr               ast.Expr
	Path               string
	PathVars           map[string]string
	HeaderVars         map[string]string
	QueryVars          map[string]string
	WrapResponse       WrapResponseHTTPTransportOption
	ServerRequestFunc  ReqRespFunc
	ServerResponseFunc ReqRespFunc
	ClientRequestFunc  ReqRespFunc
	ClientResponseFunc ReqRespFunc
}

type ErrorHTTPTransportOption struct {
	Named     *stdtypes.Named
	Code      int64
	IsPointer bool
}

type TransportOption struct {
	Protocol             string
	Prefix               string
	ServerDisabled       bool
	Client               ClientHTTPTransportOption
	Openapi              OpenapiHTTPTransportOption
	FastHTTP             bool
	JsonRPC              JsonRPCHTTPTransportOption
	MethodOptions        map[string]MethodHTTPTransportOption
	DefaultMethodOptions MethodHTTPTransportOption
	MapCodeErrors        map[string]ErrorHTTPTransportOption
}
