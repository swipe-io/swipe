package model

import stdtypes "go/types"

type VarSlice []*stdtypes.Var

func (s VarSlice) LookupField(name string) *stdtypes.Var {
	for _, p := range s {
		if p.Name() == name {
			return p
		}
	}
	return nil
}

type InstrumentingServiceOption struct {
	Enable    bool
	Namespace string
	Subsystem string
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
}

type ServiceOption struct {
	Transport     TransportOption
	Instrumenting InstrumentingServiceOption
	Logging       bool
	Methods       map[string]ServiceMethod
	Type          stdtypes.Type
	Interface     *stdtypes.Interface
	ID            string
}
