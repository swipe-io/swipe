package model

import (
	"container/list"
	"go/ast"
	stdtypes "go/types"

	"golang.org/x/tools/go/packages"
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
	name             string
	loweName         string
	nameExport       string
	nameUnExport     string
	serviceType      stdtypes.Type
	serviceTypeName  *stdtypes.Named
	serviceIface     *stdtypes.Interface
	serviceMethods   []ServiceMethod
	isNameChange     bool
	external         bool
	externalSwipePkg *packages.Package
	appName          string
}

func (g *ServiceInterface) AppName() string {
	return g.appName
}

func (g *ServiceInterface) ExternalSwipePkg() *packages.Package {
	return g.externalSwipePkg
}

func (g *ServiceInterface) External() bool {
	return g.external
}

func (g *ServiceInterface) IsNameChange() bool {
	return g.isNameChange
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

func NewServiceInterface(name, lowerName, nameExport, nameUnExport string, isNameChange bool, serviceType stdtypes.Type, serviceTypeName *stdtypes.Named, serviceIface *stdtypes.Interface, serviceMethods []ServiceMethod, external bool, externalSwipePkg *packages.Package, appName string) *ServiceInterface {
	return &ServiceInterface{
		name:             name,
		loweName:         lowerName,
		nameExport:       nameExport,
		nameUnExport:     nameUnExport,
		isNameChange:     isNameChange,
		serviceType:      serviceType,
		serviceTypeName:  serviceTypeName,
		serviceIface:     serviceIface,
		serviceMethods:   serviceMethods,
		external:         external,
		externalSwipePkg: externalSwipePkg,
		appName:          appName,
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
	Errors       HTTPErrors
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
	LoggingContext       map[string]ast.Expr
	InstrumentingEnable  bool
	Exclude              bool
}

type HTTPErrors []*HTTPError

func (e HTTPErrors) Len() int {
	return len(e)
}

func (e HTTPErrors) Less(i, j int) bool {
	return e[i].Code < e[j].Code
}

func (e HTTPErrors) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

type HTTPError struct {
	ID        uint32
	Named     *stdtypes.Named
	Code      int64
	IsPointer bool
}
