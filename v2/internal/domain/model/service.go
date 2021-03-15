package model

import (
	"container/list"
	"go/ast"
	stdtypes "go/types"

	"github.com/swipe-io/strcase"

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
	ucName           string
	lcName           string
	serviceType      stdtypes.Type
	serviceTypeName  *stdtypes.Named
	serviceIface     *stdtypes.Interface
	serviceMethods   []ServiceMethod
	external         bool
	externalSwipePkg *packages.Package
	appName          string
	ns               string
}

func (g *ServiceInterface) Namespace() string {
	return g.ns
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

func (g *ServiceInterface) UcName() string {
	return g.ucName
}

func (g *ServiceInterface) UcNameWithPrefix() string {
	if g.external {
		return g.appName + g.ucName
	}
	return g.ucName
}

func (g *ServiceInterface) LcName() string {
	return g.lcName
}

func (g *ServiceInterface) LcNameWithPrefix() string {
	if g.external {
		return strcase.ToLowerCamel(g.appName) + g.ucName
	}
	return g.lcName
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
	ucName, lcName string,
	serviceType stdtypes.Type,
	serviceTypeName *stdtypes.Named,
	serviceIface *stdtypes.Interface,
	serviceMethods []ServiceMethod,
	external bool,
	externalSwipePkg *packages.Package,
	appName, ns string,
) *ServiceInterface {
	return &ServiceInterface{
		ucName:           ucName,
		lcName:           lcName,
		serviceType:      serviceType,
		serviceTypeName:  serviceTypeName,
		serviceIface:     serviceIface,
		serviceMethods:   serviceMethods,
		external:         external,
		externalSwipePkg: externalSwipePkg,
		appName:          appName,
		ns:               ns,
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
	Type          *stdtypes.Func
	Name          string
	LcName        string
	IfaceUcName   string
	IfaceLcName   string
	NameRequest   string
	NameResponse  string
	Params        VarSlice
	Results       VarSlice
	Comments      []string
	ParamCtx      *stdtypes.Var
	ParamVariadic *stdtypes.Var
	ReturnErr     *stdtypes.Var
	ResultsNamed  bool
	Variadic      bool
	Errors        HTTPErrors
	T             stdtypes.Type
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
