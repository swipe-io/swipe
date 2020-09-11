package model

import (
	"container/list"
	stdtypes "go/types"
)

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

type InstrumentingOption struct {
	Enable    bool
	Namespace string
	Subsystem string
}

type ServiceReadme struct {
	Enable       bool
	OutputDir    string
	TemplatePath string
}

type ServiceMethod struct {
	Type         *stdtypes.Func
	Name         string
	LcName       string
	Params       VarSlice
	Results      VarSlice
	Comments     []string
	ParamCtx     *stdtypes.Var
	ReturnErr    *stdtypes.Var
	ResultsNamed bool
	Errors       map[uint32]*ErrorHTTPTransportOption
	T            stdtypes.Type
}

//type ServiceOption struct {
//	ID            string
//	RawID         string
//	Transport     TransportOption
//	Instrumenting InstrumentingOption
//	EnableLogging       bool
//	Methods       []ServiceMethod
//	Type          stdtypes.Type
//	TypeName      *stdtypes.Named
//	Interface     *stdtypes.Interface
//	Readme        ServiceReadme
//}
