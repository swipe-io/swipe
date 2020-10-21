package model

import (
	"container/list"
	"go/ast"
	stdtypes "go/types"
)

type Transport string

const (
	HTTPTransport Transport = "http"
)

type Interfaces []*ServiceInterface

func (i Interfaces) Len() int {
	return len(i)
}

func (i Interfaces) At(index int) *ServiceInterface {
	return i[index]
}

type ServiceInterface struct {
	prefix          string
	name            string
	loweName        string
	nameExport      string
	nameUnExport    string
	serviceType     stdtypes.Type
	serviceTypeName *stdtypes.Named
	serviceIface    *stdtypes.Interface
	serviceMethods  []ServiceMethod
}

func (g *ServiceInterface) Prefix() string {
	return g.prefix
}

func (g *ServiceInterface) NameExport() string {
	return g.nameExport
}

func (g *ServiceInterface) NameUnExport() string {
	return g.nameUnExport
}

func (g *ServiceInterface) Name() string {
	return g.name
}

func (g *ServiceInterface) LoweName() string {
	return g.loweName
}

func (g *ServiceInterface) Methods() []ServiceMethod {
	return g.serviceMethods
}

func (g *ServiceInterface) Type() stdtypes.Type {
	return g.serviceType
}

func (g *ServiceInterface) TypeName() *stdtypes.Named {
	return g.serviceTypeName
}

func (g *ServiceInterface) Interface() *stdtypes.Interface {
	return g.serviceIface
}

func NewServiceInterface(
	prefix, name, lowerName, nameExport, nameUnExport string,
	serviceType stdtypes.Type,
	serviceTypeName *stdtypes.Named,
	serviceIface *stdtypes.Interface,
	serviceMethods []ServiceMethod,
) *ServiceInterface {
	return &ServiceInterface{
		prefix:          prefix,
		name:            name,
		loweName:        lowerName,
		nameExport:      nameExport,
		nameUnExport:    nameUnExport,
		serviceType:     serviceType,
		serviceTypeName: serviceTypeName,
		serviceIface:    serviceIface,
		serviceMethods:  serviceMethods,
	}
}

type VarSlice []*stdtypes.Var

func (s VarSlice) LookupField(name string) *stdtypes.Var {
	for _, p := range s {
		if p.Name() == name {
			return p
		}
	}
	return nil
}

type DeclType struct {
	Obj      stdtypes.Object
	RecvType stdtypes.Type
	Links    *list.List
	Values   []stdtypes.TypeAndValue
}

type ServiceMethod struct {
	Type         *stdtypes.Func
	Name         string
	NameExport   string
	NameUnExport string
	LcName       string
	NameRequest  string
	NameResponse string
	Params       VarSlice
	Results      VarSlice
	Comments     []string
	ParamCtx     *stdtypes.Var
	ReturnErr    *stdtypes.Var
	ResultsNamed bool
	Errors       map[uint32]*HTTPError
	T            stdtypes.Type
}

type ReqRespFunc struct {
	Type stdtypes.Type
	Expr ast.Expr
}

type WrapResponseHTTPTransportOption struct {
	Enable bool
	Name   string
}

type MethodOption struct {
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
